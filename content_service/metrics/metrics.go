package metrics

import (
	"github.com/algo-data-platform/predictor/golibs/ads_common_go/common/tally"
	promReporter "github.com/algo-data-platform/predictor/golibs/ads_common_go/common/tally/prometheus"
	"strings"
	"sync"
	"time"
)

var (
	metrics               tally.Scope
	reporter              promReporter.Reporter
	timers                map[string]tally.Timer
	meters                map[string]tally.Meter
	gauges                map[string]tally.Gauge
	errorMeterMap         sync.Map
	pullModelTimerMap     sync.Map
	pullHdfsModelTimerMap sync.Map
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

// 根据上报的 service_name,error_name 来注册meter并返回
// 为了避免tag和meter频繁创建，用map实现简单的单例
func GetErrorMeter(service_name string, error_name string) tally.Meter {
	var instanceMeter interface{}
	var ok bool
	uniqKey := strings.Join([]string{service_name, error_name}, "_")
	if instanceMeter, ok = errorMeterMap.Load(uniqKey); !ok {
		tag := getErrorNameTag(service_name, error_name)
		meter := metrics.Tagged(tag).Meter(METER_SERVER_ERROR)
		instanceMeter, ok = errorMeterMap.LoadOrStore(uniqKey, meter)
	}
	return instanceMeter.(tally.Meter)
}

// 根据service_name和 error_name 生成tag
func getErrorNameTag(service_name string, error_name string) map[string]string {
	return map[string]string{
		"service_name": service_name,
		"error_name":   error_name,
	}
}

func GetServerConsumGauge(serverStartTime time.Time) tally.Gauge {
	metrics := GetMetrics()
	return metrics.Tagged(map[string]string{
		"start_time": serverStartTime.Format("20060102_150405"),
	}).Gauge(METER_CONSUMING)
}

// 根据上报的 model_name 来注册timer并返回
func GetPullSingleModelTimer(model_name string) tally.Timer {
	var instanceTimer interface{}
	var ok bool
	uniqKey := strings.Join([]string{"pull_sigle_model", model_name}, "_")
	if instanceTimer, ok = pullModelTimerMap.Load(uniqKey); !ok {
		tag := map[string]string{"model_name": model_name}
		timer := metrics.Tagged(tag).Timer(TIMER_PULL_SINGLE_MODEL_TIMER)
		instanceTimer, ok = pullModelTimerMap.LoadOrStore(uniqKey, timer)
	}
	return instanceTimer.(tally.Timer)
}

// 根据上报的 model_name 来注册timer并返回
func GetPullSingleHdfsModelTimer(model_name string) tally.Timer {
	var instanceTimer interface{}
	var ok bool
	uniqKey := strings.Join([]string{"pull_sigle_model", model_name}, "_")
	if instanceTimer, ok = pullHdfsModelTimerMap.Load(uniqKey); !ok {
		tag := map[string]string{"model_name": model_name}
		timer := metrics.Tagged(tag).Timer(TIMER_PULL_SINGLE_HDFS_MODEL_TIMER)
		instanceTimer, ok = pullHdfsModelTimerMap.LoadOrStore(uniqKey, timer)
	}
	return instanceTimer.(tally.Timer)
}
