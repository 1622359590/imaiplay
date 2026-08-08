package baota

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

var fixedBaotaTime = time.Unix(1_700_000_000, 0)

func TestRequestTimeoutForRouteAllowsLongRunningACME(t *testing.T) {
	tests := []struct {
		route string
		want  time.Duration
	}{
		{route: "/data", want: requestTimeout},
		{route: "/site", want: requestTimeout},
		{route: "/acme", want: acmeRequestTimeout},
	}

	for _, test := range tests {
		t.Run(test.route, func(t *testing.T) {
			if got := requestTimeoutForRoute(test.route); got != test.want {
				t.Fatalf("requestTimeoutForRoute(%q) = %s, want %s", test.route, got, test.want)
			}
		})
	}
}

func TestClientSelfSignedTLSIsRejectedByDefault(t *testing.T) {
	server := newBaotaTLSTestServer(t)
	defer server.Close()

	client := &Client{PanelURL: server.URL, APIKey: "test-key"}
	_, err := client.request(
		context.Background(),
		"/data",
		url.Values{"action": {"getData"}},
		false,
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("request() error = %v, want certificate verification failure", err)
	}
}

func TestClientAllowsExplicitInsecureTLS(t *testing.T) {
	server := newBaotaTLSTestServer(t)
	defer server.Close()

	client := &Client{
		PanelURL:              server.URL,
		APIKey:                "test-key",
		TLSInsecureSkipVerify: true,
	}
	if _, err := client.request(
		context.Background(),
		"/data",
		url.Values{"action": {"getData"}},
		false,
		false,
	); err != nil {
		t.Fatalf("request() error = %v", err)
	}
}

func newBaotaTLSTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/data" {
			t.Errorf("request = %s %s, want POST /data", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	}))
}

func TestAddSiteUsesOfficialRouteSignatureAndParameters(t *testing.T) {
	requests := 0
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		requests++
		if requests == 1 {
			if r.Method != http.MethodPost || r.URL.Path != "/data" {
				t.Fatalf("request = %s %s, want POST /data", r.Method, r.URL.Path)
			}
			assertSignedForm(t, r, "test-key", siteLookupForm("academy.example.com"))
			return http.StatusOK, `{"data":[]}`, nil
		}
		if r.Method != http.MethodPost || r.URL.Path != "/site" {
			t.Fatalf("request = %s %s, want POST /site", r.Method, r.URL.Path)
		}
		assertSignedForm(t, r, "test-key", url.Values{
			"action":  {"AddSite"},
			"webname": {`{"count":0,"domain":"academy.example.com","domainlist":[]}`},
			"path":    {"/www/wwwroot/academy.example.com"},
			"type":    {""},
			"version": {"00"},
			"port":    {"80"},
			"ps":      {"ImaiPlay automated tenant site"},
			"ftp":     {"false"},
			"sql":     {"false"},
			"codeing": {"utf-8"},
		})
		return http.StatusOK, `{"siteStatus":true,"siteId":37}`, nil
	})

	client := testClient(httpClient)
	siteID, err := client.AddSite("academy.example.com")
	if err != nil {
		t.Fatalf("AddSite() error = %v", err)
	}
	if siteID != 37 {
		t.Fatalf("site ID = %d, want 37", siteID)
	}
}

func TestAddSiteRejectsExistingSiteWithoutChangingIt(t *testing.T) {
	requests := 0
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		requests++
		assertSignedForm(t, r, "test-key", siteLookupForm("academy.example.com"))
		return http.StatusOK, `{"data":[{"id":37,"name":"academy.example.com","ps":"existing site"}]}`, nil
	})

	client := testClient(httpClient)
	if _, err := client.AddSite("academy.example.com"); err == nil ||
		!strings.Contains(err.Error(), "already exists") {
		t.Fatalf("AddSite() error = %v, want existing-site rejection", err)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want lookup only", requests)
	}
}

