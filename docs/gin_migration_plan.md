# Gin Framework Migration Plan

## Overview

Migrate from Gorilla Mux to Gin framework to reduce boilerplate, add automatic validation, and improve developer experience.

**Estimated Duration**: 3-4 days
**Risk Level**: Medium (significant changes to routing layer)

---

## Phase 1: Preparation & Dependencies

### 1.1 Add Gin Dependencies
```bash
cd api
go get -u github.com/gin-gonic/gin
go get -u github.com/gin-contrib/cors
go get -u github.com/gin-contrib/gzip
```

### 1.2 Remove Gorilla Dependencies (after migration complete)
```bash
go remove github.com/gorilla/mux
go remove github.com/gorilla/handlers
```

### 1.3 Create Feature Branch
```bash
git checkout -b feature/gin-migration
```

---

## Phase 2: Core Infrastructure

### 2.1 Create Gin Router Setup
**File**: `api/router.go` (new file)

Tasks:
- [ ] Create `SetupRouter() *gin.Engine` function
- [ ] Configure Gin mode (debug/release based on environment)
- [ ] Set up trusted proxies
- [ ] Configure default middleware (logger, recovery)

### 2.2 Migrate Middleware
**Files**: `api/middleware_*.go`

| Current File | New File | Changes |
|--------------|----------|---------|
| `middleware_authenticate.go` | Same | Convert to `gin.HandlerFunc` |
| `middleware_rateLimit.go` | Same | Convert to `gin.HandlerFunc` |
| `middleware_gzip.go` | Remove | Use `gin-contrib/gzip` |
| `middleware_validation.go` | Same | Convert to `gin.HandlerFunc` |
| `middleware_lastRequestTime.go` | Same | Convert to `gin.HandlerFunc` |

Middleware signature change:
```go
// Before (Gorilla)
func UserAuthMiddleware(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
        handler.ServeHTTP(w, r)
    })
}

// After (Gin)
func UserAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // ...
        c.Next()
    }
}
```

### 2.3 Update Error Handling
**File**: `api/errors.go`

Tasks:
- [ ] Add `GinResponse(c *gin.Context)` method to `APIError`
- [ ] Keep existing `RespondWithAPIError` for gradual migration
- [ ] Create helper functions for common error responses

```go
func (e APIError) GinResponse(c *gin.Context) {
    c.JSON(e.HTTPCode, gin.H{
        "code":    e.Code,
        "message": e.Message,
        "detail":  e.Detail,
    })
}
```

### 2.4 Create Context Helpers
**File**: `api/context.go` (new file)

```go
package main

import "github.com/gin-gonic/gin"

// GetUserID extracts user ID from Gin context
func GetUserID(c *gin.Context) (int32, error) {
    userID, exists := c.Get("user_id")
    if !exists {
        return 0, ErrAuthInvalidCredentials
    }
    switch v := userID.(type) {
    case int32:
        return v, nil
    case int:
        return int32(v), nil
    case string:
        // Parse string to int32
    }
    return 0, ErrAuthInvalidCredentials
}

// GetUserRole extracts user role from Gin context
func GetUserRole(c *gin.Context) string {
    role, _ := c.Get("role")
    return role.(string)
}
```

---

## Phase 3: Handler Migration

### 3.1 Handler Migration Order
Migrate handlers in order of complexity (simplest first):

| Order | Handler File | Routes | Complexity |
|-------|--------------|--------|------------|
| 1 | `chat_model_handler.go` | 2-3 | Low |
| 2 | `chat_prompt_hander.go` | 3-4 | Low |
| 3 | `chat_workspace_handler.go` | 5-6 | Low |
| 4 | `chat_comment_handler.go` | 3-4 | Low |
| 5 | `chat_file_handler.go` | 4-5 | Medium |
| 6 | `chat_snapshot_handler.go` | 5-6 | Medium |
| 7 | `chat_message_handler.go` | 5-6 | Medium |
| 8 | `chat_session_handler.go` | 8-10 | Medium |
| 9 | `chat_auth_user_handler.go` | 6-8 | Medium-High |
| 10 | `admin_handler.go` | 10+ | Medium |
| 11 | `chat_model_privilege_handler.go` | 4-5 | Medium |
| 12 | `bot_answer_history_handler.go` | 3-4 | Medium |
| 13 | `chat_main_handler.go` | 1-2 | High (SSE streaming) |

