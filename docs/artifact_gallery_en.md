# Artifact Gallery

The Artifact Gallery collects structured content created in chat sessions. Artifacts are for viewing, copying, editing, organizing, and downloading; the application does not execute artifact code.

## Supported artifact types

- Code: static source with language-aware display
- HTML: sanitized, sandboxed static preview
- SVG: sanitized image preview
- Mermaid: diagram preview using Mermaid's strict security mode
- JSON: formatted structured-data preview
- Markdown: sanitized rendered preview

## Creating artifacts

Enable Artifacts in session settings, then ask the assistant for structured output. Artifact fences use a marker on the opening line:

````markdown
```javascript <!-- artifact: Descriptive title -->
export function example() {
  return 'Source only'
}
```
````

Only the `artifact` marker is supported. Executable markers are ignored.

## Gallery actions

The gallery supports search and filters by type, language, and session. Each artifact can be previewed, edited, duplicated, or deleted. The current result page can also be exported as JSON.

HTML previews do not allow scripts, inline event handlers, embedded frames, objects, or `javascript:` URLs. Code artifacts are always treated as text.
