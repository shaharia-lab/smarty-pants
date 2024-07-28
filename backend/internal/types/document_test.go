package types

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestDocumentMarshalJSON(t *testing.T) {
	doc := &Document{
		UUID:  uuid.New(),
		URL:   &url.URL{Scheme: "https", Host: "example.com"},
		Title: "Test",
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "example.com") {
		t.Errorf("URL not found in JSON output")
	}
}
