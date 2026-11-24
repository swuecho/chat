# Chat - Multi-LLM Chat Interface

A full-stack chat application that provides a unified interface for interacting with multiple large language models (LLMs): OpenAI-compatible APIs (OpenAI, DeepSeek), Claude, Gemini, Ollama, and custom endpoints.

This file guides coding agents (Claude Code, Codex, and others) working in this repository. `CLAUDE.md` imports it.

## Project Overview

This project is a ChatGPT wrapper that extends beyond OpenAI to support multiple LLM providers. It features a Vue.js frontend with a Go backend, using PostgreSQL for data persistence.

### Key Features

- **Multi-LLM Support**: OpenAI, Claude, Gemini, Ollama, and custom models
- **Chat Sessions**: Persistent conversation history and context management
- **Workspaces**: Organize chat sessions into customizable workspaces with colors and icons
- **User Management**: Authentication, rate limiting, and admin controls
- **File Uploads**: Support for text and multimedia files (model-dependent)
- **Snapshots**: Shareable conversation snapshots with full-text search
- **Artifacts**: Render, edit, organize, and download code, HTML, SVG, Mermaid, and other artifacts (never executed)
- **Prompt Management**: Built-in prompt templates with '/' shortcut
- **Internationalization**: Support for multiple languages

## Development Commands

### Backend (Go)
```bash
cd api

# Install dependencies and the Air hot-reload tool
make install

# Run server with hot reload (Air, configured in api/.air.toml)
make serve

# Format, vet, and build
make build

# Format code
make fmt

# Run tests
go test ./...

# Regenerate SQLC code (after modifying queries or schema)
sqlc generate
```

### Frontend (Vue.js)
```bash
cd web

npm install       # Install dependencies
npm run dev       # Development server with hot reload
npm run build     # Production build
npm run lint      # Lint
npm run lint:fix  # Fix lint issues
npm test          # Vitest unit tests

# Regenerate the OpenAPI spec and typed API client after changing handler contracts
npm run generate:api
```

### E2E Tests (Playwright)
```bash
cd e2e
export DATABASE_URL=postgres://...
npx playwright test        # Run all tests
npx playwright test --ui   # Run with UI
```

Both frontend (Rsbuild) and backend (Air) reload automatically on code changes.

## Architecture

### Request Flow
```
HTTP Request → Mux Router → Handler → Service → SQLC Queries → PostgreSQL
                                ↓
                         Provider (OpenAI/Claude/Gemini/Ollama/custom)
```

### Frontend (Vue.js)
- **Framework**: Vue 3 with Composition API
- **State Management**: Pinia
- **UI Library**: Naive UI
- **Routing**: Vue Router
- **Build Tool**: Rsbuild
- **Styling**: Tailwind CSS + Less
- **API Client**: Generated from the backend OpenAPI spec into `web/src/api/generated`

#### Chat Composables

The main chat interface (`web/src/views/chat/`) is split into composables under `web/src/views/chat/composables/`:
- `useStreamHandling.ts`: LLM streaming responses
- `useConversationFlow.ts`: Conversation lifecycle
- `useRegenerate.ts`: Message regeneration
- `useSearchAndPrompts.ts`: Search and prompt templates
- `useChatActions.ts`: Snapshots, bot creation, artifact gallery toggle
- `useChatFiles.ts`: File uploads
- `useErrorHandling.ts`: Centralized error management
- `useValidation.ts`: Input validation rules
- `usePerformanceOptimizations.ts`: Debouncing, memoization

### Backend (Go)
- **Framework**: Standard Go HTTP with Gorilla Mux (routes registered per handler, wired in `main.go`)
- **Database**: PostgreSQL with SQLC for type-safe queries
- **Authentication**: JWT tokens
- **Rate Limiting**: Built-in rate limiting (100 calls/10min default)
- **File Upload**: Support for various file types

#### Backend Boundaries

