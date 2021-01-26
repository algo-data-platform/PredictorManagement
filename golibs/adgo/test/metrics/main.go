package main

import (
	"net/http"

	"github.com/algo-data-platform/predictor/golibs/adgo/common/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var metric = metrics.GetMetrics()

func testCounter() {
	counter := metric.BuildCounter(prometheus.CounterOpts{
		Namespace: "namespace",
		Subsystem: "subsystem",
		Name:      "name",
	})

	for i := 0; i < 100; i++ {
		counter.Add(1)
	}
}

func testSummaryVec() {
	summary := metric.BuildSummaryVec(prometheus.SummaryOpts{
		Namespace:  "namespace",
		Subsystem:  "subsystem",
		Name:       "summary",
		Objectives: map[float64]float64{0.5: 0.05},
	}, []string{"name", "age"})

	for i := 0; i < 100; i++ {
		summary.WithLabelValues("liubang", "26").Observe(float64(i))
		summary.WithLabelValues("zhangsan", "22").Observe(float64(i))
	}

}

func testHistogram() {
	histogram := metric.BuildHistogramVec(prometheus.HistogramOpts{
		Namespace: "namespace",
		Subsystem: "subsystem",
		Name:      "histogram",
		Buckets:   []float64{1, 2, 3},
	}, []string{"STATUS", "METHOD", "URI"})

	type req struct {
		Status string
		Method string
		Uri    string
	}

	m := []req{
		req{
			Status: "200",
			Method: "GET",
			Uri:    "/a",
		},
		req{
			Status: "404",
			Method: "POST",
			Uri:    "/bb",
		},
		req{
			Status: "200",
			Method: "DELETE",
			Uri:    "/ccc",
		},
	}

	for i := 0; i < 100; i++ {
		r := m[i%3]
		histogram.With(prometheus.Labels{
			"STATUS": r.Status,
			"METHOD": r.Method,
			"URI":    r.Uri,
		}).Observe(float64(i))
	}
}

func main() {
	go testCounter()
	go testCounter()
	go testSummaryVec()
	go testHistogram()

	server := http.NewServeMux()
	server.Handle("/metrics", metrics.NewMetricsTextExporter(prometheus.DefaultGatherer))
	server.Handle("/metrics/json", metrics.NewMetricsJsonExporter(prometheus.DefaultGatherer))
	http.ListenAndServe(":8080", server)
}
