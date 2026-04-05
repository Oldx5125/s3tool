package api

import (
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

// HandleList 处理列表请求
func HandleList(c *gin.Context) {
	req, client, ctx, cancel, ok := withClient[models.ListRequest](c, 60*time.Second)
	if !ok {
		return
	}
	defer cancel()

	limit := req.Limit
	if limit <= 0 {
		limit = 1000
	}
	if limit > 5000 {
		limit = 5000
	}

	var objects []models.ObjectInfo
	var folders []string
	var lastKey string
	var totalCount int

	for object := range client.ListObjects(ctx, req.Bucket, minio.ListObjectsOptions{
		Prefix:    req.Prefix,
		Recursive: false,
		MaxKeys:   limit,
	}) {
		if object.Err != nil {
			slog.Error("列出对象失败", "error", object.Err)
			c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_list_buckets"), object.Err.Error())})
			return
		}

		if object.Key != req.Prefix && strings.HasSuffix(object.Key, "/") {
			folders = append(folders, object.Key)
		} else if object.Key != req.Prefix {
			objects = append(objects, models.ObjectInfo{
				Key:          object.Key,
				Size:         object.Size,
				LastModified: object.LastModified.Format(time.RFC3339),
			})
		}
		lastKey = object.Key
		totalCount++
	}

	hasMore := totalCount >= limit

	slog.Info("列出对象", "bucket", req.Bucket, "prefix", req.Prefix, "folders", len(folders), "objects", len(objects), "hasMore", hasMore)
	c.JSON(http.StatusOK, models.ListResponse{
		Objects:    objects,
		Folders:    folders,
		NextMarker: lastKey,
		HasMore:    hasMore,
	})
}

// HandleMkdir 处理创建文件夹请求
func HandleMkdir(c *gin.Context) {
	req, client, ctx, cancel, ok := withClient[models.MkdirRequest](c, 30*time.Second)
	if !ok {
		return
	}
	defer cancel()

	if req.Folder == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_folder_name")})
		return
	}

	folder := strings.TrimSuffix(req.Folder, "/") + "/"

	if err := utils.ValidatePath(folder); err != nil {
		slog.Error("路径验证失败", "error", err, "folder", folder)
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_folder_path"), err.Error())})
		return
	}

	_, err := client.PutObject(ctx, req.Bucket, folder, strings.NewReader(""), 0, minio.PutObjectOptions{
		ContentType: "application/x-directory",
	})
	if err != nil {
		slog.Error("创建文件夹失败", "error", err, "folder", folder)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_folder_create"), err.Error())})
		return
	}

	slog.Info("创建文件夹成功", "bucket", req.Bucket, "folder", folder)
	c.JSON(http.StatusOK, models.Response{Success: true, Message: i18n.T("success_folder_create")})
}

// HandleRmdir 处理删除文件夹请求
func HandleRmdir(c *gin.Context) {
	req, client, ctx, cancel, ok := withClient[models.RmdirRequest](c, 60*time.Second)
	if !ok {
		return
	}
	defer cancel()

	if req.Folder == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_folder_name")})
		return
	}

	folder := strings.TrimSuffix(req.Folder, "/") + "/"

	if err := utils.ValidatePath(folder); err != nil {
		slog.Error("路径验证失败", "error", err, "folder", folder)
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_folder_path"), err.Error())})
		return
	}

	hasObjects := false
	for object := range client.ListObjects(ctx, req.Bucket, minio.ListObjectsOptions{
		Prefix:    folder,
		Recursive: true,
		MaxKeys:   2,
	}) {
		if object.Err != nil {
			slog.Error("列出对象失败", "error", object.Err)
			c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_folder_delete"), object.Err.Error())})
			return
		}
		if object.Key != folder {
			hasObjects = true
			break
		}
	}

	if hasObjects {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_folder_empty")})
		return
	}

	err := client.RemoveObject(ctx, req.Bucket, folder, minio.RemoveObjectOptions{})
	if err != nil {
		slog.Error("删除文件夹失败", "error", err, "folder", folder)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_folder_delete"), err.Error())})
		return
	}

	slog.Info("删除文件夹成功", "bucket", req.Bucket, "folder", folder)
	c.JSON(http.StatusOK, models.Response{Success: true, Message: i18n.T("success_folder_delete")})
}
