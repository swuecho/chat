# Custom Model API (Custom Provider)

This document describes the custom model API contract used by the backend when a chat model is configured with `api_type = custom`.

## When this path is used

The backend selects the custom provider when the chat model's `api_type` is set to `custom` (Admin → Models → Add Model).

## Required model config (Admin UI)

- **API Type**: `Custom`
- **URL**: Your custom API endpoint (HTTP POST).
- **API Auth Header**: The header name to pass the API key (optional).
- **API Auth Key**: The *environment variable name* that stores the API key.

Example:

- API Auth Header: `x-api-key`
- API Auth Key: `MY_CUSTOM_API_KEY`
- Then set `MY_CUSTOM_API_KEY=...` in the backend environment.

## Request format

The backend always sends a JSON POST body with the following fields:

```json
{
  "prompt": "\n\nHuman: Hello\n\nAssistant: Hi there!\n\nHuman: ...\n\nAssistant: ",
  "model": "custom-my-model",
  "max_tokens_to_sample": 2048,
  "temperature": 0.7,
  "stop_sequences": ["\n\nHuman:"],
  "stream": true
}
```

Notes:

- `prompt` is formatted in a Claude-style transcript. Each non-assistant message becomes:
  `\n\nHuman: <content>\n\nAssistant: `
- `model` is the chat model name stored in the database.
- `max_tokens_to_sample` and `temperature` come from the session settings.
- `stream` is always `true` for the custom provider.

Headers added by the backend:

- `Content-Type: application/json`
- `Accept: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`
- Optional auth header if configured (e.g. `x-api-key: <secret>`).

## Streaming response format (required)

Your API must respond with `text/event-stream` (SSE) and emit lines that start with `data: `.
Each `data:` line must contain JSON that includes a `completion` field.
The backend **expects full text so far** in `completion` on each event (not deltas).

Example stream:

```
data: {"completion":"Hello","stop":null,"stop_reason":null,"truncated":false,"log_id":"1","model":"custom-my-model","exception":null}
data: {"completion":"Hello there","stop":null,"stop_reason":null,"truncated":false,"log_id":"1","model":"custom-my-model","exception":null}
data: [DONE]
```

The stream ends when you send a line starting with `data: [DONE]`.

## Minimum response fields

The backend only uses `completion`, but it unmarshals the following fields:

```json
{
  "completion": "string",
  "stop": "string or null",
  "stop_reason": "string or null",
  "truncated": false,
  "log_id": "string",
  "model": "string",
  "exception": {}
}
```

## Quick checklist

- Endpoint accepts JSON POST with the fields listed above.
- Supports SSE with `data: {json}\n` lines.
- Sends full accumulated text in `completion` each event.
- Ends with `data: [DONE]`.
- Configure the model in Admin with `api_type = custom` and correct auth env var.
