module server

go 1.13

require (
	github.com/DeanThompson/ginpprof v0.0.0-20190408063150-3be636683586
	github.com/algo-data-platform/predictor/golibs/adgo/feature_master v0.0.0-20210106100345-133acd1a178b // indirect
	github.com/algo-data-platform/predictor/golibs/adgo/predictor_client v0.0.0-20210106100345-133acd1a178b
	github.com/algo-data-platform/predictor/golibs/ads_common_go v1.0.14
	github.com/davecgh/go-spew v1.1.1
	github.com/gin-gonic/gin v1.4.0
	github.com/jinzhu/gorm v1.9.11
	github.com/lestrrat-go/file-rotatelogs v2.2.0+incompatible
	github.com/lestrrat-go/strftime v1.0.0 // indirect
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/viper v1.4.0
	github.com/xxjwxc/gowp v0.0.0-20191121101706-4740744adc76
	go.uber.org/zap v1.10.0
)

replace github.com/algo-data-platform/predictor/golibs/ads_common_go => ../../golibs/ads_common_go

replace github.com/algo-data-platform/predictor/golibs/adgo/feature_master => ../../golibs/adgo/feature_master

replace github.com/algo-data-platform/predictor/golibs/adgo/predictor_client => ../../golibs/adgo/predictor_client
