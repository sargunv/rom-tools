package screenscraper

import (
	"errors"
	"fmt"
	"strings"
)

// APIError represents an error returned by the Screenscraper API
type APIError struct {
	StatusCode int
	Message    string
	Type       ErrorType
}

// ErrorType represents the type of API error
type ErrorType int

const (
	// ErrorTypeBadRequest represents a 400 Bad Request error
	ErrorTypeBadRequest ErrorType = iota
	// ErrorTypeUnauthorized represents a 401 Unauthorized error
	ErrorTypeUnauthorized
	// ErrorTypeForbidden represents a 403 Forbidden error
	ErrorTypeForbidden
	// ErrorTypeNotFound represents a 404 Not Found error
	ErrorTypeNotFound
	// ErrorTypeLocked represents a 423 Locked error
	ErrorTypeLocked
	// ErrorTypeUpgradeRequired represents a 426 Upgrade Required error
	ErrorTypeUpgradeRequired
	// ErrorTypeTooManyRequests represents a 429 Too Many Requests error
	ErrorTypeTooManyRequests
	// ErrorTypeQuotaExceeded represents a 430 Quota Exceeded error
	ErrorTypeQuotaExceeded
	// ErrorTypeQuotaKOExceeded represents a 431 Quota KO Exceeded error
	ErrorTypeQuotaKOExceeded
	// ErrorTypeUnknown represents an unknown error type
	ErrorTypeUnknown
)

// BadRequest error subtypes
var (
	ErrURLProblem         = &APIError{StatusCode: 400, Type: ErrorTypeBadRequest, Message: "Problem with URL"}
	ErrMissingFields      = &APIError{StatusCode: 400, Type: ErrorTypeBadRequest, Message: "Missing required fields in URL"}
	ErrROMFilenamePath    = &APIError{StatusCode: 400, Type: ErrorTypeBadRequest, Message: "Error in ROM filename: contains a path"}
	ErrHashFormat         = &APIError{StatusCode: 400, Type: ErrorTypeBadRequest, Message: "CRC, MD5 or SHA1 field error"}
	ErrROMFilenameInvalid = &APIError{StatusCode: 400, Type: ErrorTypeBadRequest, Message: "Problem in ROM filename"}
)

// Common API errors
var (
	ErrUnauthorized    = &APIError{StatusCode: 401, Type: ErrorTypeUnauthorized, Message: "API closed for non-members or inactive members"}
	ErrForbidden       = &APIError{StatusCode: 403, Type: ErrorTypeForbidden, Message: "Login error: Check your developer credentials!"}
	ErrNotFound        = &APIError{StatusCode: 404, Type: ErrorTypeNotFound, Message: "Error: Game not found! / Error: Rom/Iso/Folder not found!"}
	ErrLocked          = &APIError{StatusCode: 423, Type: ErrorTypeLocked, Message: "API completely closed"}
	ErrUpgradeRequired = &APIError{StatusCode: 426, Type: ErrorTypeUpgradeRequired, Message: "The scraping software used has been blacklisted (non-compliant / obsolete version)"}
	ErrTooManyRequests = &APIError{StatusCode: 429, Type: ErrorTypeTooManyRequests, Message: "The number of threads allowed for the member is reached"}
	ErrQuotaExceeded   = &APIError{StatusCode: 430, Type: ErrorTypeQuotaExceeded, Message: "Your scraping quota is exceeded for today!"}
	ErrQuotaKOExceeded = &APIError{StatusCode: 431, Type: ErrorTypeQuotaKOExceeded, Message: "Clean up your ROM files and come back tomorrow!"}
)

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error (HTTP %d)", e.StatusCode)
}

// Is implements the errors.Is interface for error comparison
func (e *APIError) Is(target error) bool {
	if t, ok := target.(*APIError); ok {
		return e.StatusCode == t.StatusCode && e.Type == t.Type
	}
	return false
}

// newAPIError creates a new APIError with the given status code and message
func newAPIError(statusCode int, message string) *APIError {
	errorType := mapStatusCodeToErrorType(statusCode, message)
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Type:       errorType,
	}
}

// mapStatusCodeToErrorType maps HTTP status codes and error messages to ErrorType
func mapStatusCodeToErrorType(statusCode int, message string) ErrorType {
	messageLower := strings.ToLower(message)

	switch statusCode {
	case 400:
		// Map specific 400 error subtypes based on message content
		if strings.Contains(messageLower, "url") && strings.Contains(messageLower, "no information") {
			return ErrorTypeBadRequest // ErrURLProblem
		}
		if strings.Contains(messageLower, "missing") || strings.Contains(messageLower, "required fields") {
			return ErrorTypeBadRequest // ErrMissingFields
		}
		if strings.Contains(messageLower, "rom filename") && strings.Contains(messageLower, "path") {
			return ErrorTypeBadRequest // ErrROMFilenamePath
		}
		if strings.Contains(messageLower, "crc") || strings.Contains(messageLower, "md5") || strings.Contains(messageLower, "sha1") {
			return ErrorTypeBadRequest // ErrHashFormat
		}
		if strings.Contains(messageLower, "rom filename") || strings.Contains(messageLower, "not compliant") {
			return ErrorTypeBadRequest // ErrROMFilenameInvalid
		}
		return ErrorTypeBadRequest
	case 401:
		return ErrorTypeUnauthorized
	case 403:
		return ErrorTypeForbidden
	case 404:
		return ErrorTypeNotFound
	case 423:
		return ErrorTypeLocked
	case 426:
		return ErrorTypeUpgradeRequired
	case 429:
		return ErrorTypeTooManyRequests
	case 430:
		return ErrorTypeQuotaExceeded
	case 431:
		return ErrorTypeQuotaKOExceeded
	default:
		return ErrorTypeUnknown
	}
}

// IsNotFound returns true if the error is a 404 Not Found error
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == ErrorTypeNotFound
	}
	return false
}

// IsQuotaExceeded returns true if the error is a quota exceeded error (430 or 431)
func IsQuotaExceeded(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == ErrorTypeQuotaExceeded || apiErr.Type == ErrorTypeQuotaKOExceeded
	}
	return false
}

// IsAuthenticationError returns true if the error is an authentication error (401 or 403)
func IsAuthenticationError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == ErrorTypeUnauthorized || apiErr.Type == ErrorTypeForbidden
	}
	return false
}

// IsRateLimited returns true if the error is a rate limit error (429)
func IsRateLimited(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == ErrorTypeTooManyRequests
	}
	return false
}

// GetHTTPStatusCode returns the HTTP status code from an error if it's an APIError
func GetHTTPStatusCode(err error) (int, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode, true
	}
	return 0, false
}

// IsBadRequest returns true if the error is a 400 Bad Request error
func IsBadRequest(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == ErrorTypeBadRequest
	}
	return false
}
