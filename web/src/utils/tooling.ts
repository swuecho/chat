export type ToolCall = {
  name: string
  arguments: Record<string, unknown>
}

const toolCallRegex = /```tool_call\s*([\s\S]*?)```/gi
const toolResultRegex = /```tool_result\s*([\s\S]*?)```/gi

export const extractToolCalls = (text: string) => {
  const calls: ToolCall[] = []
  let cleanedText = text

  cleanedText = cleanedText.replace(toolCallRegex, (_, jsonPayload) => {
    try {
      const parsed = JSON.parse(jsonPayload.trim())
      if (parsed && typeof parsed === 'object' && parsed.name) {
        calls.push(parsed as ToolCall)
      }
    } catch {
      // Ignore malformed tool calls.
    }
    return ''
  })

  return {
    calls,
    cleanedText: cleanedText.trim(),
  }
}

export const stripToolBlocks = (text: string) => {
  return text.replace(toolCallRegex, '').replace(toolResultRegex, '').trim()
}

export const isToolResultMessage = (text: string) => {
  const trimmed = text.trim()
  return trimmed.startsWith('[[TOOL_RESULT]]') || toolResultRegex.test(trimmed)
}
