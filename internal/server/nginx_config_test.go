package server

import (
	"os"
	"strings"
	"testing"
)

func TestDockerNginxAllowsOneGiBUploads(t *testing.T) {
	config, err := os.ReadFile("../../docker/nginx/conf/nginx.conf")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(config), "client_max_body_size 1024m;") {
		t.Fatal("Docker Nginx upload limit is not 1024m")
	}
}
