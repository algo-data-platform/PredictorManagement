#!/bin/bash
set -x
set -e
unset GOPATH
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

GOPROXY=https://goproxy.io GOOS=linux GOARCH=amd64 go build client_calculate_batch_vector_demo.go
