import { ss } from '@/utils/storage'

const LOCAL_NAME = 'chatStorage'

export function getLocalState(): Chat.ChatState {
  const localState = ss.get(LOCAL_NAME)
  return localState ?? {
    active: null,
    history: [],
    chat: [],
  }
}

export function setLocalState(state: Chat.ChatState) {
  ss.set(LOCAL_NAME, state)
}
