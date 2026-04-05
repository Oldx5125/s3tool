package models

import (
	"encoding/json"
	"testing"
)

// 数据结构 JSON 序列化测试
func TestConnectRequest_JSON(t *testing.T) {
	jsonStr := `{"endpoint":"s3.amazonaws.com","ak":"test-ak","sk":"test-sk","bucket":"test-bucket","ssl":true}`
	var req ConnectRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Endpoint != "s3.amazonaws.com" {
		t.Errorf("expected endpoint s3.amazonaws.com, got %s", req.Endpoint)
	}
	if req.AK != "test-ak" {
		t.Errorf("expected ak test-ak, got %s", req.AK)
	}
	if !req.SSL {
		t.Error("expected SSL to be true")
	}
}

func TestListRequest_JSON(t *testing.T) {
	jsonStr := `{"endpoint":"s3.amazonaws.com","ak":"test","sk":"test","bucket":"test","ssl":true,"prefix":"folder/","limit":100,"marker":"last-key"}`
	var req ListRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Prefix != "folder/" {
		t.Errorf("expected prefix folder/, got %s", req.Prefix)
	}
	if req.Limit != 100 {
		t.Errorf("expected limit 100, got %d", req.Limit)
	}
	if req.Marker != "last-key" {
		t.Errorf("expected marker last-key, got %s", req.Marker)
	}
}

func TestResponse_JSON(t *testing.T) {
	resp := Response{Success: true, Message: "操作成功"}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("expected success to be true")
	}
	if decoded.Message != "操作成功" {
		t.Errorf("expected message '操作成功', got %s", decoded.Message)
	}
}

func TestObjectInfo_JSON(t *testing.T) {
	obj := ObjectInfo{
		Key:          "folder/file.txt",
		Size:         1024,
		LastModified: "2024-01-01T00:00:00Z",
	}
	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ObjectInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Key != obj.Key {
		t.Errorf("key mismatch")
	}
	if decoded.Size != obj.Size {
		t.Errorf("size mismatch")
	}
}

func TestListResponse_JSON(t *testing.T) {
	resp := ListResponse{
		Objects: []ObjectInfo{
			{Key: "file1.txt", Size: 100, LastModified: "2024-01-01"},
			{Key: "file2.txt", Size: 200, LastModified: "2024-01-02"},
		},
		Folders:    []string{"folder1/", "folder2/"},
		NextMarker: "file2.txt",
		HasMore:    true,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ListResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.Objects) != 2 {
		t.Errorf("expected 2 objects, got %d", len(decoded.Objects))
	}
	if len(decoded.Folders) != 2 {
		t.Errorf("expected 2 folders, got %d", len(decoded.Folders))
	}
	if !decoded.HasMore {
		t.Error("expected HasMore to be true")
	}
}

func TestBucketInfo_JSON(t *testing.T) {
	info := BucketInfo{
		Name:         "my-bucket",
		CreationDate: "2024-01-01T00:00:00Z",
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded BucketInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != "my-bucket" {
		t.Errorf("expected name my-bucket, got %s", decoded.Name)
	}
}

func TestBucketsResponse_JSON(t *testing.T) {
	resp := BucketsResponse{
		Success: true,
		Buckets: []BucketInfo{
			{Name: "bucket1", CreationDate: "2024-01-01"},
			{Name: "bucket2", CreationDate: "2024-01-02"},
		},
		Message: "",
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded BucketsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("expected success to be true")
	}
	if len(decoded.Buckets) != 2 {
		t.Errorf("expected 2 buckets, got %d", len(decoded.Buckets))
	}
}

func TestCorsRequest_JSON(t *testing.T) {
	jsonStr := `{"endpoint":"s3.amazonaws.com","ak":"test-ak","sk":"test-sk","bucket":"test-bucket","ssl":true,"config":"{\"CORSRules\":[]}"}`
	var req CorsRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Endpoint != "s3.amazonaws.com" {
		t.Errorf("expected endpoint s3.amazonaws.com, got %s", req.Endpoint)
	}
	if req.Config != `{"CORSRules":[]}` {
		t.Errorf("expected config, got %s", req.Config)
	}
	if req.AK != "test-ak" {
		t.Errorf("expected ak test-ak, got %s", req.AK)
	}
	if req.SK != "test-sk" {
		t.Errorf("expected sk test-sk, got %s", req.SK)
	}
	if req.Bucket != "test-bucket" {
		t.Errorf("expected bucket test-bucket, got %s", req.Bucket)
	}
	if !req.SSL {
		t.Error("expected SSL to be true")
	}
}

func TestCorsResponse_JSON(t *testing.T) {
	resp := CorsResponse{
		Success: true,
		Message: "操作成功",
		Config:  `{"CORSRules":[{"AllowedOrigin":["*"]}]}`,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded CorsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("expected success to be true")
	}
	if decoded.Message != "操作成功" {
		t.Errorf("expected message '操作成功', got %s", decoded.Message)
	}
	if decoded.Config != `{"CORSRules":[{"AllowedOrigin":["*"]}]}` {
		t.Errorf("expected config mismatch, got %s", decoded.Config)
	}
}
