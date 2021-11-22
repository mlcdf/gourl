#!/bin/bash
set -e
    
VERSION=$(git rev-parse --short HEAD)
OS="linux"
ARCH="amd64"
    
OUTPUT="./dist/gourl-${VERSION}-${OS}-${ARCH}"
    
# https://golang.org/cmd/link/
go build -o $OUTPUT -tags release -ldflags="-X 'main.Version=${VERSION}'"
gzip --force --keep -v $OUTPUT
