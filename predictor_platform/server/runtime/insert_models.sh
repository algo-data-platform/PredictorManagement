#!/bin/sh

ip=$(ip addr | awk '/^[0-9]+: / {}; /inet.*global/ {print gensub(/(.*)\/(.*)/, "\\1", "g", $2)}' | head -n 1)
SERVER_ADDRESS=$ip":10008"
MODEL_NAME="catboost_direct_v0_demo_model"
TIMESTAMP="20200826_103000"
MD5="10234qwert"
ISLOCKED="0"
DESC="Validated"

/bin/sh ../../script/insert_model_history.sh $SERVER_ADDRESS $TIMESTAMP $MD5 $ISLOCKED $DESC