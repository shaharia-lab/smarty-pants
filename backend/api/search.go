package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/shaharia-lab/smarty-pants/internal/search"
	"github.com/sirupsen/logrus"
)

func addSearchHandler(searchSystem search.System, logging *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logging.Error("Failed to read request body: ", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		var searchReq search.Request
		if err := json.Unmarshal(body, &searchReq); err != nil {
			logging.Error("Failed to unmarshal request body: ", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		results, err := searchSystem.SearchDocument(ctx, searchReq)
		if err != nil {
			return
		}

		SendSuccessResponse(w, http.StatusOK, results, logging, nil)
	}
}
