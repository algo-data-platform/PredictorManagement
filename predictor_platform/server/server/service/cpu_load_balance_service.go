package service

import (
	"fmt"
	"math"
	"server/common"
	"server/conf"
	"server/env"
	"server/libs/logger"
	"server/metrics"
	"server/schema"
	"server/server/dao"
	"server/util"
	"time"
)

func UpdateCpuLoad(conf *conf.Conf) {
	checkTicker := time.NewTicker(time.Second * time.Duration(conf.LoadThreshold.CheckInterval))
	for {
		select {
		case <-checkTicker.C:
			checkCpuLoadBalance(conf)
		}
	}
}

func checkCpuLoadBalance(conf *conf.Conf) {
	ip_services, err := GetAllServiceList(conf)
	if err != nil {
		logger.Debugf("GetAllServiceList fail, err: %v\n", err)
		return
	}
	logger.Debugf("GetAllServiceList ip_services success: %v", ip_services)
	ip_service_lists, err := GetIPServiceList(ip_services)
	if err != nil {
		logger.Debugf("GetAllServiceList GetIPServiceList error: %v", err)
		return
	}
	logger.Debugf("checkCpuLoadBalance ip_service_lists success: %v\n", ip_service_lists)
	for _, ip_service_list := range ip_service_lists {
		// 更新数据库
		hs := &schema.HostService{
			Hid:        ip_service_list.Hid,
			Sid:        ip_service_list.Sid,
			LoadWeight: uint(ip_service_list.Load),
		}
		updateErr := updateHostServiceById(ip_service_list.Hid, ip_service_list.Sid, hs)
		if updateErr != nil {
			fmt.Errorf("err: %v", updateErr)
			continue
		}
	}
}

func GetAllServiceList(conf *conf.Conf) ([]util.IP_Service, error) {
	var ip_services []util.IP_Service = []util.IP_Service{}
	// 保存所有机器--cpu利用率的map映射
	var cpuMap map[string]map[string]util.ServiceWeight
	cpuMap = make(map[string]map[string]util.ServiceWeight)
	for _, node_info := range common.GNodeInfos {
		if len(node_info.StatusInfo) <= 0 {
			continue
		}
		logger.Debugf("GetAllServiceList common.GNodeInfos success: %v", len(node_info.StatusInfo))
		var serviceweight util.ServiceWeight
		if node_info.ResourceInfo.CoreNum == 0 {
			logger.Debugf("GetAllServiceList node_info.ResourceInfo.CoreNum is: %v", node_info.ResourceInfo.CoreNum)
			continue
		}
		serviceweight.Cpu_use = node_info.ResourceInfo.Cpu
		if serviceweight.Cpu_use <= 0 {
			continue
		}
		for _, serviceInfo := range node_info.StatusInfo {
			serviceweight.Service_weight = uint(serviceInfo.ServiceWeight)
			if _, exists := cpuMap[serviceInfo.ServiceName]; !exists {
				cpuMap[serviceInfo.ServiceName] = make(map[string]util.ServiceWeight)
			}
			cpuMap[serviceInfo.ServiceName][node_info.Host] = serviceweight
		}
	}

	logger.Debugf("GetAllServiceList cpuMap success: %v", cpuMap)

	// service 下面对应的所有机器的cpu均值和权重值
	for service_name, IPs_to_weight := range cpuMap {
		// 找到service 下所有hosts
		if !util.IsInSliceString(service_name, common.GLoadThresholdServices) {
			logger.Debugf("GetHostsByService %v is not in conf %v will not run\n", service_name, common.GLoadThresholdServices)
			continue
		}

		if len(IPs_to_weight) <= 0 {
			continue
		}

		var cpu_load_all float64
		for _, weights := range IPs_to_weight {
			cpu_load_all = cpu_load_all + weights.Cpu_use
		}

		average_cpu := cpu_load_all / float64(len(IPs_to_weight))
		logger.Debugf("GetHostsByService average_cpu is: %v, service name is %v, cpu_load_all %v\n", average_cpu, service_name, cpu_load_all)

		var weight_all_diff uint
		var count uint
		var single_ip_service []util.IP_Service
		var extra_ip_service []util.IP_Service
		for ip, weights := range IPs_to_weight {
			var load_to_update uint
			if math.Abs(weights.Cpu_use-average_cpu) <= conf.LoadThreshold.CpuLimit {
				// 小于阈值的话，不进行调整
				logger.Debugf("GetHostsByService dont need to adjust load ")
				var ip_service util.IP_Service
				ip_service.IP = ip
				ip_service.Service = service_name
				ip_service.Service_weight = weights.Service_weight
				extra_ip_service = append(extra_ip_service, ip_service)
				continue
			}

			if weights.Cpu_use <= 0 {
				continue
			}

			if conf.LoadThreshold.Method == "once" {
				load_to_update = uint(math.Ceil(average_cpu * float64(weights.Service_weight) / weights.Cpu_use))
			} else {
				if weights.Cpu_use >= average_cpu {
					load_to_update = uint(weights.Service_weight - 1)
				} else {
					load_to_update = uint(weights.Service_weight + 1)
				}
			}

			if load_to_update > 0 {
				if load_to_update >= uint(conf.LoadThreshold.Up_Gap) {
					load_to_update = uint(conf.LoadThreshold.Up_Gap)
				}
				if load_to_update <= uint(conf.LoadThreshold.Down_Gap) {
					weight_all_diff += (uint(conf.LoadThreshold.Down_Gap) - load_to_update)
					count += 1
					load_to_update = uint(conf.LoadThreshold.Down_Gap)
				}

				var ip_service util.IP_Service
				ip_service.IP = ip
				ip_service.Service = service_name
				ip_service.Service_weight = load_to_update
				single_ip_service = append(single_ip_service, ip_service)
				metrics.GetLoadChangeMeter(service_name, ip).Mark(1)
			}
		}
		// 将所有需要向下调节权重的机器的权重等比例加给同service下的其他机器
		logger.Debugf("update infos is service name is %v, weight_all_diff %v, single_ip_service %v, count %v\n", service_name, weight_all_diff, uint(len(single_ip_service)), count)
		if len(single_ip_service) <= 0 {
			logger.Debugf("no ip is need to adjust weight")
			continue
		}
		if len(single_ip_service)+len(extra_ip_service)-int(count) <= 0 {
			logger.Debugf("single_ip_service + extra_ip_service <= count")
			continue
		}
		avg_weight := uint(math.Floor(float64(weight_all_diff / (uint(len(single_ip_service)+len(extra_ip_service)) - count))))
		for index := range single_ip_service {
			if single_ip_service[index].Service_weight > uint(conf.LoadThreshold.Down_Gap) {
				single_ip_service[index].Service_weight += avg_weight
			}
		}
		for index := range extra_ip_service {
			if extra_ip_service[index].Service_weight > uint(conf.LoadThreshold.Down_Gap) {
				extra_ip_service[index].Service_weight += avg_weight
			}
		}
		ip_services = append(ip_services, single_ip_service...)
		ip_services = append(ip_services, extra_ip_service...)
	}
	logger.Debugf("GetAllServiceList need to adjust load : %v", ip_services)
	return ip_services, nil
}

