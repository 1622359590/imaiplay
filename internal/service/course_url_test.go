package service

import "testing"

func TestSafePublicContentURL(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		raw  string
		want bool
	}{
		{name: "public HTTPS domain", raw: "https://cdn.example.com/media/video.mp4", want: true},
		{name: "public HTTP domain and port", raw: "http://media.example.org:8080/guide.pdf", want: true},
		{name: "AWS lookalike domain", raw: "https://notamazonaws.com/media/video.mp4", want: true},
		{name: "Google APIs lookalike domain", raw: "https://example-googleapis.com/media/video.mp4", want: true},
		{name: "provider root below public domain", raw: "https://amazonaws.com.example.com/media/video.mp4", want: true},
		{name: "Azure root below public domain", raw: "https://core.windows.net.example.com/media/video.mp4", want: true},
		{name: "credentials", raw: "https://user:password@example.com/media.mp4"},
		{name: "localhost", raw: "http://localhost/media.mp4"},
		{name: "localhost trailing dot", raw: "http://localhost./media.mp4"},
		{name: "localhost subdomain", raw: "http://media.localhost/video.mp4"},
		{name: "localhost subdomain trailing dot", raw: "http://media.localhost./video.mp4"},
		{name: "IPv4 short loopback", raw: "http://127.1/video.mp4"},
		{name: "IPv4 integer loopback", raw: "http://2130706433/video.mp4"},
		{name: "IPv4 hexadecimal loopback", raw: "http://0x7f000001/video.mp4"},
		{name: "IPv4 octal loopback", raw: "http://0177.0.0.1/video.mp4"},
		{name: "IPv4 mixed numeric labels", raw: "http://127.0x0.0.1/video.mp4"},
		{name: "IPv4 zero padded labels", raw: "http://127.000.000.001/video.mp4"},
		{name: "IPv4 loopback", raw: "http://127.0.0.1/video.mp4"},
		{name: "IPv4 private class A", raw: "http://10.0.0.1/video.mp4"},
		{name: "IPv4 private class B", raw: "http://172.16.0.1/video.mp4"},
		{name: "IPv4 private class C", raw: "http://192.168.1.1/video.mp4"},
		{name: "IPv4 link local", raw: "http://169.254.1.1/video.mp4"},
		{name: "IPv4 unspecified", raw: "http://0.0.0.0/video.mp4"},
		{name: "IPv4 public literal", raw: "http://8.8.8.8/video.mp4"},
		{name: "IPv6 loopback", raw: "http://[::1]/video.mp4"},
		{name: "IPv6 unspecified", raw: "http://[::]/video.mp4"},
		{name: "IPv6 private", raw: "http://[fd00::1]/video.mp4"},
		{name: "IPv6 link local", raw: "http://[fe80::1]/video.mp4"},
		{name: "IPv6 public literal", raw: "http://[2001:4860:4860::8888]/video.mp4"},
		{name: "IPv4 mapped private", raw: "http://[::ffff:127.0.0.1]/video.mp4"},
		{name: "IPv4 mapped public", raw: "http://[::ffff:8.8.8.8]/video.mp4"},
		{name: "IPv6 zone", raw: "http://[fe80::1%25en0]/video.mp4"},
		{name: "empty authority", raw: "https:///media.mp4"},
		{name: "missing host", raw: "https://"},
		{name: "opaque authority", raw: "https:example.com/media.mp4"},
		{name: "scheme relative", raw: "//example.com/media.mp4"},
		{name: "empty DNS label", raw: "https://media..example.com/video.mp4"},
		{name: "leading empty DNS label", raw: "https://.example.com/video.mp4"},
		{name: "backslash authority ambiguity", raw: "https://example.com\\@127.0.0.1/video.mp4"},
		{name: "javascript scheme", raw: "javascript:alert(1)"},
		{name: "file scheme", raw: "file:///etc/passwd"},
		{name: "data scheme", raw: "data:text/html,hello"},
		{name: "S3 virtual hosted object", raw: "https://bucket.s3.amazonaws.com/private/video.mp4"},
		{name: "S3 regional object", raw: "https://bucket.s3.us-east-1.amazonaws.com/private/video.mp4"},
		{name: "S3 regional path object", raw: "https://s3.us-east-1.amazonaws.com/private-bucket/video.mp4"},
		{name: "S3 legacy regional path object", raw: "https://s3-us-west-2.amazonaws.com/private-bucket/video.mp4"},
		{name: "S3 accelerate object", raw: "https://bucket.s3-accelerate.amazonaws.com/private/video.mp4"},
		{name: "S3 Express object", raw: "https://bucket.s3express-use1-az4.us-east-1.amazonaws.com/private/video.mp4"},
		{name: "AWS provider root", raw: "https://media.execute-api.amazonaws.com/private/video.mp4"},
		{name: "S3 China virtual hosted object", raw: "https://bucket.s3.cn-north-1.amazonaws.com.cn/private/video.mp4"},
		{name: "S3 China path object", raw: "https://s3.cn-north-1.amazonaws.com.cn/private-bucket/video.mp4"},
		{name: "Aliyun OSS object", raw: "https://bucket.oss-cn-shanghai.aliyuncs.com/private/video.mp4"},
		{name: "Google Cloud Storage object", raw: "https://bucket.storage.googleapis.com/private/video.mp4"},
		{name: "Google Cloud authenticated object", raw: "https://storage.cloud.google.com/bucket/private/video.mp4"},
		{name: "Google Cloud media object", raw: "https://content-storage.googleapis.com/download/storage/v1/b/bucket/o/private?alt=media"},
		{name: "Google APIs provider root", raw: "https://maps.googleapis.com/private/video.mp4"},
		{name: "Azure Blob object", raw: "https://account.blob.core.windows.net/private/video.mp4"},
		{name: "Azure DFS object", raw: "https://account.dfs.core.windows.net/private/video.mp4"},
		{name: "Azure China Blob object", raw: "https://account.blob.core.chinacloudapi.cn/private/video.mp4"},
		{name: "Azure China DFS object", raw: "https://account.dfs.core.chinacloudapi.cn/private/video.mp4"},
		{name: "Azure US Gov Blob object", raw: "https://account.blob.core.usgovcloudapi.net/private/video.mp4"},
		{name: "Azure US Gov DFS object", raw: "https://account.dfs.core.usgovcloudapi.net/private/video.mp4"},
		{name: "Azure Germany Blob object", raw: "https://account.blob.core.cloudapi.de/private/video.mp4"},
		{name: "Azure Germany DFS object", raw: "https://account.dfs.core.cloudapi.de/private/video.mp4"},
		{name: "Tencent COS object", raw: "https://bucket.cos.ap-shanghai.myqcloud.com/private/video.mp4"},
		{name: "Tencent provider root", raw: "https://media.myqcloud.com/private/video.mp4"},
		{name: "Huawei OBS object", raw: "https://bucket.obs.cn-north-4.myhuaweicloud.com/private/video.mp4"},
		{name: "Huawei provider root", raw: "https://media.myhuaweicloud.com/private/video.mp4"},
		{name: "DigitalOcean Spaces object", raw: "https://bucket.nyc3.digitaloceanspaces.com/private/video.mp4"},
		{name: "Cloudflare R2 object", raw: "https://account.r2.cloudflarestorage.com/private/video.mp4"},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := safePublicContentURL(test.raw); got != test.want {
				t.Fatalf("safePublicContentURL(%q) = %v, want %v", test.raw, got, test.want)
			}
		})
	}
}
