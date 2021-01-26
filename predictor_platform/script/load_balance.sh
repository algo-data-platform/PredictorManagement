#/bin/bash
set -xe
time=$(date "+%Y-%m-%d %H:%M:%S")
echo "shell run on $time"
server_address="127.0.0.123:9508"
# 开启自动权重的service数组
service_arr[0]="algo_service"
service_arr[1]="upfans_service"
service_arr[2]="ocpx_service"
service_arr[3]="cpl_service"

if [ $# -ge 1 ]; then
  case "$1" in
  get)
    # 查看当前开启自动权重的服务列表
    curl "http://$server_address/load_balance/get"
    ;;
  start)
    # 开启自动权重
    for service_name in ${service_arr[@]};
    do
      curl "http://$server_address/load_balance/insert?service_name=$service_name"
    done
    ;;
  reset)
    curl "http://$server_address/load_balance/reset"
    ;;
  *)
    echo $"Usage: $0 {start|get|reset}"
    exit 1
  esac
else
    echo $"Usage: $0 {start|get|reset}"
    exit 1
fi
