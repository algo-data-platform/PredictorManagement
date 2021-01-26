package main

import (
	"time"

	"github.com/algo-data-platform/predictor/golibs/adgo/service_router"
	"github.com/hashicorp/consul/api"
)

func main() {
	config := &service_router.RouterConfig{
		Consul: api.Config{
			Address: "10.85.101.119:8500",
			Scheme:  "http",
		},
		ProjectName: "liubang_test_metrics",
	}
	router := service_router.GetRouter(config)
	router.RegisterServer(service_router.Server{
		ServiceName: "liubang_test",
		Host:        "127.0.0.1",
		Port:        4321,
		Protocol:    service_router.ServerProtocol_REDIS,
		Status:      service_router.ServerStatus_AVAILABLE,
	})

	router.GetConfigs("echo")

	for {
		time.Sleep(time.Second * 30)
	}
}
