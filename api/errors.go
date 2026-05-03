package main

import (
	"github.com/swuecho/chat_backend/dto"
)

// Re-export all error types and functions from dto for backward compatibility.
type APIError = dto.APIError

var (
	NewAPIError                 = dto.CreateAPIError
	RespondWithAPIError         = dto.RespondWithAPIError
	MapDatabaseError            = dto.MapDatabaseError
	WrapError                   = dto.WrapError
	IsErrorCode                 = dto.IsErrorCode
	createAPIError              = dto.CreateAPIError
	ErrorCatalogHandler         = dto.ErrorCatalogHandler
	ErrResourceNotFound         = dto.ErrResourceNotFound
	ErrResourceAlreadyExists    = dto.ErrResourceAlreadyExists
	ErrValidationInvalidInput   = dto.ErrValidationInvalidInput
	RespondWithJSON             = dto.RespondWithJSON
)

// Re-export error code prefixes.
const (
	ErrAuth       = dto.ErrAuth
	ErrValidation = dto.ErrValidation
	ErrResource   = dto.ErrResource
	ErrDatabase   = dto.ErrDatabase
	ErrExternal   = dto.ErrExternal
	ErrInternal   = dto.ErrInternal
	ErrModel      = dto.ErrModel
)

// Re-export all pre-defined errors.
var (
	ErrAuthInvalidCredentials       = dto.ErrAuthInvalidCredentials
	ErrAuthExpiredToken             = dto.ErrAuthExpiredToken
	ErrAuthAdminRequired            = dto.ErrAuthAdminRequired
	ErrAuthInvalidEmailOrPassword   = dto.ErrAuthInvalidEmailOrPassword
	ErrAuthAccessDenied             = dto.ErrAuthAccessDenied
	ErrResourceNotFoundGeneric      = dto.ErrResourceNotFoundGeneric
	ErrResourceAlreadyExistsGeneric = dto.ErrResourceAlreadyExistsGeneric
	ErrTooManyRequests              = dto.ErrTooManyRequests
	ErrChatSessionNotFound          = dto.ErrChatSessionNotFound
	ErrChatFileNotFound             = dto.ErrChatFileNotFound
	ErrChatModelNotFound            = dto.ErrChatModelNotFound
	ErrChatMessageNotFound          = dto.ErrChatMessageNotFound
	ErrValidationInvalidInputGeneric = dto.ErrValidationInvalidInputGeneric
	ErrChatFileTooLarge             = dto.ErrChatFileTooLarge
	ErrChatFileInvalidType          = dto.ErrChatFileInvalidType
	ErrChatSessionInvalid           = dto.ErrChatSessionInvalid
	ErrDatabaseQuery                = dto.ErrDatabaseQuery
	ErrDatabaseConnection           = dto.ErrDatabaseConnection
	ErrDatabaseForeignKey           = dto.ErrDatabaseForeignKey
	ErrExternalTimeout              = dto.ErrExternalTimeout
	ErrExternalUnavailable          = dto.ErrExternalUnavailable
	ErrInternalUnexpected           = dto.ErrInternalUnexpected
	ErrChatStreamFailed             = dto.ErrChatStreamFailed
	ErrChatRequestFailed            = dto.ErrChatRequestFailed
	ErrSystemMessageError           = dto.ErrSystemMessageError
	ErrClaudeStreamFailed           = dto.ErrClaudeStreamFailed
	ErrClaudeRequestFailed          = dto.ErrClaudeRequestFailed
	ErrClaudeInvalidResponse        = dto.ErrClaudeInvalidResponse
	ErrClaudeResponseFaild          = dto.ErrClaudeResponseFaild
	ErrOpenAIStreamFailed           = dto.ErrOpenAIStreamFailed
	ErrOpenAIRequestFailed          = dto.ErrOpenAIRequestFailed
	ErrOpenAIInvalidResponse        = dto.ErrOpenAIInvalidResponse
	ErrOpenAIConfigFailed           = dto.ErrOpenAIConfigFailed
)
