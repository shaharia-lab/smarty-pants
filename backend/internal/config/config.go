// Package config file contains the configuration options for the application. The configuration options are loaded from the environment variables.
package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config contains configuration options for the application
type Config struct {
	AppName string `envconfig:"APP_NAME" default:"smarty-pants"`

	AdminEmail string `envconfig:"ADMIN_EMAIL" default:"your@emailaddress.com"`

	DBHost string `envconfig:"DB_HOST" default:"localhost"`
	DBPort int    `envconfig:"DB_PORT" default:"5432"`
	DBUser string `envconfig:"DB_USER" default:"app"`
	DBPass string `envconfig:"DB_PASS" default:"pass"`
	DBName string `envconfig:"DB_NAME" default:"app"`

	APIPort                     int `envconfig:"API_PORT" default:"8080"`
	APIServerReadTimeoutInSecs  int `envconfig:"API_SERVER_READ_TIMEOUT_IN_SECS" default:"10"`
	APIServerWriteTimeoutInSecs int `envconfig:"API_SERVER_WRITE_TIMEOUT_IN_SECS" default:"30"`
	APIServerIdleTimeoutInSecs  int `envconfig:"API_SERVER_IDLE_TIMEOUT_IN_SECS" default:"120"`

	TracingEnabled bool   `envconfig:"TRACING_ENABLED" default:"false"`
	OTLPTracerHost string `envconfig:"OTLP_TRACER_HOST" default:"localhost"`
	OTLPTracerPort int    `envconfig:"OTLP_TRACER_PORT" default:"4317"`

	OtelMetricsEnabled     bool `envconfig:"OTEL_METRICS_ENABLED" default:"false"`
	OtelMetricsExposedPort int  `envconfig:"OTEL_METRICS_EXPOSED_PORT" default:"2223"`

	CollectorWorkerCount int `envconfig:"COLLECTOR_WORKER_COUNT" default:"1"`

	GracefulShutdownTimeoutInSecs int `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT_IN_SECS" default:"30"`

	ProcessorWorkerCount           int `envconfig:"PROCESSOR_WORKER_COUNT" default:"1"`
	ProcessorBatchSize             int `envconfig:"PROCESSOR_BATCH_SIZE" default:"2"`
	ProcessorIntervalInSecs        int `envconfig:"PROCESSOR_INTERVAL_IN_SECS" default:"10"`
	ProcessorRetryAttempts         int `envconfig:"PROCESSOR_RETRY_ATTEMPTS" default:"3"`
	ProcessorRetryDelayInSecs      int `envconfig:"PROCESSOR_RETRY_DELAY_IN_SECS" default:"5"`
	ProcessorShutdownTimeoutInSecs int `envconfig:"PROCESSOR_SHUTDOWN_TIMEOUT_IN_SECS" default:"10"`
	ProcessorRefreshIntervalInSecs int `envconfig:"PROCESSOR_REFRESH_INTERVAL_IN_SECS" default:"60"`

	EnableAuthentication bool `envconfig:"ENABLE_AUTH" default:"false"`

	EnableGoogleOAuth       bool   `envconfig:"ENABLE_GOOGLE_OAUTH" default:"false"`
	GoogleOAuthClientID     string `envconfig:"GOOGLE_OAUTH_CLIENT_ID" required:"true"`
	GoogleOAuthClientSecret string `envconfig:"GOOGLE_OAUTH_CLIENT_SECRET"`
	GoogleOAuthRedirectURL  string `envconfig:"GOOGLE_OAUTH_REDIRECT_URL" default:"http://localhost:8080/auth/google/callback"`
}

// Load loads the configuration options from the environment
func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
