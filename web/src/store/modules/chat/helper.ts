const default_chat_data: Chat.ChatState = {
  active: null, // uuid | null
  history: [], // Chat.Session[]
  chat: {}, // { [key: string]: Chat.ChatMessage[] }
}

export function getLocalState(): Chat.ChatState {
  return default_chat_data
}

export function getChatKeys(chat: Chat.ChatState['chat'], includeLength = true) {
  const keys = Object.keys(chat)
  return includeLength ? [keys, keys.length] as const : [keys]
}
