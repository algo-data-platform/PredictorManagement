##metrics 使用说明：
    目前metrics 暂时支持在server启动后添加
    在server代码中mark前，需要在registerMetrics() 方法中注册meter，例如添加一个统计modelservice耗时的timer
    
```golang
timers[TIMER_MODEL_SERVICE_CHECK_TIMER]=m.metrics.Tagged(moudleTag).Timer(TIMER_MODEL_SERVICE_CHECK_TIMER)
```
对于 GetModelGauge 调用方法 ,可减少metrics的初始化
```golang
metrics.GetModelGauge(host, model_name).Update(intervalSeconds)
```