func TestAddReverseProxyUsesOfficialCreateProxyParameters(t *testing.T) {
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		if r.URL.Path != "/site" {
			t.Fatalf("path = %s, want /site", r.URL.Path)
		}
		assertSignedForm(t, r, "test-key", url.Values{
			"action":    {"CreateProxy"},
			"sitename":  {"academy.example.com"},
			"proxyname": {"imaiplay"},
			"proxydir":  {"/"},
			"proxysite": {"http://127.0.0.1:18080"},
			"type":      {"1"},
			"cache":     {"0"},
			"cachetime": {"0"},
			"todomain":  {"academy.example.com"},
			"subfilter": {emptySubfilters},
			"advanced":  {"0"},
		})
		return http.StatusOK, `{"status":true,"msg":"添加成功"}`, nil
	})
	client := testClient(httpClient)
	if err := client.AddReverseProxy(
		"academy.example.com",
		"imaiplay",
		"http://127.0.0.1:18080",
	); err != nil {
		t.Fatalf("AddReverseProxy() error = %v", err)
	}
}

func TestEnableHTTPSRedirectUsesOfficialSiteAction(t *testing.T) {
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/site" {
			t.Fatalf("request = %s %s, want POST /site", r.Method, r.URL.Path)
		}
		assertSignedForm(t, r, "test-key", url.Values{
			"action":   {"HttpToHttps"},
			"siteName": {"academy.example.com"},
		})
		return http.StatusOK, `{"status":true,"msg":"设置成功"}`, nil
	})

	client := testClient(httpClient)
	if err := client.EnableHTTPSRedirect("academy.example.com"); err != nil {
		t.Fatalf("EnableHTTPSRedirect() error = %v", err)
	}
}

func TestApplyLetsEncryptCreatesHTTPACMEOrderAndDeploysWhilePolling(t *testing.T) {
	var requests []string
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		requests = append(requests, r.URL.Path+":"+r.PostForm.Get("action"))
		switch len(requests) {
		case 1:
			assertSignedForm(t, r, "test-key", siteLookupForm("academy.example.com"))
			return http.StatusOK, `{"data":[{"id":9,"name":"academy.example.com","path":"/www/wwwroot/academy.example.com"}]}`, nil
		case 2:
			assertSignedForm(t, r, "test-key", url.Values{
				"action":        {"apply_cert_api"},
				"id":            {"9"},
				"domains":       {`["academy.example.com"]`},
				"auth_type":     {"http"},
				"auth_to":       {"/www/wwwroot/academy.example.com"},
				"auto_wildcard": {"0"},
				"ca":            {"letsencrypt"},
			})
			return http.StatusOK, `{"status":"pending","index":"order-1"}`, nil
		case 3:
			assertSignedForm(t, r, "test-key", siteLookupForm("academy.example.com"))
			return http.StatusOK, `{"data":[{"id":9,"name":"academy.example.com","path":"/www/wwwroot/academy.example.com"}]}`, nil
		case 4:
			assertSignedForm(t, r, "test-key", url.Values{
				"action": {"get_order_find"}, "index": {"order-1"},
			})
			return http.StatusOK, `{"status":"valid","index":"order-1","cert":{}}`, nil
		case 5:
			assertSignedForm(t, r, "test-key", url.Values{
				"action": {"SetCertToSite"}, "index": {"order-1"},
				"siteName": {"academy.example.com"},
			})
			return http.StatusOK, `{"status":true,"msg":"SSL开启成功!"}`, nil
		case 6:
			assertSignedForm(t, r, "test-key", url.Values{
				"action": {"GetSSL"}, "siteName": {"academy.example.com"},
			})
			return http.StatusOK, `{"status":true,"key":true,"csr":true,"cert_data":{"issuer":"Let's Encrypt"}}`, nil
		default:
			t.Fatalf("unexpected request %d: %s", len(requests), r.URL.Path)
			return 0, "", nil
		}
	})

	client := testClient(httpClient)
	if err := client.ApplyLetsEncrypt("academy.example.com"); err != nil {
		t.Fatalf("ApplyLetsEncrypt() error = %v", err)
	}
	info, err := client.GetSiteInfo("academy.example.com")
	if err != nil {
		t.Fatalf("GetSiteInfo() error = %v", err)
	}
	if ready, _ := info["ssl_status"].(bool); !ready {
		t.Fatalf("site info = %#v, want ssl_status=true", info)
	}
	if got := strings.Join(requests, ","); got != "/data:getData,/acme:apply_cert_api,/data:getData,/acme:get_order_find,/acme:SetCertToSite,/site:GetSSL" {
		t.Fatalf("requests = %s", got)
	}
}

