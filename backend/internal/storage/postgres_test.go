package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	testCases := []struct {
		name          string
		doc           types.Document
		setupMock     func(mock sqlmock.Sqlmock)
		expectedError string
	}{
		{
			name: "Successful insertion with metadata and embeddings",
			doc: types.Document{
				UUID:      uuid.New(),
				Title:     "Test Document",
				Body:      "Test Body",
				Status:    types.DocumentStatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				URL:       func() *url.URL { u, _ := url.Parse("http://example.com"); return u }(),
				Source: types.Source{
					UUID:       uuid.New(),
					Name:       "Test Source",
					SourceType: types.DatasourceTypeSlack,
				},
				Metadata: []types.Metadata{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
				Embedding: types.Embedding{
					Embedding: []types.ContentPart{
						{
							Content:                   "Test content",
							Embedding:                 []float32{0.1, 0.2, 0.3},
							EmbeddingProviderUUID:     uuid.New(),
							GeneratedAt:               time.Now(),
							EmbeddingPromptTotalToken: 10,
						},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO documents").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO content_parts")
				mock.ExpectPrepare("INSERT INTO embeddings")
				mock.ExpectQuery("INSERT INTO content_parts").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectExec("INSERT INTO embeddings").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO metadata").WillReturnResult(sqlmock.NewResult(1, 2))
				mock.ExpectCommit()
			},
			expectedError: "",
		},
		{
			name: "Rollback on document insertion failure",
			doc: types.Document{
				UUID:      uuid.New(),
				Title:     "Test Document",
				Body:      "Test Body",
				Status:    types.DocumentStatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				URL:       func() *url.URL { u, _ := url.Parse("http://example.com"); return u }(),
				Source: types.Source{
					UUID:       uuid.New(),
					Name:       "Test Source",
					SourceType: types.DatasourceTypeSlack,
				},
				Metadata: []types.Metadata{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
				Embedding: types.Embedding{
					Embedding: []types.ContentPart{
						{
							Content:                   "Test content",
							Embedding:                 []float32{0.1, 0.2, 0.3},
							EmbeddingProviderUUID:     uuid.New(),
							GeneratedAt:               time.Now(),
							EmbeddingPromptTotalToken: 10,
						},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO documents").WillReturnError(fmt.Errorf("insertion failed"))
				mock.ExpectRollback()
			},
			expectedError: "failed to insert document: insertion failed",
		},
		{
			name: "Rollback on content part insertion failure",
			doc: types.Document{
				UUID:      uuid.New(),
				Title:     "Test Document",
				Body:      "Test Body",
				Status:    types.DocumentStatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				URL:       func() *url.URL { u, _ := url.Parse("http://example.com"); return u }(),
				Source: types.Source{
					UUID:       uuid.New(),
					Name:       "Test Source",
					SourceType: types.DatasourceTypeSlack,
				},
				Metadata: []types.Metadata{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
				Embedding: types.Embedding{
					Embedding: []types.ContentPart{
						{
							Content:                   "Test content",
							Embedding:                 []float32{0.1, 0.2, 0.3},
							EmbeddingProviderUUID:     uuid.New(),
							GeneratedAt:               time.Now(),
							EmbeddingPromptTotalToken: 10,
						},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO documents").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO content_parts")
				mock.ExpectPrepare("INSERT INTO embeddings")
				mock.ExpectQuery("INSERT INTO content_parts").WillReturnError(fmt.Errorf("content part insertion failed"))
				mock.ExpectRollback()
			},
			expectedError: "failed to insert content part: content part insertion failed",
		},
		{
			name: "Rollback on metadata insertion failure",
			doc: types.Document{
				UUID:      uuid.New(),
				Title:     "Test Document",
				Body:      "Test Body",
				Status:    types.DocumentStatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				URL:       func() *url.URL { u, _ := url.Parse("http://example.com"); return u }(),
				Source: types.Source{
					UUID:       uuid.New(),
					Name:       "Test Source",
					SourceType: types.DatasourceTypeSlack,
				},
				Metadata: []types.Metadata{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2"},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO documents").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectPrepare("INSERT INTO content_parts")
				mock.ExpectPrepare("INSERT INTO embeddings")
				mock.ExpectExec("INSERT INTO metadata").WillReturnError(fmt.Errorf("metadata insertion failed"))
				mock.ExpectRollback()
			},
			expectedError: "failed to insert metadata: metadata insertion failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			logger := logrus.New()
			logger.SetOutput(io.Discard) // Suppress log output during tests
			postgres := &Postgres{db: db, logger: logger}

			// Setup mock expectations
			tc.setupMock(mock)

			// Execute
			err = postgres.Store(context.Background(), tc.doc)

			// Assert
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Verify that all expectations were met
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