### 3.2 Handler Migration Template

For each handler, complete these tasks:

#### Step A: Update Handler Registration
```go
// Before
func (h *ChatSessionHandler) Register(router *mux.Router) {
    router.HandleFunc("/chat_sessions/user", h.getSimpleChatSessionsByUserID).Methods(http.MethodGet)
}

// After
func (h *ChatSessionHandler) Register(rg *gin.RouterGroup) {
    rg.GET("/chat_sessions/user", h.getSimpleChatSessionsByUserID)
}
```

#### Step B: Update Handler Methods
```go
// Before
func (h *ChatSessionHandler) getChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
    uuid := mux.Vars(r)["uuid"]
    // ...
    json.NewEncoder(w).Encode(response)
}

// After
func (h *ChatSessionHandler) getChatSessionByUUID(c *gin.Context) {
    uuid := c.Param("uuid")
    // ...
    c.JSON(200, response)
}
```

#### Step C: Add Request DTOs with Validation (where applicable)
```go
type CreateSessionRequest struct {
    UUID     string `json:"uuid" binding:"required"`
    Topic    string `json:"topic" binding:"required,max=200"`
    ModelID  int32  `json:"model_id" binding:"required"`
}
```

#### Step D: Update Context Access
```go
// Before
ctx := r.Context()
userID, err := getUserID(ctx)

// After
userID, err := GetUserID(c)
```

---

## Phase 4: Main.go Refactoring

### 4.1 Restructure main.go
Current `main.go` is ~330 lines. Split into:

| New File | Purpose | Lines (est.) |
|----------|---------|--------------|
| `main.go` | Entry point, DB setup, server start | ~80 |
| `router.go` | Router setup and route registration | ~100 |
| `config.go` | Configuration loading | ~60 |
| `context.go` | Context helpers | ~40 |

### 4.2 New main.go Structure
```go
func main() {
    // 1. Load config
    cfg := LoadConfig()

    // 2. Setup database
    db := SetupDatabase(cfg)

    // 3. Setup services
    services := SetupServices(db)

    // 4. Setup router
    router := SetupRouter(services)

    // 5. Start server
    StartServer(router, cfg)
}
```

### 4.3 Router.go Structure
```go
func SetupRouter(services *Services) *gin.Engine {
    r := gin.New()

    // Global middleware
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    r.Use(cors.New(CORSConfig()))

    // API routes
    api := r.Group("/api")

    // Public routes
    SetupPublicRoutes(api, services)

    // User routes (authenticated)
    user := api.Group("", UserAuthMiddleware())
    SetupUserRoutes(user, services)

    // Admin routes
    admin := api.Group("/admin", AdminAuthMiddleware())
    SetupAdminRoutes(admin, services)

    // Static files
    SetupStaticFiles(r)

    return r
}
```

---

## Phase 5: Special Cases

### 5.1 SSE Streaming (chat_main_handler.go)
The streaming handler needs special attention:

```go
func (h *ChatHandler) handleChat(c *gin.Context) {
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // Use c.Writer for streaming
    flusher, ok := c.Writer.(http.Flusher)
    if !ok {
        c.JSON(500, gin.H{"error": "Streaming unsupported"})
        return
    }

    // Stream response
    for chunk := range chunks {
        fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
        flusher.Flush()
    }
}
```

### 5.2 File Upload Handling
```go
func (h *ChatFileHandler) uploadFile(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": "File required"})
        return
    }

    // Process file...
}
```

### 5.3 Query Parameters
```go
// Before
page := r.URL.Query().Get("page")

// After
page := c.Query("page")
limit := c.DefaultQuery("limit", "10")
```

### 5.4 Path Parameters
```go
// Before
uuid := mux.Vars(r)["uuid"]

// After
uuid := c.Param("uuid")
```

---

## Phase 6: Testing

### 6.1 Update Existing Tests
- Update any tests that use `httptest` with Gorilla Mux
- Use Gin's test utilities

