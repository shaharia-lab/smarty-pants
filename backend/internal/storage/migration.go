package storage

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sirupsen/logrus"
)

type Migration struct {
	db     *sql.DB
	logger *logrus.Logger
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewMigration(db *sql.DB, logger *logrus.Logger) *Migration {
	return &Migration{db: db, logger: logger}
}

// Run executes all pending migrations
func (m *Migration) Run() error {
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		errMsg := "failed to create driver instance"
		m.logger.WithError(err).Error(errMsg)
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		errMsg := "failed to create source instance"
		m.logger.WithError(err).Error(errMsg)
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	// Create a new migrate instance
	migrator, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		errMsg := "failed to create migrator instance"
		m.logger.WithError(err).Error(errMsg)
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	// Run migrations
	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		errMsg := "failed to run migrations"
		m.logger.WithError(err).Error(errMsg)
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	m.logger.Info("Migrations completed successfully")
	return nil
}
