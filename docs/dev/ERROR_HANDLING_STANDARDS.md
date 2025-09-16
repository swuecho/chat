# Error Handling Standardization Guide

## Current Issues
- Inconsistent use of `http.Error()` vs `RespondWithAPIError()`
- Mixed error response formats
- Some areas missing proper logging

## Standards to Follow

### 1. Use RespondWithAPIError for all API responses
Replace:
```go
http.Error(w, err.Error(), http.StatusBadRequest)
```

With:
```go
RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
```

### 2. Always log errors before responding
```go
log.WithError(err).Error("Context-specific error message")
RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("User-friendly message"))
```

### 3. Use appropriate error types
- `ErrValidationInvalidInput` for bad request data
- `ErrResourceNotFound` for 404 errors
- `ErrInternalUnexpected` for 500 errors
- `ErrPermissionDenied` for 403 errors

### 4. Files that need standardization
- chat_auth_user_handler.go (partially fixed)
- admin_handler.go
- chat_prompt_hander.go 
- chat_comment_handler.go
- chat_message_handler.go
- handle_tts.go

## Implementation Priority
1. Authentication handlers (high impact)
2. Core chat functionality  
3. Admin and utility handlers

This standardization should be done gradually to avoid breaking changes.