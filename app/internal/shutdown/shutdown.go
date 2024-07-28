// Package shutdown provides a graceful shutdown manager for the application.
package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager is a shutdown manager
type Manager struct {
	logger      *logrus.Logger
	shutdownCh  chan os.Signal
	timeout     time.Duration
	wg          sync.WaitGroup
	shutdownFns []func(context.Context) error
	cancel      context.CancelFunc
}

// NewManager creates a new shutdown manager
func NewManager(logger *logrus.Logger, timeout time.Duration) *Manager {
	return &Manager{
		logger:     logger,
		shutdownCh: make(chan os.Signal, 1),
		timeout:    timeout,
	}
}

// RegisterShutdownFn registers a shutdown function
func (m *Manager) RegisterShutdownFn(fn func(context.Context) error) {
	m.shutdownFns = append(m.shutdownFns, fn)
}

// Start starts the manager
func (m *Manager) Start(ctx context.Context) {
	ctx, m.cancel = context.WithCancel(ctx)

	signal.Notify(m.shutdownCh,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	<-m.shutdownCh
	m.logger.Info("Shutdown signal received, initiating graceful shutdown")

	m.cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	for _, fn := range m.shutdownFns {
		m.wg.Add(1)
		go func(fn func(context.Context) error) {
			defer m.wg.Done()
			if err := fn(shutdownCtx); err != nil {
				m.logger.WithError(err).Error("Error during shutdown")
			}
		}(fn)
	}

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("Graceful shutdown completed")
	case <-shutdownCtx.Done():
		m.logger.Warn("Shutdown timed out, forcing exit")
	}
}

// Wait waits for the manager to finish
func (m *Manager) Wait() {
	m.wg.Wait()
}

// ShutdownChannel returns the shutdown channel
func (m *Manager) ShutdownChannel() <-chan os.Signal {
	return m.shutdownCh
}
