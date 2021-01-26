package server

import (
  "fmt"
  "net/http"
  "time"
  "content_service/env"
  "content_service/libs/logger"
  "content_service/metrics"
  "content_service/common"
)

type HttpService struct {
  StartTime time.Time
}

func NewHttpService() *HttpService {
	return &HttpService{}
}

func (service *HttpService) setRoutes(env *env.Env) {
	reporter := metrics.GetReporter()
  http.Handle("/server/prometheus", reporter.HTTPHandler())
  http.Handle("/server/json", reporter.JsonHTTPHandler())

  http.HandleFunc("/server/status", func (w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Http Service running...")
  })
  http.HandleFunc("/server/set_regression_test", func (w http.ResponseWriter, r *http.Request) {
    if !common.IsInSliceString("regressionService", env.Conf.Services) {
      w.WriteHeader(201)
      fmt.Fprintln(w, "regression switch is off")
      return
    }
    if !env.IsRegressionMode() {
      w.WriteHeader(201)
      fmt.Fprintln(w, "current mode is not regression")
      return
    }
    regression_service := GetRegressionInstance()
    query := r.URL.Query()
    packet_name := query.Get("packet_name")
    if packet_name == "" {
      w.WriteHeader(201)
      fmt.Fprintln(w, "packet_name is null")
      return
    }

    mode := query.Get("mode")
    if mode == "" {
      w.WriteHeader(201)
      fmt.Fprintln(w, "mode is null")
      return
    }

    env.Conf.RegressionService.PacketName = packet_name
    success, errs:= regression_service.Run(env, packet_name, mode)
    if success {
      w.WriteHeader(200)
      fmt.Fprintln(w, "验证成功")
    } else {
      w.WriteHeader(201)
      fmt.Fprintln(w, errs)
    }
    })
}

func (service *HttpService) Run(env *env.Env) {
  service.StartTime = time.Now()
  
  // add consuming metrics
  consumGauge := metrics.GetServerConsumGauge(service.StartTime)
  consumGauge.Update(1)

  service.setRoutes(env)
  logger.Infof("Http Service running...")
  http.ListenAndServe(":" + env.Conf.HttpPort, nil)
}