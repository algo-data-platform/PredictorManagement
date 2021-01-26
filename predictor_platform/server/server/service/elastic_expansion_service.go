package service

import (
	"fmt"
	"math"
	"server/conf"
	"server/libs/logger"
	"server/metrics"
	"server/schema"
	"server/server/dao"
	"server/server/logics"
	"server/util"
	"time"
)

func CheckElasticExpansion(conf *conf.Conf) {
	checkTicker := time.NewTicker(time.Second * time.Duration(conf.ElasticExpansion.CheckInterval))
	for {
		select {
		case <-checkTicker.C:
			elasticExpansion(conf)
		}
	}
}

// 按照比例分配机器
func elasticExpansion(conf *conf.Conf) {
	logger.Debugf("start elasticExpansion")
	// 从数据库获取待分配的扩容机器列表
	toAllocateHosts, err := dao.GetHostsToAllocate()
	if err != nil {
		logger.Errorf("GetHostsToAllocate fail, err : %v", err)
		metrics.GetMeters()[metrics.TAG_GET_HOSTS_TO_ALLOCATE_ERROR].Mark(1)
		return
	}
	if len(toAllocateHosts) == 0 {
		logger.Infof("no hosts to be allocated")
		return
	}
	// 从数据库获取已经分配的service对应的扩容机器数量
	allocatedSidHostNum, err := dao.GetAllocatedHostService()
	if err != nil {
		logger.Errorf("allocatedSidHostNum fail, err : %v", err)
		metrics.GetMeters()[metrics.TAG_GET_ALLOCATED_HOST_SERVICE_ERROR].Mark(1)
		return
	}
	// 获取service跟sid映射
	serviceMap, err := dao.GetAllServiceMap()
	if err != nil {
		logger.Errorf("GetAllServiceMap fail, err: %v", err)
		metrics.GetMeters()[metrics.TAG_GET_All_SERVICE_MAP_ERROR].Mark(1)
		return
	}
	// 获取扩容配置
	allocateConfig, sidGroupMap, err := getAllocateConfig(conf, serviceMap, allocatedSidHostNum)
	if err != nil {
		logger.Errorf("getAllocateConfig fail, err: %v", err)
		metrics.GetMeters()[metrics.TAG_GET_ALLOCATE_CONFIG_ERROR].Mark(1)
		return
	}
	if len(sidGroupMap) == 0 || len(allocateConfig) == 0 {
		logger.Infof("len(sidGroupMap) is %v, len(allocateConfig) is %v, no allocate service", len(sidGroupMap) == 0, len(allocateConfig))
		return
	}

	// 获取要扩容的service及机器数量
	toAlloateMap := getToAllocateMap(len(toAllocateHosts), allocateConfig, allocatedSidHostNum)
	logger.Debugf("totalNum: %d,allocateConfig: %+v, allocatedSidHostNum: %+v", len(toAllocateHosts), allocateConfig, allocatedSidHostNum)
	logger.Debugf("toAlloateMap : %+v", toAlloateMap)
	// 扩容
	err = alloateHostToService(toAllocateHosts, toAlloateMap, serviceMap, sidGroupMap)
	if err != nil {
		logger.Errorf("alloateHostToService fail, err: %v", err)
		metrics.GetMeters()[metrics.TAG_ALLOCATE_HOST_TO_SERVICE_ERROR].Mark(1)
	}
	metrics.GetMeters()[metrics.TAG_ELASTIC_EXPANSION_ALLOCATE].Mark(1)
	logger.Infof("elasticExpansion finish")
}

// 获取分配比例配置
func getAllocateConfig(conf *conf.Conf, serviceMap map[string]uint, allocatedSidHostNum []util.SidHostNum) ([]util.SidHostNum, map[uint][]uint, error) {
	allocateConfig := []util.SidHostNum{}
	allocateSids := []uint{}
	sidGroupMap := make(map[uint][]uint)
	for _, serviceGroup := range conf.ElasticExpansion.ServiceGroups {
		if len(serviceGroup.Services) == 0 {
			continue
		}
		var sids []uint
		for _, service_name := range serviceGroup.Services {
			if sid, ok := serviceMap[service_name]; ok {
				sids = append(sids, sid)
			}
		}
		if len(sids) == 0 {
			logger.Warnf("services config is wrong,services: %v", serviceGroup.Services)
			continue
		}
		allocateSids = append(allocateSids, sids[0])
		sidGroupMap[sids[0]] = sids
	}
	// 获取未分配前原始机器比例
	originHostNumMap, err := dao.GetOriginHostNumMap(allocateSids)
	if err != nil {
		return allocateConfig, sidGroupMap, err
	}
	for _, sid := range allocateSids {
		var hostNum int
		if _, exists := originHostNumMap[sid]; exists {
			hostNum = originHostNumMap[sid]
		}
		configRow := util.SidHostNum{
			Sid:     sid,
			HostNum: hostNum,
		}
		allocateConfig = append(allocateConfig, configRow)
	}
	return allocateConfig, sidGroupMap, nil
}

