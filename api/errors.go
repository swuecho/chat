package main

import "errors"

// ErrUsageLimitExceeded is returned when the usage limit is exceeded.
var ErrUsageLimitExceeded = errors.New("usage limit exceeded")


// auth related

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidUserID = errors.New("invalid user id")

/// token related

// ErrTokenExpired is returned when the token is expired.
var ErrTokenExpired = errors.New("token expired")

// ErrTokenNotYetValid is returned when the token is not yet valid.
var ErrTokenNotYetValid = errors.New("token not yet valid")

// ErrTokenMalformed is returned when the token is malformed.
var ErrTokenMalformed = errors.New("token malformed")

// ErrTokenInvalid is returned when the token is invalid.
var ErrTokenInvalid = errors.New("token invalid")

// openapi related

// ErrInvalidOpenAPI is returned when the openapi is invalid.
var ErrInvalidOpenAPI = errors.New("invalid openapi")

// ErrOpenAiApiError is returned when the openai api returns an error.
var ErrOpenAiApiError = errors.New("openai api error")

// http service related

// ErrBadRequest is returned when the request is malformed or contains invalid data.
var ErrBadRequest = errors.New("bad request")

// ErrNotFound is returned when the requested resource is not found.
var ErrNotFound = errors.New("not found")

// ErrInternalServerError is returned when an internal server error occurs.
var ErrInternalServerError = errors.New("internal server error")

// ErrServiceUnavailable is returned when the service is temporarily unavailable.
var ErrServiceUnavailable = errors.New("service unavailable")

// ErrRateLimitExceeded is returned when the client has sent too many requests in a given amount of time.
var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// More errors

// ErrPermissionDenied is returned when the user does not have permission to access the requested resource.
var ErrPermissionDenied = errors.New("permission denied")

// ErrResourceConflict is returned when there is a conflict with the requested resource.
var ErrResourceConflict = errors.New("resource conflict")

// ErrUnprocessableEntity is returned when the server cannot process the request due to client-side errors.
var ErrUnprocessableEntity = errors.New("unprocessable entity")

// ErrGatewayTimeout is returned when the server did not receive a timely response from an external service.
var ErrGatewayTimeout = errors.New("gateway timeout")
