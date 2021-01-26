#!/bin/bash
set -x
set -e
unset GOPATH
export GOPROXY=https://goproxy.io

# this script is only used for ci purpose, not for deployment
echo "building content_service..."
go build ./main.go
# clean db for ut
find . -name "*test.db" | xargs rm -f

go test ./...
rm ./main
