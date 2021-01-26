module content_service

go 1.12

require (
	github.com/algo-data-platform/predictor/golibs/adgo/common/tally v0.0.0-20200205103348-4f747f71689d
	github.com/algo-data-platform/predictor/golibs/adgo/feature_master v0.0.0-20210106100345-133acd1a178b
	github.com/algo-data-platform/predictor/golibs/adgo/predictor_client v0.0.0-20210106100345-133acd1a178b
	github.com/algo-data-platform/predictor/golibs/ads_common_go v1.0.14
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jinzhu/gorm v1.9.10
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.2.0+incompatible
	github.com/lestrrat-go/strftime v1.0.1 // indirect
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.2.0
	github.com/tebeka/strftime v0.1.3 // indirect
	go.uber.org/zap v1.13.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)

replace github.com/algo-data-platform/predictor/golibs/adgo/common/tally => ../golibs/adgo/common/tally

replace github.com/algo-data-platform/predictor/golibs/ads_common_go => ../golibs/ads_common_go

replace github.com/algo-data-platform/predictor/golibs/adgo/feature_master => ../golibs/adgo/feature_master

replace github.com/algo-data-platform/predictor/golibs/adgo/predictor_client => ../golibs/adgo/predictor_client
