package system

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	logger := logrus.New()
	systemInfo := Info{
		Version: "1.0.0",
		App:     App{Name: "TestApp"},
		Settings: Settings{
			AuthEnabled:    true,
			OAuthProviders: []string{"google", "github"},
		},
	}

	manager := NewManager(logger, systemInfo)

	assert.NotNil(t, manager)
	assert.Equal(t, logger, manager.logger)
	assert.Equal(t, systemInfo, manager.systemInfo)
}

func TestManager_RegisterRoutes(t *testing.T) {
	logger := logrus.New()
	systemInfo := Info{
		Version: "1.0.0",
		App:     App{Name: "TestApp"},
		Settings: Settings{
			AuthEnabled:    true,
			OAuthProviders: []string{"google", "github"},
		},
	}

	manager := NewManager(logger, systemInfo)

	r := chi.NewRouter()
	manager.RegisterRoutes(r)

	// Test /system/ping
	t.Run("Ping", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/system/ping", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response types.GenerateResponseMsg
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Pong", response.Message)
	})

	// Test /system/probes/liveness
	t.Run("Liveness", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/system/probes/liveness", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response types.GenerateResponseMsg
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "I am alive", response.Message)
	})

	// Test /system/probes/readiness
	t.Run("Readiness", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/system/probes/readiness", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response types.GenerateResponseMsg
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "I am ready", response.Message)
	})

	// Test /system/info
	t.Run("Info", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/system/info", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response Info
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, systemInfo, response)
	})
}
