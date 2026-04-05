package s3client

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"

	"s3tool/internal/models"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Create 创建 S3 客户端
func Create(req *models.ConnectRequest) (*MinioClient, error) {
	endpoint, ssl := ParseEndpoint(req.Endpoint, req.SSL)

	// 设计决策：禁用 SSL 证书验证
	// 这是为了支持内部自签名证书的存储服务（如 MinIO、Ceph RGW 等私有部署）。
	// 在企业内网环境中，使用自签名证书是常见做法，强制验证会导致连接失败。
	// 注意：在生产环境暴露公网时，建议配置正确的 CA 证书并启用验证。
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(req.AK, req.SK, ""),
		Secure:    ssl,
		Transport: transport,
	})
	if err != nil {
		return nil, err
	}
	return NewMinioClient(client), nil
}

// ParseEndpoint 解析 endpoint，处理用户输入的各种格式
// 支持: "s3.amazonaws.com", "https://s3.amazonaws.com", "http://s3.amazonaws.com:9000"
func ParseEndpoint(endpoint string, defaultSSL bool) (string, bool) {
	endpoint = strings.TrimSpace(endpoint)

	// 尝试解析为 URL
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		u, err := url.Parse(endpoint)
		if err == nil {
			// 从 URL 中提取主机名（含端口）
			host := u.Host
			// 根据协议确定 SSL
			ssl := u.Scheme == "https"
			return host, ssl
		}
	}

	// 去掉可能的路径部分（如 "s3.amazonaws.com/bucket"）
	if idx := strings.Index(endpoint, "/"); idx > 0 {
		endpoint = endpoint[:idx]
	}

	return endpoint, defaultSSL
}
