// Package logger provides a logger for the application.
package logger

import (
	"os"

	"github.com/shaharia-lab/smarty-pants-ai/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

const (
	// FormatJSON is the JSON log format
	FormatJSON types.LogFormat = "json"

	LevelDebug types.LogLevel = "debug"
	LevelInfo                 = "info"

	OutputStderr types.LogOutput = "stderr"
)

// Config is the configuration for the logger
type Config struct {
	Format types.LogFormat
	Level  types.LogLevel
	Output types.LogOutput
}

// New creates a new logger with the given configuration
func New(config Config) *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(string(config.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	switch config.Format {
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{})
	case "json":
		fallthrough
	default:
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	switch config.Output {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	default:
		file, err := os.OpenFile(string(config.Output), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.SetOutput(os.Stdout)
		} else {
			logger.SetOutput(file)
		}
	}

	return logger
}

// NoOpsLogger creates a logger that does nothing
func NoOpsLogger() *logrus.Logger {
	l, _ := test.NewNullLogger()
	return l
}

// BuildLoggerFromAppSettings creates a logger from the given application settings
func BuildLoggerFromAppSettings(appSettings types.Settings) *logrus.Logger {
	l := New(Config{
		Format: appSettings.Debugging.LogFormat,
		Level:  appSettings.Debugging.LogLevel,
		Output: appSettings.Debugging.LogOutput,
	})

	l.Info("Logger initialized successfully")
	return l
}
