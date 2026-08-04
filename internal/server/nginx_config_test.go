package server

import (
	"os"
	"strings"
	"testing"
)

func readNginxConfig(t *testing.T) string {
	t.Helper()
	config, err := os.ReadFile("../../docker/nginx/conf/nginx.conf")
	if err != nil {
		t.Fatal(err)
	}
	return string(config)
}

func nginxLocationBlock(t *testing.T, config, location string) string {
	t.Helper()
	start := strings.Index(config, location+" {")
	if start < 0 {
		t.Fatalf("missing Nginx location %q", location)
	}
	open := strings.Index(config[start:], "{")
	if open < 0 {
		t.Fatalf("location %q has no opening brace", location)
	}
	open += start
	depth := 0
	for index := open; index < len(config); index++ {
		switch config[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return config[start : index+1]
			}
		}
	}
	t.Fatalf("location %q has no closing brace", location)
	return ""
}

func TestDockerNginxAllowsOneGiBUploads(t *testing.T) {
	config := readNginxConfig(t)
	if !strings.Contains(config, "client_max_body_size 1024m;") {
		t.Fatal("Docker Nginx upload limit is not 1024m")
	}
}

func TestNginxServesTenantPortalRoutesFromPCApp(t *testing.T) {
	config := readNginxConfig(t)
	for _, location := range []string{
		"location = /login",
		"location = /select-organization",
		"location /t/",
		"location = /",
	} {
		block := nginxLocationBlock(t, config, location)
		if !strings.Contains(block, "/usr/share/nginx/html/pc/index.html") &&
			!strings.Contains(block, "/pc/index.html") {
			t.Fatalf("%s does not serve the PC app:\n%s", location, block)
		}
	}
}

func TestNginxRootPortalDoesNotAliasDirectoryURIToFile(t *testing.T) {
	config := readNginxConfig(t)
	block := nginxLocationBlock(t, config, "location = /")

	if strings.Contains(block, "alias ") {
		t.Fatalf("root directory URI aliases a file and can enter Nginx index redirects:\n%s", block)
	}
	if !strings.Contains(block, "try_files /pc/index.html =404;") {
		t.Fatalf("root portal does not safely resolve the PC entry document:\n%s", block)
	}
}

func TestNginxKeepsSPAAssetLocationsReachable(t *testing.T) {
	config := readNginxConfig(t)
	for _, forbidden := range []string{
		"location ^~ /admin/",
		"location ^~ /pc/",
		"location ^~ /h5/",
		"location ^~ /t/",
	} {
		if strings.Contains(config, forbidden) {
			t.Fatalf("%q prevents the hashed asset cache location from matching", forbidden)
		}
	}
	assetBlock := nginxLocationBlock(t, config, `location ~* ^/(admin|pc|h5)/assets/.*\.[a-z0-9]+$`)
	if !strings.Contains(assetBlock, `Cache-Control "public, immutable"`) ||
		!strings.Contains(assetBlock, "expires 1y;") {
		t.Fatalf("hashed assets do not use immutable caching:\n%s", assetBlock)
	}
}

func TestNginxDisablesCacheForEverySPAEntry(t *testing.T) {
	config := readNginxConfig(t)
	for _, location := range []string{
		"location = /admin/index.html",
		"location = /pc/index.html",
		"location = /h5/index.html",
		"location = /login",
		"location = /select-organization",
		"location /t/",
		"location = /",
	} {
		block := nginxLocationBlock(t, config, location)
		if !strings.Contains(block, `Cache-Control "no-cache, no-store, must-revalidate"`) {
			t.Fatalf("%s does not disable cache:\n%s", location, block)
		}
	}
}
