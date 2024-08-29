package analytics

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/config"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAnalyticsOverview(t *testing.T) {
	fixedTime := time.Date(2024, 7, 18, 9, 10, 6, 484502057, time.UTC)

	testCases := []struct {
		name           string
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			mockSetup: func(m *storage.StorageMock) {
				m.On("GetAnalyticsOverview", mock.Anything).Return(types.AnalyticsOverview{
					EmbeddingProviders: types.EmbeddingProvidersOverview{
						TotalProviders:       2,
						TotalActiveProviders: 1,
						ActiveProvider: types.ProviderInfo{
							Name:  "OpenAI",
							Type:  "Embedding",
							Model: "text-embedding-ada-002",
						},
					},
					LLMProviders: types.LLMProvidersOverview{
						TotalProviders:       3,
						TotalActiveProviders: 1,
						ActiveProvider: types.ProviderInfo{
							Name:  "OpenAI",
							Type:  "LLM",
							Model: "gpt-3.5-turbo",
						},
					},
					Datasources: types.DatasourcesOverview{
						ConfiguredDatasources: []types.DatasourceInfo{
							{
								Name:      "Confluence",
								Type:      "confluence",
								Status:    "active",
								CreatedAt: fixedTime,
							},
						},
						TotalDatasources: 1,
						TotalDatasourcesByType: map[string]int{
							"confluence": 1,
						},
						TotalDatasourcesByStatus: map[string]int{
							"active": 1,
						},
						TotalDocumentsFetchedByDatasourceType: map[string]int{
							"confluence": 100,
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
  "embedding_providers": {
    "total_providers": 2,
    "total_active_providers": 1,
    "active_provider": {
      "name": "OpenAI",
      "type": "Embedding",
      "model": "text-embedding-ada-002"
    }
  },
  "llm_providers": {
    "total_providers": 3,
    "total_active_providers": 1,
    "active_provider": {
      "name": "OpenAI",
      "type": "LLM",
      "model": "gpt-3.5-turbo"
    }
  },
  "datasources": {
    "configured_datasources": [
      {
        "name": "Confluence",
        "type": "confluence",
        "status": "active",
        "created_at": "2024-07-18T09:10:06.484502057Z"
      }
    ],
    "total_datasources": 1,
    "total_datasources_by_type": {
      "confluence": 1
    },
    "total_datasources_by_status": {
      "active": 1
    },
    "total_documents_fetched_by_datasource_type": {
      "confluence": 100
    }
  }
}`,
		},
		{
			name: "Internal Server Error",
			mockSetup: func(m *storage.StorageMock) {
				m.On("GetAnalyticsOverview", mock.Anything).Return(types.AnalyticsOverview{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error", "message":"Failed to get analytics overview"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _ = observability.InitTracer(context.Background(), "test", logrus.New(), &config.Config{})

			req, err := http.NewRequest("GET", "/api/v1/analytics/overview", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			mockStorage := new(storage.StorageMock)
			tc.mockSetup(mockStorage)

			logger := logrus.New()

			a := NewManager(mockStorage, logger, auth.NewACLManager(logger, false))
			handler := http.HandlerFunc(a.getAnalyticsOverview)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.JSONEq(t, tc.expectedBody, rr.Body.String())
			mockStorage.AssertExpectations(t)
		})
	}
}
