module github.com/algo-data-platform/predictor/golibs/adgo/test/echo

go 1.13

require (
	github.com/algo-data-platform/predictor/golibs/adgo/service_router v0.0.0-00010101000000-000000000000
	github.com/algo-data-platform/predictor/golibs/adgo/thirdparty/thrift v0.0.0-20191024060658-3f705a18755b
	github.com/hashicorp/consul/api v1.2.0
)

replace github.com/algo-data-platform/predictor/golibs/adgo/service_router => ../../service_router
