package migration

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRunMigrations(t *testing.T) {
	tests := []struct {
		name           string
		ensureTableErr error
		migrateErr     error
		wantErr        bool
	}{
		{
			name:           "Successful migration",
			ensureTableErr: nil,
			migrateErr:     nil,
			wantErr:        false,
		},
		{
			name:           "EnsureMigrationTableExists fails",
			ensureTableErr: errors.New("table creation failed"),
			migrateErr:     nil,
			wantErr:        true,
		},
		{
			name:           "Migrate fails",
			ensureTableErr: nil,
			migrateErr:     errors.New("migration failed"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMigrator := NewMockMigrator(t)
			mockMigrator.On("EnsureMigrationTableExists").Return(tt.ensureTableErr)

			if tt.ensureTableErr == nil {
				mockMigrator.On("Migrate", mock.Anything).Return(tt.migrateErr)
			}

			manager := NewMigrationManager(mockMigrator, []Migration{})

			err := manager.RunMigrations()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.ensureTableErr != nil {
					assert.Equal(t, tt.ensureTableErr, err)
				} else {
					assert.Equal(t, tt.migrateErr, err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockMigrator.AssertExpectations(t)
		})
	}
}

func TestRollbackMigration(t *testing.T) {
	tests := []struct {
		name        string
		rollbackErr error
		wantErr     bool
	}{
		{
			name:        "Successful rollback",
			rollbackErr: nil,
			wantErr:     false,
		},
		{
			name:        "Rollback fails",
			rollbackErr: errors.New("rollback failed"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMigrator := NewMockMigrator(t)
			mockMigrator.On("Rollback", mock.Anything).Return(tt.rollbackErr)

			manager := NewMigrationManager(mockMigrator, []Migration{})

			err := manager.RollbackMigration()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.rollbackErr, err)
			} else {
				assert.NoError(t, err)
			}

			mockMigrator.AssertExpectations(t)
		})
	}
}

func TestGetCurrentVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		versionErr  error
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "Get version successfully",
			version:     "1.0.0",
			versionErr:  nil,
			wantVersion: "1.0.0",
			wantErr:     false,
		},
		{
			name:        "Get version fails",
			version:     "",
			versionErr:  errors.New("version retrieval failed"),
			wantVersion: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMigrator := NewMockMigrator(t)
			mockMigrator.On("GetCurrentVersion").Return(tt.version, tt.versionErr)

			manager := NewMigrationManager(mockMigrator, []Migration{})

			gotVersion, err := manager.GetCurrentVersion()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.versionErr, err)
				assert.Empty(t, gotVersion)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVersion, gotVersion)
			}

			mockMigrator.AssertExpectations(t)
		})
	}
}
