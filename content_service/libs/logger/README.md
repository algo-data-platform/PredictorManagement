#logger 使用方法

##初始化：
    ```golang
    logOptions:=logger.New()
	logOptions.SetLogFile("logs/content-service.log") //设置log文件地址
	logOptions.SetEncoding("console") //日志格式 console 和 json 两种编码
    // 日志级别 debug < info < warn < error < panic < fatal
	logOptions.SetLevel("debug") //设置日志级别，建议dev 为 debug，pro 为 info
    logOptions.SetRotate(false) //是否切割日志
    //下面两个设置是基于 rotate = true 的时候
    logOptions.SetRotateCycle("day") //日志切割周期，day hour minute
    logOptions.SetRotateMaxHours(72) //日志最大保留小时数

	logOptions.InitLogger()
    ```

##使用，两种方式

###第一种是性能较好，zap官方建议，缺点较繁琐
    ```golang
    logger.Info("this is zap test",
        zap.String("url", "dummy_web.com"),
        zap.Int("num", 3),
        zap.Duration("second", time.Second),
    )
    ```

    ```golang
    conf=struct {
        m        int
        n        int
        expected int
    }{1, 0, 1}
        
    logger.Warn("this is zap test",
        zap.String("url", "dummy_web.com"),
        zap.Int("num", 3),
        zap.Any("test", conf),
    )
    ```

###另一种suggerzap 的方式，更加方便

    ```golang
    logger.Warnf("this is zap test,url: %s,num:%d,conf:%+v",
        "dummy_web.com",
        3,
        conf,
    )
    ```


    