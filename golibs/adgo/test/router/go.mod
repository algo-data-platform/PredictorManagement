module github.com/algo-data-platform/predictor/golibs/adgo/test/router

go 1.13

replace (
	github.com/algo-data-platform/predictor/golibs/adgo/common/cityhash => ../../common/cityhash
	github.com/algo-data-platform/predictor/golibs/adgo/common/tally => ../../common/tally
	github.com/algo-data-platform/predictor/golibs/adgo/service_router => ../../service_router
	github.com/algo-data-platform/predictor/golibs/adgo/thirdparty/thrift => ../../thirdparty/thrift
)

require (
	github.com/algo-data-platform/predictor/golibs/adgo/common/tally v0.0.0-20191129020730-c7c12c772507 // indirect
	github.com/algo-data-platform/predictor/golibs/adgo/service_router v0.0.0-20191022060132-dc1c3ed0407e
	github.com/algo-data-platform/predictor/golibs/adgo/thirdparty/thrift v0.0.0-20191129020730-c7c12c772507 // indirect
	github.com/hashicorp/consul/api v1.2.0
)
