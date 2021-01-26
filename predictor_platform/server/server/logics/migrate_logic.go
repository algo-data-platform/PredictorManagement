package logics

import (
	"fmt"
	"math"
	"server/common"
	"server/env"
	"server/libs/logger"
	"server/schema"
	"server/server/dao"
	"server/util"
	"sort"
	"strconv"
	"strings"
)

func GetServiceStats() ([]util.ServiceStats, error) {
	serviceStatss := []util.ServiceStats{}
	services, err := dao.GetAllServices()
	if err != nil {
		return serviceStatss, err
	}
	allIpSids, err := dao.GetAllIpSid()
	if err != nil {
		return serviceStatss, err
	}
	sidIpsMap := make(map[uint][]string)
	for _, ipSid := range allIpSids {
		sidIpsMap[ipSid.Sid] = append(sidIpsMap[ipSid.Sid], ipSid.Ip)
	}
	for _, service := range services {
		var serviceHostNum int
		var idcHostNums = make([]util.IDCHostNum, 0, len(sidIpsMap))
		var cpuHostNums = make([]util.CpuHostNum, 0, len(sidIpsMap))
		var memHostNums = make([]util.MemHostNum, 0, len(sidIpsMap))
		if _, exists := sidIpsMap[service.ID]; exists {
			serviceHostNum = len(sidIpsMap[service.ID])
			// 按照idc统计
			idcHostNums = getIDCHostNums(sidIpsMap[service.ID])
			cpuHostNums = getCpuHostNums(sidIpsMap[service.ID])
			memHostNums = getMemHostNums(sidIpsMap[service.ID])
		}

		serviceStatss = append(serviceStatss, util.ServiceStats{
			Sid:         service.ID,
			ServiceName: service.Name,
			HostNum:     serviceHostNum,
			IDCHostNums: idcHostNums,
			CpuHostNums: cpuHostNums,
			MemHostNums: memHostNums,
		})
	}
	return serviceStatss, nil
}

func GetFromServices() ([]util.SidName, error) {
	var sidNames = []util.SidName{}
	services, err := dao.GetAllServices()
	if err != nil {
		return sidNames, err
	}
	groupSidHostMap, err := getGroupSidHostMap()
	if err != nil {
		return sidNames, err
	}
	serviceMap := make(map[uint]schema.Service, len(services))
	for _, service := range services {
		serviceMap[service.ID] = service
	}
	for sidKey, ips := range groupSidHostMap {
		if sidKey == "" {
			continue
		}
		var serviceNames []string
		var sids []uint
		sidStrs := strings.Split(sidKey, "_")
		for _, sidStr := range sidStrs {
			sidInt, err := strconv.Atoi(sidStr)
			if err != nil {
				return sidNames, err
			}
			sids = append(sids, uint(sidInt))
			if s, exists := serviceMap[uint(sidInt)]; exists {
				serviceNames = append(serviceNames, s.Name)
			}
		}
		if len(serviceNames) == 1 {
			if isExcludeService(serviceNames[0]) {
				continue
			}
		}
		sidNames = append(sidNames, util.SidName{
			Sids:    sids,
			Names:   serviceNames,
			HostNum: len(ips),
		})
	}
	for _, service := range services {
		if isExcludeService(service.Name) {
			continue
		}
		sidKey := strconv.Itoa(int(service.ID))
		if _, exists := groupSidHostMap[sidKey]; !exists {
			sidNames = append(sidNames, util.SidName{
				Sids:    []uint{service.ID},
				Names:   []string{service.Name},
				HostNum: 0,
			})
		}
	}
	return sidNames, nil
}

func isExcludeService(serviceName string) bool {
	var excludeServicePrefixs = []string{
		env.Env.Conf.StressTestService,
		env.Env.Conf.ConsistenceService,
	}
	for _, prefix := range excludeServicePrefixs {
		if strings.Contains(serviceName, prefix) {
			return true
		}
	}
	return false
}

