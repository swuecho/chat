# Request Decoding and Validation Pipeline

This document describes how the Go backend decodes and validates HTTP input before invoking application services. The goal is to reject malformed or unsafe input consistently at the transport boundary while keeping business rules in the service and domain layers.

## Pipeline overview

An API request passes through these stages:

```text
HTTP request
  -> request body size limit
  -> UUID route-parameter validation
  -> authentication and rate limiting, when applicable
  -> strict JSON decoding
  -> request DTO validation
  -> handler mapping
  -> application service
  -> database/provider
```

The main pieces are:

- `api/middleware.BodyLimitMiddleware`: limits ordinary request bodies to 1 MiB.
- `api/middleware.ValidateUUIDRouteParams`: validates UUID-like Gorilla Mux route parameters.
- `api/pkg/util.DecodeJSON`: performs strict JSON decoding and invokes DTO validation.
- `api/validation`: contains reusable transport-level validation rules and shared limits.
- Handler-local or `dto` request types: describe the accepted JSON contract and implement `Validate() error` when semantic validation is required.

## Strict JSON decoding

Handlers must use the `DecodeJSON` helper re-exported by `api/handler/util.go` and return its error to the endpoint adapter:

```go
var request createWidgetRequest
if err := DecodeJSON(r, &request); err != nil {
    return err
}
```

Do not decode request bodies directly with `json.NewDecoder(r.Body).Decode(...)`.

`DecodeJSON` enforces the following rules:

1. The body must not be empty.
2. JSON syntax and field types must be valid.
3. Unknown fields are rejected through `DisallowUnknownFields`.
4. The body must contain exactly one JSON value.
5. If the target implements `Validate() error`, validation runs after decoding.

For example, each of these bodies is rejected:

```json
{"name":"example","unexpected":true}
```

```json
{"name":"first"} {"name":"second"}
```

The strict unknown-field behavior makes request DTOs part of the API contract. When a client needs to send a field, add it deliberately to the appropriate request type. Do not decode into a generated SQLC record or a response DTO merely to accept additional fields.

## Defining a validated request DTO

Request types should be named types so they can implement the validator interface:

```go
type createSessionRequest struct {
    UUID      string  `json:"uuid"`
    Topic     string  `json:"topic"`
    Model     string  `json:"model"`
    MaxTokens int32   `json:"maxTokens"`
    Temperature float64 `json:"temperature"`
}

func (r *createSessionRequest) Validate() error {
    if err := validation.UUID("uuid", r.UUID, true); err != nil {
        return err
    }
    if err := validation.Topic(r.Topic, false); err != nil {
        return err
    }
    if err := validation.ModelName("model", r.Model, true); err != nil {
        return err
    }
    if err := validation.TokenCount("maxTokens", r.MaxTokens, false); err != nil {
        return err
    }
    if r.Temperature < 0 || r.Temperature > 2 {
        return fmt.Errorf("temperature must be between 0 and 2")
    }
    return nil
}
```

Use a pointer receiver for `Validate`. Handlers pass `&request` to `DecodeJSON`, allowing both decoding and validation to operate on the same value.

Validation should return the first actionable error. The handler converts it into the standard validation error response and must return immediately, before calling a service.

## Shared validation rules

Reusable rules live in `api/validation`. Current shared limits are:

| Input | Rule |
| --- | --- |
| UUID | Must parse as a UUID; may be required or optional |
| Topic | At most 200 Unicode characters |
| Model name | At most 200 Unicode characters and no control characters |
| Token count | At most 1,000,000; minimum is zero or one depending on the field |
| Pagination limit | From 1 through 500 |
| Pagination offset | Zero or greater |
| Temperature | From 0 through 2 |
| `topP` | From 0 through 1 |
| Completion count `n` | From 1 through 128 |

Use Unicode character counts for human-readable text limits rather than byte counts. A multi-byte name such as `聊天主题` should count as four characters.

The `allowZero` argument to `validation.TokenCount` distinguishes an optional or calculated count from a required positive limit:

```go
// A calculated count may legitimately be zero.
validation.TokenCount("tokenCount", request.TokenCount, true)

// A provider output limit must be positive.
validation.TokenCount("maxTokens", request.MaxTokens, false)
```

When adding a limit used by multiple endpoints, define it in `api/validation` rather than repeating a literal in handlers.

## UUID route parameters

`ValidateUUIDRouteParams` examines matched Gorilla Mux variables whose names contain `uuid`, case-insensitively. This covers names such as:

