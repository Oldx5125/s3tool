package api

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"s3tool/internal/i18n"
	"s3tool/internal/models"
	"s3tool/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

// HandleUpload 处理上传请求
func HandleUpload(c *gin.Context) {
	// 解析 multipart form
	req := models.ConnectRequest{
		Endpoint: c.PostForm("endpoint"),
		AK:       c.PostForm("ak"),
		SK:       c.PostForm("sk"),
		Bucket:   c.PostForm("bucket"),
		SSL:      c.PostForm("ssl") == "true",
	}

	// 支持自定义 key（包含路径），否则使用文件名
	key := c.PostForm("key")

	file, err := c.FormFile("file")
	if err != nil {
		slog.Error("获取文件失败", "error", err)
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_upload_file"), err.Error())})
		return
	}

	client, err := ClientFactory(&req)
	if err != nil {
		slog.Error("创建客户端失败", "error", err)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_connection_failed"), err.Error())})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 300*time.Second)
	defer cancel()

	// 使用自定义 key 或文件名
	objectKey := key
	if objectKey == "" {
		objectKey = file.Filename
	}

	// 验证路径安全性
	if err := utils.ValidatePath(objectKey); err != nil {
		slog.Error("路径验证失败", "error", err, "key", objectKey)
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_upload_path"), err.Error())})
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		slog.Error("打开文件失败", "error", err)
		c.JSON(http.StatusInternalServerError, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_upload_open"), err.Error())})
		return
	}
	defer src.Close()

	_, err = client.PutObject(ctx, req.Bucket, objectKey, src, file.Size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		slog.Error("上传失败", "error", err, "key", objectKey)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_upload_failed"), err.Error())})
		return
	}

	slog.Info("上传成功", "bucket", req.Bucket, "key", objectKey, "size", file.Size)
	c.JSON(http.StatusOK, models.Response{Success: true, Message: i18n.T("success_upload")})
}

// HandleDownload 处理下载请求
func HandleDownload(c *gin.Context) {
	req, client, ctx, cancel, ok := withClient[models.DownloadRequest](c, 300*time.Second)
	if !ok {
		return
	}
	defer cancel()

	if req.Key == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_download_file")})
		return
	}

	object, err := client.GetObject(ctx, req.Bucket, req.Key, minio.GetObjectOptions{})
	if err != nil {
		slog.Error("获取对象失败", "error", err, "key", req.Key)
		c.JSON(http.StatusNotFound, models.Response{Success: false, Message: i18n.T("err_download_not_found")})
		return
	}
	defer object.Close()

	info, err := object.Stat()
	if err != nil {
		if isNotFoundError(err) {
			slog.Error("文件不存在", "key", req.Key)
			c.JSON(http.StatusNotFound, models.Response{Success: false, Message: i18n.T("err_download_not_found")})
			return
		}
		slog.Error("获取对象信息失败", "error", err, "key", req.Key)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: i18n.T("err_download_info")})
		return
	}

	extraHeaders := map[string]string{
		"Content-Disposition": contentDisposition(req.Key),
	}

	c.DataFromReader(http.StatusOK, info.Size, "application/octet-stream", object, extraHeaders)
	slog.Info("下载成功", "bucket", req.Bucket, "key", req.Key)
}

// HandleDelete 处理删除请求
func HandleDelete(c *gin.Context) {
	req, client, ctx, cancel, ok := withClient[models.DeleteRequest](c, 60*time.Second)
	if !ok {
		return
	}
	defer cancel()

	if req.Key == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_delete_file")})
		return
	}

	err := client.RemoveObject(ctx, req.Bucket, req.Key, minio.RemoveObjectOptions{})
	if err != nil {
		slog.Error("删除失败", "error", err, "key", req.Key)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_delete_failed"), err.Error())})
		return
	}

	utils.EnsureParentFolders(ctx, client, req.Bucket, req.Key)

	slog.Info("删除成功", "bucket", req.Bucket, "key", req.Key)
	c.JSON(http.StatusOK, models.Response{Success: true, Message: i18n.T("success_delete")})
}

// HandlePreview 处理文件预览请求
func HandlePreview(c *gin.Context) {
	req, client, ctx, cancel, ok := withClient[models.PreviewRequest](c, 60*time.Second)
	if !ok {
		return
	}
	defer cancel()

	if req.Key == "" {
		c.JSON(http.StatusBadRequest, models.PreviewResponse{Success: false, Message: i18n.T("err_preview_file")})
		return
	}

	stat, err := client.StatObject(ctx, req.Bucket, req.Key, minio.StatObjectOptions{})
	if err != nil {
		slog.Error("获取对象信息失败", "error", err, "key", req.Key)
		c.JSON(http.StatusNotFound, models.PreviewResponse{Success: false, Message: i18n.T("err_preview_not_found")})
		return
	}

	contentType := stat.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ext := strings.ToLower(strings.TrimSuffix(req.Key, "/"))
	if idx := strings.LastIndex(ext, "."); idx >= 0 {
		ext = ext[idx:]
	}

	if (imageExts[ext] || textExts[ext] || codeExts[ext] || strings.HasPrefix(contentType, "text/")) && stat.Size <= maxInlinePreviewSize {
		data, err := readObjectContent(ctx, client, req.Bucket, req.Key)
		if err != nil {
			if errors.Is(err, ErrPreviewSizeExceeded) {
				c.JSON(http.StatusBadRequest, models.PreviewResponse{Success: false, Message: err.Error()})
			} else {
				c.JSON(http.StatusBadGateway, models.PreviewResponse{Success: false, Message: "读取文件失败"})
			}
			return
		}

		var responseData string
		if imageExts[ext] {
			responseData = base64.StdEncoding.EncodeToString(data)
		} else {
			responseData = string(data)
		}

		c.JSON(http.StatusOK, models.PreviewResponse{
			Success:     true,
			ContentType: contentType,
			Data:        responseData,
		})
		return
	}

	if videoExts[ext] || audioExts[ext] || ext == ".pdf" || stat.Size > maxInlinePreviewSize {
		presignedURL, err := client.PresignedGetObject(ctx, req.Bucket, req.Key, 10*time.Minute, nil)
		if err != nil {
			c.JSON(http.StatusBadGateway, models.PreviewResponse{Success: false, Message: "生成预签名 URL 失败"})
			return
		}

		c.JSON(http.StatusOK, models.PreviewResponse{
			Success:     true,
			ContentType: contentType,
			URL:         presignedURL.String(),
		})
		return
	}

	c.JSON(http.StatusOK, models.PreviewResponse{
		Success:     false,
		ContentType: contentType,
		Message:     "该文件类型不支持预览，请下载查看",
	})
}
