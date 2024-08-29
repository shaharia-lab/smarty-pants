// Package cmd contains the start command which is used to start the application.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shaharia-lab/smarty-pants/backend/api"
	"github.com/shaharia-lab/smarty-pants/backend/internal/analytics"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/collector"
	"github.com/shaharia-lab/smarty-pants/backend/internal/config"
	"github.com/shaharia-lab/smarty-pants/backend/internal/datasource"
	"github.com/shaharia-lab/smarty-pants/backend/internal/document"
	"github.com/shaharia-lab/smarty-pants/backend/internal/embedding"
	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/processor"
	"github.com/shaharia-lab/smarty-pants/backend/internal/search"
	"github.com/shaharia-lab/smarty-pants/backend/internal/shutdown"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

func NewStartCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:     "start",
		Short:   "Start the application",
		Version: version,
		RunE:    runStart,
	}
}

func runStart(cmd *cobra.Command, _ []string) error {
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Combined context that cancels either when the command context is done or when a signal is received
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	go func() {
		select {
		case <-signalCtx.Done():
			cancel()
		case <-ctx.Done():
		}
	}()

	l := initializeLogger()
	cfg, err := loadConfig(l)
	if err != nil {
		return err
	}

	cleanup, err := initializeTracer(ctx, cfg, l)
	if err != nil {
		return err
	}
	defer cleanup()

	st, err := initializeStorage(cfg, l)
	if err != nil {
		return err
	}

	err = st.RunMigration()
	if err != nil {
		return err
	}

	_, logging, err := setupAppSettings(ctx, st, l)
	if err != nil {
		return err
	}

	userManager := auth.NewUserManager(st, logging)

	metricsServer := observability.StartMetricsServer(cfg.OtelMetricsExposedPort, logging)

	shutdownManager := initializeShutdownManager(cfg, logging)

	collectorRunner, err := setupAndStartCollector(ctx, cfg, st, logging)
	if err != nil {
		return err
	}

	processorRunner, err := setupAndStartProcessor(ctx, cfg, st, logging)
	if err != nil {
		return err
	}

	aclManager := auth.NewACLManager(logging, cfg.EnableAuthentication)
	authSkipEndpoints := []string{
		//"/api/v1/analytics/overview",
	}
	apiServer := setupAPIServer(
		cfg,
		logging,
		st,
		userManager,
		auth.NewJWTManager(auth.NewKeyManager(st, logging), userManager, logging, authSkipEndpoints),
		aclManager,
		cfg.EnableAuthentication,
		analytics.NewAnalyticsManager(st, logging, aclManager),
		datasource.NewDatasourceManager(st, logging, aclManager),
		document.NewDocumentManager(st, logging),
		embedding.NewEmbeddingManager(st, logging),
	)

	shutdownManager.RegisterShutdownFn(func(ctx context.Context) error {
		return apiServer.Shutdown(ctx)
	})
	shutdownManager.RegisterShutdownFn(func(ctx context.Context) error {
		return st.HandleShutdown(ctx)
	})
	shutdownManager.RegisterShutdownFn(func(ctx context.Context) error {
		collectorRunner.Stop()
		return nil
	})
	shutdownManager.RegisterShutdownFn(func(ctx context.Context) error {
		processorRunner.Stop()
		return nil
	})
	shutdownManager.RegisterShutdownFn(func(ctx context.Context) error {
		return metricsServer.Stop(ctx)
	})

	go shutdownManager.Start(ctx)
	go startAPIServer(cfg, apiServer, logging)

	<-ctx.Done()
	logging.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.GracefulShutdownTimeoutInSecs)*time.Second)
	defer cancel()

	if err := shutdownManager.Shutdown(shutdownCtx); err != nil {
		logging.WithError(err).Error("Error during shutdown")
		return err
	}

	logging.Info("Application shutdown complete")
	return nil
}

func initializeLogger() *logrus.Logger {
	return logger.New(logger.Config{
		Format: logger.FormatJSON,
		Level:  logger.LevelDebug,
		Output: logger.OutputStderr,
	})
}

func loadConfig(l *logrus.Logger) (*config.Config, error) {
	l.Info("Loading configuration")
	cfg, err := config.Load()
	if err != nil {
		l.WithError(err).Error("Failed to load configuration")
		return nil, err
	}
	l.Info("Configuration loaded successfully")
	return cfg, nil
}

