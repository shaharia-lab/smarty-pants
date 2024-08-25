//go:build integration
// +build integration

package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/shaharia-lab/guti"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestStartCommand(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	apiPort, err := guti.GetFreePortFromPortRange(10000, 20000)
	require.NoError(t, err, "Failed to get free port for API")
	metricsPort, err := guti.GetFreePortFromPortRange(20100, 30000)
	require.NoError(t, err, "Failed to get free port for metrics")

	os.Setenv("API_PORT", strconv.Itoa(apiPort))
	os.Setenv("OTEL_METRICS_EXPOSED_PORT", strconv.Itoa(metricsPort))
	os.Setenv("COLLECTOR_WORKER_COUNT", "1")
	os.Setenv("PROCESSOR_WORKER_COUNT", "1")
	os.Setenv("GRACEFUL_SHUTDOWN_TIMEOUT_IN_SECS", "5")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	doneChan := make(chan struct{})

	go func() {
		startCmd := NewStartCommand("test-version")
		err := startCmd.ExecuteContext(ctx)
		if err != nil {
			errChan <- err
		}
		close(doneChan)
	}()

	// Wait for the application to start
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/system/ping", apiPort))
		return err == nil && resp.StatusCode == http.StatusOK
	}, 15*time.Second, 100*time.Millisecond, "Application failed to start")

	logger.Debug("Application started successfully")

	// Run other sub-tests here...

	// Test graceful shutdown
	t.Run("Graceful_shutdown", func(t *testing.T) {
		// Send termination signal
		cancel()

		// Wait for the application to shut down
		select {
		case err := <-errChan:
			require.NoError(t, err, "Unexpected error during shutdown")
		case <-doneChan:
			// Application shut down successfully
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for application to shut down")
		}

		// Verify that the API is no longer accessible
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/system/ping", apiPort))
		require.Error(t, err, "Expected API to be inaccessible after shutdown")

		// Verify that the metrics endpoint is no longer accessible
		_, err = http.Get(fmt.Sprintf("http://localhost:%d/metrics", metricsPort))
		require.Error(t, err, "Expected metrics endpoint to be inaccessible after shutdown")
	})
}

func resetDatabase(t *testing.T, dbConn *sql.DB) {
	t.Helper()

	// Drop all tables
	_, err := dbConn.Exec(`
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = current_schema()) LOOP
				EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`)
	require.NoError(t, err, "Failed to drop all tables")

	t.Log("Database reset completed")
}
