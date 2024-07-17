export function getDataFromResponseText(responseText: string): string {
        // first data segment
        if (responseText.lastIndexOf('data:') === 0)
                return responseText.slice(5)
        // Find the last occurrence of the data segment
        const lastIndex = responseText.lastIndexOf('\n\ndata:')
        // Extract the JSON data chunk from the responseText
        const chunk = responseText.slice(lastIndex + 8)
        return chunk
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