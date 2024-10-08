// Package storage provides a storage system for the application.
package storage

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const (
	failedToBeginTxnErrFmt      = "failed to begin transaction: %w"
	failedToCommitTxnErrFmt     = "failed to commit transaction: %w"
	errorIteratingOverRowErrFmt = "error iterating over rows: %w"
	failedToMarshalConfigErrFmt = "failed to marshal configuration: %w"
	failedToGetTotalCountErrFmt = "failed to get total count: %w"
)

// PostgresConfig contains configuration for a Postgres database
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// NewPostgresDB creates a new Postgres database connection
func NewPostgresDB(cfg PostgresConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Postgres is a storage system that uses a Postgres database
type Postgres struct {
	db       *sql.DB
	logger   *logrus.Logger
	migrator *migrate.Migrate
}

// NewPostgres creates a new Postgres storage system
func NewPostgres(cfg PostgresConfig, logger *logrus.Logger) (Storage, error) {
	db, err := NewPostgresDB(cfg)
	if err != nil {
		return nil, err
	}

	return &Postgres{db: db, logger: logger}, nil
}

// HealthCheck checks the health of the storage system
func (p *Postgres) HealthCheck() error {
	return nil
}

// Store the document in the database
func (p *Postgres) Store(ctx context.Context, doc types.Document) error {
	if err := doc.Validate(); err != nil {
		p.logger.WithError(err).WithField("document_uuid", doc.UUID).Error("Document validation failed")
		return fmt.Errorf("document validation failed: %w", err)
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(failedToBeginTxnErrFmt, err)
	}
	defer p.rollbackOnError(tx, doc.UUID)

	if err := p.insertDocument(ctx, tx, doc); err != nil {
		return err
	}

	insertContentPart, insertEmbedding, err := p.prepareStatements(ctx, tx)
	if err != nil {
		return err
	}
	defer p.closeStatements(insertContentPart, insertEmbedding)

	if err := p.insertContentPartsAndEmbeddings(ctx, doc, insertContentPart, insertEmbedding); err != nil {
		return err
	}

	if err := p.insertMetadata(ctx, tx, doc); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(failedToCommitTxnErrFmt, err)
	}

	p.logger.WithField("document_uuid", doc.UUID).Info("Document stored successfully")
	return nil
}

func (p *Postgres) rollbackOnError(tx *sql.Tx, uuid uuid.UUID) {
	if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		p.logger.WithError(err).WithField("document_uuid", uuid).Error("Failed to rollback transaction")
	}
}

func (p *Postgres) insertDocument(ctx context.Context, tx *sql.Tx, doc types.Document) error {
	_, err := tx.ExecContext(ctx, `
        INSERT INTO documents (uuid, title, body, status, created_at, updated_at, datasource_uuid, url)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		doc.UUID, doc.Title, doc.Body, doc.Status, doc.CreatedAt, doc.UpdatedAt, doc.Source.UUID, doc.URL.String())
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}
	return nil
}

func (p *Postgres) prepareStatements(ctx context.Context, tx *sql.Tx) (*sql.Stmt, *sql.Stmt, error) {
	insertContentPart, err := tx.PrepareContext(ctx, `
        INSERT INTO content_parts (document_uuid, content)
        VALUES ($1, $2) RETURNING id`)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare content_parts insert statement: %w", err)
	}

	insertEmbedding, err := tx.PrepareContext(ctx, `
        INSERT INTO embeddings (content_part_id, embedding, embedding_provider_id, generated_at, embedding_prompt_token)
        VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare embeddings insert statement: %w", err)
	}

	return insertContentPart, insertEmbedding, nil
}

func (p *Postgres) closeStatements(statements ...*sql.Stmt) {
	for _, stmt := range statements {
		p.logCloseError(fmt.Sprintf("%T", stmt), stmt.Close)
	}
}

func (p *Postgres) insertContentPartsAndEmbeddings(ctx context.Context, doc types.Document, insertContentPart, insertEmbedding *sql.Stmt) error {
	for _, part := range doc.Embedding.Embedding {
		var contentPartID int
		if err := insertContentPart.QueryRowContext(ctx, doc.UUID, part.Content).Scan(&contentPartID); err != nil {
			return fmt.Errorf("failed to insert content part: %w", err)
		}

		if _, err := insertEmbedding.ExecContext(ctx,
			contentPartID,
			pgvector.NewVector(part.Embedding),
			part.EmbeddingProviderUUID,
			part.GeneratedAt,
			part.EmbeddingPromptTotalToken); err != nil {
			return fmt.Errorf("failed to insert embedding: %w", err)
		}
	}
	return nil
}

func (p *Postgres) insertMetadata(ctx context.Context, tx *sql.Tx, doc types.Document) error {
	if len(doc.Metadata) == 0 {
		return nil
	}

	values := make([]string, len(doc.Metadata))
	args := make([]interface{}, len(doc.Metadata)*3)
	for i, meta := range doc.Metadata {
		values[i] = fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		args[i*3] = doc.UUID
		args[i*3+1] = meta.Key
		args[i*3+2] = meta.Value
	}
	query := fmt.Sprintf("INSERT INTO metadata (document_uuid, key, value) VALUES %s", strings.Join(values, ","))
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to insert metadata: %w", err)
	}
	return nil
}

// Helper functions
func (p *Postgres) validateDocument(doc types.Document) error {
	if doc.UUID == uuid.Nil {
		return errors.New("document UUID is required")
	}
	if doc.Title == "" {
		return errors.New("document title is required")
	}
	if doc.Source.UUID == uuid.Nil {
		return errors.New("document source UUID is required")
	}
	// Add any other validation rules as needed
	return nil
}

func (p *Postgres) logCloseError(what string, closeFunc func() error) {
	if err := closeFunc(); err != nil {
		p.logger.WithError(err).Errorf("Failed to close %s", what)
	}
}

