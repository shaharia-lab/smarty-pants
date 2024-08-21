package util

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestSendErrorResponse(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard log output for tests

	tests := []struct {
		name           string
		statusCode     int
		message        string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Bad Request Error",
			statusCode:     http.StatusBadRequest,
			message:        "Invalid input",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid input"}`,
		},
		{
			name:           "Internal Server Error",
			statusCode:     http.StatusInternalServerError,
			message:        "Something went wrong",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Something went wrong"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SendErrorResponse(w, tt.statusCode, tt.message, logger, nil)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestSendAPIErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		apiError       *APIError
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Not Found Error",
			statusCode:     http.StatusNotFound,
			apiError:       &APIError{Message: "Resource not found", Err: "Not Found"},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"message":"Resource not found","error":"Not Found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SendAPIErrorResponse(w, tt.statusCode, tt.apiError)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestSendSuccessResponse(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard log output for tests

	tests := []struct {
		name           string
		statusCode     int
		data           interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success with data",
			statusCode:     http.StatusOK,
			data:           map[string]string{"message": "Success"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Success"}`,
		},
		{
			name:           "Created without data",
			statusCode:     http.StatusCreated,
			data:           nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SendSuccessResponse(w, tt.statusCode, tt.data, logger, nil)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

// Helper function to create a mock span for testing
func createMockSpan() trace.Span {
	return trace.SpanFromContext(context.Background())
}
