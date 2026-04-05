package api

import (
	"embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"s3tool/internal/models"
)

//go:embed frontend.html
var testFrontendFS embed.FS

func init() {
	// 设置测试用的前端文件系统
	FrontendFS = testFrontendFS
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
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
	return router
}

// 健康检查测试
func TestHandleHealth(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", resp["status"])
	}

	if resp["version"] == nil {
		t.Error("expected version to be present")
	}
}

func TestHandleIndex(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "S3 工具") {
		t.Error("response should contain 'S3 工具'")
	}
}

func TestHandleConnect_MissingFields(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name string
		body string
	}{
		{"empty", "{}"},
		{"missing endpoint", `{"ak":"test","sk":"test","bucket":"test","ssl":true}`},
		{"missing ak", `{"endpoint":"test","sk":"test","bucket":"test","ssl":true}`},
		{"missing sk", `{"endpoint":"test","ak":"test","bucket":"test","ssl":true}`},
		{"missing bucket", `{"endpoint":"test","ak":"test","sk":"test","ssl":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/connect", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}

			var resp models.Response
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Success {
				t.Error("expected success to be false")
			}
		})
	}
}

func TestHandleConnect_WrongMethod(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/connect", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleList_WrongMethod(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleUpload_WrongMethod(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/upload", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleDownload_MissingKey(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/download", strings.NewReader(`{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false when key is missing")
	}
}

func TestHandleDownload_WrongMethod(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/download", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleDelete_WrongMethod(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/delete", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleDelete_MissingKey(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/delete", strings.NewReader(`{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false when key is missing")
	}
}

func TestHandleMkdir_WrongMethod(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/mkdir", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleMkdir_MissingFolder(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/mkdir", strings.NewReader(`{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false when folder is missing")
	}
}

func TestHandleRmdir_WrongMethod(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/rmdir", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleRmdir_MissingFolder(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/rmdir", strings.NewReader(`{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false when folder is missing")
	}
}

func TestHandleConnect_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/connect", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false for invalid JSON")
	}
}

func TestHandleList_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/list", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false for invalid JSON")
	}
}

func TestHandleIndex_NotFound(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleUpload_InvalidMultipart(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("not multipart"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false for invalid multipart")
	}
}

// handleBuckets 测试
func TestHandleBuckets_MissingFields(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name string
		body string
	}{
		{"empty", "{}"},
		{"missing endpoint", `{"ak":"test","sk":"test"}`},
		{"missing ak", `{"endpoint":"test","sk":"test"}`},
		{"missing sk", `{"endpoint":"test","ak":"test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/buckets", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}

			var resp models.BucketsResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Success {
				t.Error("expected success to be false")
			}
		})
	}
}

func TestHandleBuckets_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/buckets", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.BucketsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false for invalid JSON")
	}
}

// loggerMiddleware 测试
func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LoggerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

// handleIndex 额外测试
func TestHandleIndex_WithSubPath(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

// handleConnect 完整字段测试（无 S3 连接）
func TestHandleConnect_AllFields(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"https://s3.amazonaws.com","ak":"test-ak","sk":"test-sk","bucket":"test-bucket","ssl":true}`
	req := httptest.NewRequest(http.MethodPost, "/connect", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

// handleList 完整字段测试
func TestHandleList_AllFields(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"https://s3.amazonaws.com","ak":"test","sk":"test","bucket":"test","ssl":true,"prefix":"folder/","limit":100,"marker":"key"}`
	req := httptest.NewRequest(http.MethodPost, "/list", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

// handleUpload 无文件测试
func TestHandleUpload_NoFile(t *testing.T) {
	router := setupTestRouter()

	body := "--boundary\r\n\r\n--boundary--\r\n"
	req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(body))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// 测试空请求体处理
func TestEmptyBody(t *testing.T) {
	router := setupTestRouter()

	endpoints := []string{"/connect", "/buckets", "/list", "/download", "/delete", "/mkdir", "/rmdir"}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, ep, nil)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}

			var resp models.Response
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Success {
				t.Error("expected success to be false for empty body")
			}
		})
	}
}

// 测试超大 limit 值
func TestHandleList_LargeLimit(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"limit":10000}`
	req := httptest.NewRequest(http.MethodPost, "/list", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

// 测试负数 limit 值
func TestHandleList_NegativeLimit(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"limit":-10}`
	req := httptest.NewRequest(http.MethodPost, "/list", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

// 测试 handleMkdir 路径验证
func TestHandleMkdir_PathTraversal(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"folder":"../etc"}`
	req := httptest.NewRequest(http.MethodPost, "/mkdir", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Success {
		t.Error("expected success to be false for path traversal")
	}
}

// 测试 handleRmdir 路径验证
func TestHandleRmdir_PathTraversal(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"folder":"/absolute/path"}`
	req := httptest.NewRequest(http.MethodPost, "/rmdir", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Success {
		t.Error("expected success to be false for absolute path")
	}
}

// 测试 handleMkdir 文件夹名处理
func TestHandleMkdir_WithSlash(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"folder":"myfolder/"}`
	req := httptest.NewRequest(http.MethodPost, "/mkdir", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

// 测试 handleRmdir 文件夹名处理
func TestHandleRmdir_WithSlash(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"folder":"myfolder"}`
	req := httptest.NewRequest(http.MethodPost, "/rmdir", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}

func TestHandleGetCors_MissingFields(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name string
		body string
	}{
		{"empty", "{}"},
		{"missing endpoint", `{"ak":"test","sk":"test","bucket":"test"}`},
		{"missing ak", `{"endpoint":"test","sk":"test","bucket":"test"}`},
		{"missing sk", `{"endpoint":"test","ak":"test","bucket":"test"}`},
		{"missing bucket", `{"endpoint":"test","ak":"test","sk":"test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/cors/get", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}

			var resp models.CorsResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Success {
				t.Error("expected success to be false")
			}
		})
	}
}

func TestHandleGetCors_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/cors/get", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.CorsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false for invalid JSON")
	}
}

