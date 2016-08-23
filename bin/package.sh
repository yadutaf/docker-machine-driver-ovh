#!/bin/bash

set -e

OS="darwin linux windows"
ARCH="amd64"
VERSION=${TRAVIS_TAG:-latest}
PRODUCT="docker-machine-driver-ovh"
PKG_ROOT="pkg"
BUILD_ROOT="build"

echo "Building version $VERSION of $PRODUCT in $GOPATH"
mkdir -p $PKG_ROOT
mkdir -p $BUILD_ROOT

echo "Getting build dependencies"
go get
go get -u github.com/golang/lint/golint

echo "Ensuring code quality"
pkgs=$(go list ./... | grep -v 'vendor')
go vet $pkgs
golint $pkgs

for GOOS in $OS; do
    for GOARCH in $ARCH; do
        name="${PRODUCT}-${VERSION}-${GOOS}-${GOARCH}"
        archive="${name}.tar.gz"
        build_path="${BUILD_ROOT}/${name}"
        location="${build_path}/${PRODUCT}"

        echo "Building ${name}"
        export GOOS=$GOOS
        export GOARCH=$GOARCH
        go get
        go build -o $location

        echo "Packing ${location}"
        tar -cvzf $PKG_ROOT/$archive -C $build_path $PRODUCT
    done
done

echo "Cleaning"
rm -rf $BUILD_ROOT

echo "Build completed."
