#!/bin/sh
set -ex
RUN_DIR=$(cd `dirname $0`; pwd -P)
cd $RUN_DIR/init_test_db/
go run build_test_db.go
lnDBFile=$RUN_DIR"/../../content_service/init_test_db/ad_test_db.db"
if [ ! -e $lnDBFile ]; then
    ln -s ad_test_db.db $lnDBFile
fi