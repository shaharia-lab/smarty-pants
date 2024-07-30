package collector

import "time"

// Config contains configuration options for the collector
type Config struct {
	WorkerCount               int
	CollectionInterval        time.Duration
	RetryAttempts             int
	RetryDelay                time.Duration
	ShutdownTimeout           time.Duration
	DatasourceRefreshInterval time.Duration
}

// DefaultConfig returns a default configuration for the collector
func DefaultConfig() Config {
	return Config{
		WorkerCount:               5,
		CollectionInterval:        10 * time.Second,
		RetryAttempts:             3,
		RetryDelay:                30 * time.Second,
		ShutdownTimeout:           10 * time.Second,
		DatasourceRefreshInterval: 1 * time.Minute,
	}
}
