#!/bin/bash
set -x
set -e
unset GOPATH
export GOPROXY=https://goproxy.io

# this script is only used for ci purpose, not for deployment
echo "building algo_platform..."
go build ./main.go
go test ./...
# format go code
CURRENT_DIR=$(pwd -P)
find ./ -name "*.go" |xargs gofmt -w
rm ./main
