package logger

import (
	"github.com/sirupsen/logrus"
)

// MigrationLogger wraps a logrus.Logger and implements the Logger interface
type MigrationLogger struct {
	*logrus.Logger
	verbose bool
}

// NewMigrationLogger creates a new MigrationLogger
func NewMigrationLogger(logger *logrus.Logger, verbose bool) *MigrationLogger {
	return &MigrationLogger{
		Logger:  logger,
		verbose: verbose,
	}
}

// Printf implements the Logger interface
func (ml *MigrationLogger) Printf(format string, v ...interface{}) {
	ml.Logger.Printf(format, v...)
}

// Verbose implements the Logger interface
func (ml *MigrationLogger) Verbose() bool {
	return ml.verbose
}
