package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"s3tool/internal/api"
	"s3tool/internal/config"
)

//go:embed frontend.html
var frontendFS embed.FS

func main() {
	// 解析命令行参数
	cfg := config.Parse()

	if cfg.Help {
		config.PrintHelp()
		os.Exit(0)
	}

	if cfg.Version {
		config.PrintVersion()
		os.Exit(0)
	}

	// 配置日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// 设置 Gin 为生产模式
	gin.SetMode(gin.ReleaseMode)

	// 设置前端文件系统
	api.FrontendFS = frontendFS

	// 创建路由器
	router := gin.New()
	router.Use(gin.Recovery(), api.LoggerMiddleware())

	// 注册路由
	router.GET("/", api.HandleIndex)
	router.GET("/health", api.HandleHealth)
	router.POST("/connect", api.HandleConnect)
	router.POST("/buckets", api.HandleBuckets)
	router.POST("/list", api.HandleList)
	router.POST("/upload", api.HandleUpload)
	router.POST("/download", api.HandleDownload)
	router.POST("/delete", api.HandleDelete)
	router.POST("/batch/delete", api.HandleBatchDelete)
	router.POST("/preview", api.HandlePreview)
	router.POST("/rename", api.HandleRename)
	router.POST("/move", api.HandleMove)
	router.POST("/copy", api.HandleCopy)
	router.POST("/mkdir", api.HandleMkdir)
	router.POST("/rmdir", api.HandleRmdir)
	router.POST("/cors/get", api.HandleGetCors)
	router.POST("/cors/set", api.HandleSetCors)

	// 启动服务器
	listenAddr := cfg.ListenAddress()
	slog.Info("服务器启动", "addr", listenAddr)
	if err := router.Run(listenAddr); err != nil {
		slog.Error("服务器启动失败", "error", err)
		os.Exit(1)
	}
}
