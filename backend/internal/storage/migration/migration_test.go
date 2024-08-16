package migration

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigration(t *testing.T) {
	m := Migration{
		Version: "1.0.0",
		Up: func(tx *sql.Tx) error {
			return nil
		},
		Down: func(tx *sql.Tx) error {
			return nil
		},
	}

	assert.Equal(t, "1.0.0", m.Version)
	assert.NotNil(t, m.Up)
	assert.NotNil(t, m.Down)
}

func TestDirtyMigrationError(t *testing.T) {
	err := &DirtyMigrationError{
		Version: "1.0.0",
		Message: "Migration is dirty",
	}

	assert.Equal(t, "Migration is dirty", err.Error())
}

func TestMigrationError(t *testing.T) {
	originalErr := errors.New("original error")
	err := &MigrationError{
		Version: "1.0.0",
		Err:     originalErr,
		Message: "Migration failed",
	}

	assert.Equal(t, "Migration failed: original error", err.Error())
}
