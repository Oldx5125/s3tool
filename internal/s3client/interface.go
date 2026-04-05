package s3client

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/cors"
	"s3tool/internal/models"
)

// S3Client 定义 S3 客户端接口
type S3Client interface {
	ListBuckets(ctx context.Context) ([]minio.BucketInfo, error)
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error)
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
	CopyObject(ctx context.Context, dst minio.CopyDestOptions, src minio.CopySrcOptions) (minio.UploadInfo, error)
	GetBucketCors(ctx context.Context, bucketName string) (*cors.Config, error)
	SetBucketCors(ctx context.Context, bucketName string, config *cors.Config) error
	PresignedGetObject(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values) (*url.URL, error)
}

// ClientFactory 创建 S3 客户端的工厂函数类型
type ClientFactory func(req *models.ConnectRequest) (S3Client, error)

// DefaultFactory 默认的客户端工厂
var DefaultFactory ClientFactory = func(req *models.ConnectRequest) (S3Client, error) {
	return Create(req)
}
