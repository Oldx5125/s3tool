package s3client

import (
	"context"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

// MinioClient 包装 minio.Client 实现 S3Client 接口
type MinioClient struct {
	*minio.Client
}

// NewMinioClient 创建 MinioClient 包装器
func NewMinioClient(client *minio.Client) *MinioClient {
	return &MinioClient{Client: client}
}

// Ensure MinioClient implements S3Client
var _ S3Client = (*MinioClient)(nil)

// GetObject 包装 minio.GetObject
func (m *MinioClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return m.Client.GetObject(ctx, bucketName, objectName, opts)
}

// PresignedGetObject 包装 minio.PresignedGetObject
func (m *MinioClient) PresignedGetObject(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values) (*url.URL, error) {
	return m.Client.PresignedGetObject(ctx, bucketName, objectName, expires, reqParams)
}
