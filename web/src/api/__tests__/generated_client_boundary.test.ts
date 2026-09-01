import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'
import { describe, expect, it } from 'vitest'

const sourceRoot = join(process.cwd(), 'src')

function sourceFiles(directory: string): string[] {
  return readdirSync(directory).flatMap((name) => {
    const path = join(directory, name)
    if (path.includes(`${join('api', 'generated')}`))
      return []
    return statSync(path).isDirectory() ? sourceFiles(path) : /\.(?:ts|vue)$/.test(path) ? [path] : []
  })
}

describe('generated API client boundary', () => {
  it('prevents handwritten backend transports from returning', () => {
    const violations: string[] = []
    for (const path of sourceFiles(sourceRoot)) {
      const source = readFileSync(path, 'utf8')
      const file = relative(sourceRoot, path)
      if (/from\s+['"]axios['"]|from\s+['"]@\/utils\/request/.test(source))
        violations.push(`${file}: imports the retired Axios transport`)
      if (/\bfetch\s*\(\s*['"`]\/api\//.test(source))
        violations.push(`${file}: calls a backend API with raw fetch`)
      if (/<NUpload[^>]+(?::action|action)=/.test(source))
        violations.push(`${file}: delegates upload to an implicit XMLHttpRequest`)
    }
    expect(violations).toEqual([])
  })

  it('keeps UI calls on the configured generated client or a behavior adapter', () => {
    const behaviorAdapters = new Set([
      'api_keys', // shared view-only types
      'chat_file', // upload/download and UI file metadata
      'chat_model', // default-model fallback
      'chat_session', // UI model mapping and new-session defaults
      'chat_stream', // ReadableStream parsing
      'chat_workspace', // compatibility fallback for old servers
      'content', // prompt/message dispatch
    ])
    const violations: string[] = []

    for (const path of sourceFiles(sourceRoot)) {
      const file = relative(sourceRoot, path)
      if (file.split(/[\\/]/)[0] === 'api')
        continue

      const source = readFileSync(path, 'utf8')
      for (const match of source.matchAll(/['"]@\/api\/([^'"]+)['"]/g)) {
        const module = match[1]
        if (module === 'generated_client' || behaviorAdapters.has(module))
          continue
        // Auth refresh deliberately uses an isolated raw client to avoid a
        // dependency cycle through the authenticated singleton.
        if (file === join('store', 'modules', 'auth', 'index.ts') && module.startsWith('generated/'))
          continue
        violations.push(`${file}: imports non-adapter API module ${module}`)
      }
    }

    expect(violations).toEqual([])
  })
})
