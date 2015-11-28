#!/bin/bash

set -e

OS="darwin linux windows"
ARCH="amd64"

echo "Getting build dependencies"
go get -u github.com/golang/lint/golint

echo "Ensuring code quality"
go vet ./...
golint ./...

for GOOS in $OS; do
    for GOARCH in $ARCH; do
        architecture="${GOOS}-${GOARCH}"
        echo "Building ${architecture}"
        export GOOS=$GOOS
        export GOARCH=$GOARCH
        go get
        go build -o bin/docker-machine-driver-ovh-${architecture}
    done
done
