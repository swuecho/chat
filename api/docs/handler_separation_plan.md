# Handler-REST API Separation Plan

## Current State

Handlers (`*_handler.go`) currently contain:
1. HTTP request/response handling (marshal/unmarshal, http.ResponseWriter)
2. Request/Response DTOs
3. Service instantiation and delegation
4. Route registration (`Register(*mux.Router)`)

Services (`*_service.go`) contain business logic - already well separated.

## Target State

### Directory Structure
```
api/
├── handlers/           # NEW: HTTP layer only
│   ├── chat_handler.go
│   ├── session_handler.go
│   ├── message_handler.go
│   └── ...
├── services/          # Existing: Business logic (rename from *_service.go)
│   ├── chat_service.go
│   ├── session_service.go
│   └── ...
├── middleware/        # NEW: HTTP middleware
├── dto/               # NEW: Request/Response structs
│   ├── chat.go
│   ├── session.go
│   └── ...
├── main.go
└── ...
```

### Key Changes

#### 1. Extract Request/Response DTOs
Move structs to separate `dto/` package:
```go
// dto/chat.go
type ChatRequest struct {
    Prompt      string `json:"prompt"`
    SessionUuid string `json:"session_uuid"`
    // ...
}
```

#### 2. Create Handler Interfaces
Handlers should depend on service interfaces, not implementations:
```go
type ChatServiceInterface interface {
    ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
    // ...
}

type ChatHandler struct {
    service ChatServiceInterface
}
```

#### 3. Clean Handler Methods
Handlers should only:
- Parse request
- Validate input
- Call service
- Write response

```go
func (h *ChatHandler) ChatCompletion(c *gin.Context) {
    var req dto.ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, Error{Message: err.Error()})
        return
    }

    resp, err := h.service.ChatCompletion(c.Request.Context(), req)
    if err != nil {
        c.JSON(500, Error{Message: err.Error()})
        return
    }

    c.JSON(200, resp)
}
```

#### 4. Migration Steps

1. **Create `dto/` package** - Extract all request/response structs
2. **Create `handlers/` package** - Move handler methods, use interfaces
3. **Update services** - Add interfaces for testability
4. **Update `main.go`** - Wire up handlers with services
5. **Add `middleware/` package** - Move middleware files

### Benefits

- **Testability**: Handlers can be tested with mock services
- **Separation of concerns**: HTTP vs business logic clearly separated
- **Reusability**: Services can be used by other entrypoints (CLI, gRPC)
- **Maintainability**: Clear boundaries, easier to understand
- **Gin migration**: Handlers become framework-agnostic first

### Files to Create/Modify

| Action | File |
|--------|------|
| CREATE | `dto/chat.go` |
| CREATE | `dto/session.go` |
| CREATE | `dto/message.go` |
| CREATE | `dto/workspace.go` |
| CREATE | `dto/user.go` |
| CREATE | `handlers/chat_handler.go` |
| CREATE | `handlers/session_handler.go` |
| CREATE | `handlers/message_handler.go` |
| CREATE | `handlers/workspace_handler.go` |
| CREATE | `handlers/user_handler.go` |
| RENAME | `chat_main_service.go` → `services/chat_service.go` |
| RENAME | `chat_session_service.go` → `services/session_service.go` |
| MOVE | `middleware_*.go` → `middleware/` |
| UPDATE | `main.go` - Update imports and wiring |

### Notes

- Keep existing service implementations as they are mostly clean
- Add interfaces to services for dependency injection
- Request validation can stay in handlers or move to dto with tags
