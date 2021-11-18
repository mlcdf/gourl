#!/bin/bash
set -e
    
VERSION=$(git describe --tags)
OS="linux"
ARCH="amd64"
    
OUTPUT="./dist/gourl-${VERSION}-${OS}-${ARCH}"
    
# https://golang.org/cmd/link/
go build -o $OUTPUT -tags release -ldflags "-X 'main.Version=${VERSION}'" -ldflags "-w" -ldflags "-s" ./cmd/gourl/
gzip --force --keep -v $OUTPUT
