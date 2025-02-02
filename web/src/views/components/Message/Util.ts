interface ThinkResult {
  thinkPart: string
  answerPart: string
}
function parseText(text: string): ThinkResult {
  let thinkContent = ''
  const answerContent = text.replace(/<think>(.*?)<\/think>/gs, (match, content) => {
    thinkContent = content.trim()
    return ''
  })
  return {
    thinkPart: thinkContent,
    answerPart: answerContent,
  }
}

export { parseText, ThinkResult }
