package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/analytics"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPIEndpoints(t *testing.T) {
	// Setup
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	mockStorage := new(storage.StorageMock)
	mockStorage.On("GetKeyPair").Return(privateKeyBytes, publicKeyBytes, nil)
	mockLogger := logger.NoOpsLogger()
	mockACLManager := auth.NewACLManager(mockLogger, true)
	userManager := auth.NewUserManager(mockStorage, mockLogger)
	jwtManager := auth.NewJWTManager(auth.NewKeyManager(mockStorage, mockLogger), userManager, mockLogger, []string{})

	// Create a new router
	r := chi.NewRouter()
	r.Use(jwtManager.AuthMiddleware(true))

	an := analytics.NewAnalyticsManager(mockStorage, mockLogger, mockACLManager)
	an.RegisterRoutes(r)

	// Add more routes here as needed

	// Create a test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test cases
	tests := []struct {
		name           string
		endpoint       string
		method         string
		user           *types.User
		expectedStatus int
	}{
		{
			name:     "Admin accessing analytics overview",
			endpoint: "/api/v1/analytics/overview",
			method:   http.MethodGet,
			user: &types.User{
				UUID:   uuid.New(),
				Name:   "Admin User",
				Email:  "admin@example.com",
				Status: types.UserStatusActive,
				Roles:  []types.UserRole{types.UserRoleAdmin},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Regular user accessing analytics overview",
			endpoint: "/api/v1/analytics/overview",
			method:   http.MethodGet,
			user: &types.User{
				UUID:   uuid.New(),
				Name:   "Regular User",
				Email:  "user@example.com",
				Status: types.UserStatusActive,
				Roles:  []types.UserRole{types.UserRoleUser},
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.user != nil {
				mockStorage.On("GetUser", mock.Anything, tt.user.UUID).Return(tt.user, nil)
			}

			mockStorage.On("GetAnalyticsOverview", mock.Anything).Return(types.AnalyticsOverview{}, nil)

			// Generate token for the user
			accessToken, err := jwtManager.IssueToken(context.Background(), tt.user.UUID, []string{"web"}, 10*time.Hour)
			assert.NoError(t, err)

			// Create a new request
			req, err := http.NewRequest(tt.method, ts.URL+tt.endpoint, nil)
			assert.NoError(t, err)

			// Add token to the request
			req.Header.Set("Authorization", "Bearer "+accessToken)

			// Send the request
			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Check the status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
