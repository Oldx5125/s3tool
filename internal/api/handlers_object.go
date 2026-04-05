package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"s3tool/internal/i18n"
	"s3tool/internal/models"
	"s3tool/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

const maxConcurrentDeletes = 10

// handlePartialSuccess 处理复制成功但删除源文件失败的情况
func handlePartialSuccess(c *gin.Context, operation, bucket, srcKey, dstKey string, err error, partialMsg string) {
	slog.Warn("删除源文件失败，文件可能同时存在于两个位置", "error", err, "key", srcKey)
	slog.Info(operation+"成功（部分）", "bucket", bucket, "srcKey", srcKey, "dstKey", dstKey)
	c.JSON(http.StatusOK, models.Response{Success: true, Message: fmt.Sprintf(partialMsg, err.Error())})
}

// HandleBatchDelete 处理批量删除请求
func HandleBatchDelete(c *gin.Context) {
	req, client, ctx, cancel, ok := withClient[models.BatchDeleteRequest](c, 120*time.Second)
	if !ok {
		return
	}
	defer cancel()

	if len(req.Keys) == 0 {
		c.JSON(http.StatusBadRequest, models.BatchDeleteResponse{Success: false, Message: i18n.T("err_batch_keys")})
		return
	}

	results := make([]models.DeleteResult, len(req.Keys))
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, maxConcurrentDeletes)

	for i, key := range req.Keys {
		wg.Add(1)
		go func(idx int, k string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := models.DeleteResult{Key: k, Success: true}

			if err := utils.ValidatePath(k); err != nil {
				result.Success = false
				result.Error = fmt.Sprintf(i18n.T("err_folder_path"), err.Error())
			} else if err := client.RemoveObject(ctx, req.Bucket, k, minio.RemoveObjectOptions{}); err != nil {
				result.Success = false
				result.Error = err.Error()
			}

			mu.Lock()
			results[idx] = result
			mu.Unlock()
		}(i, key)
	}
	wg.Wait()

	successCount := 0
	for _, r := range results {
		if r.Success {
			utils.EnsureParentFolders(ctx, client, req.Bucket, r.Key)
			successCount++
		}
	}

	slog.Info("批量删除完成", "bucket", req.Bucket, "total", len(req.Keys), "success", successCount)
	c.JSON(http.StatusOK, models.BatchDeleteResponse{
		Success: successCount > 0,
		Message: fmt.Sprintf(i18n.T("success_batch_delete"), successCount, len(req.Keys)),
		Results: results,
	})
}

// HandleRename 处理重命名请求
func HandleRename(c *gin.Context) {
	req, client, ctx, cancel, ok := withClient[models.RenameRequest](c, 60*time.Second)
	if !ok {
		return
	}
	defer cancel()

	if req.OldKey == "" || req.NewKey == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_rename_old")})
		return
	}

	if err := utils.ValidatePath(req.OldKey); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_upload_path"), err.Error())})
		return
	}
	if err := utils.ValidatePath(req.NewKey); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_upload_path"), err.Error())})
		return
	}

	if err := copyObject(ctx, client, req.Bucket, req.OldKey, req.NewKey); err != nil {
		slog.Error("复制文件失败", "error", err, "oldKey", req.OldKey, "newKey", req.NewKey)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_rename_failed"), err.Error())})
		return
	}

	if err := client.RemoveObject(ctx, req.Bucket, req.OldKey, minio.RemoveObjectOptions{}); err != nil {
		handlePartialSuccess(c, "重命名", req.Bucket, req.OldKey, req.NewKey, err, i18n.T("success_rename_partial"))
		return
	}

	slog.Info("重命名成功", "bucket", req.Bucket, "oldKey", req.OldKey, "newKey", req.NewKey)
	c.JSON(http.StatusOK, models.Response{Success: true, Message: i18n.T("success_rename")})
}

// HandleMove 处理移动请求
func HandleMove(c *gin.Context) {
	req, client, ctx, cancel, ok := withClient[models.MoveRequest](c, 60*time.Second)
	if !ok {
		return
	}
	defer cancel()

	if req.SrcKey == "" || req.DstKey == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_move_src")})
		return
	}

	if err := utils.ValidatePath(req.SrcKey); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_folder_path"), err.Error())})
		return
	}
	if err := utils.ValidatePath(req.DstKey); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_folder_path"), err.Error())})
		return
	}

	if err := copyObject(ctx, client, req.Bucket, req.SrcKey, req.DstKey); err != nil {
		slog.Error("复制文件失败", "error", err, "srcKey", req.SrcKey, "dstKey", req.DstKey)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_move_failed"), err.Error())})
		return
	}

	if err := client.RemoveObject(ctx, req.Bucket, req.SrcKey, minio.RemoveObjectOptions{}); err != nil {
		handlePartialSuccess(c, "移动", req.Bucket, req.SrcKey, req.DstKey, err, i18n.T("success_move_partial"))
		return
	}

	slog.Info("移动成功", "bucket", req.Bucket, "srcKey", req.SrcKey, "dstKey", req.DstKey)
	c.JSON(http.StatusOK, models.Response{Success: true, Message: i18n.T("success_move")})
}

// HandleCopy 处理复制请求
func HandleCopy(c *gin.Context) {
	req, client, ctx, cancel, ok := withClient[models.CopyRequest](c, 60*time.Second)
	if !ok {
		return
	}
	defer cancel()

	if req.SrcKey == "" || req.DstKey == "" {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: i18n.T("err_copy_same")})
		return
	}

	if err := utils.ValidatePath(req.SrcKey); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_folder_path"), err.Error())})
		return
	}
	if err := utils.ValidatePath(req.DstKey); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_folder_path"), err.Error())})
		return
	}

	if err := copyObject(ctx, client, req.Bucket, req.SrcKey, req.DstKey); err != nil {
		slog.Error("复制文件失败", "error", err, "srcKey", req.SrcKey, "dstKey", req.DstKey)
		c.JSON(http.StatusBadGateway, models.Response{Success: false, Message: fmt.Sprintf(i18n.T("err_copy_failed"), err.Error())})
		return
	}

	slog.Info("复制成功", "bucket", req.Bucket, "srcKey", req.SrcKey, "dstKey", req.DstKey)
	c.JSON(http.StatusOK, models.Response{Success: true, Message: i18n.T("success_copy")})
}
