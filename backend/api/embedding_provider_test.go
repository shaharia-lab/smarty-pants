package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddEmbeddingProviderHandler(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockBehavior   func(*storage.StorageMock)
		expectedStatus int
		expectedBody   map[string]interface{}
		expectError    bool
	}{
		{
			name: "Successful creation",
			requestBody: map[string]interface{}{
				"name":     "Test Provider",
				"provider": types.EmbeddingProviderTypeOpenAI,
				"configuration": map[string]interface{}{
					"api_key":  "test-key",
					"model_id": "test-model",
				},
			},
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("CreateEmbeddingProvider", mock.Anything, mock.AnythingOfType("types.EmbeddingProviderConfig")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"name":     "Test Provider",
				"provider": "openai",
				"configuration": map[string]interface{}{
					"api_key":  "test-key",
					"model_id": "test-model",
				},
				"status": "",
			},
			expectError: false,
		},
		{
			name:        "Invalid request body",
			requestBody: map[string]interface{}{},
			mockBehavior: func(ms *storage.StorageMock) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(mockStorage)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/embedding-provider", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			handler := addEmbeddingProviderHandler(mockStorage, logger)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var responseBody map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
			assert.NoError(t, err, "Failed to unmarshal response body")

			if !tt.expectError {

				assert.NotEmpty(t, responseBody["uuid"], "UUID should not be empty")
				delete(responseBody, "uuid")
			}

			assert.Equal(t, tt.expectedBody, responseBody)

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestUpdateEmbeddingProviderHandler(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()

	tests := []struct {
		name           string
		uuid           string
		requestBody    map[string]interface{}
		mockBehavior   func(*storage.StorageMock)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful update",
			uuid: "123e4567-e89b-12d3-a456-426614174000",
			requestBody: map[string]interface{}{
				"name":     "Updated Provider",
				"provider": types.EmbeddingProviderTypeOpenAI,
				"configuration": map[string]interface{}{
					"api_key":  "updated-key",
					"model_id": "updated-model",
				},
			},
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("UpdateEmbeddingProvider", mock.Anything, mock.AnythingOfType("types.EmbeddingProviderConfig")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"uuid":     "123e4567-e89b-12d3-a456-426614174000",
				"name":     "Updated Provider",
				"provider": "openai",
				"configuration": map[string]interface{}{
					"api_key":  "updated-key",
					"model_id": "updated-model",
				},
				"status": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(mockStorage)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/api/v1/embedding-provider/"+tt.uuid, bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler := updateEmbeddingProviderHandler(mockStorage, logger)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var responseBody map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
			assert.NoError(t, err, "Failed to unmarshal response body")

			assert.Equal(t, tt.expectedBody, responseBody)

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestDeleteEmbeddingProviderHandler(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()

	tests := []struct {
		name           string
		uuid           string
		mockBehavior   func(*storage.StorageMock)
		expectedStatus int
	}{
		{
			name: "Successful deletion",
			uuid: "123e4567-e89b-12d3-a456-426614174000",
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("DeleteEmbeddingProvider", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(mockStorage)

			req, _ := http.NewRequest("DELETE", "/api/v1/embedding-provider/"+tt.uuid, nil)
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler := deleteEmbeddingProviderHandler(mockStorage, logger)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestGetEmbeddingProviderHandler(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()

	tests := []struct {
		name           string
		uuid           string
		mockBehavior   func(*storage.StorageMock)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful retrieval",
			uuid: "123e4567-e89b-12d3-a456-426614174000",
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("GetEmbeddingProvider", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(
					&types.EmbeddingProviderConfig{
						UUID:     uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
						Name:     "Test Provider",
						Provider: types.EmbeddingProviderTypeOpenAI,
						Configuration: &types.OpenAISettings{
							APIKey:  "test-key",
							ModelID: "test-model",
						},
						Status: "",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"uuid":     "123e4567-e89b-12d3-a456-426614174000",
				"name":     "Test Provider",
				"provider": "openai",
				"configuration": map[string]interface{}{
					"api_key":  "test-key",
					"model_id": "test-model",
				},
				"status": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(mockStorage)

			req, _ := http.NewRequest("GET", "/api/v1/embedding-provider/"+tt.uuid, nil)
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler := getEmbeddingProviderHandler(mockStorage, logger)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var responseBody map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
			assert.NoError(t, err, "Failed to unmarshal response body")

			assert.Equal(t, tt.expectedBody, responseBody)

			mockStorage.AssertExpectations(t)
		})
	}
}
func TestSetActiveEmbeddingProviderHandler(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		mockBehavior   func(*storage.StorageMock)
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name: "Successful activation",
			uuid: "123e4567-e89b-12d3-a456-426614174000",
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("SetActiveEmbeddingProvider", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]string{
				"message": "Embedding provider activated successfully",
			},
		},
		{
			name: invalidUUIDMsg,
			uuid: "invalid-uuid",
			mockBehavior: func(ms *storage.StorageMock) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]string{
				"error": invalidUUIDMsg,
			},
		},
		{
			name: "storage error",
			uuid: "123e4567-e89b-12d3-a456-426614174000",
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("SetActiveEmbeddingProvider", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(errors.New("storage error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]string{
				"message": "Failed to set active embedding provider",
				"error":   "storage error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockStorage := new(storage.StorageMock)
			tt.mockBehavior(mockStorage)

			logger := logrus.New()
			logger.Out = io.Discard

			req, _ := http.NewRequest("POST", "/api/v1/embedding-provider/"+tt.uuid+"/activate", nil)
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler := setActiveEmbeddingProviderHandler(mockStorage, logger)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "DocumentStatus code mismatch")

			var responseBody map[string]string
			err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
			assert.NoError(t, err, "Failed to unmarshal response body")
			assert.Equal(t, tt.expectedBody, responseBody, "Response body mismatch")

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestGetEmbeddingProvidersHandler(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()
	logger.Out = io.Discard

	tests := []struct {
		name           string
		queryParams    string
		mockBehavior   func(*storage.StorageMock)
		expectedStatus int
		expectedBody   *types.PaginatedEmbeddingProviders
	}{
		{
			name:        "Successful retrieval with default pagination",
			queryParams: "",
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("GetAllEmbeddingProviders", mock.Anything,
					types.EmbeddingProviderFilter{},
					types.EmbeddingProviderFilterOption{Page: 1, Limit: 10}).
					Return(&types.PaginatedEmbeddingProviders{
						EmbeddingProviders: []types.EmbeddingProviderConfig{
							{
								UUID:     uuid.New(),
								Name:     "Provider1",
								Provider: types.EmbeddingProviderTypeOpenAI,
								Configuration: &types.OpenAISettings{
									APIKey:  "test-key",
									ModelID: "test-model",
								},
								Status: "active",
							},
						},
						Total:      1,
						Page:       1,
						PerPage:    10,
						TotalPages: 1,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &types.PaginatedEmbeddingProviders{
				EmbeddingProviders: []types.EmbeddingProviderConfig{
					{
						UUID:     uuid.New(),
						Name:     "Provider1",
						Provider: types.EmbeddingProviderTypeOpenAI,
						Configuration: &types.OpenAISettings{
							APIKey:  "test-key",
							ModelID: "test-model",
						},
						Status: "active",
					},
				},
				Total:      1,
				Page:       1,
				PerPage:    10,
				TotalPages: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(mockStorage)

			req, _ := http.NewRequest("GET", "/api/v1/embedding-providers"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			handler := getEmbeddingProvidersHandler(mockStorage, logger)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "DocumentStatus code mismatch")

			if tt.expectedBody != nil {
				var responseBody types.PaginatedEmbeddingProviders
				err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
				assert.NoError(t, err, "Failed to unmarshal response body")
				assert.Equal(t, tt.expectedBody.Total, responseBody.Total, "Total mismatch")
				assert.Equal(t, tt.expectedBody.Page, responseBody.Page, "Page mismatch")
				assert.Equal(t, tt.expectedBody.PerPage, responseBody.PerPage, "PerPage mismatch")
				assert.Equal(t, tt.expectedBody.TotalPages, responseBody.TotalPages, "TotalPages mismatch")
				assert.Len(t, responseBody.EmbeddingProviders, len(tt.expectedBody.EmbeddingProviders), "EmbeddingProviders length mismatch")

				if len(responseBody.EmbeddingProviders) > 0 && len(tt.expectedBody.EmbeddingProviders) > 0 {
					expectedProvider := tt.expectedBody.EmbeddingProviders[0]
					actualProvider := responseBody.EmbeddingProviders[0]
					assert.Equal(t, expectedProvider.Name, actualProvider.Name, "Provider name mismatch")
					assert.Equal(t, expectedProvider.Provider, actualProvider.Provider, "Provider type mismatch")
					assert.Equal(t, expectedProvider.Status, actualProvider.Status, "Provider status mismatch")

					expectedConfig, ok := expectedProvider.Configuration.(*types.OpenAISettings)
					assert.True(t, ok, "Expected configuration is not OpenAISettings")

					expectedJSON, err := json.Marshal(expectedConfig)
					assert.NoError(t, err, "Failed to marshal expected configuration")

					actualJSON, err := json.Marshal(actualProvider.Configuration)
					assert.NoError(t, err, "Failed to marshal actual configuration")

					assert.JSONEq(t, string(expectedJSON), string(actualJSON), "Configuration mismatch")
				}
			} else {
				var errorResponse map[string]string
				err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				assert.NoError(t, err, "Failed to unmarshal error response")
				assert.Contains(t, errorResponse["error"], "Failed to get embedding providers", "Error message mismatch")
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
