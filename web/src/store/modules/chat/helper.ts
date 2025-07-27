const default_chat_data: Chat.ChatState = {
  activeSession: {
    sessionUuid: null,
    workspaceUuid: null
  },
  workspaceActiveSessions: {}, // { [workspaceUuid: string]: string }
  workspaces: [], // Chat.Workspace[]
  workspaceHistory: {}, // { [workspaceUuid: string]: Chat.Session[] }
  chat: {}, // { [key: string]: Chat.ChatMessage[] }
}

export function getLocalState(): Chat.ChatState {
  return default_chat_data
}

export function getChatKeys(chat: Chat.ChatState['chat'], includeLength = true) {
  const keys = Object.keys(chat)
  return includeLength ? [keys, keys.length] as const : [keys]
}
