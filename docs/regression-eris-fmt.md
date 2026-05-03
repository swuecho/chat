# Regression: `eris.Wrap` → `fmt.Errorf` inline return bug

## Summary

Commit `46a1c55` introduced a regression that caused e2e browser tests to fail. The root cause was a mechanical replacement of `eris.Wrap` with `fmt.Errorf` in inline return statements, without accounting for a critical behavioral difference between the two functions when wrapping a `nil` error.

## Root Cause

### The Bug

```go
// BEFORE (correct) — eris.Wrap returns nil when err is nil
return w, eris.Wrap(err, "failed to create workspace")

// AFTER (broken) — fmt.Errorf returns non-nil even when err is nil
return w, fmt.Errorf("failed to create workspace: %w", err)
```

**`eris.Wrap(nil, "msg")` returns `nil`.**  
**`fmt.Errorf("msg: %w", nil)` returns `"msg: %!w(<nil>)"` — a non-nil error.**

This means every successful service call was returning a fake error. Handlers received this error, wrapped it via `dto.WrapError()` / `dto.MapDatabaseError()`, and returned HTTP 500 to the client.

## Scope

79 occurrences across 12 files in `api/svc/`:

| File | Occurrences |
|------|-------------|
| `chat_session_service.go` | 18 |
| `chat_workspace_service.go` | 16 |
| `bot_answer_history_service.go` | 9 |
| `chat_message_service.go` | 9 |
| `chat_auth_user_service.go` | 6 |
| `chat_main_service.go` | 5 |
| `chat_comment_service.go` | 5 |
| `chat_user_active_chat_session_service.go` | 4 |
| `jwt_secret_service.go` | 3 |
| `chat_prompt_service.go` | 2 |
| `chat_snapshot_service.go` | 1 |
| `file_upload_service.go` | 1 |

## Pattern

All affected code followed the same anti-pattern — an **inline error wrap without an `if err != nil` guard**:

```go
func (s *SomeService) SomeMethod(ctx context.Context, ...) (Result, error) {
    x, err := s.q.Query(ctx, params)
    return x, eris.Wrap(err, "operation failed")  // ← inline wrap, no nil check
}
```

This pattern relies on `eris.Wrap(nil, _)` returning `nil`. Standard library `fmt.Errorf` does not have this behavior.

## Why It Wasn't Caught by Tests

- **Unit tests**: The `svc/` package has no tests (0% coverage).
- **Compilation**: Both versions compile successfully — the bug is purely semantic.
- **`go vet`**: No warnings — both are valid Go error handling.

## Fix

The `svc/` files were restored to the parent commit (`dd12f2d`) which retains `eris`. A proper migration to `fmt.Errorf` requires adding `if err != nil` guards for every inline wrap:

```go
// Correct fmt.Errorf usage:
func (s *SomeService) SomeMethod(ctx context.Context, ...) (Result, error) {
    x, err := s.q.Query(ctx, params)
    if err != nil {
        return Result{}, fmt.Errorf("operation failed: %w", err)
    }
    return x, nil
}
```

## Prevention

1. **Never inline-wrap errors without a nil check.** Always use the `if err != nil { wrap } return val, nil` pattern.
2. **Add tests for service packages.** A simple unit test calling any service method would have caught this immediately.
3. **Treat mechanical refactoring as high-risk.** Sed-based replacements should always be followed by targeted unit tests of the affected code paths.
4. **Use `golangci-lint` with `errorlint`.** The `errorlint` linter can detect improper `%w` usage with nil errors.
