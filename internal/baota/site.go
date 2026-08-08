package baota

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const emptySubfilters = `[{"sub1":"","sub2":""},{"sub1":"","sub2":""},{"sub1":"","sub2":""}]`

// AddSite creates a static site at BaoTa's standard document-root path.
func (client *Client) AddSite(domain string) (int, error) {
	if existing, err := client.lookupSite(context.Background(), domain); err == nil {
		if id, ok := mapInt(existing, "id"); ok {
			return 0, fmt.Errorf(
				"baota site %s already exists with id %d",
				domain,
				id,
			)
		}
		return 0, fmt.Errorf("baota site %s already exists", domain)
	} else if !isMissingSiteError(err) {
		return 0, fmt.Errorf("check existing baota site: %w", err)
	}

	webname, err := json.Marshal(map[string]interface{}{
		"domain": domain, "domainlist": []string{}, "count": 0,
	})
	if err != nil {
		return 0, fmt.Errorf("encode baota site name: %w", err)
	}
	response, err := client.request(context.Background(), "/site", url.Values{
		"action":  {"AddSite"},
		"webname": {string(webname)},
		"path":    {siteRoot(domain)},
		"type":    {""},
		"version": {"00"},
		"port":    {"80"},
		"ps":      {"ImaiPlay automated tenant site"},
		"ftp":     {"false"},
		"sql":     {"false"},
		"codeing": {"utf-8"},
	}, true, false)
	if err != nil {
		if existing, lookupErr := client.lookupSite(context.Background(), domain); lookupErr == nil {
			if id, ok := mapInt(existing, "id"); ok {
				return id, nil
			}
		}
		return 0, err
	}

	for _, key := range []string{"siteId", "site_id", "id"} {
		if id, ok := response.intValue(key); ok && id > 0 {
			return id, nil
		}
	}
	if raw, ok := response["data"]; ok {
		var data apiResponse
		if json.Unmarshal(raw, &data) == nil {
			for _, key := range []string{"siteId", "site_id", "id"} {
				if id, ok := data.intValue(key); ok && id > 0 {
					return id, nil
				}
			}
		}
	}
	return 0, errors.New("baota API AddSite response has no siteId")
}

// AddReverseProxy configures a root reverse proxy for a site.
func (client *Client) AddReverseProxy(domain, proxyName, target string) error {
	_, err := client.request(context.Background(), "/site", url.Values{
		"action":    {"CreateProxy"},
		"sitename":  {domain},
		"proxyname": {proxyName},
		"proxydir":  {"/"},
		"proxysite": {target},
		"type":      {"1"},
		"cache":     {"0"},
		"cachetime": {"0"},
		"todomain":  {domain},
		"subfilter": {emptySubfilters},
		"advanced":  {"0"},
	}, true, false)
	if err == nil {
		return nil
	}
	exists, lookupErr := client.reverseProxyExists(
		context.Background(),
		domain,
		proxyName,
		target,
	)
	if lookupErr == nil && exists {
		return nil
	}
	return err
}

// ApplyLetsEncrypt creates a non-wildcard HTTP-validated ACME order.
// The order is deployed lazily while GetSiteInfo polls its status.
func (client *Client) ApplyLetsEncrypt(domain string) error {
	info, err := client.lookupSite(context.Background(), domain)
	if err != nil {
		return err
	}
	siteID, ok := mapInt(info, "id")
	if !ok {
		return errors.New("baota site has no id")
	}
	root, _ := info["path"].(string)
	if strings.TrimSpace(root) == "" {
		root = siteRoot(domain)
	}
	domains, err := json.Marshal([]string{domain})
	if err != nil {
		return fmt.Errorf("encode certificate domains: %w", err)
	}
	response, err := client.request(context.Background(), "/acme", url.Values{
		"action":        {"apply_cert_api"},
		"id":            {strconv.Itoa(siteID)},
		"domains":       {string(domains)},
		"auth_type":     {"http"},
		"auth_to":       {root},
		"auto_wildcard": {"0"},
		"ca":            {"letsencrypt"},
	}, true, false)
	if err != nil {
		return err
	}
	index, ok := response.stringValue("index")
	if ok && strings.TrimSpace(index) != "" {
		client.pendingCerts.Store(domain, index)
		return nil
	}

	certificate, hasCertificate := response.stringValue("cert")
	privateKey, hasPrivateKey := response.stringValue("private_key")
	if !hasCertificate || strings.TrimSpace(certificate) == "" ||
		!hasPrivateKey || strings.TrimSpace(privateKey) == "" {
		return errors.New("baota ACME response has neither order index nor certificate material")
	}
	certificateRoot, _ := response.stringValue("root")
	_, err = client.request(context.Background(), "/site", url.Values{
		"action":   {"SetSSL"},
		"siteName": {domain},
		"key":      {privateKey},
		"csr":      {certificate + certificateRoot},
	}, true, false)
	if err != nil {
		return fmt.Errorf("deploy baota ACME certificate: %w", err)
	}
	client.deployedCerts.Store(domain, true)
	return nil
}

