package util

import "fmt"

type APIError struct {
	Message string `json:"message"`
	Err     string `json:"error"`
	rawErr  error
}

func NewAPIError(message string, err error) *APIError {
	return &APIError{
		Message: message,
		Err:     err.Error(),
		rawErr:  err,
	}
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *APIError) GetRawError() error {
	return e.rawErr
}
