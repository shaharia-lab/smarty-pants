package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/shaharia-lab/smarty-pants/internal/config"
	"github.com/shaharia-lab/smarty-pants/internal/observability"
	"github.com/shaharia-lab/smarty-pants/internal/types"
	"github.com/sirupsen/logrus"
)

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestOpenAILLM_GetResponse(t *testing.T) {
	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {

			if req.Header.Get("Content-Type") != "application/json" {
				t.Error("Content-Type header not set correctly")
			}
			if req.Header.Get("Authorization") != "Bearer test-api-key" {
				t.Error("Authorization header not set correctly")
			}

			body, _ := io.ReadAll(req.Body)
			var gotBody map[string]interface{}
			if err := json.Unmarshal(body, &gotBody); err != nil {
				t.Fatalf("Failed to unmarshal request body: %v", err)
			}

			expectedBody := map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"messages": []interface{}{
					map[string]interface{}{"role": "system", "content": "You are a helpful assistant."},
					map[string]interface{}{"role": "user", "content": "Test prompt"},
				},
			}

			if !reflect.DeepEqual(gotBody, expectedBody) {
				t.Errorf("Unexpected request body. Got %v, want %v", gotBody, expectedBody)
			}

			return &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"choices": [
						{
							"message": {
								"content": "Test response"
							}
						}
					]
				}`)),
			}, nil
		},
	}

	settings := &types.OpenAILLMSettings{
		APIKey:  "test-api-key",
		ModelID: "gpt-3.5-turbo",
	}

	_, _ = observability.InitTracer(nil, "test-service", &logrus.Logger{}, &config.Config{})
	llm := NewOpenAILLM(settings, mockClient, nil)

	response, err := llm.GetResponse(Prompt{Text: "Test prompt"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response != "Test response" {
		t.Errorf("Unexpected response. Got %s, want %s", response, "Test response")
	}
}
