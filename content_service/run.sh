#!/bin/bash
set -ex
#运行目录
RUN_DIR=$(cd `dirname $0`; pwd)
cd $RUN_DIR

/bin/cp $RUN_DIR/runtime/content_service_dev.json $RUN_DIR/runtime/config.json
# set local ip in config file
ip=$(ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}' | head -n 1)
sed -i "s/LOCAL_IP/$ip/g" $RUN_DIR/runtime/config.json

$RUN_DIR/main --conf=$RUN_DIR/runtime/config.json