set -xe

RootDataDir=/data0/mysql_backup/data
RegressionDataDir=/data0/vad/content_service/data
NowDateTime=$(date "+%Y%m%d_%H%M%S")
OldDate=$(date -d "3 days ago" "+%Y%m%d")
echo $NowDateTime
if [ ! -d ${RootDataDir} ]; then
  /bin/mkdir -p ${RootDataDir}
fi
backup_db_file=${RootDataDir}/algo_service_db.sql.$NowDateTime
mysqldump -h127.0.0.122 -P3306 -uonline_username -pdddddddd --skip-lock-tables algo_service_db > $backup_db_file
mysql -h 127.5.96.155 -P 3306 -u dev_username -paaaaaa -e"drop database if exists algo_service_regression; create database algo_service_regression; use algo_service_regression; source $backup_db_file;"

function clear_old_backup() {
  echo "clear_old_backup"
  /bin/rm -rf ${RootDataDir}/algo_service_db.sql.${OldDate}*
}

clear_old_backup

cd /tmp
if [ "$2" == "regression" ]
then
  wget -q http://127.0.0.100:80/repository/files/deploy/Algo_Service_test/$1.tgz -O /tmp/algo_service_deploy.tgz
elif [ "$2" == "pre_online" ]
then
  wget -q http://127.0.0.100:80/repository/files/deploy/Algo_Service/$1.tgz -O /tmp/algo_service_deploy.tgz
else
  echo "mode is invalid!!!"
  exit 1
fi
tar -zxf /tmp/algo_service_deploy.tgz
sh /tmp/algo_service_deploy/deploy.sh
/bin/rm -rf ${RegressionDataDir}/summary_result*

