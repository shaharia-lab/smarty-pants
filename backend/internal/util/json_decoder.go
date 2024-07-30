package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	JsonErroFoundUnknownField = "Found unknown field"
)

func DecodeJSONBody(r *http.Request, v interface{}) *APIError {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(v)
	if err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			return NewAPIError(JsonErroFoundUnknownField, err)

		case errors.As(err, &unmarshalTypeError):
			return NewAPIError(fmt.Sprintf("Failed to decode JSON: field '%s' has incorrect type (expected %v)",
				unmarshalTypeError.Field, unmarshalTypeError.Type), err)

		default:
			return NewAPIError("failed to decode JSON", err)
		}
	}

	return nil
}
