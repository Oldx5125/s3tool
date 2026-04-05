package api

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"s3tool/internal/i18n"
	"s3tool/internal/models"
	"s3tool/internal/s3client"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

// ErrPreviewSizeExceeded 预览文件大小超过限制错误
var ErrPreviewSizeExceeded = errors.New("文件大小超过预览限制")

// FrontendFS 前端文件系统（由 main.go 设置）
var FrontendFS embed.FS

// ClientFactory S3 客户端工厂（可被测试替换）
var ClientFactory = s3client.DefaultFactory

// 文件类型扩展名映射（预览功能使用）
var (
	imageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true, ".bmp": true, ".ico": true}
	textExts  = map[string]bool{".txt": true, ".md": true, ".json": true, ".xml": true, ".yaml": true, ".yml": true, ".csv": true, ".log": true, ".ini": true, ".conf": true, ".cfg": true}
	codeExts  = map[string]bool{".js": true, ".ts": true, ".go": true, ".py": true, ".java": true, ".c": true, ".cpp": true, ".h": true, ".hpp": true, ".cs": true, ".rb": true, ".php": true, ".swift": true, ".kt": true, ".rs": true, ".sh": true, ".bat": true, ".sql": true, ".html": true, ".css": true, ".scss": true, ".less": true, ".vue": true, ".jsx": true, ".tsx": true}
	videoExts = map[string]bool{".mp4": true, ".webm": true, ".ogg": true, ".mov": true, ".avi": true, ".mkv": true}
	audioExts = map[string]bool{".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".m4a": true}
)

const maxInlinePreviewSize = 5 * 1024 * 1024

// notFoundPatterns 文件不存在错误的匹配模式
var notFoundPatterns = []string{
	"NotFound",
	"no such file",
	"不存在",
	"The specified key does not exist",
}

// isNotFoundError 判断错误是否为"文件不存在"类型
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	for _, p := range notFoundPatterns {
		if strings.Contains(errStr, p) {
			return true
		}
	}
	return false
}

// readObjectContent 读取对象内容，限制最大读取大小
func readObjectContent(ctx context.Context, client s3client.S3Client, bucket, key string) ([]byte, error) {
	object, err := client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	// 使用 LimitReader 限制读取大小，额外预留 1KB 用于检测是否超出限制
	limitedReader := io.LimitReader(object, maxInlinePreviewSize+1024)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > maxInlinePreviewSize {
		return nil, fmt.Errorf("%w (%d MB)", ErrPreviewSizeExceeded, maxInlinePreviewSize/1024/1024)
	}

	return data, nil
}

// copyObject 复制对象
func copyObject(ctx context.Context, client s3client.S3Client, bucket, srcKey, dstKey string) error {
	src := minio.CopySrcOptions{Bucket: bucket, Object: srcKey}
	dst := minio.CopyDestOptions{Bucket: bucket, Object: dstKey}
	_, err := client.CopyObject(ctx, dst, src)
	return err
}

func extractConnectRequest[T any](req *T) *models.ConnectRequest {
	v := reflect.ValueOf(req).Elem()
	if v.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Type() == reflect.TypeOf(models.ConnectRequest{}) {
			if cp, ok := field.Addr().Interface().(*models.ConnectRequest); ok {
				return cp
			}
		}
	}
	return nil
}

func withClient[T any](c *gin.Context, timeout time.Duration) (*T, s3client.S3Client, context.Context, context.CancelFunc, bool) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("解析请求失败", "error", err)
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_request_format")})
		return nil, nil, nil, nil, false
	}

	connReq := extractConnectRequest(&req)
	if connReq == nil {
		slog.Error("提取连接信息失败")
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_request_format")})
		return nil, nil, nil, nil, false
	}

	if connReq.Endpoint == "" || connReq.AK == "" || connReq.SK == "" {
		slog.Error("连接信息不完整")
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_missing_fields")})
		return nil, nil, nil, nil, false
	}

	client, err := ClientFactory(connReq)
	if err != nil {
		slog.Error("创建客户端失败", "error", err)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_connection_failed"), err.Error())})
		return nil, nil, nil, nil, false
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	return &req, client, ctx, cancel, true
}

func contentDisposition(key string) string {
	filename := key
	if idx := strings.LastIndex(key, "/"); idx >= 0 {
		filename = key[idx+1:]
	}
	return fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(filename))
}
