package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"s3tool/internal/config"
)

// HandleHealth 健康检查端点
func HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": config.Version,
	})
}

// HandleIndex 返回前端页面
func HandleIndex(c *gin.Context) {
	// 处理非根路径请求
	if c.Request.URL.Path != "/" {
		c.String(http.StatusNotFound, "Not Found")
		return
	}

	content, err := FrontendFS.ReadFile("frontend.html")
	if err != nil {
		slog.Error("读取前端文件失败", "error", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
}