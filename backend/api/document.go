package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/shaharia-lab/smarty-pants/backend/internal/observability"
	"github.com/shaharia-lab/smarty-pants/backend/internal/storage"
	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// getDocumentsHandler returns a handler function that fetches documents from storage
func getDocumentsHandler(st storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := observability.StartSpan(r.Context(), "/api/v1/documents")
		defer span.End()

		filter, option := parseDocumentQueryParams(r)

		span.SetAttributes(
			attribute.Int("query.limit", option.Limit),
			attribute.Int("query.page", option.Page),
		)

		_, storageSpan := observability.StartSpan(ctx, "storage.get")
		paginatedDocuments, err := st.Get(ctx, filter, option)
		if err != nil {
			logging.WithError(err).Error("Failed to fetch documents")
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch documents", logging, span)
			return
		}
		storageSpan.End()

		if len(paginatedDocuments.Documents) == 0 {
			paginatedDocuments = createEmptyPaginatedDocuments(option.Page, option.Limit)
		}

		SendSuccessResponse(w, http.StatusOK, paginatedDocuments, logging, span)
	}
}

// getDocumentHandler returns a handler function that fetches a single document from storage
func getDocumentHandler(st storage.Storage, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer("api").Start(r.Context(), "GetDocumentHandler")
		defer span.End()

		uuid := chi.URLParam(r, "uuid")
		span.SetAttributes(attribute.String("document.uuid", uuid))

		filter := types.DocumentFilter{UUID: uuid}
		option := types.DocumentFilterOption{Limit: 1, Page: 1}

		paginatedDocuments, err := st.Get(ctx, filter, option)
		if err != nil {
			logging.WithError(err).Error("Failed to fetch document")
			SendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch document", logging, span)
			return
		}

		if len(paginatedDocuments.Documents) == 0 {
			logging.WithField("uuid", uuid).Warning("Document not found")
			SendErrorResponse(w, http.StatusNotFound, "Document not found", logging, span)
			return
		}

		if len(paginatedDocuments.Documents) > 1 {
			logging.Error("Multiple documents found")
			SendErrorResponse(w, http.StatusInternalServerError, "Unexpected error: multiple documents found", logging, span)
			return
		}

		SendSuccessResponse(w, http.StatusOK, &paginatedDocuments.Documents[0], logging, span)
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
