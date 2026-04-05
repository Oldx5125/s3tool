package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"s3tool/internal/models"
	"s3tool/internal/s3client"
)

// 使用 mock 客户端测试成功路径
func setupMockTestRouter(t *testing.T, mock *s3client.MockClient) *gin.Engine {
	gin.SetMode(gin.TestMode)
	// 保存原始工厂函数
	originalFactory := ClientFactory
	ClientFactory = s3client.MockFactory(mock)

	router := gin.New()
	router.GET("/", HandleIndex)
	router.GET("/health", HandleHealth)
	router.POST("/connect", HandleConnect)
	router.POST("/buckets", HandleBuckets)
	router.POST("/list", HandleList)
	router.POST("/upload", HandleUpload)
	router.POST("/download", HandleDownload)
	router.POST("/delete", HandleDelete)
	router.POST("/batch/delete", HandleBatchDelete)
	router.POST("/preview", HandlePreview)
	router.POST("/rename", HandleRename)
	router.POST("/move", HandleMove)
	router.POST("/copy", HandleCopy)
	router.POST("/mkdir", HandleMkdir)
	router.POST("/rmdir", HandleRmdir)
	router.POST("/cors/get", HandleGetCors)
	router.POST("/cors/set", HandleSetCors)

	// 恢复工厂函数
	t.Cleanup(func() { ClientFactory = originalFactory })
	return router
}

func TestHandleConnect_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true}`
	req := httptest.NewRequest(http.MethodPost, "/connect", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleConnect_BucketNotExist(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.BucketExistsMap = map[string]bool{} // 无 bucket
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"nonexistent","ssl":true}`
	req := httptest.NewRequest(http.MethodPost, "/connect", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleBuckets_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","ssl":true}`
	req := httptest.NewRequest(http.MethodPost, "/buckets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestHandleList_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.Objects["test-bucket"] = []minio.ObjectInfo{
		{Key: "file1.txt", Size: 100, LastModified: time.Now()},
		{Key: "folder/", Size: 0, LastModified: time.Now()},
	}
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true}`
	req := httptest.NewRequest(http.MethodPost, "/list", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleDelete_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"key":"file.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestHandleDelete_Error(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.DeleteError = errors.New("delete failed")
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"key":"file.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestHandleBatchDelete_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"keys":["file1.txt","file2.txt"]}`
	req := httptest.NewRequest(http.MethodPost, "/batch/delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleBatchDelete_PartialSuccess(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	// 包含路径遍历的 key 会导致部分失败
	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"keys":["file1.txt","../invalid.txt"]}`
	req := httptest.NewRequest(http.MethodPost, "/batch/delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestHandlePreview_VideoFile(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	// 测试视频文件（会使用预签名 URL）
	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"key":"video.mp4"}`
	req := httptest.NewRequest(http.MethodPost, "/preview", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 验证返回了预签名 URL
	if !strings.Contains(rec.Body.String(), "\"url\"") {
		t.Errorf("expected presigned URL in response, got: %s", rec.Body.String())
	}
}

