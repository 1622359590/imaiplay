package sms

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AliyunSender struct {
	config   Config
	Client   *http.Client
	Endpoint string
}

func NewAliyunSender(config Config) *AliyunSender {
	return &AliyunSender{config: config, Client: http.DefaultClient, Endpoint: "https://dysmsapi.aliyuncs.com/"}
}

func (sender *AliyunSender) Send(ctx context.Context, phone, templateCode string, params map[string]string) error {
	if templateCode == "" {
		templateCode = sender.config.TemplateCode
	}
	paramJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}
	values := url.Values{
		"AccessKeyId": {sender.config.AccessKeyID}, "Action": {"SendSms"}, "Format": {"JSON"},
		"PhoneNumbers": {phone}, "SignName": {sender.config.SignName}, "SignatureMethod": {"HMAC-SHA1"},
		"SignatureNonce": {uuid.NewString()}, "SignatureVersion": {"1.0"},
		"TemplateCode": {templateCode}, "TemplateParam": {string(paramJSON)},
		"Timestamp": {time.Now().UTC().Format("2006-01-02T15:04:05Z")}, "Version": {"2017-05-25"},
	}
	values.Set("Signature", sign(values, sender.config.AccessKeySecret))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.Endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := sender.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("aliyun sms http status %d", response.StatusCode)
	}
	var result struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}
	if result.Code != "OK" {
		return fmt.Errorf("aliyun sms: %s", result.Message)
	}
	return nil
}

func sign(values url.Values, secret string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	canonical := make([]string, 0, len(keys))
	for _, key := range keys {
		canonical = append(canonical, percent(key)+"="+percent(values.Get(key)))
	}
	stringToSign := "POST&%2F&" + percent(strings.Join(canonical, "&"))
	hash := hmac.New(sha1.New, []byte(secret+"&"))
	_, _ = hash.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func percent(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(url.QueryEscape(value), "+", "%20"), "%7E", "~")
}
