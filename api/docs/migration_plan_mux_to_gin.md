# Migration Plan: Gorilla Mux to Gin

## Overview

Migrate the backend from `github.com/gorilla/mux` to `github.com/gin-gonic/gin` using an **incremental hybrid approach**.

## Strategy: Incremental Migration

The recommended approach is to run both mux and Gin in parallel, migrating endpoints one at a time:

1. **Keep existing mux routes** - All current handlers continue to work
2. **Add Gin router** alongside mux - New endpoints use Gin
3. **Migrate endpoints gradually** - One handler at a time
4. **Full migration** - Once all endpoints migrated, switch to Gin-only

## Changes Required

### 1. Update Dependencies

**Done:** Added `github.com/gin-gonic/gin` to `go.mod`

```bash
go get github.com/gin-gonic/gin
go mod tidy
```

### 2. Hybrid Router Setup

**File: `gin_hybrid.go`** (created)

Shows how to integrate Gin with mux:

```go
type GinHandler struct {
    engine *gin.Engine
}

func (h *GinHandler) Register(router *mux.Router) {
    router.PathPrefix("/api/gin").Handler(http.HandlerFunc(h.engine.ServeHTTP))
}
```

### 3. Update Handler Register Methods

Each handler file needs changes. Pattern for each handler:

**Before (Gorilla Mux):**
```go
import "github.com/gorilla/mux"

func (h *Handler) Register(router *mux.Router) {
    router.HandleFunc("/path", h.Handler).Methods("GET")
}
```

**After (Gin):**
```go
import "github.com/gin-gonic/gin"

func (h *Handler) Register(router *gin.RouterGroup) {
    router.POST("/path", h.Handler)
    // or router.GET("/path", h.Handler)
}
```

### 4. Update Request Handling in Handlers

| Gorilla Mux | Gin |
|-------------|-----|
| `mux.Vars(r)` | `c.Params` |
| `r.URL.Query().Get("key")` | `c.Query("key")` |
| `r.FormValue("key")` | `c.PostForm("key")` |
| `r.Header.Get("key")` | `c.GetHeader("key")` |
| `json.NewDecoder(r.Body).Decode(&v)` | `c.ShouldBindJSON(&v)` |
| `w http.ResponseWriter` | `c *gin.Context` |
| `return` (after response) | `c.JSON()` or `c.String()` then return |

### 5. Middleware Migration

**Before:**
```go
func Middleware(w http.ResponseWriter, r *http.Request) {
    // logic
    next.ServeHTTP(w, r)
}
router.Use(Middleware)
```

**After:**
```go
func Middleware(c *gin.Context) {
    // logic
    c.Next()
}
router.Use(Middleware)
```

### 6. Files to Update

| File | Changes |
|------|---------|
| `main.go` | Router init, imports |
| `chat_main_handler.go` | Register, all methods |
| `chat_session_handler.go` | Register, all methods |
| `chat_message_handler.go` | Register, all methods |
| `chat_workspace_handler.go` | Register, all methods |
| `chat_auth_user_handler.go` | Register, all methods |
| `chat_model_handler.go` | Register, all methods |
| `chat_prompt_handler.go` | Register, all methods |
| `chat_snapshot_handler.go` | Register, all methods |
| `chat_comment_handler.go` | Register, all methods |
| `file_upload_handler.go` | Register, all methods |
| `admin_handler.go` | Register, all methods |
| `chat_model_privilege_handler.go` | Register, all methods |
| `chat_user_active_chat_session_handler.go` | Register, all methods |
| `middleware_authenticate.go` | Update middleware signatures |
| `middleware_cors.go` | Update middleware signatures |
| `middleware_logging.go` | Update middleware signatures |
| `middleware_rate_limit.go` | Update middleware signatures |
| `middleware_validation.go` | Update middleware signatures |

### 7. Response Helpers

Update error/response helpers in `errors.go` and `util.go`:

```go
// Gin-friendly version
func RespondWithAPIError(c *gin.Context, err APIError) {
    c.JSON(err.HTTPCode, err)
}

func RespondWithJSON(c *gin.Context, status int, payload interface{}) {
    c.JSON(status, payload)
}
```

## Migration Order

1. Add Gin dependency
2. Create Gin router in main.go (keep mux for now)
3. Update middleware files
4. Update handlers one by one (start with simple ones)
5. Update main.go to use Gin router
6. Test and fix issues
7. Remove gorilla/mux dependency

## Benefits After Migration

- Faster routing (radix tree)
- Built-in binding/validation
- Cleaner middleware pattern
- Better error handling
- More idiomatic Go code
- Active maintenance (Gin is more popular than gorilla/mux)
