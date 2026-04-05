package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7/pkg/cors"
	"s3tool/internal/i18n"
	"s3tool/internal/models"
)

// HandleConnect 处理连接请求
func HandleConnect(c *gin.Context) {
	var req models.ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("解析请求失败", "error", err)
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_request_format")})
		return
	}

	// 验证必填字段
	if req.Endpoint == "" || req.AK == "" || req.SK == "" || req.Bucket == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_missing_fields")})
		return
	}

	// 创建客户端并测试连接
	client, err := ClientFactory(&req)
	if err != nil {
		slog.Error("创建客户端失败", "error", err, "endpoint", req.Endpoint)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_connection_failed"), err.Error())})
		return
	}

	// 检查 bucket 是否存在
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, req.Bucket)
	if err != nil {
		slog.Error("检查 Bucket 失败", "error", err, "bucket", req.Bucket)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_check_bucket"), err.Error())})
		return
	}

	if !exists {
		slog.Error("Bucket 不存在", "bucket", req.Bucket)
		c.JSON(http.StatusNotFound, models.Response{Success: false, Message: i18n.T("err_bucket_not_found")})
		return
	}

	slog.Info("连接成功", "endpoint", req.Endpoint, "bucket", req.Bucket)
	c.JSON(http.StatusOK, models.Response{Success: true, Message: i18n.T("success_connection")})
}

// HandleBuckets 处理列出存储桶请求
func HandleBuckets(c *gin.Context) {
	var req models.ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("解析请求失败", "error", err)
		c.JSON(http.StatusBadRequest, models.BucketsResponse{Success: false, Message: i18n.T("err_request_format")})
		return
	}

	if req.Endpoint == "" || req.AK == "" || req.SK == "" {
		c.JSON(http.StatusBadRequest, models.BucketsResponse{Success: false, Message: i18n.T("err_missing_endpoint")})
		return
	}

	client, err := ClientFactory(&req)
	if err != nil {
		slog.Error("创建客户端失败", "error", err, "endpoint", req.Endpoint)
		c.JSON(http.StatusBadGateway, models.BucketsResponse{Success: false, Message: fmt.Sprintf(i18n.T("err_connection_failed"), err.Error())})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		slog.Error("列出存储桶失败", "error", err)
		c.JSON(http.StatusBadGateway, models.BucketsResponse{Success: false, Message: fmt.Sprintf(i18n.T("err_list_buckets"), err.Error())})
		return
	}

	var result []models.BucketInfo
	for _, b := range buckets {
		result = append(result, models.BucketInfo{
			Name:         b.Name,
			CreationDate: b.CreationDate.Format(time.RFC3339),
		})
	}

	slog.Info("列出存储桶", "count", len(result), "endpoint", req.Endpoint)
	c.JSON(http.StatusOK, models.BucketsResponse{Success: true, Buckets: result})
}

