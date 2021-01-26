package logics

import (
	"fmt"
	"math"
	"server/common"
	"server/libs/logger"
	"server/server/dao"
)

// 获取某个service 初始化权重
func GetServiceInitWeightBySid(sid uint, add_ip string) (uint, error) {
	var initWeight uint = 0
	// 根据sid 获取service_name
	service, err := dao.GetServiceBySid(sid)
	if err != nil {
		logger.Errorf("GetServiceBySid fail, sid: %d, err: %v", sid, err)
		return initWeight, err
	}
	// 根据node_info 获取
	if len(common.GNodeInfos) > 0 {
		if add_ip != "" {
			addIpCoreNum := GetAddIpCoreNum(add_ip)
			if addIpCoreNum > 0 {
				initWeight = GetInitWeightByCoreNum(service.Name, addIpCoreNum)
			}
			if initWeight > 0 {
				return initWeight, nil
			}
		}
		if initWeight = GetInitWeightFromNodeInfo(service.Name); initWeight > 0 {
			return initWeight, nil
		}
	}
	// 根据数据库值获取
	initWeight, err = GetInitWeightFromDB(sid)
	logger.Errorf("GetInitWeightFromDB fail, sid: %d, err: %v", sid, err)
	return initWeight, err
}

// 根据node_info获取初始化权重
func GetInitWeightFromNodeInfo(serviceName string) uint {
	var totalNum int = 0
	var totalWeight uint = 0
	for _, node_info := range common.GNodeInfos {
		if len(node_info.StatusInfo) <= 0 {
			continue
		}
		for _, serviceInfo := range node_info.StatusInfo {
			if serviceName != serviceInfo.ServiceName {
				continue
			}
			weight := uint(serviceInfo.ServiceWeight)
			if weight != 0 {
				totalWeight = totalWeight + weight
				totalNum = totalNum + 1
			}
		}
	}
	if totalNum == 0 {
		return 0
	}
	var initWeight uint = uint(math.Ceil(float64(totalWeight) / float64(totalNum)))
	return initWeight

}

// 根据核数计算初始权重
func GetInitWeightByCoreNum(serviceName string, addIpCoreNum int) uint {
	var totalCoreNum int = 0
	var totalWeight uint = 0
	for _, node_info := range common.GNodeInfos {
		if len(node_info.StatusInfo) <= 0 {
			continue
		}
		for _, serviceInfo := range node_info.StatusInfo {
			if serviceName != serviceInfo.ServiceName {
				continue
			}
			weight := uint(serviceInfo.ServiceWeight)
			if weight != 0 {
				if nodeInfo, exists := NodeResMap[node_info.Host]; exists {
					if nodeInfo.CoreNum != 0 {
						totalWeight += weight
						totalCoreNum += nodeInfo.CoreNum
					}
				}
			}
		}
	}
	if totalCoreNum == 0 || totalWeight == 0 {
		return 0
	}
	var initWeight uint = uint(math.Ceil(float64(totalWeight) / float64(totalCoreNum) * float64(addIpCoreNum)))
	return initWeight
}

// 获取单个机器的核数
func GetAddIpCoreNum(add_ip string) int {
	if nodeInfo, exists := NodeResMap[add_ip]; exists {
		return nodeInfo.CoreNum
	} else {
		// 扩容机器nodeInfo还没获取，获取一次
		cur_node_metric := fmt.Sprintf("http://%s:9100/metrics", add_ip)
		resourceInfo := GetResourceInfo(cur_node_metric, 2000, add_ip)
		return resourceInfo.CoreNum
	}
}

// 根据数据库获取初始化权重
func GetInitWeightFromDB(sid uint) (uint, error) {
	hostServices, err := dao.GetHostServiceBySid(sid)
	if err != nil {
		return 0, err
	}
	var totalNum int = 0
	var totalWeight uint = 0
	for _, hostService := range hostServices {
		if hostService.LoadWeight != 0 {
			totalWeight = totalWeight + hostService.LoadWeight
			totalNum = totalNum + 1
		}
	}
	if totalNum == 0 {
		return 0, fmt.Errorf("totalNum is zero")
	}
	var initWeight uint = uint(math.Ceil(float64(totalWeight) / float64(totalNum)))
	return initWeight, nil
}
