#!/bin/bash
SERVER_ADDRESS=$1
MODEL_NAME=$2
TIMESTAMP=$3
MD5=$4
ISLOCKED=$5
DESC=$6

if [ $# -lt 3 ];then
    echo "usage:sh insert_model_history.sh ip:port model_name timestamp [md5] [is_locked] [desc]"
    exit 1
fi
insert_url="http://$SERVER_ADDRESS/mysql/insert?table=model_histories&model_name=$MODEL_NAME&timestamp=$TIMESTAMP&md5=$MD5&is_locked=$ISLOCKED&desc=$DESC"
http_code=`curl -o /dev/null -m 10 -sw %{http_code} $insert_url`
if [ "$http_code" -eq "000" ];then
	echo "request timeout or server address invalid!"
	exit 1
elif [ "$http_code" -eq "200" ];then
    echo "insert succ!"
else
	echo "invaid input data,please check data format!"
	exit 1
fi
