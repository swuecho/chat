# Full Migration Plan: Gorilla Mux to Gin

## Overview
Migrate the entire backend from Gorilla Mux to Gin framework.

## Migration Steps

### Phase 1: Preparation
- [x] Add Gin dependency
- [x] Create initial Gin router structure
- [ ] Create Gin middleware functions

### Phase 2: Middleware Migration
- [ ] Migrate `middleware_authenticate.go` → Gin middleware
- [ ] Migrate `middleware_cors.go` → Gin middleware
- [ ] Migrate `middleware_logging.go` → Gin middleware
- [ ] Migrate `middleware_rateLimit.go` → Gin middleware
- [ ] Migrate `middleware_validation.go` → Gin middleware
- [ ] Migrate `middleware_gzip.go` → Gin middleware
- [ ] Migrate `middleware_lastRequestTime.go` → Gin middleware

### Phase 3: Handler Migration
List of handlers to migrate:
1. [ ] `chat_main_handler.go` - Chat endpoints
2. [ ] `chat_session_handler.go` - Session CRUD
3. [ ] `chat_message_handler.go` - Message operations
4. [ ] `chat_workspace_handler.go` - Workspace management
5. [ ] `chat_auth_user_handler.go` - Auth & user
6. [ ] `chat_model_handler.go` - Model management
7. [ ] `chat_prompt_hander.go` - Prompt templates
8. [ ] `chat_snapshot_handler.go` - Snapshots
9. [ ] `chat_comment_handler.go` - Comments
10. [ ] `file_upload_handler.go` - File uploads
11. [ ] `admin_handler.go` - Admin operations
12. [ ] `chat_model_privilege_handler.go` - Model privileges
13. [ ] `chat_user_active_chat_session_handler.go` - Active sessions
14. [ ] `bot_answer_history_handler.go` - Bot history

### Phase 4: Main.go Update
- [ ] Replace mux router with Gin engine
- [ ] Update route registrations
- [ ] Update middleware application

### Phase 5: Cleanup
- [ ] Remove mux imports from all files
- [ ] Remove mux from go.mod
- [ ] Test full build

## Handler Migration Pattern

### Before (mux)
```go
func (h *Handler) Register(router *mux.Router) {
    router.HandleFunc("/path", h.Handler).Methods("GET")
}

func (h *Handler) Handler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]
    // ...
    json.NewEncoder(w).Encode(resp)
}
```

### After (Gin)
```go
func (h *Handler) Register(router *gin.Router) {
    router.GET("/path", h.Handler)
}

func (h *Handler) Handler(c *gin.Context) {
    id := c.Param("id")
    // ...
    c.JSON(200, resp)
}
```

## Files to Delete After Migration
- `gin_hybrid.go` (temporary file)
- All `*_handler.go` files (migrated)

## New Directory Structure
```
api/
├── dto/              # Request/Response DTOs
├── handlers/         # Migrated Gin handlers
├── middleware/       # Migrated Gin middleware
├── services/        # Business logic (unchanged)
├── main.go          # Updated
└── ...
```