- `handler` owns HTTP concerns: route parameters, JSON request/response types, status codes, and mapping to application commands.
- `svc` owns application behavior. Service inputs use `Command` or `Query` types without JSON tags, and service outputs use application models rather than transport DTOs or generated SQLC records.
- `domain` contains shared business concepts and typed application errors. Keep it independent of HTTP and SQLC.
- `sqlc_queries` is generated persistence code. Simple query-wrapper services may call SQLC directly; do not add repository interfaces solely to hide straightforward SQLC calls.
- Multi-step writes that must succeed atomically are explicit use cases backed by narrow transaction/unit-of-work contracts. Current examples include `SaveSession`, `CreateWorkspaceSession`, and `CreateSessionFromSnapshot`.

#### Backend Design Rules

- Handlers decode handler-local request DTOs and map them to service commands. Do not decode JSON directly into service types.
- Handlers return typed response structs. Avoid constructing HTTP responses with `map[string]any` or `map[string]interface{}`.
- Services must not import `dto`, expose `sqlc_queries` types, or provide a general `Q()` escape hatch.
- Use application pagination/query objects instead of transport pagination constants or DTOs.
- Scope user-owned reads and mutations in the database query whenever possible. For example, load sessions with both UUID and user ID; conceal another user's resource as not found where appropriate.
- Preserve ownership checks inside transactional use cases before changing sessions, workspaces, prompts, messages, snapshots, or active-session state.
- Keep simple CRUD thin. Introduce a focused application use case when an operation coordinates multiple records, enforces business invariants, or requires a transaction.
- Use conventional Go initialisms in application types and fields: `ID`, `UUID`, `API`, and `LLM`.
- Add or update architecture tests when establishing a boundary so later changes cannot silently bypass it.

### Database (PostgreSQL + SQLC)

- **Schema**: `api/sqlc/schema.sql`
- **Queries**: `api/sqlc/queries/*.sql`
- **Generated Go**: `api/sqlc_queries/*.go` (hand-written helpers live in `zz_custom_*.go` and `transaction.go`)
- **Config**: `api/sqlc.yaml`

SQLC workflow:
1. Define schema in `api/sqlc/schema.sql`
2. Write SQL queries in `api/sqlc/queries/*.sql`
3. Run `sqlc generate` from `api/`
4. Use the generated code in services

Key tables:
- `auth_user`: User accounts (first registered user becomes admin)
- `chat_session`: Chat sessions
- `chat_message`: Messages within sessions
- `chat_workspace`: Workspace organization
- `chat_model`: Available LLM models
- `chat_prompt`: Prompt templates
- `chat_snapshot`: Shareable conversation snapshots
- `chat_file`: File uploads

## Project Structure

```
chat/
├── api/                     # Go backend
│   ├── main.go             # Application entry point
│   ├── llm/                # LLM wire-format clients (OpenAI uses the go-openai SDK)
│   │   ├── claude/         # Claude integration
│   │   └── gemini/         # Gemini integration
│   ├── apicontract/        # Typed route registration and OpenAPI contracts
│   ├── domain/             # Business concepts and typed errors
│   ├── dto/                # Transport DTOs and API error mapping
│   ├── handler/            # HTTP routes, request/response types, mapping
│   ├── provider/           # Provider implementations and streaming
│   ├── svc/                # Application services and transactional use cases
│   ├── pkg/util/           # Framework-independent utilities such as UUIDs
│   ├── openapi/            # Generated OpenAPI spec (openapi.json)
│   ├── sqlc/               # Database schema and source queries
│   ├── sqlc_queries/       # Generated type-safe queries
│   └── static/             # Static assets
├── web/                    # Vue.js frontend
│   ├── src/
│   │   ├── components/     # Reusable components
│   │   ├── views/          # Page components
│   │   ├── store/          # Pinia stores
│   │   ├── api/            # API client (generated + wrappers)
│   │   └── utils/          # Utility functions
│   └── dist/               # Built frontend assets
├── mobile/                 # Flutter mobile client
├── docs/                   # Documentation
└── e2e/                    # End-to-end tests (Playwright)
```

## Development Setup

### Prerequisites
- Go 1.24+
- Node.js 22.18+
- PostgreSQL
- (Optional) Docker

### Database Setup

1. Create a PostgreSQL database
2. Run the schema from `api/sqlc/schema.sql`
3. Configure the database connection in environment variables

Schema changes are applied through idempotent statements in `schema.sql` (for example `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`).

## Configuration

### Environment Variables

Set in the shell or `.env`:

