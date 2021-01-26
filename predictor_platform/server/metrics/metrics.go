package metrics

import (
	"github.com/algo-data-platform/predictor/golibs/ads_common_go/common/tally"
	promReporter "github.com/algo-data-platform/predictor/golibs/ads_common_go/common/tally/prometheus"
	"strings"
	"sync"
	"time"
)

var (
	metrics           tally.Scope
	reporter          promReporter.Reporter
	timers            map[string]tally.Timer
	meters            map[string]tally.Meter
	gauges            map[string]tally.Gauge
	modelGaugeMap     sync.Map
	modelDiffGaugeMap sync.Map
	modelSizeGaugeMap sync.Map
	loadMeterMap      sync.Map
)

func InitMetrics() {
	meters = make(map[string]tally.Meter)
	timers = make(map[string]tally.Timer)
	gauges = make(map[string]tally.Gauge)
	reporter = promReporter.NewReporter(promReporter.Options{})
	metrics, _ = tally.NewRootScope(tally.ScopeOptions{
		Prefix:         PREFIX,
		Tags:           map[string]string{"name": PROJECT_NAME},
		CachedReporter: reporter,
		Separator:      promReporter.DefaultSeparator,
	}, time.Second)
	registerMetrics()
}

func GetReporter() promReporter.Reporter {
	return reporter
}

func GetMetrics() tally.Scope {
	return metrics
}

func GetMeters() map[string]tally.Meter {
	return meters
}

func GetTimers() map[string]tally.Timer {
	return timers
}

func GetGauges() map[string]tally.Gauge {
	return gauges
}

// 根据上报的 service,host,model_name 来注册meter并返回
// 为了避免tag和gauge频繁创建，用map实现简单的单例
func GetModelGauge(service string, host string, model_name string) tally.Gauge {
	var instanceGauge interface{}
	var ok bool
	uniqKey := GetModelGaugeUniqKey(service, host, model_name)
	if instanceGauge, ok = modelGaugeMap.Load(uniqKey); !ok {
		tag := getModelNameTag(service, host, model_name)
		gauge := metrics.Tagged(tag).Gauge(GAUGE_MODEL_VERSION_INTERVAL)
		instanceGauge, ok = modelGaugeMap.LoadOrStore(uniqKey, gauge)
	}
	return instanceGauge.(tally.Gauge)
}

func GetModelGaugeUniqKey(service string, host string, model_name string) string {
	return strings.Join([]string{service, host, model_name}, "_")
}

// 根据service 和 host 及 model_name 生成tag
func getModelNameTag(service string, host string, model_name string) map[string]string {
	return map[string]string{
		"service":    service,
		"host":       host,
		"model_name": model_name,
	}
}

// 清除不存在的模型gauge
func ClearNotExistModelGauge(existKeyMap map[string]bool) {
	modelGaugeMap.Range(func(uniqKey, gaugeI interface{}) bool {
		if _, ok := existKeyMap[uniqKey.(string)]; !ok {
			gaugeI.(tally.Gauge).Update(0)
		}
		return true
	})
}

// 根据上报的 service,host,diff_name来注册模型不同的Gauge并返回
func GetModelDiffGauge(service string, host string, diff_name string) tally.Gauge {
	var instanceGauge interface{}
	var ok bool
	uniqKey := GetModelDiffGaugeUniqKey(service, host, diff_name)
	if instanceGauge, ok = modelDiffGaugeMap.Load(uniqKey); !ok {
		tag := getModelDiffTag(service, host, diff_name)
		gauge := metrics.Tagged(tag).Gauge(GAUGE_MODEL_DIFF)
		instanceGauge, ok = modelDiffGaugeMap.LoadOrStore(uniqKey, gauge)
	}
	return instanceGauge.(tally.Gauge)
}

func GetModelDiffGaugeUniqKey(service string, host string, diff_name string) string {
	return strings.Join([]string{service, host, diff_name}, "_")
}

// 根据service 和 host 生成tag
func getModelDiffTag(service string, host string, diff_name string) map[string]string {
	return map[string]string{
		"service":   service,
		"host":      host,
		"diff_name": diff_name,
	}
}

// 清除不存在的模型比较gauge
func ClearNotExistModelDiffGauge(existDiffKeyMap map[string]bool) {
	modelDiffGaugeMap.Range(func(uniqKey, gaugeI interface{}) bool {
		if _, ok := existDiffKeyMap[uniqKey.(string)]; !ok {
			gaugeI.(tally.Gauge).Update(0)
		}
		return true
	})
}

// 获取模型大小gauge
func GetModelSizeGauge(model_name string) tally.Gauge {
	var instanceGauge interface{}
	var ok bool
	uniqKey := GetModelSizeGaugeUniqKey(model_name)
	if instanceGauge, ok = modelSizeGaugeMap.Load(uniqKey); !ok {
		tag := getModelSizeTag(model_name)
		gauge := metrics.Tagged(tag).Gauge(GAUGE_MODEL_SIZE)
		instanceGauge, ok = modelSizeGaugeMap.LoadOrStore(uniqKey, gauge)
	}
	return instanceGauge.(tally.Gauge)
}

func GetModelSizeGaugeUniqKey(model_name string) string {
	return model_name
}

// 根据 model_name 生成tag
func getModelSizeTag(model_name string) map[string]string {
	return map[string]string{
		"model_name": model_name,
	}
}

// 清除不存在的模型大小gauge
func ClearNotExistModelSizeGauge(existKeyMap map[string]bool) {
	modelSizeGaugeMap.Range(func(uniqKey, gaugeI interface{}) bool {
		if _, ok := existKeyMap[uniqKey.(string)]; !ok {
			gaugeI.(tally.Gauge).Update(0)
		}
		return true
	})
}

// 根据上报的 service_name,host 来注册meter并返回
// 为了避免tag和meter频繁创建，用map实现简单的单例
func GetLoadChangeMeter(service_name string, host string) tally.Meter {
	var instanceMeter interface{}
	var ok bool
	uniqKey := strings.Join([]string{service_name, host}, "_")
	if instanceMeter, ok = loadMeterMap.Load(uniqKey); !ok {
		tag := getLoadNameTag(service_name, host)
		meter := metrics.Tagged(tag).Meter(METER_LOAD_CHANGE)
		instanceMeter, ok = loadMeterMap.LoadOrStore(uniqKey, meter)
	}
	return instanceMeter.(tally.Meter)
}

// 根据service_name和 host 生成tag
func getLoadNameTag(service_name string, host string) map[string]string {
	return map[string]string{
		"service_name": service_name,
		"host":         host,
	}
}