func TestApplyLetsEncryptDeploysSynchronousCertificateResponse(t *testing.T) {
	requests := 0
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		requests++
		switch requests {
		case 1:
			assertSignedForm(t, r, "test-key", siteLookupForm("academy.example.com"))
			return http.StatusOK, `{"data":[{"id":9,"name":"academy.example.com","path":"/www/wwwroot/academy.example.com"}]}`, nil
		case 2:
			assertSignedForm(t, r, "test-key", url.Values{
				"action":        {"apply_cert_api"},
				"id":            {"9"},
				"domains":       {`["academy.example.com"]`},
				"auth_type":     {"http"},
				"auth_to":       {"/www/wwwroot/academy.example.com"},
				"auto_wildcard": {"0"},
				"ca":            {"letsencrypt"},
			})
			return http.StatusOK, `{"status":true,"msg":"申请成功!","cert":"-----BEGIN CERTIFICATE-----\nleaf\n-----END CERTIFICATE-----\n","root":"-----BEGIN CERTIFICATE-----\nroot\n-----END CERTIFICATE-----\n","private_key":"-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----\n"}`, nil
		case 3:
			assertSignedForm(t, r, "test-key", url.Values{
				"action":   {"SetSSL"},
				"siteName": {"academy.example.com"},
				"key":      {"-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----\n"},
				"csr":      {"-----BEGIN CERTIFICATE-----\nleaf\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nroot\n-----END CERTIFICATE-----\n"},
			})
			return http.StatusOK, `{"status":true,"msg":"证书已保存!"}`, nil
		default:
			t.Fatalf("unexpected request %d: %s", requests, r.URL.Path)
			return 0, "", nil
		}
	})

	client := testClient(httpClient)
	if err := client.ApplyLetsEncrypt("academy.example.com"); err != nil {
		t.Fatalf("ApplyLetsEncrypt() error = %v", err)
	}
	if requests != 3 {
		t.Fatalf("requests = %d, want lookup, apply and deploy", requests)
	}
}

func TestSynchronousCertificateDeploymentRemainsReadyWhenGetSSLIsStale(t *testing.T) {
	requests := 0
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		requests++
		switch requests {
		case 1:
			return http.StatusOK, `{"data":[{"id":9,"name":"academy.example.com","path":"/www/wwwroot/academy.example.com"}]}`, nil
		case 2:
			return http.StatusOK, `{"status":true,"cert":"-----BEGIN CERTIFICATE-----\nleaf\n-----END CERTIFICATE-----\n","root":"-----BEGIN CERTIFICATE-----\nroot\n-----END CERTIFICATE-----\n","private_key":"-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----\n"}`, nil
		case 3:
			return http.StatusOK, `{"status":true,"msg":"证书已保存!"}`, nil
		case 4:
			return http.StatusOK, `{"data":[{"id":9,"name":"academy.example.com","path":"/www/wwwroot/academy.example.com"}]}`, nil
		case 5:
			return http.StatusOK, `{"status":false,"key":false,"csr":false,"cert_data":{}}`, nil
		default:
			t.Fatalf("unexpected request %d: %s", requests, r.URL.Path)
			return 0, "", nil
		}
	})

	client := testClient(httpClient)
	if err := client.ApplyLetsEncrypt("academy.example.com"); err != nil {
		t.Fatalf("ApplyLetsEncrypt() error = %v", err)
	}
	info, err := client.GetSiteInfo("academy.example.com")
	if err != nil {
		t.Fatalf("GetSiteInfo() error = %v", err)
	}
	if ready, _ := info["ssl_status"].(bool); !ready {
		t.Fatalf("site info = %#v, want successful SetSSL to remain ready", info)
	}
}