```bash
# Database (required)
DATABASE_URL=postgres://user:pass@host:port/dbname?sslmode=disable

# LLM API keys (at least one required)
OPENAI_API_KEY=sk-...
CLAUDE_API_KEY=...
GEMINI_API_KEY=...
DEEPSEEK_API_KEY=...

# Optional
OPENAI_RATELIMIT=100  # Calls per 10 minutes (default: 100)
JWT_SECRET=...        # JWT signing secret
```

## API Endpoints

The authoritative route list is the generated spec at `api/openapi/openapi.json` (regenerate with `npm run generate:api` in `web/`). See `docs/typed-api-contracts.md` for how routes are declared.

Main route groups:
- **Auth**: `POST /api/signup`, `POST /api/login`, `POST /api/auth/refresh`
- **Sessions**: `/api/uuid/chat_sessions/...`, `/api/uuid/chat_session_from_snapshot/{uuid}`
- **Messages**: `/api/uuid/chat_messages/...`
- **Models**: `/api/chat_model`, `/api/chat_model/{id}`, `/api/chat_model/default`, `/api/chat_model/title-default`
- **Files**: `POST /api/upload`, `GET|DELETE /api/download/{id}`, `GET /api/chat_file/{uuid}/list`
- **Admin**: `/api/admin/...`

### Workspaces
- `GET /api/workspaces` - Get user workspaces
- `POST /api/workspaces` - Create new workspace
- `GET /api/workspaces/{uuid}` - Get workspace by UUID
- `PUT /api/workspaces/{uuid}` - Update workspace
- `DELETE /api/workspaces/{uuid}` - Delete workspace
- `PUT /api/workspaces/{uuid}/reorder` - Update workspace order
- `PUT /api/workspaces/{uuid}/set-default` - Set default workspace
- `POST /api/workspaces/{uuid}/sessions` - Create session in workspace
- `GET /api/workspaces/{uuid}/sessions` - Get sessions in workspace
- `POST /api/workspaces/default` - Ensure default workspace exists
- `GET|PUT /api/workspaces/{workspaceUuid}/active-session` - Get or set the workspace's active session
- `GET /api/workspaces/active-sessions` - Active sessions across workspaces

## Key Components

### Frontend Components
- `chat/index.vue`: Main chat interface
- `chat/components/Message/`: Message rendering
- `chat/components/Session/`: Session management
- `chat/components/ArtifactGallery.vue`: Artifact gallery
- `admin/index.vue`: Admin dashboard
- `prompt/creator.vue`: Prompt management

### Backend Handlers
- `chat_main.go` and `chat_stream.go`: Core chat and streaming
- `session.go`: Session HTTP endpoints
- `message.go`: Message endpoints
- `workspace*.go`: Workspace CRUD, settings, and session endpoints
- `auth_*.go`: Authentication and user management
- `admin.go`: Admin operations

### Backend Application Services

- `chat_session_service.go`: Session application model and focused operations
- `save_session.go`: Transactional normal session create/update flow
- `create_session_from_snapshot.go`: Transactional snapshot-copy flow
- `create_workspace_session.go`: Transactional workspace-session creation
- `transaction_manager.go`: SQLC transaction adapter with narrow unit-of-work interfaces
- `session_focused_services.go`: Focused history, rate-limit, snapshot, admin, and model queries

## Key Features & Patterns

### Authentication & Authorization
- JWT-based authentication
- First registered user becomes administrator (`is_superuser=true`)
- Rate limiting per user (default: 100 calls/10min, configurable via `OPENAI_RATELIMIT`)
- Per-model rate limiting can be enabled for specific models

### Chat Flow
1. The first message in a session is the system message (prompt)
2. User sends message → handler validates → service routes to the provider for the model's `api_type`
3. The provider streams the response → Server-Sent Events (SSE) → frontend renders incrementally
4. Messages are stored in PostgreSQL with full history

Session context: new sessions send the latest 10 messages (`dto.DefaultMaxLength`) plus the system prompt; users can change this per session.

Titles: generated by the model marked as the title model in the admin UI; otherwise the first 100 characters of the prompt are used.

### File Uploads
- Text files are supported for all models
- Multimedia files require model support (vision-capable OpenAI, Claude 3+, Gemini)
- Files are associated with messages via the `chat_file` table

