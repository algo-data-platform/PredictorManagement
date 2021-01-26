package common

import (
	"server/schema"
	"server/util"
)

type ModelInfoWrapper struct {
	Model    []schema.Model
	Sortfunc func(p, q *schema.Model) bool
}

func (wrapper ModelInfoWrapper) Len() int {
	return len(wrapper.Model)
}

func (wrapper ModelInfoWrapper) Swap(i, j int) {
	wrapper.Model[i], wrapper.Model[j] = wrapper.Model[j], wrapper.Model[i]
}

func (wrapper ModelInfoWrapper) Less(i, j int) bool {
	return wrapper.Sortfunc(&wrapper.Model[i], &wrapper.Model[j])
}

type ModelUpdateTimingInfoArrayWrapper struct {
	Models   []util.ModelUpdateTimingInfo
	Sortfunc func(p, q *util.ModelUpdateTimingInfo) bool
}

func (wrapper ModelUpdateTimingInfoArrayWrapper) Len() int {
	return len(wrapper.Models)
}

func (wrapper ModelUpdateTimingInfoArrayWrapper) Swap(i, j int) {
	wrapper.Models[i], wrapper.Models[j] = wrapper.Models[j], wrapper.Models[i]
}

func (wrapper ModelUpdateTimingInfoArrayWrapper) Less(i, j int) bool {
	return wrapper.Sortfunc(&wrapper.Models[i], &wrapper.Models[j])
}
