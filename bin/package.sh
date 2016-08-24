#!/bin/bash

set -e

OS="darwin linux windows"
ARCH="amd64"
VERSION=${TRAVIS_TAG:-latest}
PRODUCT="docker-machine-driver-ovh"
PKG_ROOT="pkg"
BUILD_ROOT="build"

echo "➜ Building version $VERSION of $PRODUCT in $GOPATH"
mkdir -p $PKG_ROOT
mkdir -p $BUILD_ROOT

echo "➜ Getting build dependencies"
go get
go get -u github.com/golang/lint/golint

echo "➜ Ensuring code quality"
pkgs=$(go list ./... | grep -v 'vendor')
go vet $pkgs
golint $pkgs

for GOOS in $OS; do
    for GOARCH in $ARCH; do
        name="${PRODUCT}-${VERSION}-${GOOS}-${GOARCH}"
        archive="${name}.tar.gz"
        checksum="${archive}.md5"
        build_path="${BUILD_ROOT}/${name}"
        location="${build_path}/${PRODUCT}"

        echo "➜ Releasing ${PRODUCT} for ${GOOS}-${GOARCH}"
        echo "⤷ Build"
        export GOOS=$GOOS
        export GOARCH=$GOARCH
        go build -o $location

        echo "⤷ Package"
        tar -czf $PKG_ROOT/$archive -C $build_path $PRODUCT

        echo "⤷ Checksum"
        cd $PKG_ROOT && md5sum $archive > $checksum && cd ..
    done
done

echo "➜ Cleaning"
rm -rf $BUILD_ROOT

echo "Build completed."