func GetToServices() ([]util.SidName, error) {
	var sidNames = []util.SidName{}
	services, err := dao.GetAllServices()
	if err != nil {
		return sidNames, err
	}
	for _, service := range services {
		if isExcludeService(service.Name) {
			continue
		}
		sidNames = append(sidNames, util.SidName{
			Sids:  []uint{service.ID},
			Names: []string{service.Name},
		})
	}
	serviceMap := make(map[uint]schema.Service, len(services))
	for _, service := range services {
		serviceMap[service.ID] = service
	}
	groupSids, err := getGroupServices()
	if err != nil {
		return sidNames, err
	}
	for _, groupSid := range groupSids {
		serviceNames := make([]string, 0, len(groupSid))
		for _, sid := range groupSid {
			if s, exists := serviceMap[sid]; exists {
				serviceNames = append(serviceNames, s.Name)
			}
		}
		sidNames = append(sidNames, util.SidName{
			Sids:  groupSid,
			Names: serviceNames,
		})
	}
	return sidNames, nil
}

// 获取分组机器列表
// @return map[sidKey]iplist
// eg. map[2_3_4][]{'10.133.0.1','10.133.0.2'}
func getGroupSidHostMap() (map[string][]string, error) {
	groupSidHostsMap := make(map[string][]string, 0)
	allIpSids, err := dao.GetAllIpSid()
	if err != nil {
		return groupSidHostsMap, err
	}
	var hostSidsMap = make(map[string][]int)
	for _, ipSid := range allIpSids {
		if _, exists := hostSidsMap[ipSid.Ip]; exists {
			hostSidsMap[ipSid.Ip] = append(hostSidsMap[ipSid.Ip], int(ipSid.Sid))
		} else {
			hostSidsMap[ipSid.Ip] = []int{int(ipSid.Sid)}
		}
	}

	for ip, sids := range hostSidsMap {
		sidsKey := ""
		if len(sids) > 1 {
			sort.Ints(sids)
			for idx, sid := range sids {
				sidsKey = sidsKey + strconv.Itoa(sid)
				if idx != len(sids)-1 {
					sidsKey = sidsKey + "_"
				}
			}
		} else {
			sidsKey = strconv.Itoa(sids[0])
		}
		groupSidHostsMap[sidsKey] = append(groupSidHostsMap[sidsKey], ip)
	}
	return groupSidHostsMap, nil
}

// 获取关联组的service
// @return []GroupSids
func getGroupServices() ([]util.GroupSid, error) {
	groupSids := []util.GroupSid{}
	allHostServices, err := dao.GetAllHostServices()
	if err != nil {
		return groupSids, err
	}
	var hostSidsMap = make(map[uint][]uint)
	for _, hs := range allHostServices {
		if _, exists := hostSidsMap[hs.Hid]; exists {
			hostSidsMap[hs.Hid] = append(hostSidsMap[hs.Hid], hs.Sid)
		} else {
			hostSidsMap[hs.Hid] = []uint{hs.Sid}
		}
	}
	var sidsCheckMap = make(map[string]struct{})
	for _, sids := range hostSidsMap {
		if len(sids) > 1 {
			sidInts := make([]int, 0, len(sids))
			for _, sid := range sids {
				sidInts = append(sidInts, int(sid))
			}
			sort.Ints(sidInts)
			sidsKey := ""
			for idx, sidInt := range sidInts {
				sidsKey = sidsKey + strconv.Itoa(sidInt)
				if idx != len(sids)-1 {
					sidsKey = sidsKey + "_"
				}
			}
			if _, exixts := sidsCheckMap[sidsKey]; !exixts {
				sidsCheckMap[sidsKey] = struct{}{}
				groupSids = append(groupSids, sids)
			}
		}
	}
	return groupSids, nil
}

