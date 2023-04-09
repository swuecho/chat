import { v4 as uuidv4 } from 'uuid'
import { ss } from '@/utils/storage'

import { createChatSession } from '@/api'
import { t } from '@/locales'

const LOCAL_NAME = 'chatStorage'

export function defaultState(): Chat.ChatState {
  const uuid = uuidv4()
  const new_chat_text = t('chat.new')
  createChatSession(uuid, new_chat_text)
  return { active: uuid, history: [{ uuid, title: new_chat_text, isEdit: false }], chat: [{ uuid, data: [] }] }
}

export function getLocalState(): Chat.ChatState {
  const localState = ss.get(LOCAL_NAME)
  return localState ?? defaultState()
}

export function setLocalState(state: Chat.ChatState) {
  ss.set(LOCAL_NAME, state)
}

