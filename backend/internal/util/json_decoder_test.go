package util

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeJSONBody(t *testing.T) {
	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name           string
		jsonBody       string
		expectedResult testStruct
		expectedError  string
	}{
		{
			name:           "Valid JSON",
			jsonBody:       `{"name": "John Doe", "age": 30}`,
			expectedResult: testStruct{Name: "John Doe", Age: 30},
			expectedError:  "",
		},
		{
			name:           "Unknown field",
			jsonBody:       `{"name": "John Doe", "age": 30, "unknown": "field"}`,
			expectedResult: testStruct{},
			expectedError:  JsonErroFoundUnknownField,
		},
		{
			name:           "Incorrect type",
			jsonBody:       `{"name": "John Doe", "age": "thirty"}`,
			expectedResult: testStruct{},
			expectedError:  "Failed to decode JSON: field 'age' has incorrect type (expected int)",
		},
		{
			name:           "Invalid JSON",
			jsonBody:       `{"name": "John Doe", "age": 30,}`,
			expectedResult: testStruct{},
			expectedError:  "failed to decode JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tt.jsonBody))
			req.Header.Set("Content-Type", "application/json")

			var result testStruct
			err := DecodeJSONBody(req, &result)

			if tt.expectedError == "" {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedResult, result)
			} else {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestDecodeJSONBodyWithNilPointer(t *testing.T) {
	jsonBody := `{"name": "John Doe", "age": 30}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	err := DecodeJSONBody(req, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to decode JSON")
}

func TestDecodeJSONBodyWithEmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")

	var result struct{}
	err := DecodeJSONBody(req, &result)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to decode JSON")
}