// 预览要迁移的机器数据
func PreviewMigrateHosts(fromSids []uint, toSids []uint, num int) ([]util.PreviewHost, error) {
	previewHosts := []util.PreviewHost{}
	if len(toSids) == 0 {
		return previewHosts, fmt.Errorf("the length of toSids is zero")
	}
	// 获取要迁移的机器列表
	toMigratedHosts, err := GetMigrateHosts(fromSids, toSids, num)

	if err != nil {
		return previewHosts, err
	}
	if len(toMigratedHosts) == 0 {
		return previewHosts, nil
	}
	// 获取机器配置数据
	previewHosts, err = getPreviewHosts(toMigratedHosts)
	if err != nil {
		return previewHosts, err
	}
	return previewHosts, nil
}

// 组装机器展示信息
func getPreviewHosts(toMigratedHosts []string) ([]util.PreviewHost, error) {
	previewHosts := []util.PreviewHost{}
	hostServiceInfos, err := dao.GetHostServiceInfoByIps(toMigratedHosts)

	if err != nil {
		return previewHosts, err
	}
	serviceIdMap, err := GetAllServiceIdMap()
	if err != nil {
		return previewHosts, err
	}
	var hostPosMap = make(map[string]int)
	// 聚合host
	for _, hostServiceInfo := range hostServiceInfos {
		serviceName := ""
		if _, exists := serviceIdMap[hostServiceInfo.Sid]; exists {
			serviceName = serviceIdMap[hostServiceInfo.Sid]
		}
		if pos, exists := hostPosMap[hostServiceInfo.Ip]; exists {
			previewHosts[pos].Sids = append(previewHosts[pos].Sids, hostServiceInfo.Sid)
			previewHosts[pos].ServiceNames = append(previewHosts[pos].ServiceNames, serviceName)
			continue
		} else {
			hostPosMap[hostServiceInfo.Ip] = len(previewHosts)
		}
		resourceInfo := util.NodeResourceInfo{}
		if _, exists := NodeResMap[hostServiceInfo.Ip]; exists {
			resourceInfo = NodeResMap[hostServiceInfo.Ip]
		}

		previewHost := util.PreviewHost{
			Hsid:         hostServiceInfo.Hsid,
			Hid:          hostServiceInfo.Hid,
			Ip:           hostServiceInfo.Ip,
			Sids:         []uint{hostServiceInfo.Sid},
			ServiceNames: []string{serviceName},
			ResourceInfo: resourceInfo,
			IDC:          GetIDCByIp(hostServiceInfo.Ip),
		}
		previewHosts = append(previewHosts, previewHost)
	}
	return previewHosts, nil
}

// 迁移
func DoMigrateHosts(fromSids []uint, toSids []uint, toMigrateHids []uint) error {
	toWeights := []uint{}
	for _, toSid := range toSids {
		// 获取权重
		initWeight, err := GetServiceInitWeightBySid(toSid, "")
		if err != nil {
			logger.Errorf("GetServiceInitWeightBySid fail, err: %v", err)
			return err
		}
		toWeights = append(toWeights, initWeight)
	}
	err := dao.TranscatMigrateHost(fromSids, toSids, toWeights, toMigrateHids)
	if err != nil {
		logger.Errorf("TranscatMigrateHost fail, err: %v", err)
		return err
	}
	return nil
}

// 获取服务id->name 映射
// @return map[sid]service_name
func GetAllServiceIdMap() (map[uint]string, error) {
	serviceIdMap := make(map[uint]string)
	services, err := dao.GetAllServices()
	if err != nil {
		return serviceIdMap, err
	}
	for _, service := range services {
		serviceIdMap[service.ID] = service.Name
	}
	return serviceIdMap, nil
}