func TestHandleGetCors_WrongMethod(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/cors/get", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestHandleSetCors_MissingFields(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name string
		body string
	}{
		{"empty", "{}"},
		{"missing endpoint", `{"ak":"test","sk":"test","bucket":"test"}`},
		{"missing ak", `{"endpoint":"test","sk":"test","bucket":"test"}`},
		{"missing sk", `{"endpoint":"test","ak":"test","bucket":"test"}`},
		{"missing bucket", `{"endpoint":"test","ak":"test","sk":"test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/cors/set", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}

			var resp models.CorsResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Success {
				t.Error("expected success to be false")
			}
		})
	}
}

func TestHandleSetCors_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/cors/set", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.CorsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false for invalid JSON")
	}
}

func TestHandleSetCors_InvalidCorsConfig(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"config":"not valid json"}`
	req := httptest.NewRequest(http.MethodPost, "/cors/set", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.CorsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false for invalid CORS config")
	}

	if !strings.Contains(resp.Message, "CORS配置格式错误") {
		t.Errorf("expected CORS format error message, got: %s", resp.Message)
	}
}

func TestHandleSetCors_WrongMethod(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/cors/set", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

// 并发安全测试
func TestConcurrentRequests(t *testing.T) {
	router := setupTestRouter()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			done <- rec.Code == http.StatusOK
		}()
	}

	for i := 0; i < 10; i++ {
		if !<-done {
			t.Error("concurrent request failed")
		}
	}
}

// 路由存在性测试
func TestRoutesExist(t *testing.T) {
	router := setupTestRouter()

	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/health"},
		{"POST", "/connect"},
		{"POST", "/buckets"},
		{"POST", "/list"},
		{"POST", "/upload"},
		{"POST", "/download"},
		{"POST", "/delete"},
		{"POST", "/mkdir"},
		{"POST", "/rmdir"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+route.path, func(t *testing.T) {
			var req *http.Request
			if route.method == "GET" {
				req = httptest.NewRequest(route.method, route.path, nil)
			} else {
				req = httptest.NewRequest(route.method, route.path, strings.NewReader("{}"))
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusNotFound && route.path != "/nonexistent" {
				t.Errorf("route %s %s not found", route.method, route.path)
			}
		})
	}
}

// 错误响应格式测试
func TestErrorResponseFormat(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"connect missing fields", "POST", "/connect", "{}"},
		{"list invalid json", "POST", "/list", "not json"},
		{"download missing key", "POST", "/download", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test"}`},
		{"delete missing key", "POST", "/delete", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test"}`},
		{"mkdir missing folder", "POST", "/mkdir", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test"}`},
		{"rmdir missing folder", "POST", "/rmdir", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			var resp models.Response
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Success {
				t.Errorf("expected success to be false for %s", tt.name)
			}
		})
	}
}

// HandleBatchDelete 测试
func TestHandleBatchDelete_EmptyKeys(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"keys":[]}`
	req := httptest.NewRequest(http.MethodPost, "/batch/delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.BatchDeleteResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false for empty keys")
	}
}

func TestHandleBatchDelete_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/batch/delete", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// HandlePreview 测试
func TestHandlePreview_MissingKey(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true}`
	req := httptest.NewRequest(http.MethodPost, "/preview", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}

	var resp models.PreviewResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Success {
		t.Error("expected success to be false for missing key")
	}
}

func TestHandlePreview_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/preview", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// HandleRename 测试
func TestHandleRename_MissingKeys(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name string
		body string
	}{
		{"missing oldKey", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"newKey":"new.txt"}`},
		{"missing newKey", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"oldKey":"old.txt"}`},
		{"missing both", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/rename", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
		})
	}
}

func TestHandleRename_PathTraversal(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"oldKey":"../etc/passwd","newKey":"test.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/rename", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestHandleRename_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/rename", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// HandleMove 测试
func TestHandleMove_MissingKeys(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name string
		body string
	}{
		{"missing srcKey", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"dstKey":"new.txt"}`},
		{"missing dstKey", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"srcKey":"old.txt"}`},
		{"missing both", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/move", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
		})
	}
}

func TestHandleMove_PathTraversal(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"srcKey":"../etc/passwd","dstKey":"test.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/move", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestHandleMove_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/move", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// HandleCopy 测试
func TestHandleCopy_MissingKeys(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name string
		body string
	}{
		{"missing srcKey", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"dstKey":"new.txt"}`},
		{"missing dstKey", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"srcKey":"old.txt"}`},
		{"missing both", `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/copy", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", rec.Code)
			}
		})
	}
}

func TestHandleCopy_PathTraversal(t *testing.T) {
	router := setupTestRouter()

	body := `{"endpoint":"test","ak":"test","sk":"test","bucket":"test","ssl":true,"srcKey":"../etc/passwd","dstKey":"test.txt"}`
	req := httptest.NewRequest(http.MethodPost, "/copy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestHandleCopy_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/copy", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

// 新路由存在性测试
func TestNewRoutesExist(t *testing.T) {
	router := setupTestRouter()

	routes := []struct {
		method string
		path   string
	}{
		{"POST", "/batch/delete"},
		{"POST", "/preview"},
		{"POST", "/rename"},
		{"POST", "/move"},
		{"POST", "/copy"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusNotFound {
				t.Errorf("route %s %s not found", route.method, route.path)
			}
		})
	}
}
