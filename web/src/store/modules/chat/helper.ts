import { v4 as uuidv4 } from 'uuid'
import { ss } from '@/utils/storage'

import { createChatSession } from '@/api'

const LOCAL_NAME = 'chatStorage'

export function defaultState(): Chat.ChatState {
  const uuid = uuidv4()
  createChatSession(uuid, 'New Chat')
  return { active: uuid, history: [{ uuid, title: 'New Chat', isEdit: false }], chat: [{ uuid, data: [] }] }
}

export function getLocalState(): Chat.ChatState {
  const localState = ss.get(LOCAL_NAME)
  return localState ?? defaultState()
}

export function setLocalState(state: Chat.ChatState) {
  ss.set(LOCAL_NAME, state)
}
