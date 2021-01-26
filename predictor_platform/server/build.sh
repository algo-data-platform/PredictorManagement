#!/bin/bash
set -ex
unset GOPATH
RUN_DIR=$(cd `dirname $0`; pwd -P)
BUILD_DIR=$RUN_DIR/build

mkdir -p $BUILD_DIR
mkdir -p $BUILD_DIR/frontend

cd $RUN_DIR
git version
go version

export GOPROXY=https://goproxy.io

cd $RUN_DIR/../server
go build main.go

cd $RUN_DIR/../frontend
npm -v
npm config set registry http://registry.npm.taobao.org/
npm install 
npm run build:prod
mkdir -p $RUN_DIR/frontend
/bin/cp -r dist $RUN_DIR/frontend/
