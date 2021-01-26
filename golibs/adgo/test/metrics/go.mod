module github.com/algo-data-platform/predictor/golibs/adgo/test/metrics

go 1.13

replace github.com/algo-data-platform/predictor/golibs/adgo/common/metrics => ../../common/metrics

replace github.com/algo-data-platform/predictor/golibs/adgo/common/cityhash => ../../common/cityhash

require github.com/algo-data-platform/predictor/golibs/adgo/common/metrics v0.0.0-20191022030655-4b861adfb722

require github.com/prometheus/client_golang v1.2.1
