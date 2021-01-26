package mock

import (
	"content_service/schema"
	"sort"
)

func IsEqualModelHistory(
	a schema.ModelHistory,
	b schema.ModelHistory) bool {
	return a.ModelName == b.ModelName && a.Timestamp == b.Timestamp && a.IsLocked == b.IsLocked
}

func IsEqualModelHistoryMap(
	a map[string]schema.ModelHistory,
	b map[string]schema.ModelHistory) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if vb, exists := b[k]; !exists {
			return false
		} else if !IsEqualModelHistory(v, vb) {
			return false
		}
	}
	return true
}

func IsEqualServiceModelMap(
	a map[string]map[string]schema.ModelHistory,
	b map[string]map[string]schema.ModelHistory) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if vb, exists := b[k]; !exists {
			return false
		} else if !IsEqualModelHistoryMap(v, vb) {
			return false
		}
	}
	return true
}

// 判断两个modelHistorySlice是否相同
func IsEqualModelHistorySlice(a []schema.ModelHistory, b []schema.ModelHistory) bool {
	SortModelHistoriesByID(a)
	SortModelHistoriesByID(b)
	if len(a) != len(b) {
		return false
	}
	for k, modelHistory := range a {
		if !IsStrictEqualModelHistory(modelHistory, b[k]) {
			return false
		}
	}
	return true
}

func IsStrictEqualModelHistory(
	a schema.ModelHistory,
	b schema.ModelHistory) bool {
	return a.ID == b.ID && a.ModelName == b.ModelName && a.Timestamp == b.Timestamp &&
		a.IsLocked == b.IsLocked && a.Desc == b.Desc
}

func SortModelHistoriesByID(modelHistories []schema.ModelHistory) {
	sort.SliceStable(modelHistories, func(i, j int) bool {
		if modelHistories[i].ID < modelHistories[j].ID {
			return true
		}
		return false
	})
}
