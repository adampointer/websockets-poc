package event_streamer

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsHandler() http.Handler {
	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return promhttp.HandlerFor(reg,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
}
