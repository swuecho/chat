export function extractStreamingData(streamResponse: string): string {
  const DATA_MARKER = 'data:'
  const SSE_DATA_MARKER = '\n\ndata:'
  
  // Handle single data segment at response start (most common after buffer split)
  if (streamResponse.startsWith(DATA_MARKER)) {
    return streamResponse.slice(DATA_MARKER.length).trim()
  }

  // Handle Server-Sent Events with multiple data segments - extract the last one
  const lastSSEDataPosition = streamResponse.lastIndexOf(SSE_DATA_MARKER)
  if (lastSSEDataPosition === -1) {
    return streamResponse.trim() // No SSE format detected, return original
  }

  // Extract data after the last SSE marker
  const dataStartPosition = lastSSEDataPosition + SSE_DATA_MARKER.length
  return streamResponse.slice(dataStartPosition).trim()
}

export function escapeDollarNumber(text: string) {
        let escapedText = ''
        for (let i = 0; i < text.length; i += 1) {
          let char = text[i]
          const nextChar = text[i + 1] || ' '
          if (char === '$' && nextChar >= '0' && nextChar <= '9')
            char = '\\$'
          escapedText += char
        }
        return escapedText
      }
export function escapeBrackets(text: string) {
        const pattern = /(```[\s\S]*?```|`.*?`)|\\\[([\s\S]*?[^\\])\\\]|\\\((.*?)\\\)/g
        return text.replace(pattern, (match, codeBlock, squareBracket, roundBracket) => {
          if (codeBlock)
            return codeBlock
          else if (squareBracket)
            return `$$${squareBracket}$$`
          else if (roundBracket)
            return `$${roundBracket}$`
          return match
        })
      }