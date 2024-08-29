package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/analytics"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/search"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewAPI(t *testing.T) {
	api := createTestAPI()

	assert.NotNil(t, api)
	assert.Greater(t, api.port, 0)
	assert.NotNil(t, api.router)
	assert.NotNil(t, api.logger)
	assert.NotNil(t, api.storage)
	assert.NotNil(t, api.searchSystem)
	assert.NotNil(t, api.userManager)
}

func TestSetupMiddleware(t *testing.T) {
	api := createTestAPIWithoutSetup()

	testHeaderKey := "X-Test-Header"
	testHeaderValue := "test-value"
	api.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(testHeaderKey, testHeaderValue)
			next.ServeHTTP(w, r)
		})
	})

	api.setupMiddleware()

	api.router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	api.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, testHeaderValue, rr.Header().Get(testHeaderKey))
}

func TestSetupRoutes(t *testing.T) {
	api := createTestAPIWithoutSetup()
	api.setupMiddleware()
	api.setupRoutes()

	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/system/ping"},
		{"GET", "/system/probes/liveness"},
		{"GET", "/system/probes/readiness"},
		{"GET", "/api/v1/analytics/overview"},
		{"POST", "/api/v1/datasource"},
		{"GET", "/api/v1/datasource"},
	}

	for _, tc := range testCases {
		req, _ := http.NewRequest(tc.method, tc.path, nil)
		rr := httptest.NewRecorder()
		api.router.ServeHTTP(rr, req)
		assert.NotEqual(t, http.StatusNotFound, rr.Code, "Route %s %s not found", tc.method, tc.path)
	}
}

func TestStart(t *testing.T) {
	api := createTestAPI()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	started := make(chan struct{})
	errChan := make(chan error, 1)

	go func() {
		close(started)
		err := api.Start(ctx)
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	<-started // Wait for the server to start

	// Implement a retry mechanism with timeout
	retryCtx, retryCancel := context.WithTimeout(ctx, 5*time.Second)
	defer retryCancel()

	var lastErr error
	for {
		select {
		case <-retryCtx.Done():
			t.Fatalf("Server didn't become available within the expected time. Last error: %v", lastErr)
		case <-time.After(100 * time.Millisecond):
			_, err := http.Get(fmt.Sprintf("http://localhost:%d/system/ping", api.port))
			if err == nil {
				// Successfully connected, proceed with the test
				goto serverReady
			}
			lastErr = err
		}
	}

serverReady:
	// Server is ready, proceed with the test
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/system/ping", api.port))
	assert.NoError(t, err, "Failed to connect to the server")
	if err == nil {
		resp.Body.Close()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	err = api.Shutdown(shutdownCtx)
	assert.NoError(t, err, "Failed to shut down the server")

	select {
	case err := <-errChan:
		t.Fatalf("Server error: %v", err)
	case <-ctx.Done():
	}
}

func TestShutdown(t *testing.T) {
	api := createTestAPI()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go func() {
		_ = api.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	err := api.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestDetailedRequestLogging(t *testing.T) {
	api := createTestAPIWithoutSetup()
	handler := api.detailedRequestLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestEnhancedRecoverer(t *testing.T) {
	api := createTestAPIWithoutSetup()
	handler := api.enhancedRecoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req, _ := http.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func createTestAPI() *API {
	api := createTestAPIWithoutSetup()
	api.setupMiddleware()
	api.setupRoutes()
	return api
}

func createTestAPIWithoutSetup() *API {
	logger := logrus.New()
	mockStorage := new(storage.StorageMock)
	searchSystem := search.NewSearchSystem(logger, mockStorage)
	userManager := auth.NewUserManager(mockStorage, logger)

	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(fmt.Sprintf("Failed to find an available port: %v", err))
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	config := Config{
		Port:              port,
		ServerReadTimeout: 30,
		WriteTimeout:      30,
		IdleTimeout:       60,
	}

	return &API{
		config:           config,
		router:           chi.NewRouter(),
		port:             port,
		logger:           logger,
		storage:          mockStorage,
		searchSystem:     searchSystem,
		userManager:      userManager,
		aclManager:       auth.NewACLManager(logger, false),
		analyticsManager: analytics.NewAnalytics(mockStorage, logger, auth.NewACLManager(logger, false)),
	}
}
