package mock

import (
	"server/schema"
	"server/util"
	"sort"
	"strconv"
)

func IsEqualSliceMap(
	a map[string][]string,
	b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if vb, exists := b[k]; !exists {
			return false
		} else if !IsEqualSlice(v, vb) {
			return false
		}
	}
	return true
}

func IsEqualSlice(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for idx, v := range a {
		if v != b[idx] {
			return false
		}
	}
	return true
}

func SortIPServiceByIP(ipService []util.IP_Service) {
	sort.SliceStable(ipService, func(i, j int) bool {
		if ipService[i].IP < ipService[j].IP {
			return true
		}
		return false
	})
}

func SortSidHostNumsBySid(sidHostNums []util.SidHostNum) {
	sort.SliceStable(sidHostNums, func(i, j int) bool {
		if sidHostNums[i].Sid < sidHostNums[j].Sid {
			return true
		}
		return false
	})
}

// 判断两个serviceModelSlice是否相同
func IsEqualServiceModelSlice(a []*schema.ServiceModel, b []*schema.ServiceModel) bool {
	SortServiceModelByID(a)
	SortServiceModelByID(b)
	if len(a) != len(b) {
		return false
	}
	for k, serviceModel := range a {
		if !IsStrictEqualServiceModel(serviceModel, b[k]) {
			return false
		}
	}
	return true
}

func IsStrictEqualServiceModel(
	a *schema.ServiceModel,
	b *schema.ServiceModel) bool {
	return a.ID == b.ID && a.Sid == b.Sid && a.Mid == b.Mid
}

func SortServiceModelByID(serviceModels []*schema.ServiceModel) {
	sort.SliceStable(serviceModels, func(i, j int) bool {
		if serviceModels[i].ID < serviceModels[j].ID {
			return true
		}
		return false
	})
}

// 判断两个hostServiceSlice是否相同
func IsEqualHostServiceSlice(a []*schema.HostService, b []*schema.HostService) bool {
	SortHostServiceByID(a)
	SortHostServiceByID(b)
	if len(a) != len(b) {
		return false
	}
	for k, hostService := range a {
		if !IsStrictEqualHostService(hostService, b[k]) {
			return false
		}
	}
	return true
}

func IsStrictEqualHostService(
	a *schema.HostService,
	b *schema.HostService) bool {
	return a.ID == b.ID && a.Sid == b.Sid && a.Hid == b.Hid
}

func SortHostServiceByID(hostServices []*schema.HostService) {
	sort.SliceStable(hostServices, func(i, j int) bool {
		if hostServices[i].ID < hostServices[j].ID {
			return true
		}
		return false
	})
}

// 判断两个stressInfoSlice是否相同
func IsEqualStressInfoSlice(a []*schema.StressInfo, b []*schema.StressInfo) bool {
	SortStressInfoByID(a)
	SortStressInfoByID(b)
	if len(a) != len(b) {
		return false
	}
	for k, stressInfo := range a {
		if !IsStrictEqualStressInfo(stressInfo, b[k]) {
			return false
		}
	}
	return true
}

func IsStrictEqualStressInfo(
	a *schema.StressInfo,
	b *schema.StressInfo) bool {
	return a.ID == b.ID && a.Hid == b.Hid && a.Mids == b.Mids &&
		a.Qps == b.Qps && a.OriginSids == b.OriginSids
}

func SortStressInfoByID(stressInfos []*schema.StressInfo) {
	sort.SliceStable(stressInfos, func(i, j int) bool {
		if stressInfos[i].ID < stressInfos[j].ID {
			return true
		}
		return false
	})
}

func GroupSidsToMap(groupSids []util.GroupSid) map[string]struct{} {
	groupSidMap := make(map[string]struct{}, len(groupSids))
	for _, sids := range groupSids {
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
		groupSidMap[sidsKey] = struct{}{}
	}
	return groupSidMap
}