### Snapshots
- Create shareable static pages from conversations (like ShareGPT)
- Full-text search (English) for organizing conversation history
- Conversations can be continued from snapshots

### Prompt Management
- Built-in prompt templates stored in the `chat_prompt` table
- Quick access via '/' shortcut in the chat interface

## Workspaces

The application supports organizing chat sessions into workspaces for better organization and context management.

### Workspace Features

- **Custom Organization**: Create themed workspaces for different projects or topics
- **Visual Customization**: Set custom colors and icons for easy identification
- **Session Management**: Sessions are automatically associated with workspaces
- **Default Workspace**: Each user has a default "General" workspace that's created automatically
- **Ordering**: Workspaces can be reordered for personal preference
- **Permission Control**: Users can only access their own workspaces

### Workspace Properties

- **Name**: Human-readable workspace name
- **Description**: Optional description for workspace purpose
- **Color**: Hex color code for visual theming (default: #6366f1)
- **Icon**: Icon identifier for visual representation (default: folder)
- **Default**: Boolean flag indicating if this is the user's default workspace
- **Order Position**: Integer for custom workspace ordering

### Database Schema

The `chat_workspace` table includes:
- `id`: Primary key
- `uuid`: Unique identifier for API operations
- `user_id`: Foreign key to owner user
- `name`: Workspace name (required)
- `description`: Optional description (default: empty string)
- `color`: Hex color code (default: #6366f1)
- `icon`: Icon identifier (default: folder)
- `is_default`: Boolean default flag
- `order_position`: Integer for ordering
- `created_at`/`updated_at`: Timestamps

Sessions are linked to workspaces via the `workspace_id` foreign key in the `chat_session` table.

## Model Support

The `api_type` column on `chat_model` selects the provider client:
- `openai`: OpenAI and OpenAI-compatible APIs (including DeepSeek)
- `claude`: Anthropic Claude
- `gemini`: Google Gemini
- `ollama`: Local models via Ollama
- `custom`: Custom HTTP endpoints (see `docs/custom_model_api_en.md`)

### Adding a New Model

See `docs/add_model_en.md` and `docs/add_model_zh.md`. In short:
1. Add the model to the `chat_model` table (admin UI or SQL)
2. Implement a provider in `api/provider/` (and a client in `api/llm/<provider>/` if needed) for a new provider type
3. Set the appropriate `api_type`

## Testing

```bash
# Backend
cd api && go test ./...

# Frontend
cd web && npm test

# E2E
cd e2e && npx playwright test
```

### SQLC Changes

After changing files under `api/sqlc/queries` or the schema, regenerate code and run the backend suite:

```bash
cd api
sqlc generate
go test ./...
```

## Common Gotchas

1. **SQLC code generation**: After modifying `schema.sql` or query files, run `sqlc generate` from `api/`.
2. **API contract generation**: After changing handler request/response types, run `npm run generate:api` in `web/` so the OpenAPI spec and typed client stay in sync (`npm run check:api-generated` verifies).
3. **Hot reload**: Both frontend (Rsbuild) and backend (Air) auto-reload on code changes.
4. **Rate limiting**: Applies globally (100/10min) unless per-model rate limiting is enabled.
5. **Model API types**: The `api_type` column determines which provider client is used.

## Documentation

- Local development: `docs/dev_locally_en.md`, `docs/dev_locally_zh.md`
- Deployment: `docs/deployment_en.md`, `docs/deployment_zh.md`
- Ollama integration: `docs/ollama_en.md`, `docs/ollama_zh.md`
- Custom model API: `docs/custom_model_api_en.md`
- Gateway: `docs/gateway_en.md`
- Artifact gallery: `docs/artifact_gallery_en.md`, `docs/artifact_gallery_zh.md`
- Snapshots vs ChatBots: `docs/snapshots_vs_chatbots_en.md`, `docs/snapshots_vs_chatbots_zh.md`
- Typed API contracts: `docs/typed-api-contracts.md`
- Adding models: `docs/add_model_en.md`, `docs/add_model_zh.md`
- Dev documentation: `docs/dev/` (error handling, HTTP request pipeline, request validation, SSE processing)
