package websocket_adaptor

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var latencyHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
	Namespace: "websocket_adaptor",
	Name:      "latency",
	Help:      "latency",
})

func MetricsHandler() http.Handler {
	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		latencyHistogram,
	)

	return promhttp.HandlerFor(reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
}
