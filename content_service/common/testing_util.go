package common

import (
	"content_service/api"
	"sort"
)

// 对payload 结构体里的slice 排序,在testing里对象比较前调用
func SortPayload(payload *api.PredictorPayload) {
	services := payload.Services
	sort.SliceStable(services, func(i, j int) bool {
		if services[i].ServiceName < services[j].ServiceName {
			return true
		}
		return false
	})
	serviceLen := len(payload.Services)
	for i := 0; i < serviceLen; i++ {
		records := payload.Services[i].ModelRecords
		sort.SliceStable(records, func(i, j int) bool {
			if records[i].Name < records[j].Name {
				return true
			}
			return false
		})
	}
}