func initializeTracer(ctx context.Context, cfg *config.Config, l *logrus.Logger) (func(), error) {
	l.Info("Initializing tracer")
	cleanup, err := observability.InitTracer(ctx, cfg.AppName, l, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
		return nil, err
	}
	l.Info("Tracer initialized successfully")
	return cleanup, nil
}

func initializeStorage(cfg *config.Config, l *logrus.Logger) (storage.Storage, error) {
	l.Info("Initializing storage")
	_, storageSpan := observability.StartSpan(context.Background(), "cmd.api.init-storage")
	defer storageSpan.End()

	pc := storage.PostgresConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPass,
		DBName:   cfg.DBName,
	}

	st, err := storage.NewPostgres(pc, l)
	if err != nil {
		l.Fatalf("Failed to create storage: %v", err)
		return nil, err
	}

	l.Info("Storage initialized successfully")
	return st, nil
}

func setupAppSettings(ctx context.Context, st storage.Storage, l *logrus.Logger) (types.Settings, *logrus.Logger, error) {
	l.Info("Getting app settings")
	appSettings, err := st.GetSettings(ctx)
	if err != nil {
		return appSettings, nil, fmt.Errorf("failed to get app settings: %w", err)
	}

	l.Info("Building logger from app settings")
	logging := logger.BuildLoggerFromAppSettings(appSettings)

	return appSettings, logging, nil
}

func initializeShutdownManager(cfg *config.Config, logging *logrus.Logger) *shutdown.Manager {
	return shutdown.NewManager(logging, time.Duration(cfg.GracefulShutdownTimeoutInSecs)*time.Second)
}

func setupAndStartCollector(ctx context.Context, cfg *config.Config, st storage.Storage, logging *logrus.Logger) (*collector.Collector, error) {
	logging.Info("Creating collector runner")
	collectorConfig := collector.DefaultConfig()
	meter := otel.Meter("smarty-pants-ai")
	collectorRunner, err := collector.NewCollector(collectorConfig, st, logging, meter)
	if err != nil {
		logging.WithError(err).Fatal("Failed to create collector")
		return nil, err
	}

	logging.Info("Starting collector")
	if err := collectorRunner.Start(ctx); err != nil {
		logging.WithError(err).Fatal("Failed to start collector")
		return nil, err
	}

	return collectorRunner, nil
}

func setupAndStartProcessor(ctx context.Context, cfg *config.Config, st storage.Storage, logging *logrus.Logger) (*processor.Processor, error) {
	logging.Info("Creating processor engine")
	meter := otel.Meter("smarty-pants-ai")
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
		logging.WithError(err).Fatal("Failed to create processor")
		return nil, err
	}

	logging.Info("Starting processor in the background")
	if err := processingEngine.Start(ctx); err != nil {
		logging.WithError(err).Fatal("Failed to start processor")
		return nil, err
	}

	return processingEngine, nil
}

func setupAPIServer(
	cfg *config.Config,
	logging *logrus.Logger,
	st storage.Storage,
	userManager *auth.UserManager,
	jwtmanager *auth.JWTManager,
	aclManager auth.ACLManager,
	authEnabled bool,
	analyticsManager *analytics.Analytics,
	datasourceManager *datasource.Manager,
	documentManager *document.DocumentManager,
	embeddingManager *embedding.EmbeddingManager,
) *api.API {
	logging.Info("Creating API server")
	return api.NewAPI(
		logging,
		st,
		search.NewSearchSystem(logging, st),
		api.Config{
			Port:              cfg.APIPort,
			ServerReadTimeout: cfg.APIServerReadTimeoutInSecs,
			WriteTimeout:      cfg.APIServerWriteTimeoutInSecs,
			IdleTimeout:       cfg.APIServerIdleTimeoutInSecs,
		},
		userManager,
		jwtmanager,
		aclManager,
		authEnabled,
		analyticsManager,
		datasourceManager,
		documentManager,
		embeddingManager,
	)
}

func startAPIServer(cfg *config.Config, a *api.API, logging *logrus.Logger) {
	logging.WithField("api_server_port", cfg.APIPort).Info("Starting API server")
	if err := a.Start(context.Background()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logging.WithError(err).Fatal("API server failed to start")
	}
}

func waitForShutdown(ctx context.Context, shutdownManager *shutdown.Manager, logging *logrus.Logger) {
	select {
	case <-ctx.Done():
		logging.Warn("Command execution timed out")
	case <-shutdownManager.ShutdownChannel():
		logging.Info("Shutdown signal received, initiating graceful shutdown")
	}

	shutdownManager.Wait()
	logging.Info("Application has been shutdown successfully")
}
