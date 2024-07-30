package collector

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/datasource"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
)

// Datasource is an interface for a data source that can be used by the collector
type Datasource interface {
	GetID() uuid.UUID
	GetData(ctx context.Context, currentState types.DatasourceState) ([]types.Document, types.DatasourceState, error)
	Validate() error
}

// CreateDatasourceFromConfig creates a new datasource from the given configuration
func CreateDatasourceFromConfig(config types.DatasourceConfig, logger *logrus.Logger) (Datasource, error) {
	switch config.SourceType {
	case types.DatasourceTypeSlack:
		slackSettings, ok := config.Settings.(*types.SlackSettings)
		if !ok {
			return nil, fmt.Errorf("invalid settings type for Slack datasource")
		}
		return datasource.NewSlackDatasource(config, datasource.NewConcreteSlackClient(slackSettings.Token), logger)
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", config.SourceType)
	}
}
