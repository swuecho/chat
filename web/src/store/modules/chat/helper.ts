import { ss } from '@/utils/storage'

const LOCAL_NAME = 'chatStorage'

const default_chat_data: Chat.ChatState = {
  active: null,
  history: [],
  chat: {},
}

export function getLocalState(): Chat.ChatState {
  const localState = ss.get(LOCAL_NAME)
  return localState ?? default_chat_data
}

export function setLocalState(state: Chat.ChatState) {
  ss.set(LOCAL_NAME, state)
}
