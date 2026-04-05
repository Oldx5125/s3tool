package s3client

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/cors"
	"s3tool/internal/models"
)

// MockClient mock S3 客户端
type MockClient struct {
	Buckets         []minio.BucketInfo
	Objects         map[string][]minio.ObjectInfo // bucket -> objects
	ObjectData      map[string]map[string][]byte  // bucket -> key -> data
	CorsConfig      map[string]*cors.Config       // bucket -> cors config
	BucketExistsMap map[string]bool
	CopyError       error
	DeleteError     error
	PutError        error
	GetError        error
	StatError       error
	PresignError    error
}

// NewMockClient 创建 mock 客户端
func NewMockClient() *MockClient {
	return &MockClient{
		Buckets: []minio.BucketInfo{
			{Name: "test-bucket", CreationDate: time.Now()},
		},
		Objects:    make(map[string][]minio.ObjectInfo),
		ObjectData: make(map[string]map[string][]byte),
		CorsConfig: make(map[string]*cors.Config),
		BucketExistsMap: map[string]bool{
			"test-bucket": true,
		},
	}
}

// Ensure MockClient implements S3Client
var _ S3Client = (*MockClient)(nil)

// ListBuckets mock
func (m *MockClient) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	return m.Buckets, nil
}

// BucketExists mock
func (m *MockClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return m.BucketExistsMap[bucketName], nil
}

// ListObjects mock
func (m *MockClient) ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	ch := make(chan minio.ObjectInfo, 100)
	go func() {
		defer close(ch)
		objects, ok := m.Objects[bucketName]
		if !ok {
			return
		}
		for _, obj := range objects {
			if opts.Prefix == "" || strings.HasPrefix(obj.Key, opts.Prefix) {
				ch <- obj
			}
		}
	}()
	return ch
}

// PutObject mock
func (m *MockClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	if m.PutError != nil {
		return minio.UploadInfo{}, m.PutError
	}
	if m.ObjectData[bucketName] == nil {
		m.ObjectData[bucketName] = make(map[string][]byte)
	}
	data, _ := io.ReadAll(reader)
	m.ObjectData[bucketName][objectName] = data
	return minio.UploadInfo{Key: objectName, Size: objectSize}, nil
}

// Ensure MockClient implements FolderClient
var _ interface {
	ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
} = (*MockClient)(nil)

// GetObject mock
func (m *MockClient) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	if m.ObjectData[bucketName] == nil {
		return nil, fmt.Errorf("NoSuchBucket: The specified bucket does not exist")
	}
	if _, ok := m.ObjectData[bucketName][objectName]; !ok {
		return nil, fmt.Errorf("NoSuchKey: The specified key does not exist")
	}
	return nil, nil
}

// StatObject mock
func (m *MockClient) StatObject(ctx context.Context, bucketName, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	if m.StatError != nil {
		return minio.ObjectInfo{}, m.StatError
	}

	// 根据扩展名猜测 ContentType
	contentType := "application/octet-stream"
	lowerKey := strings.ToLower(objectName)
	switch {
	case strings.HasSuffix(lowerKey, ".mp4"):
		contentType = "video/mp4"
	case strings.HasSuffix(lowerKey, ".webm"):
		contentType = "video/webm"
	case strings.HasSuffix(lowerKey, ".mp3"):
		contentType = "audio/mpeg"
	case strings.HasSuffix(lowerKey, ".wav"):
		contentType = "audio/wav"
	case strings.HasSuffix(lowerKey, ".pdf"):
		contentType = "application/pdf"
	case strings.HasSuffix(lowerKey, ".txt"):
		contentType = "text/plain"
	case strings.HasSuffix(lowerKey, ".json"):
		contentType = "application/json"
	case strings.HasSuffix(lowerKey, ".jpg") || strings.HasSuffix(lowerKey, ".jpeg"):
		contentType = "image/jpeg"
	case strings.HasSuffix(lowerKey, ".png"):
		contentType = "image/png"
	}

	return minio.ObjectInfo{
		Key:          objectName,
		Size:         100,
		LastModified: time.Now(),
		ContentType:  contentType,
	}, nil
}

// RemoveObject mock
func (m *MockClient) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	return m.DeleteError
}

// CopyObject mock
func (m *MockClient) CopyObject(ctx context.Context, dst minio.CopyDestOptions, src minio.CopySrcOptions) (minio.UploadInfo, error) {
	return minio.UploadInfo{Key: dst.Object}, m.CopyError
}

// GetBucketCors mock
func (m *MockClient) GetBucketCors(ctx context.Context, bucketName string) (*cors.Config, error) {
	return m.CorsConfig[bucketName], nil
}

// SetBucketCors mock
func (m *MockClient) SetBucketCors(ctx context.Context, bucketName string, config *cors.Config) error {
	m.CorsConfig[bucketName] = config
	return nil
}

// PresignedGetObject mock
func (m *MockClient) PresignedGetObject(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values) (*url.URL, error) {
	if m.PresignError != nil {
		return nil, m.PresignError
	}
	return &url.URL{Scheme: "https", Host: "example.com", Path: "/" + bucketName + "/" + objectName}, nil
}

// MockFactory 创建 mock 客户端的工厂
func MockFactory(mock *MockClient) ClientFactory {
	return func(req *models.ConnectRequest) (S3Client, error) {
		return mock, nil
	}
}

// MockFactoryWithError 创建返回错误的工厂
func MockFactoryWithError(err error) ClientFactory {
	return func(req *models.ConnectRequest) (S3Client, error) {
		return nil, err
	}
}
