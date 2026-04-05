package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/minio/minio-go/v7"
)

// ValidatePath 验证路径安全性，防止路径遍历攻击
func ValidatePath(path string) error {
	if path == "" {
		return nil
	}
	// 禁止路径遍历
	if strings.Contains(path, "..") {
		return fmt.Errorf("路径不能包含 '..'")
	}
	// 禁止绝对路径（S3 对象键不应该以 / 开头）
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("路径不能以 '/' 开头")
	}
	// 禁止空字节注入
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("路径不能包含空字节")
	}
	return nil
}

// MaskSecret 对敏感信息进行脱敏
// 对于长度 <= 4 的字符串，直接返回 "****"
// 对于长度 > 4 的字符串，返回 前2位 + "****" + 后2位
func MaskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	if len(s) <= 8 {
		return s[:2] + "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}

// FolderClient EnsureParentFolders 所需的接口
type FolderClient interface {
	ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
}

// EnsureParentFolders 确保父文件夹标识存在
// 只检查直接父目录，如果为空则创建文件夹标识
func EnsureParentFolders(ctx context.Context, client FolderClient, bucket, key string) {
	if !strings.Contains(key, "/") {
		return
	}

	// 获取直接父目录
	idx := strings.LastIndex(key, "/")
	if idx <= 0 {
		return
	}

	parentFolder := key[:idx+1]

	// 检查父目录下是否还有其他文件
	for obj := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    parentFolder,
		Recursive: false,
		MaxKeys:   2,
	}) {
		if obj.Err != nil {
			slog.Warn("检查父目录失败", "error", obj.Err, "folder", parentFolder)
			return
		}
		// 找到非文件夹标识的对象，说明目录不为空，无需创建
		if obj.Key != parentFolder {
			return
		}
	}

	// 父目录为空，创建文件夹标识
	_, err := client.PutObject(ctx, bucket, parentFolder, strings.NewReader(""), 0, minio.PutObjectOptions{
		ContentType: "application/x-directory",
	})
	if err != nil {
		slog.Warn("重新创建文件夹标识失败", "error", err, "folder", parentFolder)
	} else {
		slog.Info("重新创建文件夹标识", "folder", parentFolder)
	}
}
