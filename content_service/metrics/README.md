##metrics 使用说明：
    目前metrics 暂时支持在server启动后添加
    在server代码中mark前，需要在registerMetrics() 方法中注册meter，例如添加一个统计modelservice耗时的timer
    
```golang
timers[TIMER_MODEL_SERVICE_CHECK_TIMER]=m.metrics.Tagged(moudleTag).Timer(TIMER_MODEL_SERVICE_CHECK_TIMER)```
```

对于error meter 分类较多，所以单独提出一个更通用 GetErrorMeter(service_name string, error_name string) tally.Meter{} ,不需要在registerMetrics()中注册，在代码中可以指定不同error tag 的Meter

```golang
metrics.GetErrorMeter(metrics.TAG_MODEL_SERVICE, metrics.TAG_CHECK_ERROR).Mark(1)
```