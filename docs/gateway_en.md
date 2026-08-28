# OpenAI-Compatible LLM Gateway

The gateway exposes configured OpenAI-compatible models through a single API endpoint. Applications authenticate with virtual API keys created in Chat rather than receiving the upstream provider credentials.

## Features

- OpenAI-compatible model discovery and chat completions.
- Streaming and non-streaming responses.
- Transparent request and response forwarding.
- Revocable virtual API keys with per-key request limits.
- Server-managed upstream provider credentials.
- Request status, latency, byte count, token usage, hashes, and structural classifications.
- Bounded request and response samples with configurable retention.
- Admin UI for inspecting keys, usage, requests, and captured samples.

## Supported endpoints

```text
GET  /v1/models
POST /v1/chat/completions
```

The gateway currently exposes enabled models whose API type is `openai`. This includes OpenAI and providers implementing the OpenAI chat-completions protocol.

The gateway does not currently implement other OpenAI endpoints such as `/v1/responses`, `/v1/embeddings`, image generation, audio, or fine-tuning.

## Configure a provider model

Open **Admin → Model** and configure a model with:

- **Model ID**: The value clients send in the `model` field.
- **API Type**: `openai`.
- **URL**: The provider base URL or complete chat-completions URL.
- **API Auth Header**: Usually `Authorization`.
- **API Auth Key**: The name of the server environment variable containing the provider key.
- **Enabled**: The model must be enabled to appear under `/v1/models`.

For example, when **API Auth Key** is `DEEPSEEK_API_KEY`, configure the backend environment—not the database—with:

```bash
export DEEPSEEK_API_KEY='provider-secret'
```

Provider secrets are never returned to gateway users.

If outbound provider traffic requires a proxy, set:

```bash
export OPENAI_PROXY_URL='http://127.0.0.1:7890'
```

## Create a virtual API key

1. Sign in as an administrator.
2. Open **Admin → API keys**.
3. Select **Create API key**.
4. Enter a descriptive name and requests-per-minute limit.
5. Copy the generated `sk-chat-...` key immediately.

The complete key is displayed only once. The database stores its SHA-256 lookup hash and visible prefix, not the plaintext credential. Revoking a key takes effect immediately.

## Use the gateway

### curl

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $GATEWAY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "deepseek-chat",
    "messages": [{"role": "user", "content": "Hello"}],
    "stream": false
  }'
```

### OpenAI SDK

Configure an OpenAI-compatible client with:

```text
base_url = http://localhost:8080/v1
api_key  = <virtual sk-chat key>
```

No provider credential should be placed in the client application.

### Bun verification script

The repository includes `scripts/test_gateway.ts`. It verifies model discovery plus streaming and non-streaming chat completions.

```bash
GATEWAY_API_KEY='sk-chat-...' bun run scripts/test_gateway.ts
```

Optional settings:

```bash
GATEWAY_BASE_URL='http://localhost:8080/v1' \
GATEWAY_MODEL='deepseek-chat' \
GATEWAY_API_KEY='sk-chat-...' \
bun run scripts/test_gateway.ts
```

Never commit a virtual API key to the repository. Revoke any key disclosed in logs, shell history, screenshots, or chat messages.

## Transparent proxy behavior

For valid chat-completion requests, the gateway preserves the original JSON body bytes and forwards safe end-to-end headers. It does not rewrite parameters, inject streaming options, or reconstruct SSE events. Upstream status codes, safe response headers, error bodies, and response bytes are returned unchanged.

The intentional differences from calling a provider directly are:

- Virtual-key authentication and per-key rate limiting.
- Model lookup and provider routing.
- Replacement of the virtual credential with the provider credential.
- Removal of cookies, `Set-Cookie`, and hop-by-hop HTTP headers.
- Gateway-level errors when authentication, routing, configuration, or the provider connection fails.
- The model-specific configured timeout.

## Observability and content retention

Each accepted request records durable metadata:

- User, virtual key, model, and provider.
- Request UUID and provider request ID when available.
- Status, error code, latency, and streaming mode.
- Request and response byte counts.
- Prompt, completion, and total tokens when returned by the provider.
- SHA-256 hashes of the complete request and response bytes.
- Structural request classification: message count, role counts, tool presence, response-format presence, and multimodal presence.
- Response classification: HTTP status, content type, content encoding, streaming mode, and success state.

The gateway also retains a bounded prefix of the request and response bodies. Defaults:

```bash
export GATEWAY_CAPTURE_BYTES=65536
export GATEWAY_RETENTION_DAYS=7
```

After the retention period, captured request and response samples are purged. Metadata, counts, hashes, classifications, status, and latency remain available. Expired samples are cleaned periodically while gateway traffic is active.

Streaming responses are observed while their original bytes are copied to the client. Monitoring does not parse or modify the forwarded stream. Streaming token counts remain zero unless they are available through another accounting mechanism.

Captured samples may contain prompts, generated content, tool inputs, or other sensitive information. Restrict database and admin access, choose the smallest useful capture limit, and select a retention period appropriate for your privacy requirements.

## Inspect gateway activity

Open **Admin → API keys**, then:

1. Select **Requests** for a key.
2. Review its 100 most recent requests.
3. Select **Inspect** for request and response samples, hashes, classifications, token counts, byte counts, status, latency, and retention information.

The list endpoint does not load captured bodies. Samples are fetched only when an individual request is inspected.

## Management API

The admin UI uses the authenticated application API:

```text
GET    /api/api-keys
POST   /api/api-keys
DELETE /api/api-keys/{keyId}
GET    /api/api-keys/{keyId}/usage
GET    /api/api-keys/{keyId}/requests
GET    /api/api-keys/{keyId}/requests/{requestId}
```

These routes use the normal application JWT and only return keys and gateway records owned by the authenticated user. The `/v1` routes use virtual API keys instead.

## Error behavior

Authentication, rate-limit, routing, and gateway transport failures use an OpenAI-shaped error object:

```json
{
  "error": {
    "message": "Invalid API key",
    "type": "invalid_request_error",
    "param": null,
    "code": "invalid_api_key"
  }
}
```

When the provider returns an HTTP response, its status, safe headers, and body are passed through without normalization.