// HandleGetCors 获取CORS配置
func HandleGetCors(c *gin.Context) {
	var req models.CorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("解析请求失败", "error", err)
		c.JSON(http.StatusBadRequest, models.CorsResponse{
			Success: false,
			Message: i18n.T("err_request_format"),
		})
		return
	}

	// 验证必填字段
	if req.Endpoint == "" || req.AK == "" || req.SK == "" || req.Bucket == "" {
		c.JSON(http.StatusBadRequest, models.CorsResponse{
			Success: false,
			Message: i18n.T("err_missing_fields"),
		})
		return
	}

	client, err := ClientFactory(&req.ConnectRequest)
	if err != nil {
		slog.Error("创建客户端失败", "error", err, "endpoint", req.Endpoint)
		c.JSON(http.StatusBadGateway, models.CorsResponse{
			Success: false,
			Message: fmt.Sprintf(i18n.T("err_connection_failed"), err.Error()),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	corsConfig, err := client.GetBucketCors(ctx, req.Bucket)
	if err != nil {
		// 检查是否为无配置错误
		if strings.Contains(err.Error(), "NoSuchCORSConfiguration") || strings.Contains(err.Error(), "NoSuchBucket") {
			slog.Info("获取CORS配置", "bucket", req.Bucket, "result", i18n.T("err_cors_none"))
			c.JSON(http.StatusOK, models.CorsResponse{
				Success: true,
				Config:  "",
			})
			return
		}
		slog.Error("获取CORS配置失败", "error", err, "bucket", req.Bucket)
		c.JSON(http.StatusBadGateway, models.CorsResponse{
			Success: false,
			Message: fmt.Sprintf(i18n.T("err_get_cors"), err.Error()),
		})
		return
	}

	// 序列化为JSON返回
	configJSON, err := json.Marshal(corsConfig)
	if err != nil {
		slog.Error("序列化CORS配置失败", "error", err)
		c.JSON(http.StatusBadGateway, models.CorsResponse{
			Success: false,
			Message: "序列化配置失败: " + err.Error(),
		})
		return
	}

	slog.Info("获取CORS配置成功", "bucket", req.Bucket)
	c.JSON(http.StatusOK, models.CorsResponse{
		Success: true,
		Config:  string(configJSON),
	})
}

// HandleSetCors 设置CORS配置
func HandleSetCors(c *gin.Context) {
	var req models.CorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("解析请求失败", "error", err)
		c.JSON(http.StatusBadRequest, models.CorsResponse{
			Success: false,
			Message: i18n.T("err_request_format"),
		})
		return
	}

	// 验证必填字段
	if req.Endpoint == "" || req.AK == "" || req.SK == "" || req.Bucket == "" {
		c.JSON(http.StatusBadRequest, models.CorsResponse{
			Success: false,
			Message: i18n.T("err_missing_fields"),
		})
		return
	}

	client, err := ClientFactory(&req.ConnectRequest)
	if err != nil {
		slog.Error("创建客户端失败", "error", err, "endpoint", req.Endpoint)
		c.JSON(http.StatusBadGateway, models.CorsResponse{
			Success: false,
			Message: fmt.Sprintf(i18n.T("err_connection_failed"), err.Error()),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// 处理删除CORS配置（空配置）
	if req.Config == "" {
		err = client.SetBucketCors(ctx, req.Bucket, nil)
		if err != nil {
			slog.Error("删除CORS失败", "error", err, "bucket", req.Bucket)
			c.JSON(http.StatusBadGateway, models.CorsResponse{
				Success: false,
				Message: fmt.Sprintf(i18n.T("err_set_cors"), err.Error()),
			})
			return
		}
		slog.Info("删除CORS配置成功", "bucket", req.Bucket)
		c.JSON(http.StatusOK, models.CorsResponse{
			Success: true,
			Message: i18n.T("success_cors_delete"),
		})
		return
	}

	// 解析JSON配置
	var corsConfig cors.Config
	if err := json.Unmarshal([]byte(req.Config), &corsConfig); err != nil {
		slog.Error("解析CORS配置失败", "error", err)
		c.JSON(http.StatusBadRequest, models.CorsResponse{
			Success: false,
			Message: i18n.T("err_cors_format") + ": " + err.Error(),
		})
		return
	}

	// 设置CORS配置
	err = client.SetBucketCors(ctx, req.Bucket, &corsConfig)
	if err != nil {
		slog.Error("设置CORS失败", "error", err, "bucket", req.Bucket)
		c.JSON(http.StatusBadGateway, models.CorsResponse{
			Success: false,
			Message: fmt.Sprintf(i18n.T("err_set_cors"), err.Error()),
		})
		return
	}

	slog.Info("设置CORS配置成功", "bucket", req.Bucket)
	c.JSON(http.StatusOK, models.CorsResponse{
		Success: true,
		Message: i18n.T("success_cors_set"),
	})
}