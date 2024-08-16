package migration

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgreSQLMigrations(t *testing.T) {
	assert.NotEmpty(t, postgreSQLMigrations)
	assert.Equal(t, "0.0.1", postgreSQLMigrations[0].Version)

	// Test Up migration
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`
		CREATE EXTENSION IF NOT EXISTS vector;
		CREATE TABLE IF NOT EXISTS datasources.*
		CREATE TABLE IF NOT EXISTS documents.*
		CREATE TABLE IF NOT EXISTS content_parts.*
		CREATE TABLE IF NOT EXISTS embedding_providers.*
		CREATE TABLE IF NOT EXISTS embeddings.*
		CREATE TABLE IF NOT EXISTS metadata.*
		CREATE TABLE IF NOT EXISTS llm_providers.*
		CREATE TABLE IF NOT EXISTS ai_ops_usage.*
		CREATE TABLE IF NOT EXISTS app_settings.*
		CREATE TABLE IF NOT EXISTS interactions.*
		CREATE TABLE IF NOT EXISTS conversations.*
		CREATE INDEX IF NOT EXISTS.*
	`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	tx, err := db.Begin()
	assert.NoError(t, err)

	err = postgreSQLMigrations[0].Up(tx)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test Down migration
	mock.ExpectBegin()
	mock.ExpectExec(`
		DROP INDEX IF EXISTS.*
		DROP TABLE IF EXISTS.*
		DROP EXTENSION IF EXISTS vector;
	`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	tx, err = db.Begin()
	assert.NoError(t, err)

	err = postgreSQLMigrations[0].Down(tx)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
