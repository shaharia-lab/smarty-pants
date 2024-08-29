package document

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/shaharia-lab/smarty-pants/backend/internal/util"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type DocumentManager struct {
	storage storage.Storage
	logger  *logrus.Logger
}

func NewDocumentManager(storage storage.Storage, logger *logrus.Logger) *DocumentManager {
	return &DocumentManager{
		storage: storage,
		logger:  logger,
	}
}

func (dm *DocumentManager) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/document", func(r chi.Router) {
		r.Get("/", dm.GetDocumentsHandler())
		r.Get("/{uuid}", dm.GetDocumentHandler())
	})
}

func (dm *DocumentManager) GetDocumentsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "/api/v1/documents")
		defer span.End()

		filter, option := parseDocumentQueryParams(r)

		span.SetAttributes(
			attribute.Int("query.limit", option.Limit),
			attribute.Int("query.page", option.Page),
		)

		_, storageSpan := observability.StartSpan(ctx, "storage.get")
		paginatedDocuments, err := dm.storage.Get(ctx, filter, option)
		if err != nil {
			dm.logger.WithError(err).Error("Failed to fetch documents")
			util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch documents", dm.logger, span)
			return
		}
		storageSpan.End()

		if len(paginatedDocuments.Documents) == 0 {
			paginatedDocuments = createEmptyPaginatedDocuments(option.Page, option.Limit)
		}

		util.SendSuccessResponse(w, http.StatusOK, paginatedDocuments, dm.logger, span)
	}
}

func (dm *DocumentManager) GetDocumentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer("api").Start(r.Context(), "GetDocumentHandler")
		defer span.End()

		uuid := chi.URLParam(r, "uuid")
		span.SetAttributes(attribute.String("document.uuid", uuid))

		filter := types.DocumentFilter{UUID: uuid}
		option := types.DocumentFilterOption{Limit: 1, Page: 1}

		paginatedDocuments, err := dm.storage.Get(ctx, filter, option)
		if err != nil {
			dm.logger.WithError(err).Error("Failed to fetch document")
			util.SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch document", dm.logger, span)
			return
		}

		if len(paginatedDocuments.Documents) == 0 {
			dm.logger.WithField("uuid", uuid).Warning("Document not found")
			util.SendErrorResponse(w, http.StatusNotFound, "Document not found", dm.logger, span)
			return
		}

		if len(paginatedDocuments.Documents) > 1 {
			dm.logger.Error("Multiple documents found")
			util.SendErrorResponse(w, http.StatusInternalServerError, "Unexpected error: multiple documents found", dm.logger, span)
			return
		}

		util.SendSuccessResponse(w, http.StatusOK, &paginatedDocuments.Documents[0], dm.logger, span)
	}
}

func parseDocumentQueryParams(r *http.Request) (types.DocumentFilter, types.DocumentFilterOption) {
	var filter types.DocumentFilter
	var option types.DocumentFilterOption

	filter.UUID = r.URL.Query().Get("uuid")
	filter.Status = types.DocumentStatus(r.URL.Query().Get("status"))
	filter.SourceUUID = r.URL.Query().Get("source_uuid")

	option.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if option.Limit <= 0 {
		option.Limit = 10
	}

	option.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if option.Page <= 0 {
		option.Page = 1
	}

	return filter, option
}

func createEmptyPaginatedDocuments(page, limit int) types.PaginatedDocuments {
	return types.PaginatedDocuments{
		Documents:  []types.Document{},
		Total:      0,
		Page:       page,
		PerPage:    limit,
		TotalPages: 0,
	}
}
