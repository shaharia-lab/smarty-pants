package datasource

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/smarty-pants/backend/internal/auth"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type Manager struct {
	storage    storage.Storage
	logger     *logrus.Logger
	aclManager auth.ACLManager
}

func NewDatasourceManager(storage storage.Storage, logger *logrus.Logger, aclManager auth.ACLManager) *Manager {
	return &Manager{
		storage:    storage,
		logger:     logger,
		aclManager: aclManager,
	}
}

func (dm *Manager) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/datasource", func(r chi.Router) {
		r.Post("/", dm.addDatasourceHandler)
		r.Get("/", dm.getDatasourcesHandler)

		r.Route("/{uuid}", func(r chi.Router) {
			r.Delete("/", dm.deleteDatasourceHandler)
			r.Get("/", dm.getDatasourceHandler)
			r.Get("/validate", dm.validateDatasourceHandler)
			r.Put("/", dm.updateDatasourceHandler)
			r.Put("/activate", dm.setActiveDatasourceHandler)
			r.Put("/deactivate", dm.setDisableDatasourceHandler)
		})
	})
}

func (dm *Manager) getDatasourceFromRequest(w http.ResponseWriter, r *http.Request) (*types.DatasourceConfig, error) {
	dsUUIDParsed, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil || dsUUIDParsed == uuid.Nil {
		err = errors.New(types.InvalidUUIDMessage)
	}

	if err != nil {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError(types.InvalidUUIDMessage, err))
		return nil, err
	}

	ds, err := dm.storage.GetDatasource(r.Context(), dsUUIDParsed)
	if err != nil {
		util.SendAPIErrorResponse(w, http.StatusNotFound, util.NewAPIError(types.DatasourceNotFoundMsg, err))
		return nil, err
	}

	return &ds, nil
}

func (dm *Manager) addDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !dm.aclManager.IsAllowed(w, r, types.UserRoleAdmin, "", types.APiAccessOpsDatasourceAdd) {
		return
	}

	ctx, span := observability.StartSpan(ctx, "api.addDatasourceHandler")
	defer span.End()

	var payload types.DatasourcePayload
	if err := util.DecodeJSONBody(r, &payload); err != nil {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if err := dm.validatePayload(payload); err != nil {
		dm.logger.Error("Failed to validate datasource payload: ", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Failed to validate datasource payload", err))
		return
	}

	settings, err := util.ParseSettings(payload.SourceType, payload.Settings)
	if err != nil {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Failed to parse settings", err))
		return
	}

	err = settings.Validate()
	if err != nil {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Validation failed for the datasource settings", err))
		return
	}

	dsConfig := types.DatasourceConfig{
		UUID:       uuid.New(),
		Name:       payload.Name,
		SourceType: payload.SourceType,
		Settings:   settings,
		Status:     types.DatasourceStatusInactive,
		State:      &types.SlackState{},
	}

	if err := dm.storage.AddDatasource(ctx, dsConfig); err != nil {
		dm.logger.Error("Failed to add datasource: ", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, util.NewAPIError("Failed to add datasource", err))
		return
	}

	util.SendSuccessResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "Datasource added successfully",
		"uuid":    dsConfig.UUID,
	}, dm.logger, span)
}

func (dm *Manager) getDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer("api").Start(r.Context(), "getDatasourceHandler")
	defer span.End()

	ds, err := dm.getDatasourceFromRequest(w, r)
	if err != nil {
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, ds, dm.logger, span)
}

func (dm *Manager) validateDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer("api").Start(r.Context(), "api.validateDatasourceHandler")
	defer span.End()

	ds, err := dm.getDatasourceFromRequest(w, r)
	if err != nil {
		return
	}

	if ds.SourceType == types.DatasourceTypeSlack {
		slackSettings, ok := ds.Settings.(*types.SlackSettings)
		if !ok {
			util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Invalid slack settings", nil))
			return
		}

		slackDs, _ := NewSlackDatasource(*ds, NewConcreteSlackClient(slackSettings.Token), dm.logger)
		if err := slackDs.Validate(); err != nil {
			util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Datasource validation failed", err))
			return
		}

		util.SendSuccessResponse(w, http.StatusOK, map[string]string{"result": "success"}, dm.logger, span)
		return
	}

	util.SendAPIErrorResponse(
		w,
		http.StatusBadRequest,
		util.NewAPIError(
			"Unsupported datasource type",
			fmt.Errorf("unsupported datasource type: %s", ds.SourceType),
		),
	)
}

func (dm *Manager) getDatasourcesHandler(w http.ResponseWriter, r *http.Request) {
	if !dm.aclManager.IsAllowed(w, r, types.UserRoleAdmin, "", types.APiAccessOpsDatasourceGet) {
		return
	}

	ctx, span := observability.StartSpan(r.Context(), "api.getDatasourcesHandler")
	defer span.End()

	page, perPage := dm.getPaginationParams(r)

	paginatedDatasources, err := dm.storage.GetAllDatasources(ctx, page, perPage)
	if err != nil {
		dm.logger.WithError(err).Error("Failed to get datasources")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, util.NewAPIError("Failed to get datasources", err))
		return
	}

	if paginatedDatasources == nil || paginatedDatasources.Datasources == nil {
		dm.logger.Debug("No datasources found. Creating empty paginated datasources")
		paginatedDatasources = dm.createEmptyPaginatedDatasources(page, perPage)
	}

	util.SendSuccessResponse(w, http.StatusOK, paginatedDatasources, dm.logger, nil)
}

