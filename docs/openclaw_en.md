# OpenClaw Integration Guide

This guide explains how to integrate [OpenClaw](https://github.com/openclaw/openclaw) with the chat application to enable communication with OpenClaw agents through the web chat interface.

## Overview

Unlike traditional LLM integrations (OpenAI, Claude, etc.) where the chat app sends messages to an API and receives AI-generated responses, OpenClaw integration enables **bi-directional communication** with OpenClaw agents:

- **Traditional LLM**: User → Chat App → LLM API → AI Response
- **OpenClaw Agent**: User → Chat App → OpenClaw Agent → Agent Response (like Telegram/Discord)

This allows OpenClaw agents to:
- Use tools and skills
- Access memory and context
- Spawn sub-agents
- Execute long-running tasks
- Communicate through the familiar chat interface

## Architecture

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│  Web User   │◄────►│  Chat App   │◄────►│   OpenClaw  │
│   (Browser) │      │   (API)     │      │   Gateway   │
└─────────────┘      └─────────────┘      └──────┬──────┘
                                                  │
                                            ┌─────┴─────┐
                                            │  OpenClaw │
                                            │   Agent   │
                                            └───────────┘
```

The chat app communicates with OpenClaw using its **OpenAI-compatible API** (`/v1/chat/completions`), which allows:
- Streaming responses (real-time typing effect)
- Session continuity (chat session UUID is passed to OpenClaw)
- Full compatibility with existing chat infrastructure

## Prerequisites

- OpenClaw gateway running (default: `http://localhost:8080`)
- OpenClaw configured with at least one agent
- Admin access to the chat application

## Configuration

### 1. Start OpenClaw Gateway

```bash
openclaw gateway start
```

Verify it's running:
```bash
openclaw gateway status
```

### 2. Configure OpenClaw Agent

Ensure you have an agent configured. Check available agents:
```bash
openclaw agents list
```

### 3. Add OpenClaw as a Model in Chat App

Log in as admin and navigate to **Admin > Models > Add Model**.

**Configuration:**

```json
{
  "name": "openclaw-agent",
  "label": "OpenClaw Agent",
  "url": "http://localhost:8080/v1/chat/completions",
  "apiAuthHeader": "Authorization",
  "apiAuthKey": "OPENCLAW_API_KEY",
  "isDefault": false,
  "enablePerModeRatelimit": false,
  "isEnable": true,
  "orderNumber": 10,
  "defaultToken": 4096,
  "maxToken": 8192
}
```

**Important fields:**

| Field | Value | Description |
|-------|-------|-------------|
| `name` | Must start with `openclaw-` | Triggers `api_type='openclaw'` routing |
| `url` | Gateway URL + `/v1/chat/completions` | OpenClaw's OpenAI-compatible endpoint |
| `apiAuthKey` | Environment variable name | Contains the OpenClaw API key |

### 4. Set Environment Variables

```bash
# Required
export OPENCLAW_GATEWAY_URL="http://localhost:8080"

# Optional - if OpenClaw requires authentication
export OPENCLAW_API_KEY="your-api-key"
```

Or in your `.env` file:
```
OPENCLAW_GATEWAY_URL=http://localhost:8080
OPENCLAW_API_KEY=your-api-key
```

### 5. Restart the API Server

```bash
# Docker
docker-compose restart api

# Local
go run ./api
```

## Usage

1. Create a new chat session
2. Select "OpenClaw Agent" from the model dropdown
3. Start chatting - messages go to the OpenClaw agent
4. Responses stream back in real-time

## Session Continuity

The chat session UUID is passed to OpenClaw in the `session` field of each request. This enables:

- **Memory persistence**: OpenClaw agents can remember previous conversations
- **Context awareness**: Agents maintain context across messages
- **Long-running tasks**: Agents can continue work between messages

## Advanced Configuration

### Using a Specific Agent

To use a specific OpenClaw agent, configure the model name to match:

```json
{
  "name": "openclaw-coder",
  "label": "OpenClaw Coder Agent",
  "url": "http://localhost:8080/v1/chat/completions",
  "apiAuthHeader": "Authorization",
  "apiAuthKey": "OPENCLAW_API_KEY",
  "isDefault": false,
  "isEnable": true,
  "orderNumber": 11
}
```

Then configure OpenClaw to route `openclaw-coder` model requests to your specific agent.

### Remote OpenClaw Instance

```json
{
  "name": "openclaw-remote",
  "label": "OpenClaw (Remote)",
  "url": "https://openclaw.your-domain.com/v1/chat/completions",
  "apiAuthHeader": "Authorization",
  "apiAuthKey": "OPENCLAW_REMOTE_API_KEY",
  "isDefault": false,
  "isEnable": true,
  "orderNumber": 12
}
```

## Troubleshooting

### Connection Refused

```
Error: Failed to send OpenClaw request
```

**Solutions:**
1. Check OpenClaw is running: `openclaw gateway status`
2. Verify `OPENCLAW_GATEWAY_URL` is correct
3. Ensure no firewall blocks the connection

### Authentication Errors (401/403)

**Solutions:**
1. Verify `OPENCLAW_API_KEY` is set
2. Check API key in OpenClaw configuration
3. Ensure `apiAuthHeader` is `Authorization`

### Model Not Appearing

**Solutions:**
1. Check `isEnable` is `true`
2. Verify name starts with `openclaw-`
3. Restart API server after adding model

### No Streaming Response

**Solutions:**
1. Check browser console for errors
2. Verify OpenClaw supports SSE (Server-Sent Events)
3. Check API server logs

## Comparison with Other Integrations

| Feature | OpenAI/Claude | OpenClaw Agent |
|---------|--------------|----------------|
| Response type | AI-generated | Agent-processed |
| Tools access | Limited | Full skill access |
| Memory | Per-conversation | Persistent |
| Sub-agents | No | Yes |
| Long-running tasks | No | Yes |
| Custom logic | No | Yes |

## See Also

- [OpenClaw Documentation](https://docs.openclaw.ai)
- [OpenClaw GitHub](https://github.com/openclaw/openclaw)
- [OpenClaw Session API](https://docs.openclaw.ai/sessions)