- `uuid`
- `sessionUUID`
- `messageUUID`
- `bot_uuid`

Malformed values receive a `400 Bad Request` before the handler or service executes.

Use `id` for numeric database identifiers and parse it explicitly in the handler. Use a parameter name containing `uuid` only when its value is genuinely a UUID.

Handlers may still use `validateUUIDParam` when they need an explicit check close to a sensitive operation. The global middleware remains the default safety net.

When the same identifier appears in both the route and body, verify they match:

```go
pathUUID := mux.Vars(r)["uuid"]
if request.UUID != pathUUID {
    dto.RespondWithAPIError(
        w,
        dto.ErrValidationInvalidInput("request uuid must match path uuid"),
    )
    return
}
```

The path value should normally be authoritative.

## Pagination

Use the centralized typed query helper instead of parsing `limit` or `offset` directly:

```go
page, err := httpx.ParsePage(r)
if err != nil {
    return err
}
```

`httpx.ParsePage` defaults to a limit of 100 and an offset of zero. It rejects malformed integers, limits outside `1..500`, and negative offsets. Map it to `svc.PageWindow{Limit: page.Limit, Offset: page.Offset}` before invoking a service.

For endpoints that only accept a limit:

```go
limit, err := httpx.ParseLimit(r, 20)
if err != nil {
    return err
}
```

Do not silently replace malformed client input with a default. Defaults apply only when the parameter is absent.

## Transport validation versus business validation

Transport validation answers questions about the shape and safe bounds of input:

- Is this a valid UUID?
- Is the topic short enough?
- Is the temperature in the supported numeric range?
- Is the page size bounded?

Application and domain validation answers business questions:

- Does the authenticated user own this session?
- Does the selected model exist and is it enabled?
- May this workspace be deleted?
- Is this state transition currently allowed?

Services must continue enforcing business invariants even when a handler validates its DTO. Services may be called by tests, jobs, or future transports that do not pass through HTTP middleware.

Database constraints remain the final integrity boundary and are not a replacement for either transport or application validation.

## Optional fields and zero values

Go decoding cannot distinguish an omitted scalar field from a field explicitly set to its zero value. Use a pointer when that distinction matters:

```go
type updateModelRequest struct {
    Temperature *float64 `json:"temperature"`
}
```

Then validate only when present:

```go
if r.Temperature != nil && (*r.Temperature < 0 || *r.Temperature > 2) {
    return fmt.Errorf("temperature must be between 0 and 2")
}
```

Do not treat zero as “missing” when zero is a valid domain value. Conversely, document and preserve compatibility where an existing endpoint intentionally treats zero as an omitted default.

## Compatibility considerations

Enabling strict decoding can reveal clients that submit response objects back to update endpoints. Prefer dedicated request DTOs containing only writable fields.

If compatibility requires accepting server-managed fields, declare them explicitly and ignore them during command mapping. Add a comment explaining why they are accepted and ensure authoritative values still come from the path or authentication context.

Never trust client-provided ownership fields such as `userId`, `createdBy`, or `updatedBy`. Obtain identity from the authenticated request context.

## Testing requirements

Reusable validators should have table-driven boundary tests covering:

- The minimum accepted value
- The maximum accepted value
- One value below and above the range
- Empty required and optional fields
- Unicode text boundaries
- Malformed UUIDs
- Control characters where relevant

Decoder tests should cover:

- A valid body
- An empty body
- Invalid JSON syntax
- An unknown field
- Multiple JSON values
- A DTO validation failure

Handler tests should verify that invalid input returns `400` and that the mocked or real service observes no mutation.

Run the backend checks after changing request contracts:

```bash
cd api
go test ./...
go vet ./...
```

The integration suite uses PostgreSQL through its Docker-based test harness.

## New endpoint checklist

Before registering a new JSON endpoint:

1. Define a handler-local request DTO; use `dto` only when the request contract is genuinely shared.
2. Add precise JSON tags for every accepted field.
3. Implement `Validate() error` for semantic bounds and formats.
4. Decode exclusively through `DecodeJSON`.
5. Return immediately on decoding or validation failure.
6. Read the authenticated user ID from context, never from the request body.
7. Verify duplicate path/body identifiers match.
8. Use centralized pagination helpers.
9. Map the validated DTO into a service command or query.
10. Keep ownership and other business invariants in the service layer.
11. Add boundary and handler tests.

For the complete middleware, handler, service, response, and SSE lifecycle, see [HTTP Request and Response Pipeline](HTTP_REQUEST_RESPONSE_PIPELINE.md).
