#!/bin/bash
set -x
set -e
unset GOPATH
export GOPROXY=https://goproxy.io

echo "building content_service..."
go build ./main.go

