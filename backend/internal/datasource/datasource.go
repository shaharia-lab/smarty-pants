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

func AddDatasourceHandler(st storage.Storage, logging *logrus.Logger, aclManager auth.ACLManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if !aclManager.IsAllowed(w, r, types.UserRoleAdmin, "analytics") {
			return
		}

		ctx, span := observability.StartSpan(ctx, "api.addDatasourceHandler")
		defer span.End()

		var payload types.DatasourcePayload
		if err := util.DecodeJSONBody(r, &payload); err != nil {
			util.SendAPIErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		if err := ValidatePayload(payload); err != nil {
			HandleError(w, err.Error(), http.StatusBadRequest, logging, span)
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

		if err := st.AddDatasource(ctx, dsConfig); err != nil {
			logging.Error("Failed to add datasource: ", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			HandleError(w, "Failed to add datasource", http.StatusInternalServerError, logging, span)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Datasource added successfully",
			"uuid":    dsConfig.UUID,
		})
	}
}

func GetDatasourceHandler(storage storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer("api").Start(r.Context(), "getDatasourceHandler")
		defer span.End()

		dsUUID := chi.URLParam(r, "uuid")

		dsUUIDParsed, err := uuid.Parse(dsUUID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			HandleError(w, types.InvalidUUIDMessage, http.StatusBadRequest, logging, span)
			return
		}

		ds, err := storage.GetDatasource(ctx, dsUUIDParsed)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			HandleError(w, types.DatasourceNotFoundMsg, http.StatusNotFound, logging, span)
			return
		}

		util.SendSuccessResponse(w, http.StatusOK, ds, logging, span)
	}
}

func ValidateDatasourceHandler(storage storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer("api").Start(r.Context(), "api.validateDatasourceHandler")
		defer span.End()

		dsUUID := chi.URLParam(r, "uuid")
		dsUUIDParsed, err := uuid.Parse(dsUUID)
		if err != nil {
			util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError(types.InvalidUUIDMessage, err))
			return
		}

		ds, err := storage.GetDatasource(ctx, dsUUIDParsed)
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

			slackDs, _ := NewSlackDatasource(ds, NewConcreteSlackClient(slackSettings.Token), logging)
			if err := slackDs.Validate(); err != nil {
				util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Datasource validation failed", err))
				return
			}

			util.SendSuccessResponse(w, http.StatusOK, map[string]string{"result": "success"}, logging, span)
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
}

func GetDatasourcesHandler(st storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "api.getDatasourcesHandler")
		defer span.End()

		page, perPage := GetPaginationParams(r)

		paginatedDatasources, err := st.GetAllDatasources(ctx, page, perPage)
		if err != nil {
			logging.WithError(err).Error("Failed to get datasources")
			util.SendAPIErrorResponse(w, http.StatusInternalServerError, util.NewAPIError("Failed to get datasources", err))
			return
		}

		if paginatedDatasources == nil || paginatedDatasources.Datasources == nil {
			logging.Debug("No datasources found. Creating empty paginated datasources")
			paginatedDatasources = CreateEmptyPaginatedDatasources(page, perPage)
		}

		util.SendSuccessResponse(w, http.StatusOK, paginatedDatasources, logging, nil)
	}
}

func UpdateDatasourceHandler(st storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		existingDS, err := st.GetDatasource(ctx, dsUUIDParsed)
		if err != nil {
			util.SendAPIErrorResponse(w, http.StatusNotFound, util.NewAPIError(types.DatasourceNotFoundMsg, err))
			return
		}

		logging.WithField("source_type", existingDS.SourceType).Info("Updating datasource")
		newSettings, err := UpdateDatasourceSettings(existingDS, updatePayload)
		if err != nil {
			logging.Error("Failed to update datasource settings: ", err)
			util.SendAPIErrorResponse(w, http.StatusBadRequest, util.NewAPIError("Failed to update datasource settings", err))
			return
		}

		if err := st.UpdateDatasource(ctx, dsUUIDParsed, newSettings, existingDS.State); err != nil {
			logging.WithError(err).WithFields(logrus.Fields{
				"datasource_id": dsUUID,
				"source_type":   existingDS.SourceType,
			}).Error("Failed to update datasource")

			util.SendAPIErrorResponse(w, http.StatusInternalServerError, util.NewAPIError("Failed to update datasource", err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func UpdateDatasourceSettings(existingDS types.DatasourceConfig, updatePayload map[string]interface{}) (types.DatasourceSettings, error) {
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

func SetActiveDatasourceHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil || id == uuid.Nil {
			l.WithError(err).Error(types.InvalidUUIDMessage)
			SendJSONError(w, types.InvalidUUIDMessage, http.StatusBadRequest)
			return
		}

		err = s.SetActiveDatasource(r.Context(), id)
		if err != nil {
			l.WithError(err).Error(types.FailedToSetDatasourceActiveMsg)
			util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
				Message: types.FailedToSetDatasourceActiveMsg,
				Err:     err.Error(),
			})
			return
		}

		l.WithField("embedding_provider_id", id).Info("Datasource has been activated successfully")
		util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Datasource has been activated successfully"}, l, nil)
	}
}

func SetDisableDatasourceHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil || id == uuid.Nil {
			l.WithError(err).Error(types.InvalidUUIDMessage)
			SendJSONError(w, types.InvalidUUIDMessage, http.StatusBadRequest)
			return
		}

		err = s.SetDisableDatasource(r.Context(), id)
		if err != nil {
			l.WithError(err).Error("Failed to deactivate datasource")
			util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
				Message: types.FailedToSetDatasourceActiveMsg,
				Err:     err.Error(),
			})
			return
		}

		l.WithField("datasource_id", id).Info("Datasource has been deactivated successfully")
		util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Datasource has been deactivated successfully"}, l, nil)
	}
}

func DeleteDatasourceHandler(s storage.Storage, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "uuid"))
		if err != nil || id == uuid.Nil {
			l.WithError(err).Error(types.InvalidUUIDMessage)
			SendJSONError(w, types.InvalidUUIDMessage, http.StatusBadRequest)
			return
		}

		err = s.DeleteDatasource(r.Context(), id)
		if err != nil {
			l.WithError(err).Error("Failed to deactivate datasource")
			util.SendAPIErrorResponse(w, http.StatusInternalServerError, &util.APIError{
				Message: types.FailedToSetDatasourceActiveMsg,
				Err:     err.Error(),
			})
			return
		}

		l.WithField("datasource_id", id).Info(types.DatasourceDeletedSuccessfullyMsg)
		util.SendSuccessResponse(w, http.StatusOK, map[string]string{"message": "Datasource has been deleted successfully"}, l, nil)
	}
}

func ValidatePayload(payload types.DatasourcePayload) error {
	if payload.Name == "" {
		return errors.New(types.DatasourceValidationMsgNameIsRequired)
	}
	if payload.SourceType == "" {
		return errors.New(types.DatasourceValidationMsgSourceTypeIsRequired)
	}

	return nil
}

func CreateEmptyPaginatedDatasources(page, perPage int) *types.PaginatedDatasources {
	return &types.PaginatedDatasources{
		Datasources: []types.DatasourceConfig{},
		Total:       0,
		Page:        page,
		PerPage:     perPage,
		TotalPages:  0,
	}
}

func GetPaginationParams(r *http.Request) (int, int) {
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

func HandleError(w http.ResponseWriter, message string, statusCode int, logging *logrus.Logger, span trace.Span) {
	logging.Error(message)
	if span != nil {
		span.RecordError(errors.New(message))
		span.SetStatus(codes.Error, message)
	}
	SendJSONError(w, message, statusCode)
}

func SendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
