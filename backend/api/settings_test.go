package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shaharia-lab/smarty-pants/internal/logger"
	"github.com/shaharia-lab/smarty-pants/internal/storage"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSettingsHandler(t *testing.T) {
	tests := []struct {
		name           string
		mockSettings   types.Settings
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			mockSettings: types.Settings{
				General: types.GeneralSettings{ApplicationName: "smarty-pants-ai"},
				Debugging: types.DebuggingSettings{
					LogLevel:  logger.LevelInfo,
					LogFormat: logger.FormatJSON,
					LogOutput: logger.OutputStderr,
				},
				Search: types.SearchSettings{PerPage: 10},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"general":{"application_name":"smarty-pants-ai"},"debugging":{"log_level":"info","log_format":"json","log_output":"stderr"},"search":{"per_page":10}}`,
		},
		{
			name:           "Internal Server Error",
			mockSettings:   types.Settings{},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to fetch settings"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			mockStorage.On("GetSettings", mock.Anything).Return(tt.mockSettings, tt.mockError)

			l := logrus.New()
			l.Out = &bytes.Buffer{}

			handler := getSettingsHandler(mockStorage, l)

			req, err := http.NewRequest("GET", "/settings", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestUpdateSettingsHandler(t *testing.T) {
	tests := []struct {
		name           string
		inputBody      string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			inputBody:      `{"general":{"application_name":"smarty-pants-ai-update"},"debugging":{"log_level":"info","log_format":"json","log_output":"stderr"},"search":{"per_page":10}}`,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"general":{"application_name":"smarty-pants-ai-update"},"debugging":{"log_level":"info","log_format":"json","log_output":"stderr"},"search":{"per_page":10}}`,
		},
		{
			name:           "Bad Request - Invalid JSON",
			inputBody:      `{invalid json}`,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Failed to decode request body"}`,
		},
		{
			name:           "Internal Server Error",
			inputBody:      `{"general":{"log_level":"info"},"search":{"per_page":15}}`,
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to update settings"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			mockStorage.On("UpdateSettings", mock.Anything, mock.AnythingOfType("types.Settings")).Return(tt.mockError)

			l := logrus.New()
			l.Out = &bytes.Buffer{}

			handler := updateSettingsHandler(mockStorage, l)

			req, err := http.NewRequest("PUT", "/settings", bytes.NewBufferString(tt.inputBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedBody, rr.Body.String())

			if tt.expectedStatus == http.StatusOK {
				mockStorage.AssertExpectations(t)
			}
		})
	}
}
