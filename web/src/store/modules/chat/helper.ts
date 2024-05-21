const default_chat_data: Chat.ChatState = {
  active: null,
  history: [],
  chat: {},
}

export function getLocalState(): Chat.ChatState {
  return default_chat_data
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
