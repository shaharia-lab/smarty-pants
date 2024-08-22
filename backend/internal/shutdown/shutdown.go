package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type Manager struct {
	logger      *logrus.Logger
	shutdownCh  chan os.Signal
	timeout     time.Duration
	wg          sync.WaitGroup
	shutdownFns []func(context.Context) error
	cancel      context.CancelFunc
}

func NewManager(logger *logrus.Logger, timeout time.Duration) *Manager {
	return &Manager{
		logger:     logger,
		shutdownCh: make(chan os.Signal, 1),
		timeout:    timeout,
	}
}

func (m *Manager) RegisterShutdownFn(fn func(context.Context) error) {
	m.shutdownFns = append(m.shutdownFns, fn)
}

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
}

func (m *Manager) Shutdown(ctx context.Context) error {
	m.logger.Info("Executing shutdown functions")

	var wg sync.WaitGroup
	errChan := make(chan error, len(m.shutdownFns))

	for _, fn := range m.shutdownFns {
		wg.Add(1)
		go func(f func(context.Context) error) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					m.logger.WithField("panic", r).Error("Panic in shutdown function")
					errChan <- fmt.Errorf("panic in shutdown function: %v", r)
				}
			}()
			if err := f(ctx); err != nil {
				m.logger.WithError(err).Error("Error during shutdown function execution")
				errChan <- err
			}
		}(fn)
	}

	// Wait for all shutdown functions to complete or context to be done
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		m.logger.Info("All shutdown functions executed")
	case <-ctx.Done():
		m.logger.Warn("Shutdown context deadline exceeded")
		return ctx.Err()
	}

	close(errChan)
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors during shutdown: %v", len(errs), errs)
	}

	m.logger.Info("All shutdown functions executed successfully")
	return nil
}

func (m *Manager) Wait() {
	m.wg.Wait()
}

func (m *Manager) ShutdownChannel() <-chan os.Signal {
	return m.shutdownCh
}
