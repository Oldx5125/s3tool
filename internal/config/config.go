package config

import (
	"fmt"
	"os"
	"strings"
)

const Version = "1.0.0"

// Config 应用配置
type Config struct {
	Port    string
	Addr    string
	Help    bool
	Version bool
}

// Parse 解析命令行参数
func Parse() *Config {
	cfg := &Config{
		Port: "9090",
		Addr: "",
	}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "-h" || arg == "-help" || arg == "--help":
			cfg.Help = true
		case arg == "-v" || arg == "-version" || arg == "--version":
			cfg.Version = true
		case arg == "-p" && i+1 < len(os.Args):
			cfg.Port = os.Args[i+1]
			i++
		case arg == "-addr" && i+1 < len(os.Args):
			cfg.Addr = os.Args[i+1]
			i++
		case strings.HasPrefix(arg, "-p="):
			cfg.Port = strings.TrimPrefix(arg, "-p=")
		case strings.HasPrefix(arg, "-addr="):
			cfg.Addr = strings.TrimPrefix(arg, "-addr=")
		}
	}

	return cfg
}

// ListenAddress 返回完整的监听地址
func (c *Config) ListenAddress() string {
	if c.Addr != "" {
		return c.Addr + ":" + c.Port
	}
	return ":" + c.Port
}

// PrintHelp 显示帮助信息
func PrintHelp() {
	fmt.Println(`S3 工具 - 单文件、无依赖、即开即用的 S3 兼容存储管理工具

用法:
  s3tool [选项]

选项:
  -p <port>      监听端口 (默认: 9090)
  -addr <addr>   监听地址 (默认: 0.0.0.0)
  -h, -help      显示帮助信息
  -v, -version   显示版本信息

示例:
  s3tool                    # 默认监听 0.0.0.0:9090
  s3tool -p 8080            # 监听端口 8080
  s3tool -addr 127.0.0.1    # 仅监听本地
  s3tool -p 8080 -addr 127.0.0.1

访问:
  浏览器打开 http://localhost:<port>

功能:
  - 连接任意 S3 兼容存储
  - 文件夹导航
  - 上传/下载/删除文件
  - 创建/删除文件夹`)
}

// PrintVersion 显示版本信息
func PrintVersion() {
	fmt.Printf("s3tool version %s\n", Version)
}
