package s3client

import (
	"testing"

	"s3tool/internal/models"
)

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		input      string
		defaultSSL bool
		wantHost   string
		wantSSL    bool
	}{
		// 纯主机名
		{"s3.amazonaws.com", false, "s3.amazonaws.com", false},
		{"s3.amazonaws.com", true, "s3.amazonaws.com", true},
		// 带端口
		{"10.50.2.45:5080", false, "10.50.2.45:5080", false},
		// HTTPS 前缀
		{"https://s3.amazonaws.com", false, "s3.amazonaws.com", true},
		{"https://s3backup.hengrui.com", false, "s3backup.hengrui.com", true},
		// HTTP 前缀
		{"http://10.50.2.45:5080", true, "10.50.2.45:5080", false},
		// 带路径（应该被去掉）
		{"s3.amazonaws.com/bucket", true, "s3.amazonaws.com", true},
		// HTTPS 带端口
		{"https://s3.amazonaws.com:9443", false, "s3.amazonaws.com:9443", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			host, ssl := ParseEndpoint(tt.input, tt.defaultSSL)
			if host != tt.wantHost {
				t.Errorf("host: got %q, want %q", host, tt.wantHost)
			}
			if ssl != tt.wantSSL {
				t.Errorf("ssl: got %v, want %v", ssl, tt.wantSSL)
			}
		})
	}
}

// parseEndpoint 额外边界测试
func TestParseEndpoint_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantHost string
		wantSSL  bool
	}{
		{"empty string", "", "", true},
		{"whitespace only", "   ", "", true},
		{"trailing slash", "s3.amazonaws.com/", "s3.amazonaws.com", true},
		{"multiple slashes", "s3.amazonaws.com/bucket/path", "s3.amazonaws.com", true},
		{"http with port", "http://localhost:9000", "localhost:9000", false},
		{"https with port", "https://s3.region.amazonaws.com:443", "s3.region.amazonaws.com:443", true},
		{"ip address", "192.168.1.1", "192.168.1.1", false},
		{"ip with port", "192.168.1.1:9000", "192.168.1.1:9000", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, ssl := ParseEndpoint(tt.input, tt.wantSSL)
			if host != tt.wantHost {
				t.Errorf("host: got %q, want %q", host, tt.wantHost)
			}
			if ssl != tt.wantSSL {
				t.Errorf("ssl: got %v, want %v", ssl, tt.wantSSL)
			}
		})
	}
}

// parseEndpoint 更多边界测试
func TestParseEndpoint_AdditionalCases(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		defaultSSL bool
		wantHost   string
		wantSSL    bool
	}{
		{"url with path", "https://s3.amazonaws.com/bucket/path", true, "s3.amazonaws.com", true},
		{"url with query", "https://s3.amazonaws.com?query=1", true, "s3.amazonaws.com", true},
		{"url with port and path", "http://localhost:9000/bucket", false, "localhost:9000", false},
		{"domain with multiple dots", "s3.us-east-1.amazonaws.com", true, "s3.us-east-1.amazonaws.com", true},
		{"ipv4 address", "192.168.1.1", false, "192.168.1.1", false},
		{"ipv4 with port", "192.168.1.1:9000", false, "192.168.1.1:9000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, ssl := ParseEndpoint(tt.input, tt.defaultSSL)
			if host != tt.wantHost {
				t.Errorf("host: got %q, want %q", host, tt.wantHost)
			}
			if ssl != tt.wantSSL {
				t.Errorf("ssl: got %v, want %v", ssl, tt.wantSSL)
			}
		})
	}
}

// parseEndpoint Unicode 测试
func TestParseEndpoint_Unicode(t *testing.T) {
	host, ssl := ParseEndpoint("本地服务器", true)
	if host != "本地服务器" {
		t.Errorf("expected host '本地服务器', got %q", host)
	}
	if !ssl {
		t.Error("expected ssl to be true (default)")
	}
}

// createS3Client 测试
func TestCreateS3Client(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		ssl      bool
	}{
		{"https endpoint", "https://s3.amazonaws.com", true},
		{"http endpoint", "http://localhost:9000", false},
		{"plain host with ssl", "s3.amazonaws.com", true},
		{"plain host without ssl", "s3.amazonaws.com", false},
		{"with port", "localhost:9000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &models.ConnectRequest{
				Endpoint: tt.endpoint,
				AK:       "test-access-key",
				SK:       "test-secret-key",
				SSL:      tt.ssl,
			}
			client, err := Create(req)
			if err != nil {
				t.Errorf("Create failed: %v", err)
				return
			}
			if client == nil {
				t.Error("expected client to be non-nil")
			}
		})
	}
}