// 分配机器到对应的service
func alloateHostToService(toAllocateHosts []schema.Host, toAlloateMap map[uint]int, serviceMap map[string]uint, sidGroupMap map[uint][]uint) error {
	sidServiceMap := make(map[uint]string)
	for service_name, sid := range serviceMap {
		sidServiceMap[sid] = service_name
	}
	var low, high int
	for sid, hostNum := range toAlloateMap {

		high = high + hostNum
		if high > len(toAllocateHosts) {
			high = len(toAllocateHosts)
		}
		if low > len(toAllocateHosts) {
			low = len(toAllocateHosts)
		}
		serviceAlloateHosts := toAllocateHosts[low:high]
		low = high
		if len(serviceAlloateHosts) > 0 {
			// 判断是否存在
			if _, exists := sidGroupMap[sid]; !exists {
				continue
			}
			for _, childSid := range sidGroupMap[sid] {
				if _, exists := sidServiceMap[sid]; !exists {
					continue
				}
				service_name := sidServiceMap[childSid]
				for _, host := range serviceAlloateHosts {
					initWeight, _ := logics.GetServiceInitWeightBySid(sid, host.Ip)
					hostService := schema.HostService{
						Hid:        host.ID,
						Sid:        childSid,
						LoadWeight: initWeight,
						Desc:       host.Ip + " -> " + service_name,
					}
					err := dao.InsertHostService(hostService)
					if err != nil {
						return fmt.Errorf("InsertHostService fail, err: %v, ip: %s", err, host.Ip)
					}
				}
			}

		}

	}
	return nil
}

// 获取要分配的机器数量
// @return map[sid]host_num
func getToAllocateMap(totalNum int, allocateConfig []util.SidHostNum, allocatedSidHostNum []util.SidHostNum) map[uint]int {
	toAllocateMap := make(map[uint]int)
	allocatedSidHostNumMap := make(map[uint]int)
	if totalNum == 0 {
		return toAllocateMap
	}
	if len(allocatedSidHostNum) > 0 {
		for _, sidNum := range allocatedSidHostNum {
			allocatedSidHostNumMap[sidNum.Sid] = sidNum.HostNum
		}
	}
	// get origin total
	var originTotal int      // 原始未分配机器总数
	var toAllocatedTotal int // 分配机器汇总 = 已分配总数 + 待分配数
	toAllocatedTotal = totalNum
	for _, configRow := range allocateConfig {
		sid := configRow.Sid
		num := configRow.HostNum
		var allocatedNum int
		if _, exists := allocatedSidHostNumMap[sid]; exists {
			allocatedNum = allocatedSidHostNumMap[sid]
		}
		originTotal = originTotal + num
		toAllocatedTotal = toAllocatedTotal + allocatedNum
	}
	var lastToAllocatedNum = totalNum
	if originTotal == 0 {
		// 总数为0，按照1:1分配
		for index, _ := range allocateConfig {
			allocateConfig[index].HostNum = 1
			originTotal = originTotal + 1
		}
	}
	// 按照原机器比例分配
	for _, configRow := range allocateConfig {
		sid := configRow.Sid
		num := configRow.HostNum
		if lastToAllocatedNum <= 0 {
			break
		}
		var allocatedNum int
		if _, exists := allocatedSidHostNumMap[sid]; exists {
			allocatedNum = allocatedSidHostNumMap[sid]
		}

		toAllocatedNum := int(math.Ceil(float64(toAllocatedTotal)*float64(num)/float64(originTotal) - float64(allocatedNum)))
		toAllocatedNum = util.MaxInt(toAllocatedNum, 0)
		toAllocatedNum = util.MinInt(lastToAllocatedNum, toAllocatedNum)
		toAllocateMap[sid] = toAllocatedNum
		lastToAllocatedNum = lastToAllocatedNum - toAllocatedNum
	}
	// 分配完成，剩余分给配置里面第一个service
	if lastToAllocatedNum > 0 {
		toAllocateMap[allocateConfig[0].Sid] = toAllocateMap[allocateConfig[0].Sid] + lastToAllocatedNum
	}

	return toAllocateMap
}
