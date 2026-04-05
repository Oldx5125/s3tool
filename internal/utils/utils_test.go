package utils

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/minio/minio-go/v7"
)

// mockFolderClient 实现 FolderClient 接口用于测试
type mockFolderClient struct {
	objects     []minio.ObjectInfo
	putError    error
	listError   error
	putCalled   bool
	putObject   string
	listCalled  bool
	putCallCount int
}

func (m *mockFolderClient) ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	m.listCalled = true
	ch := make(chan minio.ObjectInfo, len(m.objects)+1)
	go func() {
		defer close(ch)
		if m.listError != nil {
			ch <- minio.ObjectInfo{Err: m.listError}
			return
		}
		for _, obj := range m.objects {
			ch <- obj
		}
	}()
	return ch
}

func (m *mockFolderClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	m.putCalled = true
	m.putObject = objectName
	m.putCallCount++
	return minio.UploadInfo{Key: objectName}, m.putError
}

// validatePath 单元测试
func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		// 有效路径
		{"empty path", "", false},
		{"simple filename", "file.txt", false},
		{"valid path", "folder/file.txt", false},
		{"deep nested", "a/b/c/d/file.txt", false},
		{"with chinese", "文件夹/文件.txt", false},

		// 无效路径
		{"path traversal", "../etc/passwd", true},
		{"double traversal", "foo/../bar", true},
		{"traversal at end", "folder/..", true},
		{"absolute path", "/etc/passwd", true},
		{"absolute nested", "/folder/file.txt", true},
		{"null byte", "file\x00.txt", true},
		{"null byte middle", "file\x00name.txt", true},
		{"multiple null bytes", "a\x00b\x00c", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) = %v, want error: %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

// ValidatePath 边界测试增强
func TestValidatePath_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"single dot", ".", false},
		{"trailing slash", "folder/", false},
		{"leading dot", ".hidden", false},
		{"multiple dots in name", "file...txt", true}, // 包含 ".."
		{"unicode chars", "文件/文档.txt", false},
		{"spaces", "folder name/file name.txt", false},
		{"traversal with encoding", "%2e%2e/etc/passwd", false}, // 不解码，所以有效
		{"empty after traversal check", "a/..", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) = %v, want error: %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

// ensureParentFolders 测试
func TestEnsureParentFolders_NoSlash(t *testing.T) {
	// 没有斜杠的路径，应该直接返回
	EnsureParentFolders(nil, nil, "bucket", "filename.txt")
	// 无错误即通过
}

func TestEnsureParentFolders_EmptyKey(t *testing.T) {
	// 空键，应该直接返回
	EnsureParentFolders(nil, nil, "bucket", "")
	// 无错误即通过
}

// 测试各种输入下不会 panic
func TestValidatePath_NoPanic(t *testing.T) {
	paths := []string{
		"",
		"simple",
		"path/to/file",
		"../../etc/passwd",
		"/absolute/path",
		"file\x00name",
		"normal/file.txt",
	}

	for _, p := range paths {
		_ = ValidatePath(p)
	}
}

// EnsureParentFolders 测试
func TestEnsureParentFolders_NeedToCreate(t *testing.T) {
	// 当父目录只有文件夹标识时，需要创建文件夹标识
	mock := &mockFolderClient{
		objects: []minio.ObjectInfo{
			{Key: "folder/sub/"}, // 只有文件夹标识
		},
	}

	EnsureParentFolders(context.Background(), mock, "bucket", "folder/sub/file.txt")

	if !mock.listCalled {
		t.Error("ListObjects should be called")
	}
	// 应该创建文件夹标识（因为只有文件夹标识，没有其他文件）
	if !mock.putCalled {
		t.Error("PutObject should be called to create folder marker")
	}
	if mock.putObject != "folder/sub/" {
		t.Errorf("expected putObject 'folder/sub/', got %q", mock.putObject)
	}
}

func TestEnsureParentFolders_EmptyParent(t *testing.T) {
	// 当父目录完全为空时（没有任何对象），需要创建文件夹标识
	mock := &mockFolderClient{
		objects: []minio.ObjectInfo{}, // 空
	}

	EnsureParentFolders(context.Background(), mock, "bucket", "folder/sub/file.txt")

	if !mock.listCalled {
		t.Error("ListObjects should be called")
	}
	// 应该创建文件夹标识
	if !mock.putCalled {
		t.Error("PutObject should be called to create folder marker")
	}
}

func TestEnsureParentFolders_DirectoryNotEmpty(t *testing.T) {
	// 当父目录有其他文件时，不需要创建文件夹标识
	mock := &mockFolderClient{
		objects: []minio.ObjectInfo{
			{Key: "folder/sub/"},
			{Key: "folder/sub/other.txt"}, // 有其他文件
		},
	}

	EnsureParentFolders(context.Background(), mock, "bucket", "folder/sub/file.txt")

	if !mock.listCalled {
		t.Error("ListObjects should be called")
	}
	// 不应该创建文件夹标识
	if mock.putCalled {
		t.Error("PutObject should NOT be called when directory not empty")
	}
}

func TestEnsureParentFolders_ListError(t *testing.T) {
	// 当 ListObjects 返回错误时，应该直接返回
	mock := &mockFolderClient{
		listError: errors.New("list error"),
	}

	EnsureParentFolders(context.Background(), mock, "bucket", "folder/sub/file.txt")

	if !mock.listCalled {
		t.Error("ListObjects should be called")
	}
	// 遇到错误不应创建
	if mock.putCalled {
		t.Error("PutObject should NOT be called on list error")
	}
}

func TestEnsureParentFolders_PutError(t *testing.T) {
	// 当 PutObject 失败时，应该记录警告但不 panic
	mock := &mockFolderClient{
		objects: []minio.ObjectInfo{
			{Key: "folder/sub/"}, // 只有文件夹标识，目录为空
		},
		putError: errors.New("put error"),
	}

	// 不应该 panic
	EnsureParentFolders(context.Background(), mock, "bucket", "folder/sub/file.txt")

	if !mock.putCalled {
		t.Error("PutObject should be called")
	}
}

func TestEnsureParentFolders_IdxZero(t *testing.T) {
	// 当 idx <= 0 时（如 "/file.txt"），应该直接返回
	mock := &mockFolderClient{}

	EnsureParentFolders(context.Background(), mock, "bucket", "/file.txt")

	if mock.listCalled {
		t.Error("ListObjects should NOT be called when idx <= 0")
	}
}

func TestEnsureParentFolders_DeepPath(t *testing.T) {
	// 测试深层路径
	mock := &mockFolderClient{
		objects: []minio.ObjectInfo{
			{Key: "a/b/c/d/"},
		},
	}

	EnsureParentFolders(context.Background(), mock, "bucket", "a/b/c/d/file.txt")

	if !mock.listCalled {
		t.Error("ListObjects should be called")
	}
	// 应该检查直接父目录 "a/b/c/d/"
	if mock.putObject != "a/b/c/d/" {
		t.Errorf("expected putObject 'a/b/c/d/', got %q", mock.putObject)
	}
}