// 获取可迁移的机器列表
// @return ipList
func GetMigrateHosts(fromSids []uint, toSids []uint, num int) ([]string, error) {
	var toMigratedIps = []string{}
	groupSidHostMap, err := getGroupSidHostMap()
	if err != nil {
		return toMigratedIps, err
	}
	// 获取fromService机器列表
	fromIps := []string{}
	fromSidKey := util.JoinUint(fromSids, "_")
	if _, exists := groupSidHostMap[fromSidKey]; exists {
		fromIps = groupSidHostMap[fromSidKey]
	}
	// 排除特殊机器
	fromIps = util.ExcludeSliceString(fromIps, env.Env.Conf.MigrateHosts.ExcludeHosts)
	if len(fromIps) == 0 {
		return toMigratedIps, nil
	}

	toIDCIpsMap := make(map[string][]string)
	// 获取toService机器列表
	toIps := []string{}
	toSidKey := util.JoinUint(toSids, "_")
	if _, exists := groupSidHostMap[toSidKey]; exists {
		toIps = groupSidHostMap[toSidKey]
	}
	if len(toIps) != 0 {
		// 排除toService 机器
		fromIps = util.ExcludeSliceString(fromIps, toIps)
		if len(fromIps) == 0 {
			return toMigratedIps, nil
		}
		// toService按照机房划分
		toIDCIpsMap = divideIpsByIDC(toIps)
	}
	// fromService按照机房划分
	fromIDCIpsMap := divideIpsByIDC(fromIps)
	// 根据toService机房比例，计算获取机器数量，如果机器数量不够，从其他机房平均获取
	toMigratedIDCIpsMap := getToMigratedIpsByIDC(fromIDCIpsMap, toIDCIpsMap, num)
	for _, ips := range toMigratedIDCIpsMap {
		toMigratedIps = append(toMigratedIps, ips...)
	}
	return toMigratedIps, nil
}

// 按照idc机器比例获取迁移机器列表
func getToMigratedIpsByIDC(fromIDCIpsMap map[string][]string, toIDCIpsMap map[string][]string, num int) map[string][]string {
	toMigratedIpsMap := make(map[string][]string)
	if num <= 0 {
		return toMigratedIpsMap
	}
	var fromIpsTotal int
	for _, ips := range fromIDCIpsMap {
		fromIpsTotal = fromIpsTotal + len(ips)
	}
	// 如果fromService机器不够，全部返回
	if fromIpsTotal <= num {
		return fromIDCIpsMap
	}
	// 计算按照toService IDC比例得到的机器数量
	var toIpsTotal int
	// 按照toservice 比例剩余数量
	lastNum := num
	for _, ips := range toIDCIpsMap {
		toIpsTotal = toIpsTotal + len(ips)
	}

	// 剩余未找到机房的机器数量
	var lastMigratedNum int
	// 判断toIpsTotal == 0
	if toIpsTotal > 0 {
		toMigratedNumMap := make(map[string]int)
		for idc, ips := range toIDCIpsMap {
			if lastNum == 0 {
				break
			}
			toMigrateNum := int(math.Ceil(float64(num) * float64(len(ips)) / float64(toIpsTotal)))
			toMigrateNum = util.MinInt(toMigrateNum, lastNum)
			toMigratedNumMap[idc] = toMigrateNum
			lastNum -= toMigrateNum
			if _, exists := fromIDCIpsMap[idc]; exists {
				if len(fromIDCIpsMap[idc]) >= toMigrateNum {
					toMigratedIpsMap[idc] = fromIDCIpsMap[idc][:toMigrateNum]
					fromIDCIpsMap[idc] = fromIDCIpsMap[idc][toMigrateNum:]
				} else {
					toMigratedIpsMap[idc] = fromIDCIpsMap[idc][:]
					fromIDCIpsMap[idc] = []string{}
					lastMigratedNum += (toMigrateNum - len(toMigratedIpsMap[idc]))
				}
			} else {
				lastMigratedNum += toMigrateNum
			}
		}
		if lastMigratedNum == 0 {
			return toMigratedIpsMap
		}
	} else {
		lastMigratedNum = lastNum
	}

	// 剩余未分配完的机器根据比例从其他idc获取
	var lastFromIpsTotal int
	var fromLastNum = lastMigratedNum
	for _, ips := range fromIDCIpsMap {
		lastFromIpsTotal = lastFromIpsTotal + len(ips)
	}
	for idc, ips := range fromIDCIpsMap {
		if fromLastNum == 0 {
			break
		}
		toMigrateNum := int(math.Ceil(float64(lastMigratedNum) * float64(len(ips)) / float64(lastFromIpsTotal)))
		toMigrateNum = util.MinInt(toMigrateNum, fromLastNum)
		fromLastNum = fromLastNum - toMigrateNum
		toMigratedIpsMap[idc] = append(toMigratedIpsMap[idc], fromIDCIpsMap[idc][:toMigrateNum]...)
	}
	return toMigratedIpsMap
}

