# HTTP Request and Response Pipeline

This document defines the backend HTTP boundary. Its purpose is to make every endpoint follow the same visible flow while keeping HTTP, application, and persistence concerns separate.

## Pipeline at a glance

```text
request
  -> access log / request ID / recovery / CORS / body limit
  -> authentication / authorization / rate limit
  -> route and query parsing
  -> strict request DTO decoding and validation
  -> request DTO -> application command or query
  -> service / use case
  -> application result -> typed HTTP response DTO
  -> centralized JSON response or error writer
```

Streaming endpoints use a separate final stage:

```text
validated request -> service/provider -> commit SSE -> typed events -> terminal event
```

Before SSE is committed, an error uses the normal JSON error envelope. After commitment, the handler must emit a `failed` or `canceled` event; it must never try to append a JSON error response.

## Responsibilities

### Middleware

Middleware establishes cross-cutting request state and runs in the explicit order composed in `main.go` with `middleware.Chain`. It owns request IDs, access logging, recovery, CORS, body limits, authentication, authorization, and rate limiting. Authentication stores a typed `requestctx.Principal`; handlers should not inspect raw context keys.

The access logger records request ID, method, path, status, response bytes, and duration. Its response wrapper preserves flushing so SSE continues to work.

### Handler

A handler owns HTTP only:

1. Parse route and query input with `httpx.UUIDParam`, `Int32Param`, `Int64Param`, `ParsePage`, or `ParseLimit`.
2. Decode a handler-local request DTO with `DecodeJSON`.
3. Map the validated DTO and authenticated principal into a service command/query.
4. Invoke one focused service or use case.
5. Map the application result into a typed response DTO.
6. Return `respondJSON`, `noContent`, or an error.

Ordinary handlers have the signature:

```go
func (h *WidgetHandler) create(w http.ResponseWriter, r *http.Request) error
```

and are registered through `endpoint(h.create)`. Do not call `dto.RespondWithAPIError`, `json.NewEncoder(w)`, or `w.WriteHeader` inside an ordinary handler.

### Application service

Services accept `Command` or `Query` values without JSON tags. They enforce ownership, permissions, state transitions, transactions, and other business invariants. A service must not depend on `http.Request`, handler DTOs, or response DTOs.

Service results are application models. Generated SQLC records should be mapped at the persistence/application boundary rather than becoming the public API contract.

### Response boundary

`httpx.Adapt` converts a returned error into the standard error envelope once. `httpx.JSON` marshals before committing headers, so encoding failures can still be handled correctly. Successful payloads use named response structs; avoid `map[string]any` and anonymous transport contracts.

The standard offset pagination envelope is:

```json
{
  "items": [],
  "total": 0,
  "limit": 100,
  "offset": 0
}
```

Use `httpx.PageResponse[T]` (via the handler helper) and carry `svc.PageWindow` into the application layer.

## Example

```go
type createWidgetRequest struct {
    Name string `json:"name"`
}

func (r *createWidgetRequest) Validate() error {
    return validation.Topic(r.Name, true)
}

type widgetHTTPResponse struct {
    UUID string `json:"uuid"`
    Name string `json:"name"`
}

func (h *WidgetHandler) create(w http.ResponseWriter, r *http.Request) error {
    principal, err := authenticatedPrincipal(r)
    if err != nil {
        return err
    }
    var request createWidgetRequest
    if err := DecodeJSON(r, &request); err != nil {
        return err
    }

    result, err := h.service.Create(r.Context(), svc.CreateWidgetCommand{
        UserID: principal.UserID,
        Name: request.Name,
    })
    if err != nil {
        return err
    }
    return respondJSON(w, http.StatusCreated, widgetHTTPResponse{
        UUID: result.UUID,
        Name: result.Name,
    })
}
```

The mapping is intentionally visible. It prevents JSON tags and client-controlled ownership fields from leaking into services, and prevents database schema changes from silently changing the API.

## Streaming and proxy endpoints

Register JSON/SSE endpoints with `streamEndpoint`. Validate and authorize before the first flush whenever possible. Once streaming begins:

- deltas use typed `provider.AnswerEvent` values;
- success ends with `completed`;
- cancellation ends with `canceled`;
- provider or persistence failure ends with `failed`;
- no JSON error envelope is written after commitment.

Gateway-compatible, file download, and TTS proxy handlers are explicit protocol boundaries. They may preserve upstream status, headers, binary bodies, or OpenAI-compatible error shapes and should not be registered through the ordinary JSON adapter unless their protocol is changed deliberately.

## Review checklist

For every new or changed endpoint, verify:

1. The route uses the ordinary or streaming adapter explicitly.
2. Body, route, query, and pagination input are parsed centrally.
3. The request DTO is strict and validates safe bounds.
4. Identity comes from the typed principal, not client input.
5. DTO-to-command and result-to-response mappings are explicit.
6. Services do not import HTTP or transport DTOs.
7. Responses use named types and one centralized writer.
8. Errors are returned, not written in multiple layers.
9. SSE errors respect the pre-commit/post-commit boundary.
10. Tests cover invalid input, mapper behavior, error status, and no service mutation on rejection.

The architecture test in `api/handler/pipeline_architecture_test.go` prevents ordinary handler files from reintroducing ad hoc response writes. Run `go test ./...` and `go vet ./...` after pipeline changes.