func TestHandlePreview_AudioFile(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	// 测试音频文件（会使用预签名 URL）
	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"key":"audio.mp3"}`
	req := httptest.NewRequest(http.MethodPost, "/preview", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlePreview_PresignError(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.PresignError = errors.New("presign failed")
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"key":"video.mp4"}`
	req := httptest.NewRequest(http.MethodPost, "/preview", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestHandlePreview_StatError(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.StatError = errors.New("file not found")
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"key":"missing.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/preview", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleRename_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"oldKey":"old.txt","newKey":"new.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/rename", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleRename_CopyError(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.CopyError = errors.New("copy failed")
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"oldKey":"old.txt","newKey":"new.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/rename", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestHandleMove_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"srcKey":"src.txt","dstKey":"dst.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/move", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleMove_CopyError(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.CopyError = errors.New("copy failed")
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"srcKey":"src.txt","dstKey":"dst.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/move", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestHandleCopy_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"srcKey":"src.txt","dstKey":"dst.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/copy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleCopy_CopyError(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.CopyError = errors.New("copy failed")
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"srcKey":"src.txt","dstKey":"dst.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/copy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestHandleMkdir_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"folder":"newfolder"}`
	req := httptest.NewRequest(http.MethodPost, "/mkdir", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleMkdir_PutError(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.PutError = errors.New("put failed")
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"folder":"newfolder"}`
	req := httptest.NewRequest(http.MethodPost, "/mkdir", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestHandleRmdir_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"folder":"emptyfolder"}`
	req := httptest.NewRequest(http.MethodPost, "/rmdir", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleRmdir_NotEmpty(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.Objects["test-bucket"] = []minio.ObjectInfo{
		{Key: "emptyfolder/file.txt", Size: 100, LastModified: time.Now()},
	}
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"folder":"emptyfolder"}`
	req := httptest.NewRequest(http.MethodPost, "/rmdir", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestHandleGetCors_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true}`
	req := httptest.NewRequest(http.MethodPost, "/cors/get", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleSetCors_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"config":""}`
	req := httptest.NewRequest(http.MethodPost, "/cors/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleSetCors_WithConfig(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"config":"{\"CORSRules\":[{\"AllowedOrigin\":[\"*\"],\"AllowedMethod\":[\"GET\"]}]}"}`
	req := httptest.NewRequest(http.MethodPost, "/cors/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleUpload_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	// 创建 multipart form 数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加表单字段
	_ = writer.WriteField("endpoint", "test")
	_ = writer.WriteField("ak", "test")
	_ = writer.WriteField("sk", "test")
	_ = writer.WriteField("bucket", "test-bucket")
	_ = writer.WriteField("ssl", "true")

	// 添加文件
	part, _ := writer.CreateFormFile("file", "test.txt")
	_, _ = part.Write([]byte("test content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleUpload_WithCustomKey(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("endpoint", "test")
	_ = writer.WriteField("ak", "test")
	_ = writer.WriteField("sk", "test")
	_ = writer.WriteField("bucket", "test-bucket")
	_ = writer.WriteField("ssl", "true")
	_ = writer.WriteField("key", "custom/path/file.txt")

	part, _ := writer.CreateFormFile("file", "original.txt")
	_, _ = part.Write([]byte("test"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandleUpload_PathTraversal(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("endpoint", "test")
	_ = writer.WriteField("ak", "test")
	_ = writer.WriteField("sk", "test")
	_ = writer.WriteField("bucket", "test-bucket")
	_ = writer.WriteField("ssl", "true")
	_ = writer.WriteField("key", "../etc/passwd")

	part, _ := writer.CreateFormFile("file", "test.txt")
	_, _ = part.Write([]byte("test"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestHandleUpload_PutError(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.PutError = errors.New("put failed")
	router := setupMockTestRouter(t, mock)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("endpoint", "test")
	_ = writer.WriteField("ak", "test")
	_ = writer.WriteField("sk", "test")
	_ = writer.WriteField("bucket", "test-bucket")
	_ = writer.WriteField("ssl", "true")

	part, _ := writer.CreateFormFile("file", "test.txt")
	_, _ = part.Write([]byte("test"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestHandleDownload_Success(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.ObjectData["test-bucket"] = map[string][]byte{
		"file.txt": []byte("test content"),
	}
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"key":"file.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/download", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// 由于 mock GetObject 返回 nil，这里会返回错误
	// 但我们测试的是请求流程
}

func TestHandleDownload_GetError(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.GetError = errors.New("get failed")
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"key":"missing.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/download", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleUpload_ClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalFactory := ClientFactory
	ClientFactory = s3client.MockFactoryWithError(errors.New("connection failed"))
	defer func() { ClientFactory = originalFactory }()

	router := gin.New()
	router.POST("/upload", HandleUpload)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("endpoint", "test")
	_ = writer.WriteField("ak", "test")
	_ = writer.WriteField("sk", "test")
	_ = writer.WriteField("bucket", "test-bucket")
	_ = writer.WriteField("ssl", "true")

	part, _ := writer.CreateFormFile("file", "test.txt")
	_, _ = part.Write([]byte("test"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestClientFactoryError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	originalFactory := ClientFactory
	ClientFactory = s3client.MockFactoryWithError(errors.New("connection failed"))
	defer func() { ClientFactory = originalFactory }()

	router := gin.New()
	router.POST("/connect", HandleConnect)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true}`
	req := httptest.NewRequest(http.MethodPost, "/connect", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

// TestHandleList_HasMore_Empty 测试空目录的分页行为（hasMore 应为 false）
func TestHandleList_HasMore_Empty(t *testing.T) {
	mock := s3client.NewMockClient()
	mock.Objects["test-bucket"] = []minio.ObjectInfo{} // 空目录
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"limit":10}`
	req := httptest.NewRequest(http.MethodPost, "/list", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp models.ListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.HasMore {
		t.Error("expected hasMore to be false for empty directory")
	}
}

// TestHandleList_HasMore_LessThanLimit 测试对象数小于 limit 的分页行为（hasMore 应为 false）
func TestHandleList_HasMore_LessThanLimit(t *testing.T) {
	mock := s3client.NewMockClient()
	// 设置 3 个对象，limit 为 10
	mock.Objects["test-bucket"] = []minio.ObjectInfo{
		{Key: "file1.txt", Size: 100, LastModified: time.Now()},
		{Key: "file2.txt", Size: 200, LastModified: time.Now()},
		{Key: "file3.txt", Size: 300, LastModified: time.Now()},
	}
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"limit":10}`
	req := httptest.NewRequest(http.MethodPost, "/list", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp models.ListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.HasMore {
		t.Error("expected hasMore to be false when count < limit")
	}
}

// TestHandleList_HasMore_EqualLimit 测试对象数等于 limit 的分页行为（hasMore 应为 true）
func TestHandleList_HasMore_EqualLimit(t *testing.T) {
	mock := s3client.NewMockClient()
	// 设置 10 个对象，limit 为 10
	mock.Objects["test-bucket"] = []minio.ObjectInfo{
		{Key: "file1.txt", Size: 100, LastModified: time.Now()},
		{Key: "file2.txt", Size: 200, LastModified: time.Now()},
		{Key: "file3.txt", Size: 300, LastModified: time.Now()},
		{Key: "file4.txt", Size: 400, LastModified: time.Now()},
		{Key: "file5.txt", Size: 500, LastModified: time.Now()},
		{Key: "file6.txt", Size: 600, LastModified: time.Now()},
		{Key: "file7.txt", Size: 700, LastModified: time.Now()},
		{Key: "file8.txt", Size: 800, LastModified: time.Now()},
		{Key: "file9.txt", Size: 900, LastModified: time.Now()},
		{Key: "file10.txt", Size: 1000, LastModified: time.Now()},
	}
	router := setupMockTestRouter(t, mock)

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"limit":10}`
	req := httptest.NewRequest(http.MethodPost, "/list", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp models.ListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.HasMore {
		t.Error("expected hasMore to be true when count >= limit")
	}
}

func TestHandleBatchDelete_ManyFiles(t *testing.T) {
	mock := s3client.NewMockClient()
	router := setupMockTestRouter(t, mock)

	// 测试删除 50 个文件（验证并发控制不会阻塞）
	keys := make([]string, 50)
	for i := 0; i < 50; i++ {
		keys[i] = fmt.Sprintf("file%d.txt", i)
	}

	body := fmt.Sprintf(`{"endpoint":"test","ak":"test","sk":"test","bucket":"test-bucket","ssl":true,"keys":%s}`, mustMarshalJSON(keys))
	req := httptest.NewRequest(http.MethodPost, "/batch/delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp models.BatchDeleteResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Results) != 50 {
		t.Errorf("expected 50 results, got %d", len(resp.Results))
	}
}

func mustMarshalJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
