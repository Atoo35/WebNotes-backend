package resterror

import "net/http"

type RestErrorI interface {
	Error() string
	Code() int
}

type RestError struct {
	ErrorMessage string `json:"error"`
	StatusCode   int    `json:"status"`
}

func (re *RestError) Error() string { return re.ErrorMessage }

func (re *RestError) Code() int { return re.StatusCode }
func NewRestError(error string, statusCode int) *RestError {
	return &RestError{ErrorMessage: error, StatusCode: statusCode}
}

func BadRequest(error string) *RestError {
	return NewRestError(error, http.StatusBadRequest)
}

func InternalServerError(error string) *RestError {
	return NewRestError(error, http.StatusInternalServerError)
}

func StandardInternalServerError() *RestError {
	return NewRestError("Internal Server Error", http.StatusInternalServerError)
}

func Forbidden(error string) *RestError {
	return NewRestError(error, http.StatusForbidden)
}

func NotFound(error string) *RestError {
	return NewRestError(error, http.StatusNotFound)
}
