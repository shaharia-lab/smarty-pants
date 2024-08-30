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
	"go.opentelemetry.io/otel/trace"
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

func (dm *Manager) addDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !dm.aclManager.IsAllowed(w, r, types.UserRoleAdmin, "datasource_add") {
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
		dm.handleError(w, err.Error(), http.StatusBadRequest, span)
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
		dm.handleError(w, "Failed to add datasource", http.StatusInternalServerError, span)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Datasource added successfully",
		"uuid":    dsConfig.UUID,
	})
}

func (dm *Manager) getDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("api").Start(r.Context(), "getDatasourceHandler")
	defer span.End()

	dsUUID := chi.URLParam(r, "uuid")

	dsUUIDParsed, err := uuid.Parse(dsUUID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		dm.handleError(w, types.InvalidUUIDMessage, http.StatusBadRequest, span)
		return
	}

	ds, err := dm.storage.GetDatasource(ctx, dsUUIDParsed)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		dm.handleError(w, types.DatasourceNotFoundMsg, http.StatusNotFound, span)
		return
	}

	util.SendSuccessResponse(w, http.StatusOK, ds, dm.logger, span)
}

func (dm *Manager) validateDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("api").Start(r.Context(), "api.validateDatasourceHandler")
	defer span.End()

	dsUUID := chi.URLParam(r, "uuid")
	dsUUIDParsed, err := uuid.Parse(dsUUID)
	if err != nil {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError(types.InvalidUUIDMessage, err))
		return
	}

	ds, err := dm.storage.GetDatasource(ctx, dsUUIDParsed)
	if err != nil {
		util.SendAPIErrorResponse(w, http.StatusNotFound, util.NewAPIError(types.DatasourceNotFoundMsg, err))
		return
	}

	if ds.SourceType == types.DatasourceTypeSlack {
		slackSettings, ok := ds.Settings.(*types.SlackSettings)
		if !ok {
			util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Invalid slack settings", nil))
			return
		}

		slackDs, _ := NewSlackDatasource(ds, NewConcreteSlackClient(slackSettings.Token), dm.logger)
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
	ctx, span := observability.StartSpan(r.Context(), "api.updateDatasourceHandler")
	defer span.End()

	dsUUID := chi.URLParam(r, "uuid")
	if dsUUID == "" {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Missing UUID", nil))
		return
	}

	var updatePayload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Invalid request body", err))
		return
	}

	dsUUIDParsed, err := uuid.Parse(dsUUID)
	if err != nil {
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError(types.InvalidUUIDMessage, err))
		return
	}

	existingDS, err := dm.storage.GetDatasource(ctx, dsUUIDParsed)
	if err != nil {
		util.SendAPIErrorResponse(w, http.StatusNotFound, util.NewAPIError(types.DatasourceNotFoundMsg, err))
		return
	}

	dm.logger.WithField("source_type", existingDS.SourceType).Info("Updating datasource")
	newSettings, err := dm.updateDatasourceSettings(existingDS, updatePayload)
	if err != nil {
		dm.logger.Error("Failed to update datasource settings: ", err)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Failed to update datasource settings", err))
		return
	}

	if err := dm.storage.UpdateDatasource(ctx, dsUUIDParsed, newSettings, existingDS.State); err != nil {
		dm.logger.WithError(err).WithFields(logrus.Fields{
			"datasource_id": dsUUID,
			"source_type":   existingDS.SourceType,
		}).Error("Failed to update datasource")

		util.SendAPIErrorResponse(w, http.StatusInternalServerError, util.NewAPIError("Failed to update datasource", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	id, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err == nil && id == uuid.Nil {
		err = errors.New(types.InvalidUUIDMessage)
	}

	if err != nil || id == uuid.Nil {
		dm.logger.WithError(err).Error(types.InvalidUUIDMessage)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{
			Message: types.InvalidUUIDMessage,
			Err:     err.Error(),
		})
		return
	}

	err = dm.storage.SetActiveDatasource(r.Context(), id)
	if err != nil {
		dm.logger.WithError(err).Error(types.FailedToSetDatasourceActiveMsg)
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
			Message: types.FailedToSetDatasourceActiveMsg,
			Err:     err.Error(),
		})
		return
	}

	dm.logger.WithField("embedding_provider_id", id).Info("Datasource has been activated successfully")
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Datasource has been activated successfully"}, dm.logger, nil)
}

func (dm *Manager) setDisableDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	datasourceUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err == nil && datasourceUUID == uuid.Nil {
		err = errors.New(types.InvalidUUIDMessage)
	}

	if err != nil || datasourceUUID == uuid.Nil {
		dm.logger.WithError(err).Error(types.InvalidUUIDMessage)
		util.SendAPIErrorResponse(w, http.StatusBadRequest, &util.APIError{
			Message: types.InvalidUUIDMessage,
			Err:     err.Error(),
		})
		return
	}

	err = dm.storage.SetDisableDatasource(r.Context(), datasourceUUID)
	if err != nil {
		dm.logger.WithError(err).Error("Failed to deactivate datasource")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
			Message: types.FailedToSetDatasourceActiveMsg,
			Err:     err.Error(),
		})
		return
	}

	dm.logger.WithField("datasource_id", datasourceUUID).Info("Datasource has been deactivated successfully")
	util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Datasource has been deactivated successfully"}, dm.logger, nil)
}

func (dm *Manager) deleteDatasourceHandler(w http.ResponseWriter, r *http.Request) {
	datasourceUUID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil || datasourceUUID == uuid.Nil {
		dm.logger.WithError(err).Error(types.InvalidUUIDMessage)
		dm.sendJSONError(w, types.InvalidUUIDMessage, http.StatusBadRequest)
		return
	}

	err = dm.storage.DeleteDatasource(r.Context(), datasourceUUID)
	if err != nil {
		dm.logger.WithError(err).Error("Failed to deactivate datasource")
		util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
			Message: types.FailedToSetDatasourceActiveMsg,
			Err:     err.Error(),
		})
		return
	}

	dm.logger.WithField("datasource_id", datasourceUUID).Info(types.DatasourceDeletedSuccessfullyMsg)
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

func (dm *Manager) handleError(w http.ResponseWriter, message string, statusCode int, span trace.Span) {
	dm.logger.Error(message)
	if span != nil {
		span.RecordError(errors.New(message))
		span.SetStatus(codes.Error, message)
	}
	dm.sendJSONError(w, message, statusCode)
}

// Deprecated: sendJSONError is deprecated and should be replaced with util.SendAPIErrorResponse
func (dm *Manager) sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