// Update the document in the database
func (p *Postgres) Update(ctx context.Context, doc types.Document) error {
	if err := doc.Validate(); err != nil {
		p.logger.WithError(err).WithField("document_uuid", doc.UUID).Error("Document validation failed")
		return fmt.Errorf("document validation failed: %w", err)
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Always update the document's status and updated_at fields
	err = updateDocumentFields(tx, doc)
	if err != nil {
		return err
	}

	// Optionally update content parts and embeddings if they are provided
	if len(doc.Embedding.Embedding) > 0 {
		err = updateContentPartsAndEmbeddings(tx, doc)
		if err != nil {
			return err
		}
	}

	// Optionally update metadata if it is provided
	if len(doc.Metadata) > 0 {
		err = updateMetadata(tx, doc)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func updateDocumentFields(tx *sql.Tx, doc types.Document) error {
	_, err := tx.Exec(`
		UPDATE documents
		SET title = $1, status = $2, updated_at = $3
		WHERE uuid = $4`,
		doc.Title, doc.Status, doc.UpdatedAt, doc.UUID)
	return err
}

func updateContentPartsAndEmbeddings(tx *sql.Tx, doc types.Document) error {
	// Delete existing content parts and embeddings
	_, err := tx.Exec(`
		DELETE FROM embeddings
		WHERE content_part_id IN (
			SELECT id FROM content_parts WHERE document_uuid = $1
		)`, doc.UUID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM content_parts
		WHERE document_uuid = $1`, doc.UUID)
	if err != nil {
		return err
	}

	// Insert new content parts and embeddings
	for _, part := range doc.Embedding.Embedding {
		var contentPartID int
		err = tx.QueryRow(`
			INSERT INTO content_parts (document_uuid, content)
			VALUES ($1, $2) RETURNING id`,
			doc.UUID, part.Content).Scan(&contentPartID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO embeddings (content_part_id, embedding, embedding_provider_id, generated_at, embedding_prompt_token)
			VALUES ($1, $2, $3, $4, $5)`,
			contentPartID, pgvector.NewVector(part.Embedding), part.EmbeddingProviderUUID, time.Now().UTC(), part.EmbeddingPromptTotalToken)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateMetadata(tx *sql.Tx, doc types.Document) error {
	// Delete existing metadata
	_, err := tx.Exec(`
		DELETE FROM metadata
		WHERE document_uuid = $1`, doc.UUID)
	if err != nil {
		return err
	}

	// Insert new metadata
	for _, meta := range doc.Metadata {
		_, err = tx.Exec(`
			INSERT INTO metadata (document_uuid, key, value)
			VALUES ($1, $2, $3)`,
			doc.UUID, meta.Key, meta.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get the document from the database
func (p *Postgres) Get(ctx context.Context, filter types.DocumentFilter, options types.DocumentFilterOption) (types.PaginatedDocuments, error) {
	// Step 1: Get basic document information
	docs, total, err := p.getBasicDocuments(filter, options)
	if err != nil {
		return types.PaginatedDocuments{}, err
	}

	// Step 2: Fetch related data for each document
	for i, doc := range docs {
		// Get metadata
		metadata, err := p.getDocumentMetadata(doc.UUID)
		if err != nil {
			return types.PaginatedDocuments{}, err
		}
		docs[i].Metadata = metadata

		// Get content parts and embeddings
		contentParts, err := p.getDocumentContentParts(doc.UUID)
		if err != nil {
			return types.PaginatedDocuments{}, err
		}
		docs[i].Embedding.Embedding = contentParts
	}

	totalPages := (total + options.Limit - 1) / options.Limit

	return types.PaginatedDocuments{
		Documents:  docs,
		Total:      total,
		Page:       options.Page,
		PerPage:    options.Limit,
		TotalPages: totalPages,
	}, nil
}

func (p *Postgres) getBasicDocuments(filter types.DocumentFilter, options types.DocumentFilterOption) ([]types.Document, int, error) {
	query, args := p.buildQuery(filter, options)

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying documents: %w", err)
	}
	defer rows.Close()

	documents, total, err := p.scanRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

func (p *Postgres) buildQuery(filter types.DocumentFilter, options types.DocumentFilterOption) (string, []interface{}) {
	baseQuery := `
        WITH filtered_docs AS (
            SELECT d.*, s.name as source_name, s.source_type
            FROM documents d
            LEFT JOIN datasources s ON d.datasource_uuid = s.uuid
            WHERE 1=1
    `

	conditions, args := p.buildConditions(filter)
	query := baseQuery + conditions

	query += `
        ),
        count_docs AS (
            SELECT COUNT(*) as total FROM filtered_docs
        )
        SELECT 
            fd.uuid, fd.title, fd.body, fd.status, fd.url, fd.created_at, fd.updated_at, fd.fetched_at,
            fd.source_name, fd.source_type, fd.datasource_uuid,
            cd.total
        FROM filtered_docs fd
        CROSS JOIN count_docs cd
        ORDER BY fd.created_at DESC
        LIMIT $%d OFFSET $%d
    `

	limit, offset := p.calculateLimitAndOffset(options)
	args = append(args, limit, offset)
	query = fmt.Sprintf(query, len(args)-1, len(args))

	return query, args
}

func (p *Postgres) buildConditions(filter types.DocumentFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	placeholderIndex := 1

	if filter.UUID != "" {
		args = append(args, filter.UUID)
		conditions = append(conditions, fmt.Sprintf("d.uuid = $%d", placeholderIndex))
		placeholderIndex++
	}
	if filter.Status != "" {
		args = append(args, string(filter.Status))
		conditions = append(conditions, fmt.Sprintf("d.status = $%d", placeholderIndex))
		placeholderIndex++
	}
	if filter.SourceUUID != "" {
		args = append(args, filter.SourceUUID)
		conditions = append(conditions, fmt.Sprintf("d.datasource_uuid = $%d", placeholderIndex))
		placeholderIndex++
	}

	var queryConditions string
	if len(conditions) > 0 {
		queryConditions = " AND " + strings.Join(conditions, " AND ")
	}

	return queryConditions, args
}

func (p *Postgres) calculateLimitAndOffset(options types.DocumentFilterOption) (int, int) {
	limit := options.Limit
	if limit <= 0 {
		limit = 10 // Default limit
	}
	offset := (options.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func (p *Postgres) scanRows(rows *sql.Rows) ([]types.Document, int, error) {
	var documents []types.Document
	var total int

	for rows.Next() {
		doc, rowTotal, err := p.scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		documents = append(documents, doc)
		total = rowTotal
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating document rows: %w", err)
	}

	return documents, total, nil
}

func (p *Postgres) scanRow(row *sql.Rows) (types.Document, int, error) {
	var doc types.Document
	var urlString sql.NullString
	var sourceName, sourceType sql.NullString
	var total int

	err := row.Scan(
		&doc.UUID, &doc.Title, &doc.Body, &doc.Status, &urlString, &doc.CreatedAt, &doc.UpdatedAt, &doc.FetchedAt,
		&sourceName, &sourceType, &doc.Source.UUID,
		&total,
	)
	if err != nil {
		return types.Document{}, 0, fmt.Errorf("error scanning document row: %w", err)
	}

	if urlString.Valid {
		parsedURL, err := url.Parse(urlString.String)
		if err == nil {
			doc.URL = parsedURL
		}
	}

	doc.Source.Name = sourceName.String
	doc.Source.SourceType = types.DatasourceType(sourceType.String)

	return doc, total, nil
}

func (p *Postgres) getDocumentMetadata(docUUID uuid.UUID) ([]types.Metadata, error) {
	query := `
        SELECT key, value
        FROM metadata
        WHERE document_uuid = $1
    `

	rows, err := p.db.Query(query, docUUID)
	if err != nil {
		return nil, fmt.Errorf("error querying document metadata: %w", err)
	}
	defer rows.Close()

	var metadata []types.Metadata

	for rows.Next() {
		var m types.Metadata
		if err := rows.Scan(&m.Key, &m.Value); err != nil {
			return nil, fmt.Errorf("error scanning metadata row: %w", err)
		}
		metadata = append(metadata, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metadata rows: %w", err)
	}

	return metadata, nil
}

func (p *Postgres) getDocumentContentParts(docUUID uuid.UUID) ([]types.ContentPart, error) {
	query := `
        SELECT cp.content, e.embedding, e.embedding_provider_id, e.generated_at, e.embedding_prompt_token
        FROM content_parts cp
        LEFT JOIN embeddings e ON cp.id = e.content_part_id
        WHERE cp.document_uuid = $1
    `

	rows, err := p.db.Query(query, docUUID)
	if err != nil {
		return nil, fmt.Errorf("error querying document content parts: %w", err)
	}
	defer rows.Close()

	var contentParts []types.ContentPart

	for rows.Next() {
		var cp types.ContentPart
		var embedding pgvector.Vector
		var embeddingProviderID sql.NullString
		var generatedAt sql.NullTime
		var embeddingPromptToken sql.NullInt32

		err := rows.Scan(&cp.Content, &embedding, &embeddingProviderID, &generatedAt, &embeddingPromptToken)
		if err != nil {
			return nil, fmt.Errorf("error scanning content part row: %w", err)
		}

		if len(embedding.Slice()) > 0 {
			cp.Embedding = embedding.Slice()
		}

		if embeddingProviderID.Valid {
			cp.EmbeddingProviderUUID, _ = uuid.Parse(embeddingProviderID.String)
		}
		if generatedAt.Valid {
			cp.GeneratedAt = generatedAt.Time
		}
		if embeddingPromptToken.Valid {
			cp.EmbeddingPromptTotalToken = embeddingPromptToken.Int32
		}

		contentParts = append(contentParts, cp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating content part rows: %w", err)
	}

	return contentParts, nil
}

// GetForProcessing retrieves documents for processing
func (p *Postgres) GetForProcessing(ctx context.Context, _ types.DocumentFilter, batchLimit int) ([]uuid.UUID, error) {
	// Validate batchLimit
	if batchLimit <= 0 {
		return nil, fmt.Errorf("batchLimit must be greater than 0")
	}

	// Start a transaction
	tx, err := p.db.Begin()
	if err != nil {
		return nil, fmt.Errorf(failedToBeginTxnErrFmt, err)
	}
	defer tx.Rollback()

	// Construct the query
	query := `
        WITH cte AS (
            SELECT uuid
            FROM documents
            WHERE status = $1
            ORDER BY created_at
            FOR UPDATE SKIP LOCKED
            LIMIT $2
        )
        UPDATE documents d
        SET status = $3
        FROM cte
        WHERE d.uuid = cte.uuid
        RETURNING d.uuid;`

	// Execute the query
	rows, err := tx.Query(query, types.DocumentStatusPending, batchLimit, types.DocumentStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("failed to update documents: %w", err)
	}
	defer rows.Close()

	// Collect the UUIDs
	var docUUIDs []uuid.UUID
	for rows.Next() {
		var docUUID uuid.UUID
		if err := rows.Scan(&docUUID); err != nil {
			return nil, fmt.Errorf("failed to scan document UUID: %w", err)
		}
		docUUIDs = append(docUUIDs, docUUID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(errorIteratingOverRowErrFmt, err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf(failedToCommitTxnErrFmt, err)
	}

	return docUUIDs, nil
}

// AddDatasource adds a new datasource to the database
func (p *Postgres) AddDatasource(ctx context.Context, dsConfig types.DatasourceConfig) error {
	settingsJSON, err := json.Marshal(dsConfig.Settings)
	if err != nil {
		return err
	}

	stateJSON, err := json.Marshal(dsConfig.State)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO datasources (uuid, name, source_type, settings, status, state)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = p.db.Exec(query, dsConfig.UUID, dsConfig.Name, dsConfig.SourceType, settingsJSON, dsConfig.Status, stateJSON)
	return err
}

// GetDatasource retrieves a datasource from the database
func (p *Postgres) GetDatasource(ctx context.Context, uuid uuid.UUID) (types.DatasourceConfig, error) {
	var dsConfig types.DatasourceConfig
	var settingsBytes, stateBytes []byte

	query := `
        SELECT uuid, name, source_type, settings, status, state
        FROM datasources
        WHERE uuid = $1
    `
	err := p.db.QueryRow(query, uuid).Scan(
		&dsConfig.UUID,
		&dsConfig.Name,
		&dsConfig.SourceType,
		&settingsBytes,
		&dsConfig.Status,
		&stateBytes,
	)
	if err != nil {
		return types.DatasourceConfig{}, err
	}

	// Parse settings
	settings, err := util.ParseSettings(dsConfig.SourceType, settingsBytes)
	if err != nil {
		return types.DatasourceConfig{}, fmt.Errorf("failed to parse settings: %w", err)
	}
	dsConfig.Settings = settings

	// Parse state
	state, err := types.ParseDatasourceStateFromRawJSON(dsConfig.SourceType, stateBytes)
	if err != nil {
		return types.DatasourceConfig{}, fmt.Errorf("failed to parse state: %w", err)
	}
	dsConfig.State = state

	return dsConfig, nil
}

// GetAllDatasources retrieves all datasources from the database
func (p *Postgres) GetAllDatasources(ctx context.Context, page, perPage int) (*types.PaginatedDatasources, error) {
	offset := (page - 1) * perPage

	query := `
        SELECT uuid, name, source_type, settings, status, state
        FROM datasources
        ORDER BY name
        LIMIT $1 OFFSET $2
    `
	rows, err := p.db.Query(query, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasources []types.DatasourceConfig

	for rows.Next() {
		var ds types.DatasourceConfig
		var settingsBytes, stateBytes []byte

		if err := rows.Scan(&ds.UUID, &ds.Name, &ds.SourceType, &settingsBytes, &ds.Status, &stateBytes); err != nil {
			return nil, err
		}

		// Parse settings based on source type
		settings, err := util.ParseSettings(ds.SourceType, settingsBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse settings for datasource %s: %w", ds.UUID, err)
		}
		ds.Settings = settings

		// Parse state based on source type
		state, err := types.ParseDatasourceStateFromRawJSON(ds.SourceType, stateBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse state for datasource %s: %w", ds.UUID, err)
		}
		ds.State = state

		datasources = append(datasources, ds)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get total count
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM datasources"
	if err := p.db.QueryRow(countQuery).Scan(&totalCount); err != nil {
		return nil, err
	}

	totalPages := (totalCount + perPage - 1) / perPage

	return &types.PaginatedDatasources{
		Datasources: datasources,
		Total:       totalCount,
		Page:        page,
		PerPage:     perPage,
		TotalPages:  totalPages,
	}, nil
}

// UpdateDatasource updates a datasource in the database
func (p *Postgres) UpdateDatasource(ctx context.Context, uuid uuid.UUID, settings types.DatasourceSettings, state types.DatasourceState) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}

	query := `
        UPDATE datasources
        SET settings = $1, state = $2
        WHERE uuid = $3
    `
	_, err = p.db.ExecContext(ctx, query, settingsJSON, stateJSON, uuid)
	return err
}

// SetActiveDatasource sets a datasource to active status
func (p *Postgres) SetActiveDatasource(ctx context.Context, uuid uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, "UPDATE datasources SET status = $1 WHERE uuid = $2 AND status = $3", types.DatasourceStatusActive, uuid, types.DatasourceStatusInactive)
	if err != nil {
		return fmt.Errorf("failed to activate datasource: %w", err)
	}

	return nil
}

// SetDisableDatasource sets a datasource to inactive status
func (p *Postgres) SetDisableDatasource(ctx context.Context, uuid uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, "UPDATE datasources SET status = $1 WHERE uuid = $2 AND status = $3", types.DatasourceStatusInactive, uuid, types.DatasourceStatusActive)
	if err != nil {
		return fmt.Errorf("failed to deactivate datasource: %w", err)
	}

	return nil
}

// DeleteDatasource deletes a datasource from the database
func (p *Postgres) DeleteDatasource(ctx context.Context, uuid uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM datasources WHERE uuid = $1", uuid)
	if err != nil {
		return fmt.Errorf("failed to delete datasource: %w", err)
	}

	return nil
}

// CreateEmbeddingProvider creates a new embedding provider in the database
func (p *Postgres) CreateEmbeddingProvider(ctx context.Context, provider types.EmbeddingProviderConfig) error {
	configJSON, err := json.Marshal(provider.Configuration)
	if err != nil {
		return fmt.Errorf(failedToMarshalConfigErrFmt, err)
	}

	_, err = p.db.ExecContext(ctx,
		"INSERT INTO embedding_providers (id, name, provider, status, configuration) VALUES ($1, $2, $3, $4, $5)",
		provider.UUID, provider.Name, provider.Provider, "inactive", configJSON)
	if err != nil {
		return fmt.Errorf("failed to insert embedding provider: %w", err)
	}

	return nil
}

// UpdateEmbeddingProvider updates an existing embedding provider in the database
func (p *Postgres) UpdateEmbeddingProvider(ctx context.Context, provider types.EmbeddingProviderConfig) error {
	configJSON, err := json.Marshal(provider.Configuration)
	if err != nil {
		return fmt.Errorf(failedToMarshalConfigErrFmt, err)
	}

	_, err = p.db.ExecContext(ctx,
		"UPDATE embedding_providers SET name = $2, provider = $3, configuration = $4 WHERE id = $1",
		provider.UUID, provider.Name, provider.Provider, configJSON)
	if err != nil {
		return fmt.Errorf("failed to update embedding provider: %w", err)
	}

	return nil
}

// DeleteEmbeddingProvider deletes an embedding provider from the database
func (p *Postgres) DeleteEmbeddingProvider(ctx context.Context, id uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM embedding_providers WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete embedding provider: %w", err)
	}

	return nil
}

// GetEmbeddingProvider retrieves an embedding provider from the database
func (p *Postgres) GetEmbeddingProvider(ctx context.Context, id uuid.UUID) (*types.EmbeddingProviderConfig, error) {
	var provider types.EmbeddingProviderConfig
	var configJSON []byte

	err := p.db.QueryRowContext(ctx,
		"SELECT id, name, provider, status, configuration FROM embedding_providers WHERE id = $1",
		id).Scan(&provider.UUID, &provider.Name, &provider.Provider, &provider.Status, &configJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.ErrEmbeddingProviderNotFound
		}
		return nil, fmt.Errorf("failed to get embedding provider: %w", err)
	}

	settings, err := util.ParseEmbeddingProviderSettings(provider.Provider, configJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedding provider settings: %w", err)
	}

	provider.Configuration = settings

	return &provider, nil
}

// GetAllEmbeddingProviders retrieves all embedding providers from the database
func (p *Postgres) GetAllEmbeddingProviders(ctx context.Context, filter types.EmbeddingProviderFilter, option types.EmbeddingProviderFilterOption) (*types.PaginatedEmbeddingProviders, error) {
	query := "SELECT id, name, provider, status, configuration FROM embedding_providers WHERE 1=1"
	var args []interface{}
	argCount := 1

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filter.Status)
		argCount++
	}

	// Add more filter conditions here as needed

	// Count total before applying pagination
	countQuery := strings.Replace(query, "SELECT id, name, provider, status, configuration", "SELECT COUNT(*)", 1)
	var total int
	err := p.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf(failedToGetTotalCountErrFmt, err)
	}

	// Apply sorting and pagination
	query += " ORDER BY name LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, option.Limit, (option.Page-1)*option.Limit)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query embedding providers: %w", err)
	}
	defer rows.Close()

	var providers []types.EmbeddingProviderConfig
	for rows.Next() {
		var provider types.EmbeddingProviderConfig
		var configJSON []byte
		err := rows.Scan(&provider.UUID, &provider.Name, &provider.Provider, &provider.Status, &configJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan embedding provider: %w", err)
		}

		settings, err := util.ParseEmbeddingProviderSettings(provider.Provider, configJSON)
		if err != nil {
			p.logger.WithFields(logrus.Fields{
				"provider": provider.Provider,
				"uuid":     provider.UUID,
				"error":    err,
			}).Error("Failed to parse embedding provider settings")
			continue // Skip this provider but continue processing others
		}

		provider.Configuration = settings
		providers = append(providers, provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(errorIteratingOverRowErrFmt, err)
	}

	totalPages := (total + option.Limit - 1) / option.Limit

	return &types.PaginatedEmbeddingProviders{
		EmbeddingProviders: providers,
		Total:              total,
		Page:               option.Page,
		PerPage:            option.Limit,
		TotalPages:         totalPages,
	}, nil
}

// SetActiveEmbeddingProvider sets an embedding provider to active status
func (p *Postgres) SetActiveEmbeddingProvider(ctx context.Context, uuid uuid.UUID) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(failedToBeginTxnErrFmt, err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE embedding_providers SET status = 'inactive' WHERE status = 'active'")
	if err != nil {
		return fmt.Errorf("failed to deactivate current active provider: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE embedding_providers SET status = 'active' WHERE id = $1", uuid)
	if err != nil {
		return fmt.Errorf("failed to activate new provider: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(failedToCommitTxnErrFmt, err)
	}

	return nil
}

// SetDisableEmbeddingProvider sets an embedding provider to inactive status
func (p *Postgres) SetDisableEmbeddingProvider(ctx context.Context, uuid uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, "UPDATE embedding_providers SET status = $1 WHERE id = $2", types.DatasourceStatusInactive, uuid)
	if err != nil {
		return fmt.Errorf("failed to deactivate new provider: %w", err)
	}

	return nil
}

// SetActiveLLMProvider sets an LLM provider to active status
func (p *Postgres) SetActiveLLMProvider(ctx context.Context, id uuid.UUID) error {
	p.logger.WithField("llm_provider_id", id).Info("Setting LLM provider active")

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		p.logger.WithError(err).Error("Failed to begin transaction during setting LLM provider active")
		return fmt.Errorf(failedToBeginTxnErrFmt, err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE llm_providers SET status = $1 WHERE status = $2", types.LLMProviderStatusInactive, types.LLMProviderStatusActive)
	if err != nil {
		return fmt.Errorf("failed to update current providers: %w", err)
	}

	_, err = tx.ExecContext(ctx, "UPDATE llm_providers SET status = $1 WHERE id = $2", types.LLMProviderStatusActive, id)
	if err != nil {
		return fmt.Errorf("failed to activate new provider: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(failedToCommitTxnErrFmt, err)
	}

	p.logger.WithField("llm_provider_id", id).Info("LLM provider set active")

	return nil
}

// SetDisableLLMProvider sets an LLM provider to inactive status
func (p *Postgres) SetDisableLLMProvider(ctx context.Context, id uuid.UUID) error {
	p.logger.WithField("llm_provider_id", id).Info("Deactivating LLM provider")

	_, err := p.db.ExecContext(ctx, "UPDATE llm_providers SET status = $1 WHERE id = $2", types.LLMProviderStatusInactive, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate new provider: %w", err)
	}

	p.logger.WithField("llm_provider_id", id).Info("LLM provider deactivated")

	return nil
}

// RecordAIOpsUsage records AI operations usage in the database
func (p *Postgres) RecordAIOpsUsage(ctx context.Context, usage types.AIUsage) error {
	_, err := p.db.ExecContext(ctx, `
        INSERT INTO ai_ops_usage (
            ops_provider_id,
            document_id,
            input_tokens,
            output_tokens,
            dimensions,
            operation_type,
            cost_per_thousands_token,
            created_at,
            total_latency
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		usage.OpsProviderID,
		usage.DocumentID,
		usage.InputTokens,
		usage.OutputTokens,
		usage.Dimensions,
		usage.OperationType,
		usage.CostPerThousandsToken,
		usage.CreatedAt,
		usage.TotalLatency,
	)
	return err
}

// Search searches for documents in the database
func (p *Postgres) Search(ctx context.Context, config types.SearchConfig) (*types.SearchResults, error) {
	// Validate input
	if config.Limit <= 0 {
		config.Limit = 10 // Default limit
	}
	if config.Page <= 0 {
		config.Page = 1 // Default page
	}
	offset := (config.Page - 1) * config.Limit

	// Construct the base query
	query := `
        SELECT 
            cp.id AS content_part_id,
            cp.content AS content_part,
            d.uuid AS original_document_uuid,
            1 - (e.embedding <=> $1) AS cosine_similarity
        FROM 
            embeddings e
            JOIN content_parts cp ON e.content_part_id = cp.id
            JOIN documents d ON cp.document_uuid = d.uuid
        WHERE 1 = 1
    `
	args := []interface{}{pgvector.NewVector(config.Embedding)}
	argCount := 2

	if config.Status == "" {
		config.Status = "ready_to_search"
	}

	if config.Status != "" {
		query += fmt.Sprintf(" AND d.status = $%d", argCount)
		args = append(args, config.Status)
		argCount++
	}

	// Add source type filter if provided
	if config.SourceType != "" {
		query += fmt.Sprintf(" AND d.source_type = $%d", argCount)
		args = append(args, config.SourceType)
		argCount++
	}

	// Add order by, limit and offset
	// Note: We order by cosine_similarity DESC because higher values indicate more similarity
	query += fmt.Sprintf(`
        ORDER BY cosine_similarity DESC
        LIMIT $%d OFFSET $%d
    `, argCount, argCount+1)
	args = append(args, config.Limit, offset)

	// Execute the query
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	// Process the results
	var results []types.SearchResultsDocument

	for rows.Next() {
		var result struct {
			ContentPart          string
			ContentPartID        int
			OriginalDocumentUUID uuid.UUID
			RelevantScore        float64
		}
		if err := rows.Scan(&result.ContentPartID, &result.ContentPart, &result.OriginalDocumentUUID, &result.RelevantScore); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(errorIteratingOverRowErrFmt, err)
	}

	// Count total results
	countQuery := `
        SELECT COUNT(*)
        FROM 
            embeddings e
            JOIN content_parts cp ON e.content_part_id = cp.id
            JOIN documents d ON cp.document_uuid = d.uuid
        WHERE 
            d.status = 'ready_to_search'
    `
	var countArgs []interface{}
	if config.SourceType != "" {
		countQuery += " AND d.source_type = $1"
		countArgs = append(countArgs, config.SourceType)
	}

	var totalCount int
	err = p.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf(failedToGetTotalCountErrFmt, err)
	}

	totalPages := (totalCount + config.Limit - 1) / config.Limit

	return &types.SearchResults{
		Documents:    results,
		QueryText:    config.QueryText,
		Limit:        config.Limit,
		Page:         config.Page,
		TotalPages:   totalPages,
		TotalResults: totalCount,
	}, nil
}

// UpdateSettings updates the application settings in the database
func (p *Postgres) UpdateSettings(ctx context.Context, settings types.Settings) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		UPDATE app_settings
		SET settings = $1, last_updated_at = NOW()
		WHERE id = 1
	`
	_, err = p.db.Exec(query, settingsJSON)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}

// GetSettings retrieves the application settings from the database
func (p *Postgres) GetSettings(ctx context.Context) (types.Settings, error) {
	var appSettings types.AppSettings
	var settings types.Settings

	// Try to get existing settings
	query := "SELECT id, settings, last_updated_at FROM app_settings LIMIT 1"
	row := p.db.QueryRow(query)

	var settingsJSON []byte
	err := row.Scan(&appSettings.ID, &settingsJSON, &appSettings.LastUpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no settings exist, insert default settings
			return p.insertDefaultSettings()
		}
		return settings, fmt.Errorf("failed to scan settings: %w", err)
	}

	// Unmarshal existing settings
	err = json.Unmarshal(settingsJSON, &settings)
	if err != nil {
		return settings, fmt.Errorf("failed to unmarshal settings JSON: %w", err)
	}

	// Check if settings are empty (application name is a good indicator)
	if settings.General.ApplicationName == "" {
		// If settings are empty, insert default settings
		return p.insertDefaultSettings()
	}

	return settings, nil
}

func (p *Postgres) insertDefaultSettings() (types.Settings, error) {
	defaultSettings := types.Settings{
		General: types.GeneralSettings{ApplicationName: "SmartyPants AI"},
		Debugging: types.DebuggingSettings{
			LogLevel:  logger.LevelDebug,
			LogFormat: logger.FormatJSON,
			LogOutput: logger.OutputStderr,
		},
		Search: types.SearchSettings{PerPage: 10},
	}

	defaultSettingsJSON, err := json.Marshal(defaultSettings)
	if err != nil {
		return types.Settings{}, fmt.Errorf("failed to marshal default settings: %w", err)
	}

	insertQuery := `
		INSERT INTO app_settings (id, settings, last_updated_at)
		VALUES (1, $1, NOW())
		ON CONFLICT (id) DO UPDATE
		SET settings = $1, last_updated_at = NOW()
	`
	_, err = p.db.Exec(insertQuery, defaultSettingsJSON)
	if err != nil {
		return types.Settings{}, fmt.Errorf("failed to insert default settings: %w", err)
	}

	return defaultSettings, nil
}

// CreateLLMProvider creates a new LLM provider in the database
func (p *Postgres) CreateLLMProvider(ctx context.Context, provider types.LLMProviderConfig) error {
	configJSON, err := json.Marshal(provider.Configuration)
	if err != nil {
		return fmt.Errorf(failedToMarshalConfigErrFmt, err)
	}

	_, err = p.db.ExecContext(ctx,
		"INSERT INTO llm_providers (id, name, provider, status, configuration) VALUES ($1, $2, $3, $4, $5)",
		provider.UUID, provider.Name, provider.Provider, provider.Status, configJSON)
	if err != nil {
		return fmt.Errorf("failed to insert LLM provider: %w", err)
	}

	return nil
}

// UpdateLLMProvider updates an existing LLM provider in the database
func (p *Postgres) UpdateLLMProvider(ctx context.Context, provider types.LLMProviderConfig) error {
	configJSON, err := json.Marshal(provider.Configuration)
	if err != nil {
		return fmt.Errorf(failedToMarshalConfigErrFmt, err)
	}

	_, err = p.db.ExecContext(ctx,
		"UPDATE llm_providers SET name = $2, provider = $3, configuration = $4 WHERE id = $1",
		provider.UUID, provider.Name, provider.Provider, configJSON)
	if err != nil {
		return fmt.Errorf("failed to update LLM provider: %w", err)
	}

	return nil
}

// DeleteLLMProvider deletes an LLM provider from the database
func (p *Postgres) DeleteLLMProvider(ctx context.Context, uuid uuid.UUID) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM llm_providers WHERE id = $1", uuid)
	if err != nil {
		return fmt.Errorf("failed to delete LLM provider: %w", err)
	}

	return nil
}

// GetLLMProvider retrieves an LLM provider from the database
func (p *Postgres) GetLLMProvider(ctx context.Context, uuid uuid.UUID) (*types.LLMProviderConfig, error) {
	var provider types.LLMProviderConfig
	var configJSON []byte

	err := p.db.QueryRowContext(ctx,
		"SELECT id, name, provider, status, configuration FROM llm_providers WHERE id = $1",
		uuid).Scan(&provider.UUID, &provider.Name, &provider.Provider, &provider.Status, &configJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.ErrLLMProviderNotFound
		}
		return nil, fmt.Errorf("failed to get embedding provider: %w", err)
	}

	settings, err := util.ParseLLMProviderSettings(provider.Provider, configJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM provider settings: %w", err)
	}

	provider.Configuration = settings

	return &provider, nil
}

// GetAllLLMProviders retrieves all LLM providers from the database
func (p *Postgres) GetAllLLMProviders(ctx context.Context, filter types.LLMProviderFilter, option types.LLMProviderFilterOption) (*types.PaginatedLLMProviders, error) {
	query := "SELECT id, name, provider, status, configuration FROM llm_providers WHERE 1=1"
	var args []interface{}
	argCount := 1

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filter.Status)
		argCount++
	}

	// Add more filter conditions here as needed

	// Count total before applying pagination
	countQuery := strings.Replace(query, "SELECT id, name, provider, status, configuration", "SELECT COUNT(*)", 1)
	var total int
	err := p.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf(failedToGetTotalCountErrFmt, err)
	}

	// Apply sorting and pagination
	query += " ORDER BY name LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, option.Limit, (option.Page-1)*option.Limit)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query LLM providers: %w", err)
	}
	defer rows.Close()

	var providers []types.LLMProviderConfig
	for rows.Next() {
		var provider types.LLMProviderConfig
		var configJSON []byte
		err := rows.Scan(&provider.UUID, &provider.Name, &provider.Provider, &provider.Status, &configJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan embedding provider: %w", err)
		}

		settings, err := util.ParseLLMProviderSettings(provider.Provider, configJSON)
		if err != nil {
			p.logger.WithFields(logrus.Fields{
				"provider": provider.Provider,
				"uuid":     provider.UUID,
				"error":    err,
			}).Error("Failed to parse LLM provider settings")
			continue
		}

		provider.Configuration = settings
		providers = append(providers, provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(errorIteratingOverRowErrFmt, err)
	}

	totalPages := (total + option.Limit - 1) / option.Limit

	return &types.PaginatedLLMProviders{
		LLMProviders: providers,
		Total:        total,
		Page:         option.Page,
		PerPage:      option.Limit,
		TotalPages:   totalPages,
	}, nil
}

// CreateInteraction creates a new interaction in the database
func (p *Postgres) CreateInteraction(ctx context.Context, interaction types.Interaction) (types.Interaction, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return interaction, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if interaction.UUID == uuid.Nil {
		interaction.UUID = uuid.New()
	}

	// Use the transaction to insert the interaction
	_, err = tx.Exec("INSERT INTO interactions (uuid, query, created_at) VALUES ($1, $2, $3)",
		interaction.UUID, interaction.Query, time.Now().UTC())
	if err != nil {
		return interaction, err
	}

	// Add conversation using the transaction
	for _, conversation := range interaction.Conversations {
		if conversation.UUID == uuid.Nil {
			conversation.UUID = uuid.New()
		}
		_, err = p.AddConversationTx(ctx, tx, interaction.UUID.String(), conversation)
		if err != nil {
			return interaction, err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return interaction, err
	}

	return interaction, nil
}

// AddConversationTx adds a new conversation to the database using the provided transaction
func (p *Postgres) AddConversationTx(ctx context.Context, tx *sql.Tx, interactionUUID string, conversation types.Conversation) (types.Conversation, error) {
	if conversation.UUID == uuid.Nil {
		conversation.UUID = uuid.New()
	}

	_, err := tx.Exec("INSERT INTO conversations (uuid, interaction_uuid, role, text, created_at) VALUES ($1, $2, $3, $4, $5)",
		conversation.UUID, interactionUUID, conversation.Role, conversation.Text, time.Now().UTC())
	if err != nil {
		return conversation, err
	}

	return conversation, nil
}

// AddConversation store a new conversation in the database
func (p *Postgres) AddConversation(ctx context.Context, interactionUUID uuid.UUID, role string, message string) (types.Conversation, error) {
	conversation := types.Conversation{
		UUID: uuid.New(),
		Role: types.InteractionRole(role),
		Text: message,
	}

	_, err := p.db.ExecContext(ctx, "INSERT INTO conversations (uuid, interaction_uuid, role, text, created_at) VALUES ($1, $2, $3, $4, $5)",
		conversation.UUID, interactionUUID.String(), conversation.Role, conversation.Text, time.Now().UTC())
	if err != nil {
		return conversation, err
	}

	return conversation, nil
}

// GetInteraction retrieves an interaction from the database
func (p *Postgres) GetInteraction(ctx context.Context, uuid uuid.UUID) (types.Interaction, error) {
	var interaction types.Interaction
	err := p.db.QueryRow("SELECT uuid, query, created_at FROM interactions WHERE uuid = $1", uuid).
		Scan(&interaction.UUID, &interaction.Query, &interaction.CreatedAt)
	if err != nil {
		return interaction, err
	}

	interaction.Conversations, err = p.GetConversation(ctx, uuid)

	return interaction, nil
}

// GetAllInteractions retrieves all interactions from the database
func (p *Postgres) GetAllInteractions(ctx context.Context, page, perPage int) (*types.PaginatedInteractions, error) {
	offset := (page - 1) * perPage

	query := `
		SELECT i.uuid, i.query, i.created_at
		FROM interactions i
		INNER JOIN conversations c ON i.uuid = c.interaction_uuid
		GROUP BY i.uuid, i.query, i.created_at
		HAVING COUNT(c.uuid) > 0
		ORDER BY i.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := p.db.QueryContext(ctx, query, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query interactions: %w", err)
	}
	defer rows.Close()

	var interactions []types.Interaction

	for rows.Next() {
		var interaction types.Interaction
		if err := rows.Scan(&interaction.UUID, &interaction.Query, &interaction.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan interaction: %w", err)
		}

		interaction.Conversations, err = p.GetConversation(ctx, interaction.UUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation for interaction %s: %w", interaction.UUID, err)
		}

		interactions = append(interactions, interaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	// Get total count of interactions with conversations
	var totalCount int
	countQuery := `
		SELECT COUNT(DISTINCT i.uuid)
		FROM interactions i
		INNER JOIN conversations c ON i.uuid = c.interaction_uuid
	`
	if err := p.db.QueryRowContext(ctx, countQuery).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	totalPages := (totalCount + perPage - 1) / perPage

	return &types.PaginatedInteractions{
		Interactions: interactions,
		Total:        totalCount,
		Page:         page,
		PerPage:      perPage,
		TotalPages:   totalPages,
	}, nil
}

// GetConversation retrieves all conversations for an interaction from the database
func (p *Postgres) GetConversation(ctx context.Context, interactionUUID uuid.UUID) ([]types.Conversation, error) {
	var conversations []types.Conversation

	query := `select uuid, role, text, created_at from conversations where interaction_uuid = $1 order by created_at`
	rows, err := p.db.Query(query, interactionUUID)
	if err != nil {
		return conversations, fmt.Errorf("failed to query conversations: %w", err)
	}

	for rows.Next() {
		var conversation types.Conversation
		if err := rows.Scan(&conversation.UUID, &conversation.Role, &conversation.Text, &conversation.CreatedAt); err != nil {
			return conversations, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conversation)
	}

	return conversations, nil
}

// GetAnalyticsOverview retrieves an overview of analytics data from the database
func (p *Postgres) GetAnalyticsOverview(ctx context.Context) (types.AnalyticsOverview, error) {
	ctx, span := observability.StartSpan(ctx, "postgres.GetAnalyticsOverview")
	defer span.End()

	var overview types.AnalyticsOverview

	_, epOverviewQuerySpan := observability.StartSpan(ctx, "postgres.query.get_embedding_providers_overview")

	var activeName, activeType, activeModel sql.NullString
	var activeProviders sql.NullInt64
	err := p.db.QueryRowContext(ctx, `
        SELECT 
            COUNT(*) AS total_providers,
            SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) AS active_providers,
            (SELECT name FROM embedding_providers WHERE status = 'active' LIMIT 1) AS active_name,
            (SELECT provider FROM embedding_providers WHERE status = 'active' LIMIT 1) AS active_type,
            (SELECT configuration->>'model_id' FROM embedding_providers WHERE status = 'active' LIMIT 1) AS active_model
        FROM embedding_providers
    `).Scan(
		&overview.EmbeddingProviders.TotalProviders,
		&activeProviders,
		&activeName,
		&activeType,
		&activeModel,
	)
	if err != nil {
		return overview, err
	}
	epOverviewQuerySpan.End()

	overview.EmbeddingProviders.TotalActiveProviders = int(activeProviders.Int64)
	overview.EmbeddingProviders.ActiveProvider.Name = activeName.String
	overview.EmbeddingProviders.ActiveProvider.Type = activeType.String
	overview.EmbeddingProviders.ActiveProvider.Model = activeModel.String

	_, llmOverviewQuerySpan := observability.StartSpan(ctx, "postgres.query.get_llm_providers_overview")
	activeName, activeType, activeModel = sql.NullString{}, sql.NullString{}, sql.NullString{}
	activeProviders = sql.NullInt64{}
	err = p.db.QueryRowContext(ctx, `
        SELECT 
            COUNT(*) AS total_providers,
            SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) AS active_providers,
            (SELECT name FROM llm_providers WHERE status = 'active' LIMIT 1) AS active_name,
            (SELECT provider FROM llm_providers WHERE status = 'active' LIMIT 1) AS active_type,
            (SELECT configuration->>'model_id' FROM llm_providers WHERE status = 'active' LIMIT 1) AS active_model
        FROM llm_providers
    `).Scan(
		&overview.LLMProviders.TotalProviders,
		&activeProviders,
		&activeName,
		&activeType,
		&activeModel,
	)
	if err != nil {
		return overview, err
	}
	llmOverviewQuerySpan.End()

	overview.LLMProviders.TotalActiveProviders = int(activeProviders.Int64)
	overview.LLMProviders.ActiveProvider.Name = activeName.String
	overview.LLMProviders.ActiveProvider.Type = activeType.String
	overview.LLMProviders.ActiveProvider.Model = activeModel.String

	_, dsOverviewQuerySpan := observability.StartSpan(ctx, "postgres.query.get_datasources_overview")
	rows, err := p.db.QueryContext(ctx, `
        SELECT 
            name, 
            source_type, 
            status, 
            created_at 
        FROM datasources
    `)
	if err != nil {
		return overview, err
	}
	defer rows.Close()

	dsOverviewQuerySpan.End()

	overview.Datasources.TotalDatasourcesByType = make(map[string]int)
	overview.Datasources.TotalDatasourcesByStatus = make(map[string]int)
	overview.Datasources.TotalDocumentsFetchedByDatasourceType = make(map[string]int)

	for rows.Next() {
		var ds types.DatasourceInfo
		err := rows.Scan(&ds.Name, &ds.Type, &ds.Status, &ds.CreatedAt)
		if err != nil {
			return overview, err
		}
		overview.Datasources.ConfiguredDatasources = append(overview.Datasources.ConfiguredDatasources, ds)
		overview.Datasources.TotalDatasources++
		overview.Datasources.TotalDatasourcesByType[ds.Type]++
		overview.Datasources.TotalDatasourcesByStatus[ds.Status]++
	}

	if err = rows.Err(); err != nil {
		return overview, err
	}

	// If there are no datasources, we can return early
	if overview.Datasources.TotalDatasources == 0 {
		return overview, nil
	}

	// Get total documents fetched by datasource type
	_, docOverviewQuerySpan := observability.StartSpan(ctx, "postgres.query.get_documents_overview")
	rows, err = p.db.QueryContext(ctx, `
        SELECT 
            d.source_type, 
            COUNT(doc.uuid) 
        FROM datasources d
        LEFT JOIN documents doc ON d.uuid = doc.datasource_uuid
        GROUP BY d.source_type
    `)
	if err != nil {
		return overview, err
	}
	defer rows.Close()
	docOverviewQuerySpan.End()

	for rows.Next() {
		var sourceType string
		var count int
		err := rows.Scan(&sourceType, &count)
		if err != nil {
			return overview, err
		}
		overview.Datasources.TotalDocumentsFetchedByDatasourceType[sourceType] = count
	}

	if err = rows.Err(); err != nil {
		return overview, err
	}

	return overview, nil
}

// RunMigration run database migration
func (p *Postgres) RunMigration() error {
	p.logger.Info("Preparing to run migration")

	if p.migrator == nil {
		p.logger.Debug("creating new database migrator")
		migrator, err := p.newMigrator()
		if err != nil {
			return fmt.Errorf("failed to create database migrator")
		}

		p.logger.Debug("migrator has been configured")
		p.migrator = migrator
	}

	p.logger.Debug("Running migration.up")
	err := p.migrator.Up()

	if err != nil && errors.Is(err, migrate.ErrNoChange) {
		p.logger.Info("No new changes found for migration")
		return nil
	}

	if err != nil {
		errMsg := "failed to run migrations"
		p.logger.WithError(err).Error(errMsg)
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	p.logger.Info("database migration has been completed successfully")
	return nil
}

func (p *Postgres) newMigrator() (*migrate.Migrate, error) {
	p.logger.Debug("creating driver instance for migration")
	driver, err := postgres.WithInstance(p.db, &postgres.Config{})
	if err != nil {
		errMsg := "failed to create driver instance"
		p.logger.WithError(err).Error(errMsg)
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}

	p.logger.Debug("preparing source for migration using in-memory file systems")
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		errMsg := "failed to create source instance"
		p.logger.WithError(err).Error(errMsg)
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}

	p.logger.Debug("configuring migrator instance")
	migrator, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		errMsg := "failed to create migrator instance"
		p.logger.WithError(err).Error(errMsg)
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}

	var verboseLogging bool
	if p.logger.Level == logrus.DebugLevel {
		verboseLogging = true
	}

	migrator.Log = logger.NewMigrationLogger(p.logger, verboseLogging)

	p.logger.Debug("getting current migration status")
	version, isDirty, err := migrator.Version()
	if err != nil && errors.Is(err, migrate.ErrNilVersion) {
		p.logger.Debug("no migration exists. this will be the first migration")
		return migrator, nil
	}

	if err != nil {
		errMsg := "failed to get current migration version"
		p.logger.WithError(err).Error(errMsg)
		return nil, fmt.Errorf("%s : %w", errMsg, err)
	}
	p.logger.WithFields(logrus.Fields{"migration_version": version, "migration_is_dirty": isDirty, "migration_verbose_logging": verboseLogging}).Debug("current migration status")
	return migrator, nil
}

// HandleShutdown handle graceful shutdown for db migration
func (p *Postgres) HandleShutdown(ctx context.Context) error {
	if p.migrator != nil {
		p.logger.Info("Gracefully stopping migration process")
		p.migrator.GracefulStop <- true
		select {
		case <-ctx.Done():
			p.logger.Warn("Migration shutdown timed out")
			return ctx.Err()
		case <-p.migrator.GracefulStop:
			p.logger.Info("Migration process stopped gracefully")
		}
	}
	return nil
}

func (p *Postgres) GetKeyPair() (privateKey, publicKey []byte, err error) {
	query := "SELECT private_key, public_key FROM key_pairs LIMIT 1"
	err = p.db.QueryRow(query).Scan(&privateKey, &publicKey)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, errors.New("no key pair found")
	}
	return
}

func (p *Postgres) UpdateKeyPair(privateKey, publicKey []byte) error {
	query := `
        INSERT INTO key_pairs (private_key, public_key) 
        VALUES ($1, $2) 
        ON CONFLICT (id) DO UPDATE 
        SET private_key = $1, public_key = $2
    `
	_, err := p.db.Exec(query, privateKey, publicKey)
	return err
}

func (p *Postgres) CreateUser(ctx context.Context, user *types.User) error {
	user.UUID = uuid.New()
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()

	query := `
        INSERT INTO users (uuid, name, email, status, roles, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	_, err := p.db.ExecContext(ctx, query,
		user.UUID,
		user.Name,
		user.Email,
		user.Status,
		pq.Array(user.Roles),
		user.CreatedAt,
		user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (p *Postgres) GetUser(ctx context.Context, uuid uuid.UUID) (*types.User, error) {
	var user types.User
	var roleStrings []string

	err := p.db.QueryRowContext(ctx, `
		SELECT uuid, name, email, status, roles, created_at, updated_at 
		FROM users 
		WHERE uuid = $1
	`, uuid).Scan(
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Status,
		pq.Array(&roleStrings),
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, types.UserNotFoundError
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Roles = make([]types.UserRole, len(roleStrings))
	for i, roleStr := range roleStrings {
		user.Roles[i] = types.UserRole(roleStr)
	}

	return &user, nil
}

func (p *Postgres) UpdateUserStatus(ctx context.Context, uuid uuid.UUID, status types.UserStatus) error {
	query := `
        UPDATE users
        SET status = $1, updated_at = $2
        WHERE uuid = $3
    `

	result, err := p.db.ExecContext(ctx, query, status, time.Now().UTC(), uuid)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return types.UserNotFoundError
	}

	return nil
}

func (p *Postgres) GetPaginatedUsers(ctx context.Context, filter types.UserFilter, option types.UserFilterOption) (types.PaginatedUsers, error) {
	var result types.PaginatedUsers
	var whereClause []string
	var args []interface{}
	argCount := 1

	// Build WHERE clause based on filter
	if filter.NameContains != "" {
		whereClause = append(whereClause, fmt.Sprintf("name ILIKE $%d", argCount))
		args = append(args, "%"+filter.NameContains+"%")
		argCount++
	}
	if filter.EmailContains != "" {
		whereClause = append(whereClause, fmt.Sprintf("email ILIKE $%d", argCount))
		args = append(args, "%"+filter.EmailContains+"%")
		argCount++
	}
	if filter.Status != "" {
		whereClause = append(whereClause, fmt.Sprintf("status = $%d", argCount))
		args = append(args, filter.Status)
		argCount++
	}
	if len(filter.Roles) > 0 {
		roleStrings := make([]string, len(filter.Roles))
		for i, role := range filter.Roles {
			roleStrings[i] = string(role)
		}
		whereClause = append(whereClause, fmt.Sprintf("roles && $%d", argCount))
		args = append(args, pq.Array(roleStrings))
		argCount++
	}

	// Construct the WHERE clause string
	var whereStr string
	if len(whereClause) > 0 {
		whereStr = "WHERE " + strings.Join(whereClause, " AND ")
	}

	// Count total matching users
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereStr)
	var total int
	err := p.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return result, fmt.Errorf("error counting users: %w", err)
	}

	// Calculate pagination
	if option.Page < 1 {
		option.Page = 1
	}
	if option.PerPage < 1 {
		option.PerPage = 10 // Default to 10 per page
	}
	offset := (option.Page - 1) * option.PerPage

	// Fetch users
	query := fmt.Sprintf(`
		SELECT uuid, name, email, status, roles, created_at, updated_at
		FROM users
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereStr, argCount, argCount+1)
	args = append(args, option.PerPage, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return result, fmt.Errorf("error querying users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user types.User
		var userUUID uuid.UUID
		var roles pq.StringArray
		err := rows.Scan(&userUUID, &user.Name, &user.Email, &user.Status, &roles, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return result, fmt.Errorf("error scanning user: %w", err)
		}
		user.UUID = userUUID
		user.Roles = make([]types.UserRole, len(roles))
		for i, role := range roles {
			user.Roles[i] = types.UserRole(role)
		}
		result.Users = append(result.Users, user)
	}

	if err = rows.Err(); err != nil {
		return result, fmt.Errorf("error iterating users: %w", err)
	}

	// Populate pagination info
	result.Total = total
	result.Page = option.Page
	result.PerPage = option.PerPage
	result.TotalPages = (total + option.PerPage - 1) / option.PerPage

	return result, nil
}

func (p *Postgres) UpdateUserRoles(ctx context.Context, uuid uuid.UUID, roles []types.UserRole) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update roles
	_, err = tx.ExecContext(ctx, `
		UPDATE users 
		SET roles = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE uuid = $2
	`, pq.Array(roles), uuid)
	if err != nil {
		return fmt.Errorf("failed to update user roles: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
