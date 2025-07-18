// Use the Chat namespace type
export type Artifact = Chat.Artifact

// Generate a simple UUID for frontend use
function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0
    const v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

// Check if a language is supported for code execution
function isExecutableLanguage(language: string): boolean {
  const executableLanguages = [
    'javascript', 'js', 'typescript', 'ts',
    'python', 'py'
  ]
  
  const normalizedLanguage = language.toLowerCase().trim()
  return executableLanguages.includes(normalizedLanguage)
}

// Check if code contains patterns that suggest it should be executable
function containsExecutablePatterns(content: string): boolean {
  // Patterns that suggest the code is meant to be executed
  const executablePatterns = [
    // JavaScript patterns
    'console.log',
    'console.error',
    'console.warn',
    'function',
    'const ',
    'let ',
    'var ',
    '=>',
    'if (',
    'for (',
    'while (',
    'return ',
    
    // Python patterns
    'print(',
    'import ',
    'from ',
    'def ',
    'if __name__',
    'class ',
    'for ',
    'while '
  ]
  
  const contentLower = content.toLowerCase()
  return executablePatterns.some(pattern => contentLower.includes(pattern))
}

// Extract artifacts from message content (mirrors backend logic)
export function extractArtifacts(content: string): Artifact[] {
  const artifacts: Artifact[] = []

  // Pattern for HTML artifacts (check specific types first)
  const htmlArtifactRegex = /```html\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n```/gs
  const htmlMatches = content.matchAll(htmlArtifactRegex)

  for (const match of htmlMatches) {
    const title = match[1].trim()
    const artifactContent = match[2].trim()

    const artifact: Artifact = {
      uuid: generateUUID(),
      type: 'html',
      title,
      content: artifactContent,
      language: 'html'
    }
    artifacts.push(artifact)
  }

  // Pattern for SVG artifacts
  const svgArtifactRegex = /```svg\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n```/gs
  const svgMatches = content.matchAll(svgArtifactRegex)

  for (const match of svgMatches) {
    const title = match[1].trim()
    const artifactContent = match[2].trim()

    const artifact: Artifact = {
      uuid: generateUUID(),
      type: 'svg',
      title,
      content: artifactContent,
      language: 'svg'
    }
    artifacts.push(artifact)
  }

  // Pattern for Mermaid diagrams
  const mermaidArtifactRegex = /```mermaid\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n```/gs
  const mermaidMatches = content.matchAll(mermaidArtifactRegex)

  for (const match of mermaidMatches) {
    const title = match[1].trim()
    const artifactContent = match[2].trim()

    const artifact: Artifact = {
      uuid: generateUUID(),
      type: 'mermaid',
      title,
      content: artifactContent,
      language: 'mermaid'
    }
    artifacts.push(artifact)
  }

  // Pattern for JSON artifacts
  const jsonArtifactRegex = /```json\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n```/gs
  const jsonMatches = content.matchAll(jsonArtifactRegex)

  for (const match of jsonMatches) {
    const title = match[1].trim()
    const artifactContent = match[2].trim()

    const artifact: Artifact = {
      uuid: generateUUID(),
      type: 'json',
      title,
      content: artifactContent,
      language: 'json'
    }
    artifacts.push(artifact)
  }

  // Pattern for executable code artifacts
  // Example: ```javascript <!-- executable: Calculator -->
  const executableArtifactRegex = /```(\w+)?\s*<!--\s*executable:\s*([^>]+?)\s*-->\s*\n(.*?)\n```/gs
  const executableMatches = content.matchAll(executableArtifactRegex)

  for (const match of executableMatches) {
    const language = match[1] || 'javascript'
    const title = match[2].trim()
    const artifactContent = match[3].trim()

    // Skip if already processed as HTML, SVG, Mermaid, or JSON
    if (language === 'html' || language === 'svg' || language === 'mermaid' || language === 'json') {
      continue
    }

    // Only create executable artifacts for supported languages
    if (isExecutableLanguage(language)) {
      const artifact: Artifact = {
        uuid: generateUUID(),
        type: 'executable-code',
        title,
        content: artifactContent,
        language
      }
      artifacts.push(artifact)
    }
  }

  // Pattern for general code artifacts (exclude html, svg, mermaid, json which are handled above)
  const codeArtifactRegex = /```(\w+)?\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n```/gs
  const codeMatches = content.matchAll(codeArtifactRegex)

  for (const match of codeMatches) {
    const language = match[1] || 'text'
    const title = match[2].trim()
    const artifactContent = match[3].trim()

    // Skip if already processed as HTML, SVG, Mermaid, JSON, or executable
    if (language === 'html' || language === 'svg' || language === 'mermaid' || language === 'json') {
      continue
    }

    // Check if this should be an executable artifact for supported languages
    let artifactType = 'code'
    if (isExecutableLanguage(language)) {
      // For supported languages, make them executable by default if they contain certain patterns
      if (containsExecutablePatterns(artifactContent)) {
        artifactType = 'executable-code'
      }
    }

    const artifact: Artifact = {
      uuid: generateUUID(),
      type: artifactType,
      title,
      content: artifactContent,
      language
    }
    artifacts.push(artifact)
  }

  return artifacts
}