package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shaharia-lab/smarty-pants/internal/logger"
	"github.com/shaharia-lab/smarty-pants/internal/util"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Response is a generic response structure for API responses
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func sendResponse(w http.ResponseWriter, statusCode int, data interface{}, err error, logger *logrus.Logger, span trace.Span) {
	if err != nil {
		if span != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		logger.WithError(err).Error("API error")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if encodeErr := json.NewEncoder(w).Encode(data); encodeErr != nil {
		logger.WithError(encodeErr).Error("Failed to encode API response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// SendErrorResponse sends an error response with a message
func SendErrorResponse(w http.ResponseWriter, statusCode int, message string, logger *logrus.Logger, span trace.Span) {
	response := Response{
		Error: message,
	}
	sendResponse(w, statusCode, response, errors.New(message), logger, span)
}

// SendSuccessResponse sends a success response with data
func SendAPIErrorResponse(w http.ResponseWriter, statusCode int, err *util.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if encodeErr := json.NewEncoder(w).Encode(err); encodeErr != nil {
		l := logger.New(logger.Config{})
		l.WithError(encodeErr).Error("Failed to encode API response")
		_, _ = w.Write([]byte(`{"message":"Something went wrong", error": "Internal server error"}`))
	}
}

// SendSuccessResponse sends a success response with data
func SendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}, logger *logrus.Logger, span trace.Span) {
	sendResponse(w, statusCode, data, nil, logger, span)
}