func (dm *Manager) updateDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	if !dm.aclManager.IsAllowed(w, r, types.UserRoleAdmin, "", types.APiAccessOpsDatasourceUpdate) {
		return
	}

	if !dm.aclManager.IsAllowed(w, r, types.UserRoleAdmin, "", "datasource_update") {
		return
	}

	ctx, span := observability.StartSpan(r.Context(), "api.updateDatasourceHandler")
	defer span.End()

	existingDS, err := dm.getDatasourceFromRequest(w, r)
	if err != nil {
		return
	}

	var updatePayload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Invalid request body", err))
		return
	}

	dm.logger.WithField("source_type", existingDS.SourceType).Info("Updating datasource")
	newSettings, err := dm.updateDatasourceSettings(*existingDS, updatePayload)
	if err != nil {
		dm.logger.Error("Failed to update datasource settings: ", err)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Failed to update datasource settings", err))
		return
	}

	if err := dm.storage.UpdateDatasource(ctx, existingDS.UUID, newSettings, existingDS.State); err != nil {
		dm.logger.WithError(err).WithFields(logrus.Fields{
			"datasource_id": existingDS.UUID,
			"source_type":   existingDS.SourceType,
		}).Error("Failed to update datasource")

		util.SendAPIErrorResponse(w, http.StatusInternalServerError, util.NewAPIError("Failed to update datasource", err))
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Datasource has been updated successfully"}, dm.logger, span)
}

func (dm *Manager) updateDatasourceSettings(existingDS types.DatasourceConfig, updatePayload map[string]interface{}) (types.DatasourceSettings, error) {
	switch existingDS.SourceType {
	case "slack":
		slackSettings, ok := existingDS.Settings.(*types.SlackSettings)
		if !ok {
			return nil, errors.New("invalid settings type for Slack datasource")
		}
		if channelID, ok := updatePayload["channel_id"].(string); ok {
			slackSettings.ChannelID = channelID
		}
		return slackSettings, nil
	case "github":
		githubSettings, ok := existingDS.Settings.(*types.GitHubSettings)
		if !ok {
			return nil, errors.New("invalid settings type for GitHub datasource")
		}
		if org, ok := updatePayload["org"].(string); ok {
			githubSettings.Org = org
		}
		return githubSettings, nil
	default:
		return nil, errors.New("unsupported datasource type")
	}
}

func (dm *Manager) setActiveDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	if !dm.aclManager.IsAllowed(w, r, types.UserRoleAdmin, "", types.APiAccessOpsDatasourceActivate) {
		return
	}

	ds, err := dm.getDatasourceFromRequest(w, r)
	if err != nil {
		return
	}

	if ds.Status == types.DatasourceStatusActive {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Datasource is already active", nil))
		return
	}

	err = dm.storage.SetActiveDatasource(r.Context(), ds.UUID)
	if err != nil {
		dm.logger.WithError(err).Error(types.FailedToSetDatasourceActiveMsg)
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
			Message: types.FailedToSetDatasourceActiveMsg,
			Err:     err.Error(),
		})
		return
	}

	dm.logger.WithField("embedding_provider_id", ds.UUID).Info("Datasource has been activated successfully")
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Datasource has been activated successfully"}, dm.logger, nil)
}

func (dm *Manager) setDisableDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	if !dm.aclManager.IsAllowed(w, r, types.UserRoleAdmin, "", types.APiAccessOpsDatasourceDeactivate) {
		return
	}

	ds, err := dm.getDatasourceFromRequest(w, r)
	if err != nil {
		return
	}

	if ds.Status == types.DatasourceStatusInactive {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Datasource is already inactive", nil))
		return
	}

	err = dm.storage.SetDisableDatasource(r.Context(), ds.UUID)
	if err != nil {
		dm.logger.WithError(err).Error("Failed to deactivate datasource")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
			Message: types.FailedToSetDatasourceActiveMsg,
			Err:     err.Error(),
		})
		return
	}

	dm.logger.WithField("datasource_id", ds.UUID).Info("Datasource has been deactivated successfully")
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Datasource has been deactivated successfully"}, dm.logger, nil)
}

func (dm *Manager) deleteDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	if !dm.aclManager.IsAllowed(w, r, types.UserRoleAdmin, "", types.APiAccessOpsDatasourceDelete) {
		return
	}

	_, span := otel.Tracer("api").Start(r.Context(), "deleteDatasourceHandler")
	defer span.End()

	if !dm.aclManager.IsAllowed(w, r, types.UserRoleAdmin, "", "datasource_delete") {
		return
	}

	ds, err := dm.getDatasourceFromRequest(w, r)
	if err != nil {
		return
	}

	err = dm.storage.DeleteDatasource(r.Context(), ds.UUID)
	if err != nil {
		dm.logger.WithError(err).Error("Failed to deactivate datasource")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
			Message: types.FailedToSetDatasourceActiveMsg,
			Err:     err.Error(),
		})
		return
	}

	dm.logger.WithField("datasource_id", ds.UUID).Info(types.DatasourceDeletedSuccessfullyMsg)
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Datasource has been deleted successfully"}, dm.logger, nil)
}

func (dm *Manager) validatePayload(payload types.DatasourcePayload) error {
	if payload.Name == "" {
		return errors.New(types.DatasourceValidationMsgNameIsRequired)
	}
	if payload.SourceType == "" {
		return errors.New(types.DatasourceValidationMsgSourceTypeIsRequired)
	}

	return nil
}

func (dm *Manager) createEmptyPaginatedDatasources(page, perPage int) *types.PaginatedDatasources {
	return &types.PaginatedDatasources{
		Datasources: []types.DatasourceConfig{},
		Total:       0,
		Page:        page,
		PerPage:     perPage,
		TotalPages:  0,
	}
}

func (dm *Manager) getPaginationParams(r *http.Request) (int, int) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(r.URL.Query().Get("per_page"))
	if err != nil || perPage < 1 {
		perPage = 10
	}

	return page, perPage
}
