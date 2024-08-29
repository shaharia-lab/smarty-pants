package llm

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

func TestAddLLMProviderHandler(t *testing.T) {
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
				"provider": types.LLMProviderTypeOpenAI,
				"configuration": map[string]interface{}{
					"api_key":  "test-key",
					"model_id": "test-model",
				},
			},
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("CreateLLMProvider", mock.Anything, mock.AnythingOfType("types.LLMProviderConfig")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"name":     "Test Provider",
				"provider": "openai",
				"configuration": map[string]interface{}{
					"api_key":  "test-key",
					"model_id": "test-model",
				},
				"status": "inactive",
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
			req, _ := http.NewRequest("POST", "/api/v1/llm-provider", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			handler := AddLLMProviderHandler(mockStorage, logger)
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

func TestUpdateLLMProviderHandler(t *testing.T) {
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
				"provider": types.LLMProviderTypeOpenAI,
				"configuration": map[string]interface{}{
					"api_key":  "updated-key",
					"model_id": "updated-model",
				},
			},
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("UpdateLLMProvider", mock.Anything, mock.AnythingOfType("types.LLMProviderConfig")).Return(nil)
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
			req, _ := http.NewRequest("PUT", "/api/v1/llm-provider/"+tt.uuid, bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler := UpdateLLMProviderHandler(mockStorage, logger)
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

func TestDeleteLLMProviderHandler(t *testing.T) {
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
				ms.On("DeleteLLMProvider", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior(mockStorage)

			req, _ := http.NewRequest("DELETE", "/api/v1/llm-provider/"+tt.uuid, nil)
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler := DeleteLLMProviderHandler(mockStorage, logger)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestGetLLMProviderHandler(t *testing.T) {
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
				ms.On("GetLLMProvider", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(
					&types.LLMProviderConfig{
						UUID:     uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
						Name:     "Test Provider",
						Provider: types.LLMProviderTypeOpenAI,
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

			req, _ := http.NewRequest("GET", "/api/v1/llm-provider/"+tt.uuid, nil)
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler := GetLLMProviderHandler(mockStorage, logger)
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

func TestGetLLMProvidersHandler(t *testing.T) {
	mockStorage := new(storage.StorageMock)
	logger := logrus.New()
	logger.Out = io.Discard

	tests := []struct {
		name           string
		queryParams    string
		mockBehavior   func(*storage.StorageMock)
		expectedStatus int
		expectedBody   *types.PaginatedLLMProviders
	}{
		{
			name:        "Successful retrieval with default pagination",
			queryParams: "",
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("GetAllLLMProviders", mock.Anything,
					types.LLMProviderFilter{},
					types.LLMProviderFilterOption{Page: 1, Limit: 10}).
					Return(&types.PaginatedLLMProviders{
						LLMProviders: []types.LLMProviderConfig{
							{
								UUID:     uuid.New(),
								Name:     "Provider1",
								Provider: types.LLMProviderTypeOpenAI,
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
			expectedBody: &types.PaginatedLLMProviders{
				LLMProviders: []types.LLMProviderConfig{
					{
						UUID:     uuid.New(),
						Name:     "Provider1",
						Provider: types.LLMProviderTypeOpenAI,
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

			req, _ := http.NewRequest("GET", "/api/v1/llm-providers"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			handler := GetLLMProvidersHandler(mockStorage, logger)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "DocumentStatus code mismatch")

			if tt.expectedBody != nil {
				var responseBody types.PaginatedLLMProviders
				err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
				assert.NoError(t, err, "Failed to unmarshal response body")
				assert.Equal(t, tt.expectedBody.Total, responseBody.Total, "Total mismatch")
				assert.Equal(t, tt.expectedBody.Page, responseBody.Page, "Page mismatch")
				assert.Equal(t, tt.expectedBody.PerPage, responseBody.PerPage, "PerPage mismatch")
				assert.Equal(t, tt.expectedBody.TotalPages, responseBody.TotalPages, "TotalPages mismatch")
				assert.Len(t, responseBody.LLMProviders, len(tt.expectedBody.LLMProviders), "LLMProviders length mismatch")

				if len(responseBody.LLMProviders) > 0 && len(tt.expectedBody.LLMProviders) > 0 {
					expectedProvider := tt.expectedBody.LLMProviders[0]
					actualProvider := responseBody.LLMProviders[0]
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
				assert.Contains(t, errorResponse["error"], "Failed to get llm providers", "Error message mismatch")
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestSetActiveLLMProviderHandler(t *testing.T) {
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
				ms.On("SetActiveLLMProvider", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]string{
				"message": "LLM provider activated successfully",
			},
		},
		{
			name: types.InvalidUUIDMessage,
			uuid: "invalid-uuid",
			mockBehavior: func(ms *storage.StorageMock) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]string{
				"error": types.InvalidUUIDMessage,
			},
		},
		{
			name: "storage error",
			uuid: "123e4567-e89b-12d3-a456-426614174000",
			mockBehavior: func(ms *storage.StorageMock) {
				ms.On("SetActiveLLMProvider", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(errors.New("storage error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]string{
				"error": "Failed to set active LLM provider",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockStorage := new(storage.StorageMock)
			tt.mockBehavior(mockStorage)

			logger := logrus.New()
			logger.Out = io.Discard

			req, _ := http.NewRequest("POST", "/api/v1/llm-provider/"+tt.uuid+"/activate", nil)
			rr := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler := SetActiveLLMProviderHandler(mockStorage, logger)
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
