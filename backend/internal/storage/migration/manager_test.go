package migration

import (
	"errors"
	"testing"

	"github.com/shaharia-lab/smarty-pants/backend/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRunMigrations(t *testing.T) {
	tests := []struct {
		name              string
		acquireLockResult bool
		acquireLockErr    error
		ensureTableErr    error
		migrateErr        error
		wantErr           bool
		expectedErrMsg    string
	}{
		{
			name:              "Successful migration",
			acquireLockResult: true,
			acquireLockErr:    nil,
			ensureTableErr:    nil,
			migrateErr:        nil,
			wantErr:           false,
		},
		{
			name:              "Failed to acquire lock",
			acquireLockResult: false,
			acquireLockErr:    nil,
			ensureTableErr:    nil,
			migrateErr:        nil,
			wantErr:           false,
		},
		{
			name:              "Error acquiring lock",
			acquireLockResult: false,
			acquireLockErr:    errors.New("failed to acquire lock"),
			ensureTableErr:    nil,
			migrateErr:        nil,
			wantErr:           true,
			expectedErrMsg:    "failed to acquire migration lock: failed to acquire lock",
		},
		{
			name:              "EnsureMigrationTableExists fails",
			acquireLockResult: true,
			acquireLockErr:    nil,
			ensureTableErr:    errors.New("table creation failed"),
			migrateErr:        nil,
			wantErr:           true,
			expectedErrMsg:    "table creation failed",
		},
		{
			name:              "Migrate fails",
			acquireLockResult: true,
			acquireLockErr:    nil,
			ensureTableErr:    nil,
			migrateErr:        errors.New("migration failed"),
			wantErr:           true,
			expectedErrMsg:    "migration failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMigrator := NewMockMigrator(t)
			mockMigrator.On("AcquireMigrationLock").Return(tt.acquireLockResult, tt.acquireLockErr)

			if tt.acquireLockResult {
				mockMigrator.On("ReleaseMigrationLock").Return(nil)
				mockMigrator.On("EnsureMigrationTableExists").Return(tt.ensureTableErr)

				if tt.ensureTableErr == nil {
					mockMigrator.On("Migrate", mock.Anything).Return(tt.migrateErr)
				}
			}

			manager := NewMigrationManager(mockMigrator, []Migration{}, logger.NoOpsLogger())

			err := manager.RunMigrations()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
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

			manager := NewMigrationManager(mockMigrator, []Migration{}, logger.NoOpsLogger())

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

			manager := NewMigrationManager(mockMigrator, []Migration{}, logger.NoOpsLogger())

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
