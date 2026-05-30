package apperrors

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	CodeInvalidRequest Code = "invalid_request"
	CodeRateLimited    Code = "rate_limited"
	CodeOpenAI         Code = "openai_error"
	CodeNotFound       Code = "not_found"
	CodeInternal       Code = "internal_error"
)

type AppError struct {
	Code       Code
	Message    string
	HTTPStatus int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err == nil {
		return string(e.Code) + ": " + e.Message
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code Code, status int, message string) *AppError {
	return &AppError{Code: code, HTTPStatus: status, Message: message}
}

func Wrap(code Code, status int, message string, err error) *AppError {
	return &AppError{Code: code, HTTPStatus: status, Message: message, Err: err}
}

func Invalid(message string) *AppError {
	return New(CodeInvalidRequest, http.StatusBadRequest, message)
}

func OpenAI(message string, err error) *AppError {
	return Wrap(CodeOpenAI, http.StatusBadGateway, message, err)
}

func Internal(message string, err error) *AppError {
	return Wrap(CodeInternal, http.StatusInternalServerError, message, err)
}

func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
