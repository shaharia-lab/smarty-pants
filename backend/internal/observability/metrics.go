package observability

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type MetricsServer struct {
	server *http.Server
	logger *logrus.Logger
}

func NewMetricsServer(port int, logger *logrus.Logger) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return &MetricsServer{
		server: server,
		logger: logger,
	}
}

func (ms *MetricsServer) Start() {
	go func() {
		ms.logger.Infof("Serving metrics at localhost%s/metrics", ms.server.Addr)
		if err := ms.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			ms.logger.WithError(err).Error("Metrics server failed to start")
		}
	}()
}

func (ms *MetricsServer) Stop(ctx context.Context) error {
	ms.logger.Info("Shutting down metrics server")
	return ms.server.Shutdown(ctx)
}

func StartMetricsServer(port int, logger *logrus.Logger) *MetricsServer {
	ms := NewMetricsServer(port, logger)
	ms.Start()
	return ms
}