func TestGetSiteInfoLeavesPendingACMEOrderForNextPoll(t *testing.T) {
	requests := 0
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		requests++
		switch requests {
		case 1:
			return http.StatusOK, `{"data":[{"id":9,"name":"academy.example.com"}]}`, nil
		case 2:
			return http.StatusOK, `{"status":"pending","index":"order-1"}`, nil
		case 3:
			return http.StatusOK, `{"status":false,"key":false,"csr":false}`, nil
		case 4:
			return http.StatusOK, `{"data":[{"id":9,"name":"academy.example.com"}]}`, nil
		case 5:
			return http.StatusOK, `{"status":"pending","index":"order-1"}`, nil
		case 6:
			return http.StatusOK, `{"status":false,"key":false,"csr":false}`, nil
		default:
			t.Fatalf("unexpected request %d", requests)
			return 0, "", nil
		}
	})
	client := testClient(httpClient)
	client.pendingCerts.Store("academy.example.com", "order-1")

	for poll := 0; poll < 2; poll++ {
		info, err := client.GetSiteInfo("academy.example.com")
		if err != nil {
			t.Fatalf("GetSiteInfo() poll %d error = %v", poll+1, err)
		}
		if ready, _ := info["ssl_status"].(bool); ready {
			t.Fatalf("site info poll %d = %#v, want ssl_status=false", poll+1, info)
		}
	}
	if requests != 6 {
		t.Fatalf("requests = %d, want 6 without SetCertToSite", requests)
	}
}

func TestNginxSnippetWritesTestsAndReloadsSiteConfig(t *testing.T) {
	requests := 0
	savedConfig := ""
	snippet := `# ImaiPlay tenant admin guard
location ^~ /admin {
    return 404;
}`
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		requests++
		switch requests {
		case 1:
			assertSignedForm(t, r, "test-key", url.Values{
				"action": {"GetFileBody"},
				"path":   {"/www/server/panel/vhost/nginx/academy.example.com.conf"},
			})
			return http.StatusOK, `{"status":true,"data":"server {\n    listen 80;\n}\n"}`, nil
		case 2, 3:
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.URL.Path != "/files" ||
				r.PostForm.Get("action") != "SaveFileBody" ||
				r.PostForm.Get("path") != "/www/server/panel/vhost/nginx/academy.example.com.conf" ||
				r.PostForm.Get("encoding") != "utf-8" {
				t.Fatalf("unexpected SaveFileBody request: %s %v", r.URL.Path, r.PostForm)
			}
			savedConfig = r.PostForm.Get("data")
			if !strings.Contains(savedConfig, snippet) {
				t.Fatalf("saved config does not contain admin guard: %q", savedConfig)
			}
			if requests == 2 {
				return http.StatusBadGateway, "temporary", nil
			}
			return http.StatusOK, `{"status":true}`, nil
		case 4:
			assertSignedForm(t, r, "test-key", url.Values{
				"action": {"GetFileBody"},
				"path":   {"/www/server/panel/vhost/nginx/academy.example.com.conf"},
			})
			body := fmt.Sprintf(`{"status":true,"data":%q}`, savedConfig)
			return http.StatusOK, body, nil
		case 5:
			assertSignedForm(t, r, "test-key", url.Values{
				"action": {"ServiceAdmin"}, "name": {"nginx"}, "type": {"test"},
			})
			return http.StatusOK, `{"status":true}`, nil
		case 6:
			assertSignedForm(t, r, "test-key", url.Values{
				"action": {"ServiceAdmin"}, "name": {"nginx"}, "type": {"reload"},
			})
			return http.StatusOK, `{"status":true}`, nil
		default:
			t.Fatalf("unexpected request %d: %s", requests, r.URL.Path)
			return 0, "", nil
		}
	})

	client := testClient(httpClient)
	if err := client.AddNginxSnippet(
		"academy.example.com",
		snippet,
	); err != nil {
		t.Fatalf("AddNginxSnippet() error = %v", err)
	}
	if requests != 6 {
		t.Fatalf("requests = %d, want read/save retry/verify/test/reload", requests)
	}
}

