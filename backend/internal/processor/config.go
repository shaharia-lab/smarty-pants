// Package processor provides a processor for the application.
package processor

import "time"

// Config contains configuration options for the processor
type Config struct {
	WorkerCount              int
	BatchSize                int
	ProcessInterval          time.Duration
	RetryAttempts            int
	RetryDelay               time.Duration
	ShutdownTimeout          time.Duration
	ProcessorRefreshInterval time.Duration
}

// DefaultConfig returns a default configuration for the processor
func DefaultConfig() Config {
	return Config{
		WorkerCount:              5,
		BatchSize:                10,
		ProcessInterval:          10 * time.Second,
		RetryAttempts:            3,
		RetryDelay:               5 * time.Second,
		ShutdownTimeout:          30 * time.Second,
		ProcessorRefreshInterval: 5 * time.Minute,
	}
}