func updateHostServiceById(hid uint, sid uint, hs *schema.HostService) error {
	logger.Debugf("updateHostServiceById start")
	errs := env.Env.MysqlDB.Model(&schema.HostService{}).Where("hid = ? and sid = ?", hid, sid).Update(hs).GetErrors()
	if len(errs) != 0 {
		return fmt.Errorf("updateHostServiceById err: %+v", errs)
	}
	return nil
}

func GetIPServiceList(ip_services []util.IP_Service) ([]util.UpdateInfo, error) {
	var updateInfos []util.UpdateInfo = make([]util.UpdateInfo, 0)
	hosts, err := GetAllHostsMap()
	if err != nil {
		return updateInfos, fmt.Errorf("GetAllHostsMap, err: %+v", err)
	}

	services, err := dao.GetAllServiceMap()
	if err != nil {
		return updateInfos, fmt.Errorf("dao.GetAllServiceMap, err: %+v", err)
	}

	for _, ip_service := range ip_services {
		HID, ok := hosts[ip_service.IP]
		if !ok {
			logger.Debugf("GetIPServiceList has not current IP %v\n", ip_service.IP)
			continue
		}

		SID, ok := services[ip_service.Service]
		if !ok {
			logger.Debugf("GetIPServiceList has not current service %v\n", ip_service.Service)
			continue
		}

		var updateinfo util.UpdateInfo
		updateinfo.Hid = HID
		updateinfo.Sid = SID
		updateinfo.Load = ip_service.Service_weight
		updateInfos = append(updateInfos, updateinfo)
	}
	return updateInfos, nil
}

// 拿到当前service下面所有的host
func GetAllHostsMap() (map[string]uint, error) {
	var hosts_map map[string]uint
	hosts_map = make(map[string]uint)
	hosts, err := dao.GetAllHosts()
	if err != nil {
		return hosts_map, fmt.Errorf("GetAllHosts, err: %+v", err)
	}
	for _, host := range hosts {
		hosts_map[host.Ip] = host.ID
	}
	return hosts_map, nil
}
