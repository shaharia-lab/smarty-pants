// Package observability provides observability tools for the application.
package observability

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// StartMetricsEndpoint starts an HTTP server that serves metrics
func StartMetricsEndpoint(port int, log *logrus.Logger) {
	log.Printf("serving metrics at localhost:%d/metrics", port)
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}
