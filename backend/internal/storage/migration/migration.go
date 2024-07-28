package migration

import "github.com/shaharia-lab/smarty-pants-ai/internal/storage"

type Provider interface {
	Up() error
	Down() error
}

type Migration struct {
	storage storage.Storage
}

func (m *Migration) Up() error {
	return m.storage.MigrationUp()
}

func (m *Migration) Down() error {
	return m.storage.MigrationDown()
}
