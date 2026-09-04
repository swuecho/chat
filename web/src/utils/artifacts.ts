import { v7 as uuid } from 'uuid'
// Use the Chat namespace type

export type Artifact = Chat.Artifact

// Generate a simple UUID for frontend use
function generateUUID(): string {
  return uuid()
}

// Extract artifacts from message content (mirrors backend logic)
export function extractArtifacts(content: string, previousArtifacts: Artifact[] = []): Artifact[] {
  const artifacts: Artifact[] = []

  const addArtifact = (artifact: Omit<Artifact, 'uuid'>) => {
    const ordinal = artifacts.filter(item => item.type === artifact.type && item.title === artifact.title).length
    const previous = previousArtifacts.filter(item => item.type === artifact.type && item.title === artifact.title)[ordinal]
    artifacts.push({ ...artifact, uuid: previous?.uuid || generateUUID() })
  }

  // Pattern for HTML artifacts (check specific types first)
  const htmlArtifactRegex = /```html\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\r?\n(.*?)\r?\n```/gsi
  const htmlMatches = content.matchAll(htmlArtifactRegex)

  for (const match of htmlMatches) {
    const title = match[1].trim()
    const artifactContent = match[2].trim()

    addArtifact({
      type: 'html',
      title,
      content: artifactContent,
      language: 'html',
    })
  }

  // Pattern for SVG artifacts
  const svgArtifactRegex = /```svg\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\r?\n(.*?)\r?\n```/gsi
  const svgMatches = content.matchAll(svgArtifactRegex)

  for (const match of svgMatches) {
    const title = match[1].trim()
    const artifactContent = match[2].trim()

    addArtifact({
      type: 'svg',
      title,
      content: artifactContent,
      language: 'svg',
    })
  }

  // Pattern for Mermaid diagrams
  const mermaidArtifactRegex = /```mermaid\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\r?\n(.*?)\r?\n```/gsi
  const mermaidMatches = content.matchAll(mermaidArtifactRegex)

  for (const match of mermaidMatches) {
    const title = match[1].trim()
    const artifactContent = match[2].trim()

    addArtifact({
      type: 'mermaid',
      title,
      content: artifactContent,
      language: 'mermaid',
    })
  }

  // Pattern for JSON artifacts
  const jsonArtifactRegex = /```json\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\r?\n(.*?)\r?\n```/gsi
  const jsonMatches = content.matchAll(jsonArtifactRegex)

  for (const match of jsonMatches) {
    const title = match[1].trim()
    const artifactContent = match[2].trim()

    addArtifact({
      type: 'json',
      title,
      content: artifactContent,
      language: 'json',
    })
  }

  // Pattern for general code artifacts (exclude html, svg, mermaid, json which are handled above)
  const codeArtifactRegex = /```([\w+-]+)?\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\r?\n(.*?)\r?\n```/gsi
  const codeMatches = content.matchAll(codeArtifactRegex)

  for (const match of codeMatches) {
    const language = (match[1] || 'text').toLowerCase()
    const title = match[2].trim()
    const artifactContent = match[3].trim()

    // Skip formats handled by their specialized renderers.
    if (language === 'html' || language === 'svg' || language === 'mermaid' || language === 'json')
      continue

    addArtifact({
      type: 'code',
      title,
      content: artifactContent,
      language,
    })
  }

  return artifacts
}
