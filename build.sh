#!/bin/bash

# LLMGATE 构建脚本
# 使用方法: ./build.sh [linux|darwin|windows] [amd64|arm64]
# 默认构建当前平台的可执行文件

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== LLMGATE 构建脚本 ===${NC}"

# 获取版本号
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "版本: $VERSION"
echo "构建时间: $BUILD_TIME"
echo "提交: $COMMIT"
echo ""

# 解析参数
GOOS=${1:-$(go env GOOS)}
GOARCH=${2:-$(go env GOARCH)}

echo -e "${YELLOW}目标平台: $GOOS/$GOARCH${NC}"

# 设置输出文件名
if [ "$GOOS" = "windows" ]; then
    OUTPUT="llmgate.exe"
else
    OUTPUT="llmgate"
fi

# 步骤 1: 构建前端
echo -e "${GREEN}[1/3] 构建前端...${NC}"
cd web

# 检查 node_modules 是否存在
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}安装前端依赖...${NC}"
    npm install
fi

# 构建生产版本
echo "构建生产版本..."
npm run build

cd ..

# 步骤 2: 复制前端构建产物到嵌入目录
echo -e "${GREEN}[2/3] 准备嵌入文件...${NC}"
mkdir -p internal/static/dist
cp -r web/dist/* internal/static/dist/
echo "已复制 $(ls internal/static/dist | wc -l) 个文件到嵌入目录"

# 步骤 3: 构建 Go 可执行文件
echo -e "${GREEN}[3/3] 构建 Go 可执行文件...${NC}"

# 构建参数
LDFLAGS="-s -w \
    -X main.Version=$VERSION \
    -X main.BuildTime=$BUILD_TIME \
    -X main.Commit=$COMMIT"

# 执行构建
GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
    -ldflags "$LDFLAGS" \
    -o $OUTPUT \
    ./cmd/server

echo -e "${GREEN}构建完成!${NC}"
echo ""
echo "输出文件: $OUTPUT"
echo "文件大小: $(du -h $OUTPUT | cut -f1)"
echo ""
echo "使用方法:"
echo "  1. 确保 config.yaml 配置文件存在"
echo "  2. 运行: ./$OUTPUT"
echo ""
echo "配置文件说明:"
echo "  - 复制 config.yaml 到运行目录"
echo "  - 修改数据库路径、JWT密钥等配置"
echo ""

# 可选：创建发布包
if [ "$3" = "release" ]; then
    RELEASE_NAME="llmgate-${VERSION}-${GOOS}-${GOARCH}"
    mkdir -p releases/$RELEASE_NAME
    cp $OUTPUT releases/$RELEASE_NAME/
    cp config.yaml releases/$RELEASE_NAME/config.example.yaml
    cp README.md releases/$RELEASE_NAME/

    # 创建压缩包
    if [ "$GOOS" = "windows" ]; then
        (cd releases && zip -r ${RELEASE_NAME}.zip $RELEASE_NAME)
        echo "发布包: releases/${RELEASE_NAME}.zip"
    else
        (cd releases && tar -czf ${RELEASE_NAME}.tar.gz $RELEASE_NAME)
        echo "发布包: releases/${RELEASE_NAME}.tar.gz"
    fi
fi
