#!/bin/bash

set -e  # 发生错误时退出

APP_NAME="cftun"
VERSION="2.1.4"
BUILD_TYPE="release"
BUILD_DIR="build"
PLATFORMS=("linux/amd64" "linux/mipsle" "windows/amd64")

# 创建 build 目录
mkdir -p $BUILD_DIR

# 交叉编译
for PLATFORM in "${PLATFORMS[@]}"; do
    OS=${PLATFORM%%/*}
    ARCH=${PLATFORM##*/}

    EXT=""
    if [ "$OS" == "windows" ]; then
        EXT=".exe"
    fi

    echo "Building server for $OS/$ARCH..."
    OUTPUT_NAME="$APP_NAME-server-$OS-$ARCH$EXT"
    LDFLAGS="-s -w -X main.Version=$VERSION -X main.BuildDate=$(date '+%Y-%m-%d_%H:%M:%S_%Z') -X main.BuildType=$BUILD_TYPE"
    env CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -trimpath -ldflags "$LDFLAGS" -o $BUILD_DIR/$OUTPUT_NAME main_server.go

    echo "Building client for $OS/$ARCH..."
    OUTPUT_NAME="$APP_NAME-client-$OS-$ARCH$EXT"
    env CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -trimpath -ldflags "$LDFLAGS" -o $BUILD_DIR/$OUTPUT_NAME main_client.go
done

echo "Build completed! Files are in the '$BUILD_DIR' directory."
