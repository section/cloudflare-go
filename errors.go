package cloudflare

import (
	"fmt"
	"net/http"
	"strings"
)

// Error messages.
const (
	errEmptyCredentials          = "invalid credentials: key & email must not be empty" //nolint:gosec,unused
	errEmptyAPIToken             = "invalid credentials: API Token must not be empty"   //nolint:gosec,unused
	errInternalServiceError      = "internal service error"
	errMakeRequestError          = "error from makeRequest"
	errUnmarshalError            = "error unmarshalling the JSON response"
	errUnmarshalErrorBody        = "error unmarshalling the JSON response error body"
	errRequestNotSuccessful      = "error reported by API"
	errMissingAccountID          = "account ID is empty and must be provided"
	errOperationStillRunning     = "bulk operation did not finish before timeout"
	errOperationUnexpectedStatus = "bulk operation returned an unexpected status"
	errResultInfo                = "incorrect pagination info (result_info) in responses"
	errManualPagination          = "unexpected pagination options passed to functions that handle pagination automatically"
	errInvalidZoneIdentifer      = "invalid zone identifier: %s"
)

type Error struct {
	StatusCode int

	Errors     []ResponseInfo
	ErrorCodes []int

	RayID string
}

func (e Error) Error() string {
	var errString string
	errMessages := []string{}
	for _, err := range e.Errors {
		m := ""
		if err.Message != "" {
			m += err.Message
		}

		if err.Code != 0 {
			m += fmt.Sprintf(" (%d)", err.Code)
		}

		errMessages = append(errMessages, m)
	}

	return errString + strings.Join(errMessages, ", ")
}

// RequestError is for 4xx errors that we encounter not covered elsewhere
// (generally bad payloads).
type RequestError struct {
	cloudflareError *Error
}

func (e RequestError) Error() string {
	return e.cloudflareError.Error()
}

// RatelimitError is for HTTP 429s where the service is telling the client to
// slow down.
type RatelimitError struct {
	cloudflareError *Error
}

func (e RatelimitError) Error() string {
	return e.cloudflareError.Error()
}

// ServiceError is a handler for 5xx errors returned to the client.
type ServiceError struct {
	cloudflareError *Error
}

func (e ServiceError) Error() string {
	return e.cloudflareError.Error()
}

// AuthenticationError is for HTTP 401 responses.
type AuthenticationError struct {
	cloudflareError *Error
}

func (e AuthenticationError) Error() string {
	return e.cloudflareError.Error()
}

// AuthorizationError is for HTTP 403 responses.
type AuthorizationError struct {
	cloudflareError *Error
}

func (e AuthorizationError) Error() string {
	return e.cloudflareError.Error()
}

// NotFoundError is for HTTP 404 responses.
type NotFoundError struct {
	cloudflareError *Error
}

func (e NotFoundError) Error() string {
	return e.cloudflareError.Error()
}

// HTTPStatusCode exposes the HTTP status from the error response encountered.
func (e Error) HTTPStatusCode() int {
	return e.StatusCode
}

// ErrorMessages exposes the error messages as a slice of strings from the error
// response encountered.
func (e *Error) ErrorMessages() []string {
	messages := []string{}

	for _, e := range e.Errors {
		messages = append(messages, e.Message)
	}

	return messages
}

// InternalErrorCodes exposes the internal error codes as a slice of int from
// the error response encountered.
func (e *Error) InternalErrorCodes() []int {
	ec := []int{}

	for _, e := range e.Errors {
		ec = append(ec, e.Code)
	}

	return ec
}

// ServiceError returns a boolean whether or not the raised error was caused by
// an internal service.
func (e *Error) ServiceError() bool {
	return e.StatusCode >= http.StatusInternalServerError &&
		e.StatusCode < 600
}

// ClientError returns a boolean whether or not the raised error was caused by
// something client side.
func (e *Error) ClientError() bool {
	return e.StatusCode >= http.StatusBadRequest &&
		e.StatusCode < http.StatusInternalServerError
}

// ClientRateLimited returns a boolean whether or not the raised error was
// caused by too many requests from the client.
func (e *Error) ClientRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// InternalErrorCodeIs returns a boolean whether or not the desired internal
// error code is present in `e.InternalErrorCodes`.
func (e *Error) InternalErrorCodeIs(code int) bool {
	for _, errCode := range e.InternalErrorCodes() {
		if errCode == code {
			return true
		}
	}

	return false
}

// ErrorMessageContains returns a boolean whether or not a substring exists in
// any of the `e.ErrorMessages` slice entries.
func (e *Error) ErrorMessageContains(s string) bool {
	for _, errMsg := range e.ErrorMessages() {
		if strings.Contains(errMsg, s) {
			return true
		}
	}
	return false
}
