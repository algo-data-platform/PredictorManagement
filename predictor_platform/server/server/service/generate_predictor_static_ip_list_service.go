package service

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"server/conf"
	"server/libs/logger"
	"server/server/dao"
	"server/util"
	"time"
)

func GeneratePredictorStaticIpList(conf *conf.Conf) {
	checkTicker := time.NewTicker(time.Second * time.Duration(conf.ServiceStaticList.UpdateInterval))
	for {
		select {
		case <-checkTicker.C:
			service_map, err := dao.GetAllServiceMap()
			if err != nil {
				logger.Errorf("dao.GetAllServiceMap fail, err:%v", err)
				return
			}
			service_ip_list_map := GetServiceIpList(service_map)
			GenerateIpListFileAndPush(service_ip_list_map, conf)
		}
	}
}

func GetServiceIpList(service_map map[string]uint) map[string][]string {
	service_ip_list_map := make(map[string][]string)
	for service_name, service_id := range service_map {
		ipWeights, err := dao.GetIPWeightsBySid(service_id)
		if err != nil {
			return service_ip_list_map
		}
		for _, ipWeight := range ipWeights {
			if ipWeight.LoadWeight < 100 {
				ipWeight.LoadWeight = 100
			}
			service_ip_list_map[service_name] = append(service_ip_list_map[service_name], fmt.Sprintf("thrift,%v:9537,default,%v", ipWeight.HostIp, ipWeight.LoadWeight))
		}
	}
	return service_ip_list_map
}

func GenerateIpListFileAndPush(service_ip_list_map map[string][]string, conf *conf.Conf) {
	if len(service_ip_list_map) == 0 {
		return
	}
	timestamp := fmt.Sprintf(time.Now().Format("20060102_150405"))
	output_dir := filepath.Join(conf.DataDir, conf.ServiceStaticList.SubDataDir, timestamp)
	cmd := exec.Command("/bin/mkdir", "-p", output_dir)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("mkdir failed: err=%v, output=%s", err, stdout)
	}
	predictor_service_list_file := filepath.Join(output_dir, "predictor_service_list")
	var service_list_str string
	for service := range service_ip_list_map {
		service_list_str += service + "\n"
		ret := GenerateIpListFile(service, service_ip_list_map[service], output_dir, conf.ServiceStaticList.AdServerRelayHost)
		if !ret {
			return
		}
	}
	_, err_file := util.WriteFile(predictor_service_list_file, []byte(service_list_str))
	if err_file != nil {
		logger.Errorf("WriteFile %v failed, err:%v", predictor_service_list_file, err)
		return
	}
	rsync_dir := output_dir + "/"
	logger.Infof("rsync_dir:%v", rsync_dir)
	util.RsyncFile(rsync_dir, conf.ServiceStaticList.AdServerRelayHost)
	util.RsyncFile(rsync_dir, filepath.Join(conf.ServiceStaticList.PredictorRelayHost, conf.ServiceStaticList.SubDataDir))
	ClearOldData(conf)
}

func GenerateIpListFile(service_name string, ip_list []string, data_dir string, ad_server_relay_host string) bool {
	if len(ip_list) == 0 {
		logger.Errorf("service:%v ip_list is empty", service_name)
		return false
	}
	file_name := filepath.Join(data_dir, "predictor_"+service_name+"_static_ip_list")
	var ip_list_str string
	for _, ip := range ip_list {
		ip_list_str += ip + "\n"
	}
	_, err := util.WriteFile(file_name, []byte(ip_list_str))
	if err != nil {
		logger.Errorf("WriteFile %v failed, err:%v", file_name, err)
		return false
	}
	return true
}

func ClearOldData(conf *conf.Conf) {
	timestamp_1day_ago := fmt.Sprintf(time.Now().AddDate(0, 0, -1).Format("20060102"))
	files_to_remove := filepath.Join(conf.DataDir, conf.ServiceStaticList.SubDataDir, timestamp_1day_ago, "*")
	cmd := exec.Command("/bin/rm", "-r", files_to_remove)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		logger.Errorf("remove old data failed: err=%v, output=%s", err, stdout)
	}
}
