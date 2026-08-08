// Package baota provides a small client for the BaoTa panel API.
package baota

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const requestTimeout = 10 * time.Second

// Client calls the BaoTa panel API.
type Client struct {
	PanelURL              string
	APIKey                string
	TLSInsecureSkipVerify bool
	HTTPClient            *http.Client

	now          func() time.Time
	pendingCerts sync.Map
}

// BaotaClient is the explicit product-facing name for Client.
type BaotaClient = Client

type apiResponse map[string]json.RawMessage

func (client *Client) request(
	ctx context.Context,
	route string,
	form url.Values,
	retry bool,
	allowFalseStatus bool,
) (apiResponse, error) {
	endpoint, err := client.endpoint(route)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(client.APIKey) == "" {
		return nil, errors.New("baota API key is required")
	}
	if form.Get("action") == "" {
		return nil, errors.New("baota API action is required")
	}

	attempts := 1
	if retry {
		attempts = 2
	}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		signed := cloneValues(form)
		client.sign(signed)
		response, retryable, err := client.do(ctx, endpoint, signed)
		if err == nil {
			if operationErr := response.operationError(allowFalseStatus); operationErr == nil {
				return response, nil
			} else {
				err = operationErr
				retryable = retry
			}
		}
		lastErr = err
		if !retryable {
			break
		}
	}
	return nil, lastErr
}

func (client *Client) endpoint(route string) (string, error) {
	panelURL := strings.TrimRight(strings.TrimSpace(client.PanelURL), "/")
	if panelURL == "" {
		return "", errors.New("baota panel URL is required")
	}
	if route == "" || route[0] != '/' {
		return "", fmt.Errorf("invalid baota API route %q", route)
	}
	endpoint, err := url.Parse(panelURL + route)
	if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Host == "" {
		return "", fmt.Errorf("invalid baota panel URL %q", client.PanelURL)
	}
	return endpoint.String(), nil
}

func (client *Client) sign(form url.Values) {
	now := time.Now
	if client.now != nil {
		now = client.now
	}
	requestTime := strconv.FormatInt(now().Unix(), 10)
	keyDigest := md5.Sum([]byte(client.APIKey))
	tokenDigest := md5.Sum([]byte(requestTime + hex.EncodeToString(keyDigest[:])))
	form.Set("request_time", requestTime)
	form.Set("request_token", hex.EncodeToString(tokenDigest[:]))
}

func (client *Client) do(
	ctx context.Context,
	endpoint string,
	form url.Values,
) (apiResponse, bool, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewBufferString(form.Encode()),
	)
	if err != nil {
		return nil, false, fmt.Errorf("create baota API request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: requestTimeout}
		if client.TLSInsecureSkipVerify {
			transport := http.DefaultTransport.(*http.Transport).Clone()
			// This is an explicit operator opt-in for a trusted same-host BaoTa panel.
			// It remains isolated from every other HTTP client in the application.
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
			httpClient.Transport = transport
		}
	} else if httpClient.Timeout <= 0 {
		configured := *httpClient
		configured.Timeout = requestTimeout
		httpClient = &configured
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, true, fmt.Errorf("call baota API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode >= http.StatusInternalServerError, fmt.Errorf("read baota API response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = resp.Status
		}
		return nil, resp.StatusCode >= http.StatusInternalServerError, fmt.Errorf(
			"baota API HTTP %d: %s",
			resp.StatusCode,
			message,
		)
	}

	var response apiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		var list []json.RawMessage
		if json.Unmarshal(body, &list) == nil {
			return apiResponse{"_list": append(json.RawMessage(nil), body...)}, false, nil
		}
		var message string
		if json.Unmarshal(body, &message) == nil && strings.TrimSpace(message) != "" {
			return nil, false, fmt.Errorf("baota API: %s", message)
		}
		return nil, true, fmt.Errorf("invalid JSON from baota API: %w", err)
	}
	return response, false, nil
}

func (response apiResponse) operationError(allowFalseStatus bool) error {
	for _, key := range []string{"success", "siteStatus"} {
		if value, ok := response.boolValue(key); ok && !value {
			return response.error()
		}
	}
	if value, ok := response.boolValue("status"); ok && !value && !allowFalseStatus {
		return response.error()
	}
	if value, ok := response.stringValue("status"); ok {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "false", "failed", "failure", "error":
			return response.error()
		}
	}
	if code, ok := response.intValue("code"); ok && code != 0 {
		return response.error()
	}
	return nil
}

func (response apiResponse) error() error {
	for _, key := range []string{"message", "msg", "error"} {
		if value, ok := response.stringValue(key); ok && strings.TrimSpace(value) != "" {
			return fmt.Errorf("baota API: %s", value)
		}
	}
	return errors.New("baota API: operation failed")
}

func (response apiResponse) boolValue(key string) (bool, bool) {
	raw, exists := response[key]
	if !exists {
		return false, false
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return false, false
	}
	return value, true
}

func (response apiResponse) stringValue(key string) (string, bool) {
	raw, exists := response[key]
	if !exists {
		return "", false
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false
	}
	return value, true
}

func (response apiResponse) intValue(key string) (int, bool) {
	raw, exists := response[key]
	if !exists {
		return 0, false
	}
	var value int
	if err := json.Unmarshal(raw, &value); err == nil {
		return value, true
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return 0, false
	}
	value, err := strconv.Atoi(text)
	return value, err == nil
}

func cloneValues(values url.Values) url.Values {
	clone := make(url.Values, len(values))
	for key, value := range values {
		clone[key] = append([]string(nil), value...)
	}
	return clone
}
