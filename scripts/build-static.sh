#!/bin/sh

VERSION=${VERSION:-latest}
GARM_PROVIDER_NAME=${GARM_PROVIDER_NAME:-garm-provider-hetzner}
GARM_SOURCE="/build/$GARM_PROVIDER_NAME"
git config --global --add safe.directory "$GARM_SOURCE"
cd $GARM_SOURCE

OUTPUT_DIR="/build/output"
BUILD_DIR="$OUTPUT_DIR/$VERSION"


[ ! -d "$BUILD_DIR/linux" ] && mkdir -p "$BUILD_DIR/linux"
[ ! -d "$BUILD_DIR/windows" ] && mkdir -p "$BUILD_DIR/windows"

export CGO_ENABLED=1
USER_ID=${USER_ID:-$UID}
USER_GROUP=${USER_GROUP:-$(id -g)}

# Linux
GOOS=linux GOARCH=amd64 go build -mod vendor \
    -o $BUILD_DIR/linux/amd64/$GARM_PROVIDER_NAME \
    -tags osusergo,netgo,sqlite_omit_load_extension \
    -ldflags "-extldflags '-static' -s -w -X github.com/imtf-group/garm-provider-hetzner/provider.Version=$VERSION" .
GOOS=linux GOARCH=arm64 CC=aarch64-linux-musl-gcc go build \
    -mod vendor \
    -o $BUILD_DIR/linux/arm64/$GARM_PROVIDER_NAME \
    -tags osusergo,netgo,sqlite_omit_load_extension \
    -ldflags "-extldflags '-static' -s -w -X github.com/imtf-group/garm-provider-hetzner/provider.Version=$VERSION" .

# Windows
GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-cc go build -mod vendor \
    -o $BUILD_DIR/windows/amd64/$GARM_PROVIDER_NAME.exe \
    -tags osusergo,netgo,sqlite_omit_load_extension \
    -ldflags "-s -w -X github.com/imtf-group/garm-provider-hetzner/provider.Version=$VERSION" .

chown $USER_ID:$USER_GROUP -R "$OUTPUT_DIR"
