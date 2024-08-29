package document

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/config"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDocumentsHandler(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - Fully populated documents",
			queryParams: map[string]string{
				"limit": "2",
				"page":  "1",
			},
			mockSetup: func(ms *storage.StorageMock) {
				docUUID1 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				docUUID2 := uuid.MustParse("223e4567-e89b-12d3-a456-426614174000")
				sourceUUID := uuid.MustParse("323e4567-e89b-12d3-a456-426614174000")
				embeddingProviderUUID := uuid.MustParse("423e4567-e89b-12d3-a456-426614174000")
				createdAt := time.Date(2023, 5, 15, 10, 0, 0, 0, time.UTC)
				updatedAt := time.Date(2023, 5, 15, 11, 0, 0, 0, time.UTC)
				fetchedAt := time.Date(2023, 5, 15, 9, 0, 0, 0, time.UTC)
				generatedAt := time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC)

				ms.On("Get",
					mock.Anything,
					types.DocumentFilter{},
					types.DocumentFilterOption{Limit: 2, Page: 1}).
					Return(types.PaginatedDocuments{
						Documents: []types.Document{
							{
								UUID:  docUUID1,
								URL:   &url.URL{Scheme: "https", Host: "example.com", Path: "/document1"},
								Title: "Fully Populated Document 1",
								Body:  "This is the full body of document 1 with all fields populated.",
								Embedding: types.Embedding{
									Embedding: []types.ContentPart{
										{
											Content:                   "Content part 1",
											Embedding:                 []float32{0.1, 0.2, 0.3},
											EmbeddingProviderUUID:     embeddingProviderUUID,
											EmbeddingPromptTotalToken: 100,
											GeneratedAt:               generatedAt,
										},
									},
								},
								Metadata: []types.Metadata{
									{Key: "author", Value: "John Doe"},
									{Key: "category", Value: "Test"},
								},
								Status:    types.DocumentStatusReadyToSearch,
								CreatedAt: createdAt,
								UpdatedAt: updatedAt,
								FetchedAt: fetchedAt,
								Source: types.Source{
									UUID:       sourceUUID,
									Name:       "Test Source",
									SourceType: types.DatasourceTypeSlack,
								},
							},
							{
								UUID:  docUUID2,
								URL:   &url.URL{Scheme: "https", Host: "example.com", Path: "/document2"},
								Title: "Fully Populated Document 2",
								Body:  "This is the full body of document 2 with all fields populated.",
								Embedding: types.Embedding{
									Embedding: []types.ContentPart{
										{
											Content:                   "Content part 2",
											Embedding:                 []float32{0.4, 0.5, 0.6},
											EmbeddingProviderUUID:     embeddingProviderUUID,
											EmbeddingPromptTotalToken: 120,
											GeneratedAt:               generatedAt,
										},
									},
								},
								Metadata: []types.Metadata{
									{Key: "author", Value: "Jane Smith"},
									{Key: "category", Value: "Example"},
								},
								Status:    types.DocumentStatusProcessing,
								CreatedAt: createdAt.Add(time.Hour),
								UpdatedAt: updatedAt.Add(time.Hour),
								FetchedAt: fetchedAt.Add(time.Hour),
								Source: types.Source{
									UUID:       sourceUUID,
									Name:       "Test Source",
									SourceType: types.DatasourceTypeSlack,
								},
							},
						},
						Total:      2,
						Page:       1,
						PerPage:    2,
						TotalPages: 1,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"documents": [
					{
						"uuid": "123e4567-e89b-12d3-a456-426614174000",
						"url": "https://example.com/document1",
						"title": "Fully Populated Document 1",
						"body": "This is the full body of document 1 with all fields populated.",
						"embedding": {
							"embedding": [
								{
									"content": "Content part 1",
									"embedding": [0.1, 0.2, 0.3],
									"embedding_provider_uuid": "423e4567-e89b-12d3-a456-426614174000",
									"embedding_prompt_token": 100,
									"generated_at": "2023-05-15T10:30:00Z"
								}
							]
						},
						"metadata": [
							{"key": "author", "value": "John Doe"},
							{"key": "category", "value": "Test"}
						],
						"status": "ready_to_search",
						"created_at": "2023-05-15T10:00:00Z",
						"updated_at": "2023-05-15T11:00:00Z",
						"fetched_at": "2023-05-15T09:00:00Z",
						"source": {
							"uuid": "323e4567-e89b-12d3-a456-426614174000",
							"name": "Test Source",
							"type": "slack"
						}
					},
					{
						"uuid": "223e4567-e89b-12d3-a456-426614174000",
						"url": "https://example.com/document2",
						"title": "Fully Populated Document 2",
						"body": "This is the full body of document 2 with all fields populated.",
						"embedding": {
							"embedding": [
								{
									"content": "Content part 2",
									"embedding": [0.4, 0.5, 0.6],
									"embedding_provider_uuid": "423e4567-e89b-12d3-a456-426614174000",
									"embedding_prompt_token": 120,
									"generated_at": "2023-05-15T10:30:00Z"
								}
							]
						},
						"metadata": [
							{"key": "author", "value": "Jane Smith"},
							{"key": "category", "value": "Example"}
						],
						"status": "processing",
						"created_at": "2023-05-15T11:00:00Z",
						"updated_at": "2023-05-15T12:00:00Z",
						"fetched_at": "2023-05-15T10:00:00Z",
						"source": {
							"uuid": "323e4567-e89b-12d3-a456-426614174000",
							"name": "Test Source",
							"type": "slack"
						}
					}
				],
				"total": 2,
				"page": 1,
				"per_page": 2,
				"total_pages": 1
			}`,
		},
		{
			name:        "Success - Default params",
			queryParams: map[string]string{},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("Get",
					mock.Anything,
					types.DocumentFilter{},
					types.DocumentFilterOption{Limit: 10, Page: 1}).
					Return(types.PaginatedDocuments{
						Documents: []types.Document{
							{
								UUID:  uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
								Title: "Document 1",
								Source: types.Source{
									UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-526614174000"),
									Name:       "Slack Default",
									SourceType: types.DatasourceTypeSlack,
								},
								URL: &url.URL{
									Scheme: "http",
									Host:   "example.com",
								},
							},
							{
								UUID:  uuid.MustParse("223e4567-e89b-12d3-a456-426614174000"),
								Title: "Document 2",
								Source: types.Source{
									UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-526614174000"),
									Name:       "Slack Default",
									SourceType: types.DatasourceTypeSlack,
								},
								URL: &url.URL{
									Scheme: "http",
									Host:   "example.com",
								},
							},
						},
						Total:      2,
						Page:       1,
						PerPage:    10,
						TotalPages: 1,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
  "documents": [
    {
      "body": "",
      "created_at": "0001-01-01T00:00:00Z",
      "embedding": {
        "embedding": null
      },
      "fetched_at": "0001-01-01T00:00:00Z",
      "metadata": null,
      "source": {
        "name": "Slack Default",
        "type": "slack",
        "uuid": "123e4567-e89b-12d3-a456-526614174000"
      },
      "status": "",
      "title": "Document 1",
      "updated_at": "0001-01-01T00:00:00Z",
      "url": "http://example.com",
      "uuid": "123e4567-e89b-12d3-a456-426614174000"
    },
    {
      "body": "",
      "created_at": "0001-01-01T00:00:00Z",
      "embedding": {
        "embedding": null
      },
      "fetched_at": "0001-01-01T00:00:00Z",
      "metadata": null,
      "source": {
        "name": "Slack Default",
        "type": "slack",
        "uuid": "123e4567-e89b-12d3-a456-526614174000"
      },
      "status": "",
      "title": "Document 2",
      "updated_at": "0001-01-01T00:00:00Z",
      "url": "http://example.com",
      "uuid": "223e4567-e89b-12d3-a456-426614174000"
    }
  ],
  "page": 1,
  "per_page": 10,
  "total": 2,
  "total_pages": 1
}`,
		},
		{
			name: "Success - Custom params",
			queryParams: map[string]string{
				"uuid":   "123e4567-e89b-12d3-a456-426614174000",
				"status": "pending",
				"limit":  "5",
				"page":   "2",
			},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("Get",
					mock.Anything,
					types.DocumentFilter{UUID: "123e4567-e89b-12d3-a456-426614174000", Status: types.DocumentStatusPending},
					types.DocumentFilterOption{Limit: 5, Page: 2}).
					Return(types.PaginatedDocuments{
						Documents: []types.Document{
							{
								UUID:  uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
								Title: "Document 1", Status: types.DocumentStatusPending,
								Source: types.Source{
									UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-526614174000"),
									Name:       "Slack Default",
									SourceType: types.DatasourceTypeSlack,
								},
								URL: &url.URL{
									Scheme: "http",
									Host:   "example.com",
								},
							},
						},
						Total:      1,
						Page:       2,
						PerPage:    5,
						TotalPages: 1,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
  "documents": [
    {
      "body": "",
      "created_at": "0001-01-01T00:00:00Z",
      "embedding": {
        "embedding": null
      },
      "fetched_at": "0001-01-01T00:00:00Z",
      "metadata": null,
      "source": {
        "name": "Slack Default",
        "type": "slack",
        "uuid": "123e4567-e89b-12d3-a456-526614174000"
      },
      "status": "pending",
      "title": "Document 1",
      "updated_at": "0001-01-01T00:00:00Z",
      "url": "http://example.com",
      "uuid": "123e4567-e89b-12d3-a456-426614174000"
    }
  ],
  "page": 2,
  "per_page": 5,
  "total": 1,
  "total_pages": 1
}`,
		},
		{
			name:        "Success - Empty result",
			queryParams: map[string]string{},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("Get",
					mock.Anything,
					types.DocumentFilter{},
					types.DocumentFilterOption{Limit: 10, Page: 1}).
					Return(types.PaginatedDocuments{
						Documents:  []types.Document{},
						Total:      0,
						Page:       1,
						PerPage:    10,
						TotalPages: 0,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"documents":[],"total":0,"page":1,"per_page":10,"total_pages":0}`,
		},
		{
			name:        "Error - storage error",
			queryParams: map[string]string{},
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("Get",
					mock.Anything,
					types.DocumentFilter{},
					types.DocumentFilterOption{Limit: 10, Page: 1}).
					Return(types.PaginatedDocuments{}, errors.New("storage error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to fetch documents"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			tt.mockSetup(mockStorage)

			_, _ = observability.InitTracer(context.Background(), "test-service", logrus.New(), &config.Config{})

			dm := NewManager(mockStorage, logrus.New())
			handler := dm.GetDocumentsHandler()

			req := httptest.NewRequest("GET", "/api/v1/documents", nil)
			q := req.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestGetDocumentHandler(t *testing.T) {
	tests := []struct {
		name           string
		uuid           string
		mockSetup      func(*storage.StorageMock)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success - Fully populated document",
			uuid: "123e4567-e89b-12d3-a456-426614174000",
			mockSetup: func(ms *storage.StorageMock) {
				docUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				sourceUUID := uuid.MustParse("223e4567-e89b-12d3-a456-426614174000")
				embeddingProviderUUID := uuid.MustParse("323e4567-e89b-12d3-a456-426614174000")
				createdAt := time.Date(2023, 5, 15, 10, 0, 0, 0, time.UTC)
				updatedAt := time.Date(2023, 5, 15, 11, 0, 0, 0, time.UTC)
				fetchedAt := time.Date(2023, 5, 15, 9, 0, 0, 0, time.UTC)
				generatedAt := time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC)

				ms.On("Get",
					mock.Anything,
					types.DocumentFilter{UUID: "123e4567-e89b-12d3-a456-426614174000"},
					types.DocumentFilterOption{Limit: 1, Page: 1}).
					Return(types.PaginatedDocuments{
						Documents: []types.Document{
							{
								UUID:  docUUID,
								URL:   &url.URL{Scheme: "https", Host: "example.com", Path: "/document"},
								Title: "Fully Populated Document",
								Body:  "This is the full body of the document with all fields populated.",
								Embedding: types.Embedding{
									Embedding: []types.ContentPart{
										{
											Content:                   "Content part 1",
											Embedding:                 []float32{0.1, 0.2, 0.3},
											EmbeddingProviderUUID:     embeddingProviderUUID,
											EmbeddingPromptTotalToken: 100,
											GeneratedAt:               generatedAt,
										},
									},
								},
								Metadata: []types.Metadata{
									{Key: "author", Value: "John Doe"},
									{Key: "category", Value: "Test"},
								},
								Status:    types.DocumentStatusReadyToSearch,
								CreatedAt: createdAt,
								UpdatedAt: updatedAt,
								FetchedAt: fetchedAt,
								Source: types.Source{
									UUID:       sourceUUID,
									Name:       "Test Source",
									SourceType: types.DatasourceTypeSlack,
								},
							},
						},
						Total:      1,
						Page:       1,
						PerPage:    1,
						TotalPages: 1,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"uuid": "123e4567-e89b-12d3-a456-426614174000",
				"url": "https://example.com/document",
				"title": "Fully Populated Document",
				"body": "This is the full body of the document with all fields populated.",
				"embedding": {
					"embedding": [
						{
							"content": "Content part 1",
							"embedding": [0.1, 0.2, 0.3],
							"embedding_provider_uuid": "323e4567-e89b-12d3-a456-426614174000",
							"embedding_prompt_token": 100,
							"generated_at": "2023-05-15T10:30:00Z"
						}
					]
				},
				"metadata": [
					{"key": "author", "value": "John Doe"},
					{"key": "category", "value": "Test"}
				],
				"status": "ready_to_search",
				"created_at": "2023-05-15T10:00:00Z",
				"updated_at": "2023-05-15T11:00:00Z",
				"fetched_at": "2023-05-15T09:00:00Z",
				"source": {
					"uuid": "223e4567-e89b-12d3-a456-426614174000",
					"name": "Test Source",
					"type": "slack"
				}
			}`,
		},
		{
			name: "Success - Minimal document",
			uuid: "123e4567-e89b-12d3-a456-426614174000",
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("Get",
					mock.Anything,
					types.DocumentFilter{UUID: "123e4567-e89b-12d3-a456-426614174000"},
					types.DocumentFilterOption{Limit: 1, Page: 1}).
					Return(types.PaginatedDocuments{
						Documents: []types.Document{
							{
								UUID:  uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
								Title: "Document 1",
								Source: types.Source{
									UUID:       uuid.MustParse("123e4567-e89b-12d3-a456-526614174000"),
									Name:       "Slack Default",
									SourceType: types.DatasourceTypeSlack,
								},
								URL: &url.URL{
									Scheme: "http",
									Host:   "example.com",
								},
							},
						},
						Total:      1,
						Page:       1,
						PerPage:    1,
						TotalPages: 1,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `{
  "body": "",
  "created_at": "0001-01-01T00:00:00Z",
  "embedding": {
    "embedding": null
  },
  "fetched_at": "0001-01-01T00:00:00Z",
  "metadata": null,
  "source": {
        "name": "Slack Default",
        "type": "slack",
        "uuid": "123e4567-e89b-12d3-a456-526614174000"
      },
  "status": "",
  "title": "Document 1",
  "updated_at": "0001-01-01T00:00:00Z",
  "url": "http://example.com",
  "uuid": "123e4567-e89b-12d3-a456-426614174000"
}`,
		},
		{
			name: "Not Found",
			uuid: "nonexistent",
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("Get",
					mock.Anything,
					types.DocumentFilter{UUID: "nonexistent"},
					types.DocumentFilterOption{Limit: 1, Page: 1}).
					Return(types.PaginatedDocuments{
						Documents:  []types.Document{},
						Total:      0,
						Page:       1,
						PerPage:    1,
						TotalPages: 0,
					}, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Document not found"}`,
		},
		{
			name: "Error - storage error",
			uuid: "123e4567-e89b-12d3-a456-426614174000",
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("Get",
					mock.Anything,
					types.DocumentFilter{UUID: "123e4567-e89b-12d3-a456-426614174000"},
					types.DocumentFilterOption{Limit: 1, Page: 1}).
					Return(types.PaginatedDocuments{}, errors.New("storage error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to fetch document"}`,
		},
		{
			name: "Error - Unexpected multiple documents",
			uuid: "123e4567-e89b-12d3-a456-426614174000",
			mockSetup: func(ms *storage.StorageMock) {
				ms.On("Get",
					mock.Anything,
					types.DocumentFilter{UUID: "123e4567-e89b-12d3-a456-426614174000"},
					types.DocumentFilterOption{Limit: 1, Page: 1}).
					Return(types.PaginatedDocuments{
						Documents: []types.Document{
							{UUID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), Title: "Document 1"},
							{UUID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), Title: "Document 1 Duplicate"},
						},
						Total:      2,
						Page:       1,
						PerPage:    1,
						TotalPages: 2,
					}, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Unexpected error: multiple documents found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(storage.StorageMock)
			tt.mockSetup(mockStorage)

			dm := NewManager(mockStorage, logrus.New())
			handler := dm.GetDocumentHandler()

			req := httptest.NewRequest("GET", "/api/v1/document/"+tt.uuid, nil)
			w := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("uuid", tt.uuid)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.JSONEq(t, tt.expectedBody, w.Body.String())

			mockStorage.AssertExpectations(t)
		})
	}
}
