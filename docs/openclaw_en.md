# Using OpenClaw with Chat

This document explains how to configure OpenClaw as a model provider in the chat application.

## What is OpenClaw?

OpenClaw is an AI assistant platform that provides an **OpenAI-compatible API**. This means you can use OpenClaw models through the same interface as OpenAI models - no code changes needed!

## Getting Your OpenClaw API Key

### Option 1: From OpenClaw Gateway

1. OpenClaw runs a Gateway service (default: `http://localhost:7488`)
2. The API key is typically stored in your OpenClaw configuration
3. Check your `~/.openclaw/config.yaml` or environment variables

### Option 2: Environment Variable

OpenClaw often uses environment variables for API keys. Check for:
- `OPENAI_API_KEY` - If OpenClaw is configured to use OpenAI
- `ZHIPU_API_KEY` - For GLM models (zai/glm-5)
- Other provider-specific keys

### Option 3: Direct Configuration

If you're self-hosting OpenClaw, you can generate API keys through:
```bash
# Check OpenClaw status
openclaw status

# View configuration
cat ~/.openclaw/config.yaml
```

## Configuration

Since OpenClaw uses an OpenAI-compatible API, simply add it as an OpenAI model:

### Via Admin UI

Navigate to the model management page and add a new model:

| Field | Value |
|-------|-------|
| **Name** | The model name (e.g., `zai/glm-5`, `zai/gpt-4o`) |
| **Label** | Display name (e.g., "OpenClaw GLM-5") |
| **URL** | Your OpenClaw API endpoint |
| **API Type** | `openai` (uses OpenAI-compatible interface) |
| **API Auth Key** | Your API key or environment variable name |

### Example Configuration

```json
{
  "name": "zai/glm-5",
  "label": "OpenClaw GLM-5",
  "url": "http://localhost:7488/v1/chat/completions",
  "apiType": "openai",
  "apiAuthKey": "ZHIPU_API_KEY"
}
```

### Common OpenClaw Endpoints

| Endpoint | Description |
|----------|-------------|
| `http://localhost:7488/v1/chat/completions` | Local OpenClaw Gateway |
| `http://your-server:7488/v1/chat/completions` | Remote OpenClaw instance |

### Using Environment Variables

Set the API key via environment variable:

```bash
# For GLM models
export ZHIPU_API_KEY=your_zhipu_key

# Or use a generic name
export OPENCLAW_API_KEY=your_key
```

Then in the model configuration, set **API Auth Key** to the variable name (e.g., `ZHIPU_API_KEY`).

### Using Direct API Key

You can also enter the API key directly in the **API Auth Key** field without using environment variables.

## Available Models

OpenClaw supports various models depending on your configuration:

| Model Name | Description |
|------------|-------------|
| `zai/glm-5` | GLM-5 (Zhipu AI) |
| `zai/gpt-4o` | GPT-4o (via OpenClaw proxy) |
| `zai/claude-3-5-sonnet` | Claude 3.5 Sonnet (via OpenClaw proxy) |

Check your OpenClaw configuration for available models.

## Features

- **Streaming Support**: Full streaming response support for real-time chat
- **Reasoning Content**: Supports models that output reasoning/thinking content
- **File Uploads**: Supports text and multimedia file uploads (if model supports)
- **Rate Limiting**: Can be configured per-model

## Troubleshooting

### Connection Refused

```
Error: dial tcp 127.0.0.1:7488: connection refused
```

**Solution:**
1. Ensure OpenClaw Gateway is running: `openclaw gateway status`
2. Start if needed: `openclaw gateway start`
3. Check the URL in model configuration

### Authentication Errors

```
Error: 401 Unauthorized
```

**Solution:**
1. Verify your API key is correct
2. Check if the environment variable is set: `echo $ZHIPU_API_KEY`
3. Ensure the key has proper permissions

### Model Not Found

```
Error: model not found
```

**Solution:**
1. Check available models in OpenClaw configuration
2. Verify the model name matches exactly (including prefix like `zai/`)
3. Check OpenClaw logs: `openclaw gateway logs`

### Timeout Errors

**Solution:**
1. Increase `HttpTimeOut` in model configuration (default: 120 seconds)
2. Check network connectivity
3. Verify OpenClaw Gateway is responsive

## Architecture

```
Chat App → OpenAI-compatible API → OpenClaw Gateway → LLM Provider
                                    ↓
                              API Key Management
                              Request Routing
                              Response Streaming
```

The Chat application treats OpenClaw exactly like OpenAI since they share the same API format. OpenClaw Gateway handles:
- API key management for multiple providers
- Request routing to appropriate LLM backends
- Response streaming back to clients

## Related Documentation

- [Adding New Models Guide](./add_model_en.md)
- [Local Development Guide](./dev_locally_en.md)
- [Deployment Guide](./deployment_en.md)
