//go:build integration
// +build integration

package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/shaharia-lab/smarty-pants/backend/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartCommand(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Find available ports for API and metrics
	apiPort, err := getFreePort()
	require.NoError(t, err, "Failed to get free port for API")
	metricsPort, err := getFreePort()
	require.NoError(t, err, "Failed to get free port for metrics")

	// Set environment variables for the test
	os.Setenv("API_PORT", strconv.Itoa(apiPort))
	os.Setenv("OTEL_METRICS_EXPOSED_PORT", strconv.Itoa(metricsPort))
	os.Setenv("COLLECTOR_WORKER_COUNT", "1")
	os.Setenv("PROCESSOR_WORKER_COUNT", "1")
	os.Setenv("GRACEFUL_SHUTDOWN_TIMEOUT_IN_SECS", "10")

	// Run the start command in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		startCmd := NewStartCommand("test-version")
		err := startCmd.ExecuteContext(ctx)
		if err != nil {
			errChan <- err
		}
	}()

	// Wait for the application to start
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/system/ping", apiPort))
		return err == nil && resp.StatusCode == http.StatusOK
	}, 15*time.Second, 100*time.Millisecond, "Application failed to start")

	logger.Debug("Application started successfully")

	// Perform your integration tests here
	t.Run("Application is running", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/system/ping", apiPort))
		assert.NoError(t, err, "Failed to send request to /system/ping")
		if resp != nil {
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Unexpected status code from /system/ping")
			resp.Body.Close()
		}
	})

	t.Run("Health endpoints are accessible", func(t *testing.T) {
		endpoints := []string{"/system/probes/liveness", "/system/probes/readiness"}
		for _, endpoint := range endpoints {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", apiPort, endpoint))
			assert.NoError(t, err, fmt.Sprintf("Failed to send request to %s", endpoint))
			if resp != nil {
				assert.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Unexpected status code from %s", endpoint))
				resp.Body.Close()
			}
		}
	})

	t.Run("Metrics endpoint is accessible", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", metricsPort))
		assert.NoError(t, err, "Failed to send request to /metrics")
		if resp != nil {
			assert.Equal(t, http.StatusOK, resp.StatusCode, "Unexpected status code from /metrics")
			resp.Body.Close()
		}
	})

	t.Run("Database migration completed successfully", func(t *testing.T) {
		cfg, _ := config.Load()
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBUser,
			cfg.DBPass,
			cfg.DBName,
		)

		dbConn, err := sql.Open("postgres", dsn)
		require.NoError(t, err, "Failed to connect to the database")
		defer dbConn.Close()

		// Check if the schema_migrations table exists
		var exists bool
		err = dbConn.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&exists)
		require.NoError(t, err, "Failed to check if schema_migrations table exists")
		assert.True(t, exists, "Expected table 'schema_migrations' to exist after migrations")

		// Check that dirty is false and version is not null
		var version sql.NullInt64
		var dirty bool
		err = dbConn.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
		require.NoError(t, err, "Failed to query schema_migrations table")

		assert.False(t, dirty, "Expected 'dirty' to be false in schema_migrations")
		assert.True(t, version.Valid, "Expected 'version' to be not null in schema_migrations")
		assert.Greater(t, version.Int64, int64(0), "Expected 'version' to be greater than 0 in schema_migrations")

		t.Logf("Current migration version: %d", version.Int64)

		// reset the database
		resetDatabase(t, dbConn)
	})

	// Check if the application started without errors
	select {
	case err := <-errChan:
		require.NoError(t, err, "Start command failed to execute")
	default:
		// No error, continue
	}

	// Cleanup
	logger.Debug("Signaling application to shutdown")
	cancel() // Signal shutdown

	// Forcefully kill the API server
	logger.Debug("Forcibly killing the API server")
	// No need to wait for graceful shutdown; end the test here
}

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
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