// EnableHTTPSRedirect forces all HTTP requests for a site to HTTPS.
func (client *Client) EnableHTTPSRedirect(domain string) error {
	_, err := client.request(context.Background(), "/site", url.Values{
		"action":   {"HttpToHttps"},
		"siteName": {domain},
	}, true, false)
	return err
}

// AddNginxSnippet inserts a rule into the site's actual Nginx configuration,
// verifies the saved content, then tests and reloads Nginx.
func (client *Client) AddNginxSnippet(domain, snippet string) error {
	if strings.ContainsAny(domain, `/\`) {
		return errors.New("invalid baota site domain")
	}
	ctx := context.Background()
	path := nginxSiteConfigPath(domain)
	current, err := client.readFile(ctx, path)
	if err != nil {
		return err
	}
	if !strings.Contains(current, snippet) {
		closingBrace := strings.LastIndex(current, "}")
		if closingBrace < 0 {
			return errors.New("baota Nginx site config has no server closing brace")
		}
		updated := current[:closingBrace] +
			"\n" + strings.TrimSpace(snippet) +
			"\n" + current[closingBrace:]
		if _, err := client.request(ctx, "/files", url.Values{
			"action":   {"SaveFileBody"},
			"path":     {path},
			"data":     {updated},
			"encoding": {"utf-8"},
		}, true, false); err != nil {
			return err
		}
	}
	saved, err := client.readFile(ctx, path)
	if err != nil {
		return err
	}
	if !strings.Contains(saved, snippet) {
		return errors.New("baota Nginx admin guard was not saved")
	}
	if _, err := client.request(ctx, "/system", url.Values{
		"action": {"ServiceAdmin"},
		"name":   {"nginx"},
		"type":   {"test"},
	}, true, false); err != nil {
		return fmt.Errorf("test baota Nginx config: %w", err)
	}
	if _, err := client.request(ctx, "/system", url.Values{
		"action": {"ServiceAdmin"},
		"name":   {"nginx"},
		"type":   {"reload"},
	}, true, false); err != nil {
		return fmt.Errorf("reload baota Nginx config: %w", err)
	}
	return nil
}

// DeleteSite removes a site by exact domain name.
func (client *Client) DeleteSite(domain string) error {
	info, err := client.lookupSite(context.Background(), domain)
	if err != nil {
		if isMissingSiteError(err) {
			return nil
		}
		return err
	}
	siteID, ok := mapInt(info, "id")
	if !ok {
		return errors.New("baota site has no id")
	}
	_, err = client.request(context.Background(), "/site", url.Values{
		"action":  {"DeleteSite"},
		"id":      {strconv.Itoa(siteID)},
		"webname": {domain},
	}, true, false)
	if err == nil {
		client.pendingCerts.Delete(domain)
		client.deployedCerts.Delete(domain)
		return nil
	}
	if _, lookupErr := client.lookupSite(context.Background(), domain); isMissingSiteError(lookupErr) {
		client.pendingCerts.Delete(domain)
		client.deployedCerts.Delete(domain)
		return nil
	}
	return err
}

// GetSiteInfo returns panel site data combined with its SSL deployment state.
func (client *Client) GetSiteInfo(domain string) (map[string]interface{}, error) {
	ctx := context.Background()
	info, err := client.lookupSite(ctx, domain)
	if err != nil {
		return nil, err
	}

	var certificateErr error
	if rawIndex, exists := client.pendingCerts.Load(domain); exists {
		index, _ := rawIndex.(string)
		if index != "" {
			var order apiResponse
			order, certificateErr = client.request(ctx, "/acme", url.Values{
				"action": {"get_order_find"},
				"index":  {index},
			}, true, false)
			if certificateErr == nil {
				status, _ := order.stringValue("status")
				switch strings.ToLower(strings.TrimSpace(status)) {
				case "valid":
					_, certificateErr = client.request(ctx, "/acme", url.Values{
						"action":   {"SetCertToSite"},
						"index":    {index},
						"siteName": {domain},
					}, true, false)
					if certificateErr == nil {
						client.pendingCerts.Delete(domain)
						client.deployedCerts.Store(domain, true)
					}
				case "pending", "processing":
					// ACME validation is asynchronous; the next poll checks again.
				case "expired":
					certificateErr = errors.New("baota ACME order expired")
				default:
					certificateErr = fmt.Errorf(
						"baota ACME order has unexpected status %q",
						status,
					)
				}
			}
		}
	}

	sslResponse, sslErr := client.request(ctx, "/site", url.Values{
		"action":   {"GetSSL"},
		"siteName": {domain},
	}, true, true)
	if sslErr != nil {
		return nil, sslErr
	}
	sslStatus, _ := sslResponse.boolValue("status")
	hasKey, _ := sslResponse.boolValue("key")
	hasCertificate, _ := sslResponse.boolValue("csr")
	_, deploymentConfirmed := client.deployedCerts.Load(domain)
	info["ssl_status"] = deploymentConfirmed || (sslStatus && hasKey && hasCertificate)
	if raw, ok := sslResponse["cert_data"]; ok {
		var certData map[string]interface{}
		if json.Unmarshal(raw, &certData) == nil {
			info["certificate"] = certData
		}
	}
	if certificateErr != nil && !sslStatus {
		return nil, certificateErr
	}
	return info, nil
}

func (client *Client) lookupSite(
	ctx context.Context,
	domain string,
) (map[string]interface{}, error) {
	response, err := client.request(ctx, "/data", url.Values{
		"action": {"getData"},
		"table":  {"sites"},
		"type":   {"-1"},
		"search": {domain},
		"p":      {"1"},
		"limit":  {"20"},
	}, true, false)
	if err != nil {
		return nil, err
	}
	raw, ok := response["data"]
	if !ok {
		return nil, errors.New("baota site not found")
	}
	var sites []map[string]interface{}
	if err := json.Unmarshal(raw, &sites); err != nil {
		return nil, fmt.Errorf("invalid site list from baota API: %w", err)
	}
	for _, site := range sites {
		name, _ := site["name"].(string)
		if strings.EqualFold(strings.TrimSpace(name), domain) {
			return site, nil
		}
	}
	return nil, errors.New("baota site not found")
}

func (client *Client) reverseProxyExists(
	ctx context.Context,
	domain, proxyName, target string,
) (bool, error) {
	response, err := client.request(ctx, "/site", url.Values{
		"action":   {"GetProxyList"},
		"sitename": {domain},
	}, true, false)
	if err != nil {
		return false, err
	}
	raw, ok := response["_list"]
	if !ok {
		return false, errors.New("baota proxy list response is not an array")
	}
	var proxies []map[string]interface{}
	if err := json.Unmarshal(raw, &proxies); err != nil {
		return false, fmt.Errorf("invalid proxy list from baota API: %w", err)
	}
	for _, proxy := range proxies {
		name, _ := proxy["proxyname"].(string)
		site, _ := proxy["proxysite"].(string)
		path, _ := proxy["proxydir"].(string)
		if name == proxyName && site == target && path == "/" {
			return true, nil
		}
	}
	return false, nil
}

func siteRoot(domain string) string {
	return "/www/wwwroot/" + domain
}

func nginxSiteConfigPath(domain string) string {
	return "/www/server/panel/vhost/nginx/" + domain + ".conf"
}

func (client *Client) readFile(
	ctx context.Context,
	path string,
) (string, error) {
	response, err := client.request(ctx, "/files", url.Values{
		"action": {"GetFileBody"},
		"path":   {path},
	}, true, false)
	if err != nil {
		return "", err
	}
	content, ok := response.stringValue("data")
	if !ok {
		return "", errors.New("baota file response has no text content")
	}
	return content, nil
}

func mapInt(values map[string]interface{}, key string) (int, bool) {
	value, exists := values[key]
	if !exists {
		return 0, false
	}
	switch current := value.(type) {
	case float64:
		return int(current), current > 0
	case int:
		return current, current > 0
	case string:
		id, err := strconv.Atoi(current)
		return id, err == nil && id > 0
	default:
		return 0, false
	}
}

func isMissingSiteError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}
