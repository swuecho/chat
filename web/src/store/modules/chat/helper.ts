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

export function check_chat(chat: Chat.ChatState['chat'], need_length = true) {
  const keys = Object.keys(chat)
  const data: [Array<string>, number?] = [keys]
  if (need_length) {
    const keys_length = keys.length
    data.push(keys_length)
  }
  return data
}
