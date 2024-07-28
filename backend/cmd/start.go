// Package cmd contains the start command which is used to start the application.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/shaharia-lab/smarty-pants-ai/api"
	"github.com/shaharia-lab/smarty-pants-ai/internal/collector"
	"github.com/shaharia-lab/smarty-pants-ai/internal/config"
	"github.com/shaharia-lab/smarty-pants-ai/internal/logger"
	"github.com/shaharia-lab/smarty-pants-ai/internal/observability"
	"github.com/shaharia-lab/smarty-pants-ai/internal/processor"
	"github.com/shaharia-lab/smarty-pants-ai/internal/search"
	"github.com/shaharia-lab/smarty-pants-ai/internal/shutdown"
	"github.com/shaharia-lab/smarty-pants-ai/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

// NewStartCommand creates a new start command
func NewStartCommand() *cobra.Command {
	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the application",
		RunE: func(_ *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			l := logger.New(logger.Config{
				Format: logger.FormatJSON,
				Level:  logger.LevelDebug,
				Output: logger.OutputStderr,
			})

			l.Info("Loading configuration")
			cfg, err := config.Load()
			if err != nil {
				l.WithError(err).Error("Failed to load configuration")
				return err
			}
			l.Info("Configuration loaded successfully")

			l.Info("Initializing tracer")
			cleanup, err := observability.InitTracer(ctx, cfg.AppName, l, cfg)
			if err != nil {
				log.Fatalf("Failed to initialize tracer: %v", err)
			}
			defer cleanup()
			l.Info("Tracer initialized successfully")

			l.Info("Initializing storage")
			_, storageSpan := observability.StartSpan(ctx, "cmd.api.init-storage")
			st := initializeStorage(cfg, l)
			storageSpan.End()
			l.Info("Storage initialized successfully")

			l.Info("Running migration for database")
			err = st.MigrationUp()
			if err != nil {
				l.WithError(err).Error("Failed to migrate")
				return fmt.Errorf("failed to migrate: %w", err)
			}
			l.Info("Database migration completed successfully")

			l.Info("Getting app settings")
			appSettings, err := st.GetSettings(ctx)
			if err != nil {
				return fmt.Errorf("failed to get app settings: %w", err)
			}

			l.Info("Building logger from app settings")
			logging := logger.BuildLoggerFromAppSettings(appSettings)

			l.WithField("metrics_server_port", cfg.OtelMetricsExposedPort).Info("Starting metrics server in the background")
			go func() {
				observability.StartMetricsEndpoint(cfg.OtelMetricsExposedPort, logging)
			}()

			shutdownManager := shutdown.NewManager(logging, time.Duration(cfg.GracefulShutdownTimeoutInSecs)*time.Second)
			meter := otel.Meter("smarty-pants-ai")

			l.Info("Creating collector runner")
			collectorConfig := collector.DefaultConfig()
			collectorRunner, err := collector.NewCollector(collectorConfig, st, logging, meter)
			if err != nil {
				logging.WithError(err).Fatal("Failed to create collector")
			}

			shutdownManager.RegisterShutdownFn(func(ctx context.Context) error {
				collectorRunner.Stop()
				return nil
			})

			l.Info("Starting collector")
			if err := collectorRunner.Start(ctx); err != nil {
				logging.WithError(err).Fatal("Failed to start collector")
			}

			l.Info("Creating processor engine")
			processingEngine, err := processor.NewProcessor(processor.Config{
				WorkerCount:              cfg.ProcessorWorkerCount,
				BatchSize:                cfg.ProcessorBatchSize,
				ProcessInterval:          time.Duration(cfg.ProcessorIntervalInSecs) * time.Second,
				RetryAttempts:            cfg.ProcessorRetryAttempts,
				RetryDelay:               time.Duration(cfg.ProcessorRetryDelayInSecs) * time.Second,
				ShutdownTimeout:          time.Duration(cfg.ProcessorShutdownTimeoutInSecs) * time.Second,
				ProcessorRefreshInterval: time.Duration(cfg.ProcessorRefreshIntervalInSecs) * time.Second,
			}, st, logging, meter)

			if err != nil {
				logging.WithError(err).Fatal("Failed to create collector")
			}

			shutdownManager.RegisterShutdownFn(func(_ context.Context) error {
				processingEngine.Stop()
				return nil
			})

			l.Info("Starting processor in the background")
			if err := processingEngine.Start(ctx); err != nil {
				logging.WithError(err).Fatal("Failed to start collector")
			}

			l.Info("Creating API server")
			a := api.NewAPI(
				logging,
				st,
				search.NewSearchSystem(logging, st),
				api.Config{
					Port:              cfg.APIPort,
					ServerReadTimeout: cfg.APIServerReadTimeoutInSecs,
					WriteTimeout:      cfg.APIServerWriteTimeoutInSecs,
					IdleTimeout:       cfg.APIServerIdleTimeoutInSecs,
				},
			)

			shutdownManager.RegisterShutdownFn(func(ctx context.Context) error {
				return a.Shutdown(ctx)
			})

			l.Info("Starting shutdown manager in the background")
			go shutdownManager.Start(ctx)

			l.WithField("api_server_port", cfg.APIPort).Info("Starting API server")
			go func() {
				if err := a.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logging.WithError(err).Fatal("API server failed to start")
				}
			}()

			select {
			case <-ctx.Done():
				logging.Warn("Command execution timed out")
			case <-shutdownManager.ShutdownChannel():
				logging.Info("Shutdown signal received, initiating graceful shutdown")
			}

			cancel()
			shutdownManager.Wait()

			logging.Info("Application has been shutdown successfully")
			return nil
		},
	}

	return startCmd
}

func initializeStorage(cfg *config.Config, log *logrus.Logger) storage.Storage {
	pc := storage.PostgresConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPass,
		DBName:   cfg.DBName,
		Config:   postgres.Config{},
	}

	st, err := storage.NewPostgres(pc, cfg.DBMigrationPath, log)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	log.Info("Storage connected successfully")
	return st
}
