package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// S3 is a small path-style SigV4 client. It keeps the project dependency-free
// while working with AWS S3, Aliyun OSS S3 compatibility, and MinIO.
type S3 struct {
	config S3Config
	client *http.Client
}

func NewS3(config S3Config) (*S3, error) {
	if strings.TrimSpace(config.Endpoint) == "" || strings.TrimSpace(config.Bucket) == "" || strings.TrimSpace(config.AccessKey) == "" || strings.TrimSpace(config.SecretKey) == "" {
		return nil, errors.New("s3 endpoint, bucket, access key and secret are required")
	}
	if config.Region == "" {
		config.Region = "us-east-1"
	}
	if _, err := url.ParseRequestURI(strings.TrimRight(config.Endpoint, "/")); err != nil {
		return nil, fmt.Errorf("invalid s3 endpoint: %w", err)
	}
	config.Endpoint = strings.TrimRight(config.Endpoint, "/")
	config.Prefix = strings.Trim(config.Prefix, "/")
	return &S3{config: config, client: http.DefaultClient}, nil
}

func (s3 *S3) URL(key string) string {
	key = s3.objectKey(key)
	if key == "" {
		return s3.config.Endpoint + "/" + s3.config.Bucket
	}
	return s3.config.Endpoint + "/" + s3.config.Bucket + "/" + key
}

func (s3 *S3) Put(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	if size < 0 {
		return "", errors.New("file size cannot be negative")
	}
	request, err := s3.signedRequest(ctx, http.MethodPut, key, reader)
	if err != nil {
		return "", err
	}
	request.ContentLength = size
	response, err := s3.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("s3 put: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("s3 put returned %s", response.Status)
	}
	return s3.URL(key), nil
}

func (s3 *S3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	request, err := s3.signedRequest(ctx, http.MethodGet, key, nil)
	if err != nil {
		return nil, err
	}
	response, err := s3.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("s3 get: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		response.Body.Close()
		return nil, fmt.Errorf("s3 get returned %s", response.Status)
	}
	return response.Body, nil
}

func (s3 *S3) Delete(ctx context.Context, key string) error {
	request, err := s3.signedRequest(ctx, http.MethodDelete, key, nil)
	if err != nil {
		return err
	}
	response, err := s3.client.Do(request)
	if err != nil {
		return fmt.Errorf("s3 delete: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("s3 delete returned %s", response.Status)
	}
	return nil
}

func (s3 *S3) Test(ctx context.Context) error {
	request, err := s3.signedRequest(ctx, http.MethodHead, "", nil)
	if err != nil {
		return err
	}
	response, err := s3.client.Do(request)
	if err != nil {
		return fmt.Errorf("s3 connection: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("s3 connection returned %s", response.Status)
	}
	return nil
}

func (s3 *S3) objectKey(key string) string {
	key = strings.Trim(key, "/")
	if s3.config.Prefix == "" {
		return key
	}
	if key == "" {
		return s3.config.Prefix
	}
	return s3.config.Prefix + "/" + key
}

func (s3 *S3) signedRequest(ctx context.Context, method, key string, body io.Reader) (*http.Request, error) {
	object := s3.objectKey(key)
	requestURL := s3.config.Endpoint + "/" + s3.config.Bucket
	if object != "" {
		requestURL += "/" + object
	}
	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	date := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")
	host := request.URL.Host
	payloadHash := "UNSIGNED-PAYLOAD"
	request.Header.Set("Host", host)
	request.Header.Set("x-amz-content-sha256", payloadHash)
	request.Header.Set("x-amz-date", amzDate)
	canonicalURI := pathEscape(request.URL.EscapedPath())
	canonicalHeaders := "host:" + host + "\n" + "x-amz-content-sha256:" + payloadHash + "\n" + "x-amz-date:" + amzDate + "\n"
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonical := method + "\n" + canonicalURI + "\n\n" + canonicalHeaders + "\n" + signedHeaders + "\n" + payloadHash
	scope := date + "/" + s3.config.Region + "/s3/aws4_request"
	toSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hashHex([]byte(canonical))
	signature := hex.EncodeToString(hmacSHA256(signingKey(s3.config.SecretKey, date, s3.config.Region, "s3"), toSign))
	request.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+s3.config.AccessKey+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+signature)
	return request, nil
}

func pathEscape(value string) string { return strings.ReplaceAll(value, "+", "%20") }
func hashHex(value []byte) string    { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }
func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
func signingKey(secret, date, region, service string) []byte {
	k := hmacSHA256([]byte("AWS4"+secret), date)
	k = hmacSHA256(k, region)
	k = hmacSHA256(k, service)
	return hmacSHA256(k, "aws4_request")
}
