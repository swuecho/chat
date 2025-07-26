const default_chat_data: Chat.ChatState = {
  activeSession: {
    sessionUuid: null,
    workspaceUuid: null
  },
  workspaces: [], // Chat.Workspace[]
  history: [], // Chat.Session[]
  chat: {}, // { [key: string]: Chat.ChatMessage[] }

  // Legacy compatibility - kept for now while auto-migration handles the transition
  active: null, // uuid | null
  activeWorkspace: null, // workspace uuid | null
  workspaceActiveSessions: {}, // { [workspaceUuid: string]: string | null }
}

export function getLocalState(): Chat.ChatState {
  return default_chat_data
}

export function getChatKeys(chat: Chat.ChatState['chat'], includeLength = true) {
  const keys = Object.keys(chat)
  return includeLength ? [keys, keys.length] as const : [keys]
}
