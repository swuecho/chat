# Artifacts

Artifacts are structured blocks created in chat for viewing, editing, copying, downloading, and organizing. They are optional per session and appear inline with messages and in the Artifact Gallery.

## Safety boundary

The application never executes artifact code.

- Code is displayed as static, syntax-highlighted text.
- HTML is sanitized and shown in a sandboxed preview. Scripts, inline event handlers, embedded frames, objects, and unsafe URLs are removed.
- SVG is sanitized before rendering.
- Mermaid diagrams use Mermaid's strict security mode.
- Legacy `executable` markers are ignored.

## Supported formats

- `html`: sanitized static preview
- `svg`: sanitized graphic preview
- `mermaid`: rendered diagram
- `json`: formatted structured data
- Any other fenced language: static code artifact

## Creating an artifact

Enable Artifacts in chat settings and use an artifact marker on the opening fence:

````markdown
```javascript <!-- artifact: Descriptive title -->
export function greeting(name) {
  return `Hello, ${name}`
}
```
````

The marker must use `<!-- artifact: Title -->` on the same line as the opening fence. Blocks without that marker remain ordinary message content.

## Persistence and gallery

Artifacts are extracted from completed assistant messages and stored with their originating message and session. The gallery supports:

- Search and filtering by type, language, or session
- Preview, editing, duplication, and deletion
- Copying and downloading individual artifacts
- Exporting the current result page as JSON

Artifact access is scoped to the authenticated owner. Titles are limited to 200 bytes and content to 1 MB.

## Main implementation areas

- Extraction and persistence: `api/svc/chat_artifact.go`
- Artifact operations: `api/svc/artifact_service.go`
- HTTP endpoints: `api/handler/artifact.go`
- Client-side streaming extraction: `web/src/utils/artifacts.ts`
- Rendering: `web/src/views/chat/components/Message/ArtifactContent.vue`
- Gallery: `web/src/views/chat/components/ArtifactGallery.vue`

The authenticated API provides listing, updating, deleting, and duplication under `/artifacts`.
