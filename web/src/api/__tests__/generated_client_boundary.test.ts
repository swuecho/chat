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
})
