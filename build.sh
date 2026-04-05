#!/bin/bash
# S3 Tool 自动打包脚本
# 生成 Linux 和 Windows AMD64 平台的可执行文件

set -e

VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_DIR="dist"
BINARY_NAME="s3tool"

echo "=== S3 Tool 打包脚本 ==="
echo "版本: $VERSION"
echo ""

# 清理旧的构建目录
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# 显示 Go 版本
echo "Go 版本:"
go version
echo ""

# Linux AMD64
echo "构建 Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$VERSION" -o "$BUILD_DIR/${BINARY_NAME}-linux-amd64" .
echo "✓ $BUILD_DIR/${BINARY_NAME}-linux-amd64"

# Windows AMD64
echo "构建 Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$VERSION" -o "$BUILD_DIR/${BINARY_NAME}-windows-amd64.exe" .
echo "✓ $BUILD_DIR/${BINARY_NAME}-windows-amd64.exe"

echo ""
echo "=== 构建完成 ==="
echo ""
ls -lh "$BUILD_DIR"
echo ""

# 计算校验和
echo "计算 SHA256 校验和..."
cd "$BUILD_DIR"
sha256sum * > checksums.sha256
echo "✓ checksums.sha256"
cd ..

echo ""
echo "打包完成！文件位于 $BUILD_DIR 目录"