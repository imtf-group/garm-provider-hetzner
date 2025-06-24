#!/bin/bash

VERSION=${GITHUB_REF_NAME:-latest}
GARM_PROVIDER_NAME=${GARM_PROVIDER_NAME:-garm-provider-hetzner}

RELEASE="$PWD/release"

[ ! -d "$RELEASE" ] && mkdir -p "$RELEASE"

if [ ! -d "build/$VERSION" ]; then
    echo "missing build/$VERSION"
    exit 1
fi

# Windows

if [ ! -d "build/$VERSION/windows/amd64" ];then
    echo "missing build/$VERSION/windows/amd64"
    exit 1
fi

if [ ! -f "build/$VERSION/windows/amd64/$GARM_PROVIDER_NAME.exe" ];then
    echo "missing build/$VERSION/windows/amd64/$GARM_PROVIDER_NAME.exe"
    exit 1
fi

pushd build/$VERSION/windows/amd64
zip $GARM_PROVIDER_NAME-windows-amd64.zip $GARM_PROVIDER_NAME.exe
sha256sum $GARM_PROVIDER_NAME-windows-amd64.zip > $GARM_PROVIDER_NAME-windows-amd64.zip.sha256
mv $GARM_PROVIDER_NAME-windows-amd64.zip $RELEASE
mv $GARM_PROVIDER_NAME-windows-amd64.zip.sha256 $RELEASE
popd

# Linux
OS_ARCHES=("amd64" "arm64")

for arch in ${OS_ARCHES[@]};do
    if [ ! -f "build/$VERSION/linux/$arch/$GARM_PROVIDER_NAME" ];then
        echo "missing build/$VERSION/linux/$arch/$GARM_PROVIDER_NAME"
        exit 1
    fi

    pushd build/$VERSION/linux/$arch
    tar czf $GARM_PROVIDER_NAME-linux-$arch.tgz $GARM_PROVIDER_NAME
    sha256sum $GARM_PROVIDER_NAME-linux-$arch.tgz > $GARM_PROVIDER_NAME-linux-$arch.tgz.sha256
    mv $GARM_PROVIDER_NAME-linux-$arch.tgz $RELEASE
    mv $GARM_PROVIDER_NAME-linux-$arch.tgz.sha256 $RELEASE
    popd
done
