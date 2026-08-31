# Typed API contracts

The `api/apicontract` package keeps ordinary JSON route registration, strict
request decoding, response encoding, and OpenAPI documentation connected to the
same Go request and response types.

```go
apicontract.RegisterJSON(router, registry, apicontract.Operation{
    Method:        http.MethodPut,
    Path:          "/uuid/chat_sessions/topic/{uuid}",
    OperationID:   "updateChatSessionTopic",
    Summary:       "Update a chat session topic",
    SuccessStatus: http.StatusOK,
    Parameters:    []apicontract.Parameter{apicontract.UUIDPathParameter("uuid")},
}, handler)
```

The handler has a compile-time request and response contract:

```go
func (h *Handler) updateTopic(
    r *http.Request,
    input updateSessionTopicRequest,
) (chatSessionHTTPResponse, error)
```

`RegisterJSON` performs strict single-value JSON decoding, invokes the existing
`Validate() error` method when implemented, calls the handler, and serializes the
declared response type. `invopop/jsonschema` reflects those same types into the
OpenAPI 3.1 document served at `GET /api/openapi.json`. The interactive Scalar
reference is available at `GET /api/docs`; its API client can persist and send
the JWT bearer token declared by protected operations.

Use `jsonschema` tags for constraints clients can understand:

```go
type updateSessionTopicRequest struct {
    Topic string `json:"topic" jsonschema:"required,minLength=1,maxLength=200"`
}
```

Keep conditional and business validation in `Validate()`. Authorization,
ownership, and persistence invariants remain in application services.

Migrate regular JSON endpoints incrementally. SSE streams, uploads, downloads,
and transparent provider proxies should continue using their specialized
registration paths. `apicontract.NoBody` represents an absent JSON request or
response body.

All chat-session and workspace JSON endpoints use the registry. Those resources
are the reference implementation for migrating another handler group.

## OpenAPI to TypeScript pipeline

Generate the checked-in OpenAPI document and TypeScript SDK together:

```bash
cd web
npm run generate:api
```

The backend generator builds the contract-only router without PostgreSQL and
writes `api/openapi/openapi.json`. Hey API then reads that artifact and replaces
`web/src/api/generated`. Both files are generated outputs and must not be edited
manually. CI repeats the pipeline and fails when the committed output differs.

`web/src/api/generated_client.ts` configures the generated Fetch client with the
same JWT initialization, proactive refresh, cookie, 401 retry, and thrown-error
semantics as the existing Axios boundary. Import generated operations through
that module rather than directly from `api/generated`.

Migration is incremental: `api/chat_session.ts` and `api/chat_workspace.ts` use
generated operations for the endpoints present in OpenAPI while their
message-clearing and active-session calls remain on Axios until those backend
endpoints have typed contracts.