// 按照机房划分ip
func divideIpsByIDC(ips []string) map[string][]string {
	IDCIpsMap := make(map[string][]string)
	for _, ip := range ips {
		idc := GetIDCByIp(ip)
		if idc == "" {
			continue
		}
		IDCIpsMap[idc] = append(IDCIpsMap[idc], ip)
	}
	return IDCIpsMap
}

func GetIDCByIp(ip string) string {
	idc := ""
	ipFields := strings.Split(ip, ".")
	if len(ipFields) != 4 {
		return idc
	}
	ipABField := ipFields[0] + "." + ipFields[1]
	if idc, exists := common.GIpToIDCMap[ipABField]; exists {
		return idc
	}
	return "unknown"
}

// 按照机房统计数量
func getIDCHostNums(ips []string) []util.IDCHostNum {
	IDCHostNumMap := make(map[string]int)
	for _, ip := range ips {
		idc := GetIDCByIp(ip)
		if idc == "" {
			continue
		}
		IDCHostNumMap[idc] += 1
	}
	idcHostNums := make([]util.IDCHostNum, 0, len(IDCHostNumMap))
	for idc, num := range IDCHostNumMap {
		idcHostNums = append(idcHostNums, util.IDCHostNum{
			IDC:     idc,
			HostNum: num,
		})
	}
	sort.SliceStable(idcHostNums, func(i, j int) bool {
		if idcHostNums[i].IDC < idcHostNums[j].IDC {
			return true
		}
		return false
	})
	return idcHostNums
}

// 按照cpu 核数统计
func getCpuHostNums(ips []string) []util.CpuHostNum {
	cpuHostNumMap := make(map[int]int)
	var resourceInfo = util.NodeResourceInfo{}
	for _, ip := range ips {
		if _, exists := NodeResMap[ip]; exists {
			resourceInfo = NodeResMap[ip]
		} else {
			resourceInfo = util.NodeResourceInfo{}
		}
		cpuHostNumMap[resourceInfo.CoreNum] += 1
	}
	cpuHostNums := make([]util.CpuHostNum, 0, len(cpuHostNumMap))
	for coreNum, num := range cpuHostNumMap {
		cpuHostNums = append(cpuHostNums, util.CpuHostNum{
			CoreNum: coreNum,
			HostNum: num,
		})
	}
	sort.SliceStable(cpuHostNums, func(i, j int) bool {
		if cpuHostNums[i].CoreNum < cpuHostNums[j].CoreNum {
			return true
		}
		return false
	})
	return cpuHostNums
}

// 按照cpu 核数统计
func getMemHostNums(ips []string) []util.MemHostNum {
	memHostNumMap := make(map[int64]int)
	var resourceInfo = util.NodeResourceInfo{}
	for _, ip := range ips {
		if _, exists := NodeResMap[ip]; exists {
			resourceInfo = NodeResMap[ip]
		} else {
			resourceInfo = util.NodeResourceInfo{}
		}
		memHostNumMap[resourceInfo.TotalMem] += 1
	}
	memHostNums := make([]util.MemHostNum, 0, len(memHostNumMap))
	for memNum, num := range memHostNumMap {
		memHostNums = append(memHostNums, util.MemHostNum{
			TotalMem: int(memNum),
			HostNum:  num,
		})
	}
	sort.SliceStable(memHostNums, func(i, j int) bool {
		if memHostNums[i].TotalMem < memHostNums[j].TotalMem {
			return true
		}
		return false
	})
	return memHostNums
}
