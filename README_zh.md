# S3 工具

[English](README.md) | 简体中文

一个**单文件、无依赖、即开即用**的 S3 兼容存储管理工具。

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 功能特性

### 核心功能
- :white_check_mark: 连接任意 S3 兼容存储（AWS S3、MinIO、华为 OBS 等）
- :white_check_mark: 列出账户中的所有存储桶
- :white_check_mark: 存储桶选择（无需手动输入存储桶名称）
- :white_check_mark: 快速切换存储桶（无需退出重连）
- :white_check_mark: 文件夹导航（面包屑、双击进入）
- :white_check_mark: 自签名 SSL 证书支持
- :white_check_mark: 刷新页面自动恢复连接
- :white_check_mark: 退出登录清除凭证
- :white_check_mark: 命令行参数配置

### 文件操作
- :white_check_mark: 上传文件（支持上传到指定目录）
- :white_check_mark: 下载文件
- :white_check_mark: 删除文件
- :white_check_mark: 创建文件夹
- :white_check_mark: 删除文件夹（仅空文件夹）
- :white_check_mark: 重命名文件
- :white_check_mark: 移动文件（跨目录移动）
- :white_check_mark: 复制文件

### 批量操作
- :white_check_mark: 多选文件（复选框 + 全选）
- :white_check_mark: 批量上传（多文件选择 + 队列进度）
- :white_check_mark: 批量下载（逐个下载选中文件）
- :white_check_mark: 批量删除（并发删除）

### 文件预览
- :white_check_mark: 图片预览（JPEG、PNG、GIF、WebP、SVG）
- :white_check_mark: 文本预览（自动检测编码）
- :white_check_mark: 视频播放（MP4、WebM）
- :white_check_mark: 音频播放（MP3、WAV、OGG）
- :white_check_mark: PDF 预览

### 存储桶管理
- :white_check_mark: 配置存储桶 CORS 规则

## 安装

### 编译安装

```bash
# GitHub
git clone https://github.com/Oldx5125/s3tool.git
# 或 Gitee（国内更快）
git clone https://gitee.com/kylingx/s3tool.git
cd s3tool
go build -o s3tool .

## 快速开始

### 启动服务

```bash
# 默认监听 0.0.0.0:9090
./s3tool

# 指定端口
./s3tool -p 8080

# 仅本地访问
./s3tool -addr 127.0.0.1

# 指定端口和地址
./s3tool -p 8080 -addr 127.0.0.1
```

### 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-p <port>` | 监听端口 | 9090 |
| `-addr <addr>` | 监听地址 | 0.0.0.0 |
| `-h, -help` | 显示帮助信息 | - |
| `-v, -version` | 显示版本信息 | - |

### 访问界面

浏览器打开 http://localhost:9090

### 连接存储

在界面中输入：

| 字段 | 说明 | 示例 |
|------|------|------|
| Endpoint | S3 服务地址 | `s3.amazonaws.com` 或 `10.50.2.45:5080` |
| AK | Access Key | - |
| SK | Secret Key | - |
| SSL | 是否使用 HTTPS | 根据实际情况勾选 |

**连接步骤：**
1. 填写 Endpoint、AK、SK
2. 点击"列出存储桶"
3. 从下拉列表选择存储桶
4. 点击"连接"

> 支持 URL 格式的 Endpoint，如 `https://s3.amazonaws.com`，会自动解析。

## 国际化

本工具支持**中文**和**英文**两种语言（前端实现）。

### 语言检测方式

| 方式 | 说明 |
|------|------|
| URL 参数 | 在 URL 后添加 `?lang=en` 或 `?lang=zh` |
| 浏览器设置 | 前端自动检测浏览器语言 |

### 示例

前端访问：
- 中文（默认）：http://localhost:9090
- 英文：http://localhost:9090?lang=en

## 使用指南

### 文件列表操作

- **单击** - 选中文件/文件夹
- **双击文件夹** - 进入文件夹
- **表头排序** - 点击表头按文件名、大小或修改时间排序（再次点击切换升序/降序）
- **面包屑导航** - 点击路径层级快速跳转
- **返回上级** - 返回上一级目录

### 文件操作

| 操作 | 说明 |
|------|------|
| 列出存储桶 | 获取账户中的所有存储桶 |
| 切换存储桶 | 连接后可切换到其他存储桶 |
| 上传文件 | 上传到当前目录（支持多选） |
| 批量下载 | 下载所有选中的文件 |
| 批量删除 | 并发删除所有选中的文件 |
| 新建文件夹 | 在当前目录创建子文件夹 |
| 重命名 | 重命名文件或文件夹 |
| 移动 | 移动文件到其他目录 |
| 复制 | 复制文件到其他目录 |
| 预览 | 预览图片、文本、视频、音频、PDF |
| CORS配置 | 配置存储桶的跨域访问规则 |

### 多选操作

- 点击文件行左侧的复选框选中文件
- 点击表头的复选框全选当前目录所有文件
- 选中文件后可进行批量下载、批量删除

### CORS 配置

点击"CORS配置"按钮可以管理存储桶的跨域访问规则：

1. **查看配置** - 打开弹窗时自动加载当前 CORS 配置
2. **加载示例** - 填入常用的 CORS 配置模板
3. **保存配置** - 保存编辑器中的 JSON 配置
4. **删除配置** - 清除存储桶的 CORS 配置

**CORS 配置 JSON 示例：**
```json
{
    "CORSRules": [
        {
            "AllowedOrigin": ["*"],
            "AllowedMethod": ["PUT", "POST", "DELETE"],
            "AllowedHeader": ["*"]
        },
        {
            "AllowedOrigin": ["*"],
            "AllowedMethod": ["GET", "HEAD"]
        }
    ]
}
```

> **注意**：字段名使用单数形式：`AllowedOrigin`、`AllowedMethod`、`AllowedHeader`

## 安全特性

- 凭证仅保存在浏览器 localStorage
- 服务端不存储任何凭证
- 刷新页面自动恢复连接
- 退出登录清除所有凭证

## 技术栈

- **Go 1.21+** - 运行时
- **Gin** - HTTP 框架
- **minio-go v7** - S3 客户端
- **embed** - 前端资源内嵌
- **slog** - 结构化日志

## 构建

```bash
# 编译
go build -o s3tool .

# 运行测试
go test -v ./...

# 运行测试并查看覆盖率
go test -v -cover -coverprofile=coverage.out .
go tool cover -func=coverage.out

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o s3tool-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build -o s3tool-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build -o s3tool-windows-amd64.exe .
```

## 项目结构

```
s3tool/
├── main.go           # 应用入口
├── frontend.html     # 前端界面（内嵌）
├── internal/
│   ├── api/          # HTTP 处理器和中间件
│   ├── config/       # 配置管理
│   ├── i18n/         # 国际化
│   ├── models/       # 数据模型
│   ├── s3client/     # S3 客户端封装
│   └── utils/        # 工具函数
├── go.mod
├── go.sum
└── README.md
```

## 开源协议

MIT License - 详见 [LICENSE](LICENSE) 文件。
