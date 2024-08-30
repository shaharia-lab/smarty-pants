package datasource

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/config"
	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddDatasourceHandler(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Valid Slack Datasource",
			payload: map[string]interface{}{
				"name":        "Test Slack",
				"source_type": "slack",
				"settings":    map[string]interface{}{"token": "testtoken", "channel_id": "testchannel", "workspace": "testworkspace"},
			},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("AddDatasource", mock.Anything, mock.AnythingOfType("types.DatasourceConfig")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"message": "Datasource added successfully",
				"uuid":    mock.Anything,
			},
		},
		{
			name: "Invalid Payload - Missing Name",
			payload: map[string]interface{}{
				"source_type": "github",
				"settings":    map[string]interface{}{"org": "testorg"},
			},
			mockSetup:      func(ms *storage.StorageMock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "name is required",
			},
		},
		{
			name: "Invalid Payload - Missing Source Type",
			payload: map[string]interface{}{
				"name":     "Test Invalid",
				"settings": map[string]interface{}{"org": "testorg"},
			},
			mockSetup:      func(ms *storage.StorageMock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "source_type is required",
			},
		},
		{
			name: "Invalid Payload - Unsupported Source Type",
			payload: map[string]interface{}{
				"name":        "Test Invalid",
				"source_type": "unsupported",
				"settings":    map[string]interface{}{"key": "value"},
			},
			mockSetup:      func(ms *storage.StorageMock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "unsupported source type: unsupported",
				"message": "Failed to parse settings",
			},
		},
		{
			name: "Invalid Payload - Missing Slack Token",
			payload: map[string]interface{}{
				"name":        "Test Invalid Slack",
				"source_type": "slack",
				"settings":    map[string]interface{}{"channel_id": "testchannel", "workspace": "workspace"},
			},
			mockSetup:      func(ms *storage.StorageMock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "slack token is required",
				"message": "Failed to parse settings",
			},
		},
		{
			name: "Invalid Payload - Missing Slack Workspace",
			payload: map[string]interface{}{
				"name":        "Test Invalid Slack",
				"source_type": "slack",
				"settings":    map[string]interface{}{"channel_id": "testchannel", "token": "xxxx"},
			},
			mockSetup:      func(ms *storage.StorageMock) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error":   "slack workspace is required",
				"message": "Failed to parse settings",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			logger.Out = &bytes.Buffer{}

			_, _ = observability.InitTracer(context.Background(), "test", logger, &config.Config{})

			mockStorage := new(storage.StorageMock)
			tt.mockSetup(mockStorage)

			dm := NewDatasourceManager(mockStorage, logger, auth.NewACLManager(logger, false))
			handler := http.HandlerFunc(dm.addDatasourceHandler)

			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/v1/datasources", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var responseBody map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, tt.expectedBody["message"], responseBody["message"])
				assert.NotEmpty(t, responseBody["uuid"])
			} else {
				assert.Equal(t, tt.expectedBody, responseBody)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestGetDatasourceHandler(t *testing.T) {
	tests := []struct {
		name           string
		uuid           uuid.UUID
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(types.DatasourceConfig{
					UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Name:       "Test Datasource",
					SourceType: "slack",
					Settings:   &types.SlackSettings{ChannelID: "123", Workspace: "workspace", Token: "xxxx"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
  "uuid": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Test Datasource",
  "source_type": "slack",
  "settings": {
    "token": "xxxx",
    "channel_id": "123",
    "workspace": "workspace"
  }
}`,
		},
		{
			name: "Not Found",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(types.DatasourceConfig{}, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Datasource not found"}` + "\n",
		},
		{
			name: "Server Error",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(types.DatasourceConfig{}, errors.New("internal server error"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Datasource not found"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			tt.mockSetup(mockStorage)

			l := logger.NoOpsLogger()

			dm := NewDatasourceManager(mockStorage, l, auth.NewACLManager(l, false))
			handler := http.HandlerFunc(dm.getDatasourceHandler)

			req := httptest.NewRequest("GET", "/api/v1/datasource/"+tt.uuid.String(), nil)
			w := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {

				var got, expected types.DatasourceConfig
				err := json.Unmarshal(w.Body.Bytes(), &got)
				assert.NoError(t, err)
				err = json.Unmarshal([]byte(tt.expectedBody), &expected)
				assert.NoError(t, err)
				assert.Equal(t, expected, got)
			} else {

				assert.Equal(t, tt.expectedBody, w.Body.String())
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestGetDatasourcesHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - Multiple Datasources",
			queryParams: map[string]string{"page": "1", "per_page": "10"},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetAllDatasources", mock.Anything, 1, 10).Return(&types.PaginatedDatasources{
					Datasources: []types.DatasourceConfig{
						{
							UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
							Name:       "Test Datasource 1",
							SourceType: "slack",
							Settings:   &types.SlackSettings{ChannelID: "123", Workspace: "workspace", Token: "xxxx"},
							State:      &types.SlackState{Type: "slack", NextCursor: ""},
						},
					},
					Total:      1,
					Page:       1,
					PerPage:    10,
					TotalPages: 1,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
  "datasources": [
    {
      "uuid": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Test Datasource 1",
      "source_type": "slack",
      "settings": {
        "channel_id": "123",
        "token": "xxxx",
        "workspace": "workspace"
      },
      "status": "",
      "state": {
        "type": "slack",
        "next_cursor": ""
      }
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 10,
  "total_pages": 1
}`,
		},
		{
			name:        "Success - Empty Datasources",
			queryParams: map[string]string{"page": "1", "per_page": "10"},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetAllDatasources", mock.Anything, 1, 10).Return(&types.PaginatedDatasources{
					Datasources: []types.DatasourceConfig{},
					Total:       0,
					Page:        1,
					PerPage:     10,
					TotalPages:  0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"datasources":[],"total":0,"page":1,"per_page":10,"total_pages":0}`,
		},
		{
			name:        "Error - Failed to fetch",
			queryParams: map[string]string{"page": "1", "per_page": "10"},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetAllDatasources", mock.Anything, 1, 10).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"Failed to get datasources","error":"database error"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			tt.mockSetup(mockStorage)

			_, _ = observability.InitTracer(context.Background(), "test", &logrus.Logger{}, &config.Config{})

			l := logger.NoOpsLogger()
			dm := NewDatasourceManager(mockStorage, l, auth.NewACLManager(l, false))
			handler := http.HandlerFunc(dm.getDatasourcesHandler)

			req := httptest.NewRequest("GET", "/api/v1/datasources", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if w.Code == http.StatusOK {

				var got, expected types.PaginatedDatasources
				err := json.Unmarshal(w.Body.Bytes(), &got)
				assert.NoError(t, err)
				err = json.Unmarshal([]byte(tt.expectedBody), &expected)
				assert.NoError(t, err)
				assert.Equal(t, expected, got)
			} else {

				assert.Equal(t, tt.expectedBody, w.Body.String())
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
func TestUpdateDatasourceHandler(t *testing.T) {
	tests := []struct {
		name           string
		uuid           uuid.UUID
		payload        interface{}
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			payload: map[string]interface{}{
				"channel_id": "456",
			},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(types.DatasourceConfig{
					UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Name:       "Test Datasource",
					SourceType: "slack",
					Settings:   &types.SlackSettings{ChannelID: "123", Workspace: "workspace"},
					State:      &types.SlackState{},
				}, nil)
				ms.On("UpdateDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					mock.AnythingOfType("*types.SlackSettings"),
					&types.SlackState{}).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:    "Invalid JSON",
			uuid:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			payload: `{"invalid": json`,
			mockSetup: func(ms *storage.StorageMock) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"Invalid request body","error":"invalid character 'j' looking for beginning of value"}` + "\n",
		},
		{
			name: "Datasource Not Found",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			payload: map[string]interface{}{
				"channel_id": "456",
			},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(types.DatasourceConfig{}, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"message":"Datasource not found","error":"not found"}` + "\n",
		},
		{
			name: "Update Error",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			payload: map[string]interface{}{
				"channel_id": "456",
			},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(types.DatasourceConfig{
					UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Name:       "Test Datasource",
					SourceType: "slack",
					Settings:   &types.SlackSettings{ChannelID: "123"},
					State:      &types.SlackState{},
				}, nil)
				ms.On("UpdateDatasource",
					mock.Anything,
					uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					mock.AnythingOfType("*types.SlackSettings"),
					&types.SlackState{},
				).Return(errors.New("update error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"Failed to update datasource","error":"update error"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			tt.mockSetup(mockStorage)
			l := logger.NoOpsLogger()
			dm := NewDatasourceManager(mockStorage, l, auth.NewACLManager(l, false))
			handler := http.HandlerFunc(dm.updateDatasourceHandler)

			var body []byte
			var err error
			if jsonStr, ok := tt.payload.(string); ok {
				body = []byte(jsonStr)
			} else {
				body, err = json.Marshal(tt.payload)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("PUT", "/api/v1/datasource/"+tt.uuid.String(), bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestValidateDatasourceHandler(t *testing.T) {
	tests := []struct {
		name           string
		uuid           uuid.UUID
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - Valid Slack Datasource",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(types.DatasourceConfig{
					UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Name:       "Test Slack Datasource",
					SourceType: "slack",
					Settings:   &types.SlackSettings{Token: "valid_token", ChannelID: "valid_channel", Workspace: "valid_workspace"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"result":"success"}`,
		},
		{
			name: "Error - Datasource Not Found",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")).Return(types.DatasourceConfig{}, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"message":"Datasource not found","error":"not found"}`,
		},
		{
			name: "Error - Unsupported Datasource Type",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("GetDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174002")).Return(types.DatasourceConfig{
					UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
					Name:       "Test Unsupported Datasource",
					SourceType: "unsupported",
					Settings:   &SlackDatasource{},
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"unsupported datasource type: unsupported", "message":"Unsupported datasource type"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			tt.mockSetup(mockStorage)

			l := logger.NoOpsLogger()
			dm := NewDatasourceManager(mockStorage, l, auth.NewACLManager(l, false))
			handler := http.HandlerFunc(dm.validateDatasourceHandler)

			req := httptest.NewRequest("POST", "/api/v1/datasource/"+tt.uuid.String()+"/validate", nil)
			w := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestSetDisableDatasourceHandler(t *testing.T) {
	tests := []struct {
		name           string
		uuid           uuid.UUID
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - Datasource Disabled",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("SetDisableDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Datasource has been deactivated successfully"}`,
		},
		{
			name: "Error - Invalid UUID",
			uuid: uuid.Nil,
			mockSetup: func(ms *storage.StorageMock) {
				// No mock setup needed for invalid UUID
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid UUID"}`,
		},
		{
			name: "Error - Failed to Deactivate",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("SetDisableDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")).Return(errors.New("deactivation failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"Failed to set datasource active","error":"deactivation failed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			tt.mockSetup(mockStorage)

			l := logger.NoOpsLogger()
			dm := NewDatasourceManager(mockStorage, l, auth.NewACLManager(l, false))
			handler := http.HandlerFunc(dm.setDisableDatasourceHandler)

			req := httptest.NewRequest("POST", "/api/v1/datasource/"+tt.uuid.String()+"/disable", nil)
			w := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestSetActiveDatasourceHandler(t *testing.T) {
	tests := []struct {
		name           string
		uuid           uuid.UUID
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - Datasource Activated",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("SetActiveDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Datasource has been activated successfully"}`,
		},
		{
			name: "Error - Invalid UUID",
			uuid: uuid.Nil,
			mockSetup: func(ms *storage.StorageMock) {
				// No mock setup needed for invalid UUID
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid UUID"}`,
		},
		{
			name: "Error - Failed to Activate",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("SetActiveDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")).Return(errors.New("activation failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"Failed to set datasource active","error":"activation failed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			tt.mockSetup(mockStorage)

			l := logger.NoOpsLogger()
			dm := NewDatasourceManager(mockStorage, l, auth.NewACLManager(l, false))
			handler := http.HandlerFunc(dm.setActiveDatasourceHandler)

			req := httptest.NewRequest("POST", "/api/v1/datasource/"+tt.uuid.String()+"/activate", nil)
			w := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestDeleteDatasourceHandler(t *testing.T) {
	tests := []struct {
		name           string
		uuid           uuid.UUID
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - Datasource Deleted",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("DeleteDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Datasource has been deleted successfully"}`,
		},
		{
			name: "Error - Invalid UUID",
			uuid: uuid.Nil,
			mockSetup: func(ms *storage.StorageMock) {
				// No mock setup needed for invalid UUID
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid UUID"}`,
		},
		{
			name: "Error - Failed to Delete",
			uuid: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("DeleteDatasource", mock.Anything, uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")).Return(errors.New("deletion failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"Failed to set datasource active","error":"deletion failed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			tt.mockSetup(mockStorage)

			l := logger.NoOpsLogger()
			dm := NewDatasourceManager(mockStorage, l, auth.NewACLManager(l, false))
			handler := http.HandlerFunc(dm.deleteDatasourceHandler)

			req := httptest.NewRequest("DELETE", "/api/v1/datasource/"+tt.uuid.String(), nil)
			w := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockStorage.AssertExpectations(t)
		})
	}
}