```go
func TestGetChatSession(t *testing.T) {
    router := setupTestRouter()

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/api/uuid/chat_sessions/123", nil)
    req.Header.Set("Authorization", "Bearer "+testToken)

    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
}
```

### 6.2 Add Integration Tests
- Test middleware chain
- Test route groups
- Test authentication flows

---

## Phase 7: Cleanup

### 7.1 Remove Old Dependencies
```bash
go mod tidy
```

### 7.2 Remove Deprecated Files
- [ ] Remove `middleware_gzip.go` (replaced by gin-contrib/gzip)
- [ ] Remove unused helper functions
- [ ] Clean up imports

### 7.3 Update Documentation
- [ ] Update CLAUDE.md with new architecture
- [ ] Update deployment docs if needed
- [ ] Update development setup docs

---

## Phase 8: Deployment Checklist

### 8.1 Pre-deployment
- [ ] All tests passing
- [ ] No deprecated imports
- [ ] Environment variables unchanged
- [ ] Database queries unchanged
- [ ] API endpoints unchanged (same paths, same request/response formats)

### 8.2 Deployment
- [ ] Deploy to staging first
- [ ] Run smoke tests
- [ ] Monitor logs for errors
- [ ] Check response times (should be similar or better)

### 8.3 Post-deployment
- [ ] Monitor error rates
- [ ] Check memory usage
- [ ] Verify SSE streaming works
- [ ] Verify file uploads work

---

## Rollback Plan

If issues arise:

1. **Immediate rollback**: Revert to previous commit
2. **Database**: No changes required (SQLC layer unchanged)
3. **Frontend**: No changes required (API contract unchanged)

---

## Detailed Task Checklist

### Week 1: Foundation

#### Day 1
- [ ] Add Gin dependencies
- [ ] Create `router.go` with basic setup
- [ ] Create `context.go` with helpers
- [ ] Migrate CORS middleware

#### Day 2
- [ ] Migrate `middleware_authenticate.go`
- [ ] Migrate `middleware_rateLimit.go`
- [ ] Migrate `middleware_lastRequestTime.go`
- [ ] Test authentication flow

#### Day 3
- [ ] Migrate `chat_model_handler.go`
- [ ] Migrate `chat_prompt_hander.go`
- [ ] Migrate `chat_workspace_handler.go`
- [ ] Test these endpoints

### Week 1 Continued

#### Day 4
- [ ] Migrate `chat_comment_handler.go`
- [ ] Migrate `chat_file_handler.go`
- [ ] Migrate `chat_snapshot_handler.go`
- [ ] Test these endpoints

#### Day 5
- [ ] Migrate `chat_message_handler.go`
- [ ] Migrate `chat_session_handler.go`
- [ ] Test these endpoints

### Week 2: Complex Handlers

#### Day 6
- [ ] Migrate `chat_auth_user_handler.go`
- [ ] Migrate `admin_handler.go`
- [ ] Test these endpoints

#### Day 7
- [ ] Migrate `chat_model_privilege_handler.go`
- [ ] Migrate `bot_answer_history_handler.go`
- [ ] Migrate `chat_main_handler.go` (SSE streaming)
- [ ] Test SSE streaming

#### Day 8
- [ ] Refactor `main.go`
- [ ] Remove old middleware
- [ ] Clean up imports
- [ ] Run all tests

#### Day 9
- [ ] Update documentation
- [ ] Final testing
- [ ] Code review

#### Day 10
- [ ] Deploy to staging
- [ ] Run smoke tests
- [ ] Deploy to production

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Breaking API changes | Keep API contract identical, add integration tests |
| SSE streaming issues | Thorough testing, keep flusher logic |
| Auth middleware bugs | Comprehensive auth tests, session tests |
| Performance regression | Benchmark before/after |
| Missing routes | Compare route lists before/after |

---

## Success Metrics

- [ ] All existing tests pass
- [ ] Code coverage maintained or improved
- [ ] ~30-40% reduction in handler code
- [ ] Response times unchanged or improved
- [ ] No increase in error rates
- [ ] All API endpoints functional

---

## References

- [Gin Documentation](https://gin-gonic.com/docs/)
- [Gin GitHub](https://github.com/gin-gonic/gin)
- [Migration from net/http](https://gin-gonic.com/docs/quick-start/)
