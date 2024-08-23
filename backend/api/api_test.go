package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/search"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("Server didn't start within the expected time")
	}

	_, err := http.Get(fmt.Sprintf("http://localhost:%d/system/ping", api.port))
	assert.NoError(t, err, "Failed to connect to the server")

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
		config:       config,
		router:       chi.NewRouter(),
		port:         port,
		logger:       logger,
		storage:      mockStorage,
		searchSystem: searchSystem,
		userManager:  userManager,
		aclManager:   auth.NewACLManager(logger, false),
	}
}

func TestAnalyticsOverviewEndpoint(t *testing.T) {
	user := &types.User{
		UUID:      uuid.MustParse("bc1d183a-1003-436c-97b8-2937b34dd0f4"),
		Name:      "Test User",
		Email:     "test@hello-world.com",
		Status:    types.UserStatusActive,
		Roles:     []types.UserRole{types.UserRoleAdmin},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	mockStorage := new(storage.StorageMock)
	mockStorage.On("GetKeyPair").Return(privateKeyBytes, publicKeyBytes, nil)
	mockStorage.On("GetUser", mock.Anything, user.UUID).Return(user, nil)
	mockStorage.On("GetAnalyticsOverview", mock.Anything).Return(types.AnalyticsOverview{}, nil).Once()

	mockLogger := logger.NoOpsLogger()
	mockACLManager := auth.NewACLManager(mockLogger, true)
	userManager := auth.NewUserManager(mockStorage, mockLogger)

	jwtManager := auth.NewJWTManager(auth.NewKeyManager(mockStorage, mockLogger), userManager, mockLogger)
	accessToken, err := jwtManager.IssueTokenForUser(context.Background(), user.UUID, []string{"web"}, 10*time.Hour)
	assert.NoError(t, err)

	// Create a new router
	r := chi.NewRouter()
	r.Use(jwtManager.AuthMiddleware(true))

	// Set up the route
	r.Get("/api/v1/analytics/overview", getAnalyticsOverview(mockStorage, mockLogger, mockACLManager))

	// Create a test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test cases
	tests := []struct {
		name           string
		withToken      bool
		hasPermission  bool
		expectedStatus int
	}{
		{"With valid token and permission", true, true, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request
			req, err := http.NewRequest("GET", ts.URL+"/api/v1/analytics/overview", nil)
			assert.NoError(t, err)

			// Add token if needed
			if tt.withToken {
				req.Header.Set("Authorization", "Bearer "+accessToken)
			}

			// Send the request
			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Check the status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			mockStorage.AssertExpectations(t)
		})
	}
}