func TestDeleteSiteLooksUpIDAndUsesOfficialParameters(t *testing.T) {
	requests := 0
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		requests++
		switch requests {
		case 1:
			assertSignedForm(t, r, "test-key", siteLookupForm("academy.example.com"))
			return http.StatusOK, `{"data":[{"id":19,"name":"academy.example.com"}]}`, nil
		case 2:
			assertSignedForm(t, r, "test-key", url.Values{
				"action": {"DeleteSite"}, "id": {"19"},
				"webname": {"academy.example.com"},
			})
			return http.StatusOK, `{"status":true}`, nil
		default:
			t.Fatalf("unexpected request %d", requests)
			return 0, "", nil
		}
	})
	client := testClient(httpClient)
	if err := client.DeleteSite("academy.example.com"); err != nil {
		t.Fatalf("DeleteSite() error = %v", err)
	}
}

func TestGetSiteInfoRetriesSafeQueriesOnce(t *testing.T) {
	attempts := 0
	httpClient := mockHTTPClient(func(r *http.Request) (int, string, error) {
		attempts++
		if attempts == 1 {
			return 0, "", errors.New("connection reset")
		}
		if attempts == 2 {
			return http.StatusOK, `{"data":[{"id":9,"name":"academy.example.com"}]}`, nil
		}
		return http.StatusOK, `{"status":false,"key":false,"csr":false,"cert_data":{}}`, nil
	})
	client := testClient(httpClient)
	info, err := client.GetSiteInfo("academy.example.com")
	if err != nil {
		t.Fatalf("GetSiteInfo() error = %v", err)
	}
	if info["id"] != float64(9) || info["ssl_status"] != false {
		t.Fatalf("info = %#v", info)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestClientReportsHTTPAndBusinessErrors(t *testing.T) {
	tests := []struct {
		name, body, want string
		status           int
	}{
		{name: "http", status: http.StatusUnauthorized, body: `invalid key`, want: "HTTP 401"},
		{name: "business", status: http.StatusOK, body: `{"status":false,"msg":"permission denied"}`, want: "permission denied"},
		{name: "string response", status: http.StatusOK, body: `"invalid token"`, want: "invalid token"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := testClient(mockHTTPClient(func(*http.Request) (int, string, error) {
				return test.status, test.body, nil
			}))
			err := client.AddReverseProxy("academy.example.com", "imaiplay", "http://127.0.0.1:18080")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func testClient(httpClient *http.Client) Client {
	return Client{
		PanelURL:   "http://baota.test",
		APIKey:     "test-key",
		HTTPClient: httpClient,
		now: func() time.Time {
			return fixedBaotaTime
		},
	}
}

func siteLookupForm(domain string) url.Values {
	return url.Values{
		"action": {"getData"}, "table": {"sites"}, "type": {"-1"},
		"search": {domain}, "p": {"1"}, "limit": {"20"},
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func mockHTTPClient(
	handler func(*http.Request) (int, string, error),
) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		status, body, err := handler(request)
		if err != nil {
			return nil, err
		}
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    request,
		}, nil
	})}
}

func assertSignedForm(
	t *testing.T,
	request *http.Request,
	apiKey string,
	want url.Values,
) {
	t.Helper()
	if err := request.ParseForm(); err != nil {
		t.Fatalf("ParseForm() error = %v", err)
	}
	got := cloneValues(request.PostForm)
	requestTime := got.Get("request_time")
	requestToken := got.Get("request_token")
	got.Del("request_time")
	got.Del("request_token")
	if requestTime != "1700000000" {
		t.Fatalf("request_time = %q", requestTime)
	}
	keyDigest := md5.Sum([]byte(apiKey))
	tokenDigest := md5.Sum([]byte(requestTime + hex.EncodeToString(keyDigest[:])))
	if wantToken := hex.EncodeToString(tokenDigest[:]); requestToken != wantToken {
		t.Fatalf("request_token = %q, want %q", requestToken, wantToken)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("form = %v, want %v", got, want)
	}
}
