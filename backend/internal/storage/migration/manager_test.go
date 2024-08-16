package migration

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name              string
		ensureTableErr    error
		acquireLockResult bool
		acquireLockErr    error
		migrateErr        error
		releaseLockErr    error
		wantErr           bool
		expectedErrMsg    string
		expectAcquireLock bool
		expectMigrate     bool
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
			expectAcquireLock: true,
			expectMigrate:     true,
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
			expectAcquireLock: true,
			expectMigrate:     false,
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
			expectAcquireLock: true,
			expectMigrate:     false,
			expectLockRelease: false,
		},
		{
			name:              "EnsureMigrationTableExists fails",
			ensureTableErr:    errors.New("table creation failed"),
			acquireLockResult: false,
			acquireLockErr:    nil,
			migrateErr:        nil,
			releaseLockErr:    nil,
			wantErr:           true,
			expectedErrMsg:    "failed to ensure migration table exists: table creation failed",
			expectAcquireLock: false,
			expectMigrate:     false,
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
			expectAcquireLock: true,
			expectMigrate:     true,
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
			expectAcquireLock: true,
			expectMigrate:     true,
			expectLockRelease: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMigrator := new(MockMigrator)
			mockMigrator.On("EnsureMigrationTableExists").Return(tt.ensureTableErr)

			if tt.expectAcquireLock {
				mockMigrator.On("AcquireMigrationLock").Return(tt.acquireLockResult, tt.acquireLockErr)
			}

			if tt.expectMigrate {
				mockMigrator.On("Migrate", mock.Anything).Return(tt.migrateErr)
			}
			if tt.expectLockRelease {
				mockMigrator.On("ReleaseMigrationLock").Return(tt.releaseLockErr)
			}

			logger, hook := test.NewNullLogger()
			logger.SetLevel(logrus.DebugLevel)

			manager := NewMigrationManager(mockMigrator, logger)

			err := manager.Run()

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
