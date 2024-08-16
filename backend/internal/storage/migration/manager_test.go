package migration

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRunMigrations(t *testing.T) {
	tests := []struct {
		name              string
		ensureTableErr    error
		acquireLockResult bool
		acquireLockErr    error
		migrateErr        error
		releaseLockErr    error
		wantErr           bool
		expectedErrMsg    string
		expectLockRelease bool
	}{
		{
			name:              "Successful migration",
			ensureTableErr:    nil,
			acquireLockResult: true,
			acquireLockErr:    nil,
			migrateErr:        nil,
			releaseLockErr:    nil,
			wantErr:           false,
			expectLockRelease: true,
		},
		{
			name:              "Failed to acquire lock",
			ensureTableErr:    nil,
			acquireLockResult: false,
			acquireLockErr:    nil,
			migrateErr:        nil,
			releaseLockErr:    nil,
			wantErr:           false,
			expectLockRelease: false,
		},
		{
			name:              "Error acquiring lock",
			ensureTableErr:    nil,
			acquireLockResult: false,
			acquireLockErr:    errors.New("failed to acquire lock"),
			migrateErr:        nil,
			releaseLockErr:    nil,
			wantErr:           true,
			expectedErrMsg:    "failed to acquire migration lock: failed to acquire lock",
			expectLockRelease: false,
		},
		{
			name:              "EnsureMigrationTableExists fails",
			ensureTableErr:    errors.New("table creation failed"),
			acquireLockResult: true,
			acquireLockErr:    nil,
			migrateErr:        nil,
			releaseLockErr:    nil,
			wantErr:           true,
			expectedErrMsg:    "Failed to ensure migration table exists: table creation failed",
			expectLockRelease: false,
		},
		{
			name:              "Migrate fails",
			ensureTableErr:    nil,
			acquireLockResult: true,
			acquireLockErr:    nil,
			migrateErr:        errors.New("migration failed"),
			releaseLockErr:    nil,
			wantErr:           true,
			expectedErrMsg:    "failed to run migrations: migration failed",
			expectLockRelease: true,
		},
		{
			name:              "ReleaseMigrationLock fails",
			ensureTableErr:    nil,
			acquireLockResult: true,
			acquireLockErr:    nil,
			migrateErr:        nil,
			releaseLockErr:    errors.New("failed to release lock"),
			wantErr:           false,
			expectLockRelease: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMigrator := new(MockMigrator)
			mockMigrator.On("EnsureMigrationTableExists").Return(tt.ensureTableErr)

			if tt.ensureTableErr == nil {
				mockMigrator.On("AcquireMigrationLock").Return(tt.acquireLockResult, tt.acquireLockErr)

				if tt.acquireLockResult {
					mockMigrator.On("Migrate", mock.Anything).Return(tt.migrateErr)
					if tt.expectLockRelease {
						mockMigrator.On("ReleaseMigrationLock").Return(tt.releaseLockErr)
					}
				}
			}

			logger, hook := test.NewNullLogger()
			logger.SetLevel(logrus.DebugLevel)

			manager := NewMigrationManager(mockMigrator, []Migration{}, logger)

			err := manager.RunMigrations()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			mockMigrator.AssertExpectations(t)

			// Check log messages if needed
			for _, entry := range hook.AllEntries() {
				t.Logf("LOG [%s]: %s", entry.Level, entry.Message)
			}
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

			testLogger := logrus.New()
			testLogger.Out = nil // Disable logging output for tests

			manager := NewMigrationManager(mockMigrator, []Migration{}, testLogger)

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

			testLogger := logrus.New()
			testLogger.Out = nil // Disable logging output for tests

			manager := NewMigrationManager(mockMigrator, []Migration{}, testLogger)

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
