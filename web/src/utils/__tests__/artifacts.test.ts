import { describe, expect, it } from 'vitest'
import { extractArtifacts } from '../artifacts'

describe('extractArtifacts', () => {
  it('keeps artifact identity stable while a response streams', () => {
    const initial = extractArtifacts('```html <!-- artifact: Demo -->\n<p>one</p>\n```')
    const updated = extractArtifacts('```html <!-- artifact: Demo -->\n<p>two</p>\n```', initial)

    expect(updated).toHaveLength(1)
    expect(updated[0].uuid).toBe(initial[0].uuid)
    expect(updated[0].content).toBe('<p>two</p>')
  })

  it('ignores legacy executable markers', () => {
    const artifacts = extractArtifacts('```python <!-- executable: Report -->\nprint("hello")\n```')

    expect(artifacts).toEqual([])
  })

  it('accepts case-insensitive languages and CRLF fences', () => {
    const artifacts = extractArtifacts('```HTML <!-- artifact: Demo -->\r\n<p>Hello</p>\r\n```')

    expect(artifacts).toMatchObject([{ type: 'html', language: 'html', title: 'Demo' }])
  })
})
