import { defineStore } from 'pinia'
import { v4 as uuidv4 } from 'uuid'

import { check_chat, getLocalState, setLocalState } from './helper'
import { router } from '@/router'
import {
  clearSessionChatMessages,
  createChatSession,
  createOrUpdateUserActiveChatSession,
  deleteChatData,
  deleteChatSession,
  updateChatSession as fetchUpdateChatByUuid,
  getChatSessionDefault,
  getChatMessagesBySessionUUID as getChatSessionHistory,
  getChatSessionsByUser,
  getUserActiveChatSession,
  renameChatSession,
} from '@/api'

import { t } from '@/locales'

export const useChatStore = defineStore('chat-store', {
  state: (): Chat.ChatState => getLocalState(),

  getters: {
    getChatSessionByCurrentActive(state: Chat.ChatState) {
      const index = state.history.findIndex(
        item => item.uuid === state.active,
      )
      if (index !== -1)
        return state.history[index]
      return null
    },

    getChatSessionByUuid(state: Chat.ChatState) {
      return (uuid?: string) => {
        if (uuid)
          return state.history.find(item => item.uuid === uuid)
        return (
          state.history.find(item => item.uuid === state.active)
        )
      }
    },

    getChatSessionDataByUuid(state: Chat.ChatState) {
      return (uuid?: string) => {
        if (uuid)
          return state.chat[uuid] ?? []
        if (state.active)
          return state.chat[state.active] ?? []
        return []
      }
    },
  },

  actions: {
    recordState() {
      setLocalState(this.$state)
    },

    async reloadRoute(uuid?: string) {
      this.recordState()
      await router.push({ name: 'Chat', params: { uuid } })
    },

    async syncChatSessions() {
      const sessions = await getChatSessionsByUser()
      this.history = []
      await sessions.forEach(async (r: Chat.Session) => {
        this.history.unshift(r)
      })
      if (this.history.length === 0) {
        const new_chat_text = t('chat.new')
        this.addChatSession(await getChatSessionDefault(new_chat_text))
      }

      let active_session_uuid = this.history[0].uuid

      const active_session = await getUserActiveChatSession()
      if (active_session)
        active_session_uuid = active_session.chatSessionUuid

      this.active = active_session_uuid
      this.reloadRoute(this.active)
    },

    async syncChatMessages(need_uuid: string) {
      if (need_uuid) {
        const messageData = await getChatSessionHistory(need_uuid)
        this.chat[need_uuid] = messageData
        this.reloadRoute(need_uuid)
      }
    },

    addChatSession(history: Chat.Session, chatData: Chat.Message[] = []) {
      createChatSession(history.uuid, history.title, history.model)
      this.history.unshift(history)
      this.chat[history.uuid] = chatData
      this.active = history.uuid
      this.reloadRoute(history.uuid)
    },

    async updateChatSession(uuid: string, edit: Partial<Chat.Session>) {
      const index = this.history.findIndex(item => item.uuid === uuid)
      if (index !== -1) {
        this.history[index] = { ...this.history[index], ...edit }
        // update chat session
        await fetchUpdateChatByUuid(uuid, this.history[index])
        this.recordState()
      }
    },

    deleteChatSession(index: number) {
      deleteChatSession(this.history[index].uuid)
      delete this.chat[this.history[index].uuid]
      this.history.splice(index, 1)

      if (this.history.length === 0) {
        this.active = null
        this.reloadRoute()
        return
      }

      if (index > 0 && index <= this.history.length) {
        const uuid = this.history[index - 1].uuid
        this.setActive(uuid)
        return
      }

      if (index === 0) {
        if (this.history.length > 0) {
          const uuid = this.history[0].uuid
          this.setActive(uuid)
        }
      }

      if (index > this.history.length) {
        const uuid = this.history[this.history.length - 1].uuid
        this.setActive(uuid)
      }
    },

    async setActive(uuid: string) {
      this.active = uuid
      await createOrUpdateUserActiveChatSession(uuid)
      await this.reloadRoute(uuid)
    },

    getChatByUuidAndIndex(uuid: string, index: number) {
      const [keys, keys_length] = check_chat(this.chat)
      if (!uuid) {
        if (keys_length)
          return this.chat[uuid][index]
        return null
      }
      // const chatIndex = this.chat.findIndex(item => item.uuid === uuid)
      if (keys.includes(uuid))
        return this.chat[uuid][index]
      return null
    },

    async addChatByUuid(uuid: string, chat: Chat.Message) {
      const new_chat_text = t('chat.new')
      const [keys] = check_chat(this.chat, false)
      if (!uuid) {
        if (this.history.length === 0) {
          const uuid = uuidv4()
          const default_model_parameters = await getChatSessionDefault(new_chat_text)

          createChatSession(uuid, chat.text, default_model_parameters.model)
          this.history.push({ uuid, title: chat.text, isEdit: false })
          // first chat message is prompt
          // this.chat.push({ uuid, data: [{ ...chat, isPrompt: true, isPin: false }] })
          this.chat[uuid] = [{ ...chat, isPrompt: true, isPin: false }]
          this.active = uuid
          this.recordState()
        }
        else {
          // this.chat[0].data.push(chat)
          this.chat[keys[0]].push(chat)
          if (this.history[0].title === new_chat_text) {
            this.history[0].title = chat.text
            renameChatSession(this.history[0].uuid, chat.text.substring(0, 20))
          }
          this.recordState()
        }
      }

      // const index = this.chat.findIndex(item => item.uuid === uuid)
      if (keys.includes(uuid)) {
        if (this.chat[uuid].length === 0)
          this.chat[uuid].push({ ...chat, isPrompt: true, isPin: false })
        else
          this.chat[uuid].push(chat)

        if (this.history[0].title === new_chat_text) {
          this.history[0].title = chat.text
          renameChatSession(this.history[0].uuid, chat.text.substring(0, 20))
        }
        this.recordState()
      }
    },

    async updateChatByUuid(uuid: string, index: number, chat: Chat.Message) {
      // TODO: sync with server
      const [keys, keys_length] = check_chat(this.chat)
      if (!uuid) {
        if (keys_length) {
          this.chat[keys[0]][index] = chat
          this.recordState()
        }
        return
      }

      // const chatIndex = this.chat.findIndex(item => item.uuid === uuid)
      if (keys.includes(uuid)) {
        this.chat[uuid][index] = chat
        this.recordState()
      }
    },

    updateChatPartialByUuid(
      uuid: string,
      index: number,
      chat: Partial<Chat.Message>,
    ) {
      const [keys, keys_length] = check_chat(this.chat)
      if (!uuid) {
        if (keys_length) {
          this.chat[keys[0]][index] = { ...this.chat[keys[0]][index], ...chat }
          this.recordState()
        }
        return
      }

      // const chatIndex = this.chat.findIndex(item => item.uuid === uuid)
      if (keys.includes(uuid)) {
        this.chat[uuid][index] = {
          ...this.chat[uuid][index],
          ...chat,
        }
        this.recordState()
      }
    },

    async deleteChatByUuid(uuid: string, index: number) {
      const [keys, keys_length] = check_chat(this.chat)
      if (!uuid) {
        if (keys_length) {
          const chatData = this.chat[keys[0]]
          const chat = chatData[index]
          chatData.splice(index, 1)
          this.recordState()
          if (chat)
            await deleteChatData(chat)
        }
        return
      }

      // const chatIndex = this.chat.findIndex(item => item.uuid === uuid)
      if (keys.includes(uuid)) {
        const chatData = this.chat[uuid]
        const chat = chatData[index]
        chatData.splice(index, 1)
        this.recordState()
        if (chat)
          await deleteChatData(chat)
      }
    },

    clearChatByUuid(uuid: string) {
      // does this every happen?
      const [keys, keys_length] = check_chat(this.chat)
      if (!uuid) {
        if (keys_length) {
          this.chat[keys[0]] = []
          this.recordState()
        }
        return
      }

      // const index = this.chat.findIndex(item => item.uuid === uuid)
      if (keys.includes(uuid)) {
        const data: Chat.Message[] = []
        for (const chat of this.chat[uuid]) {
          if (chat.isPin || chat.isPrompt)
            data.push(chat)
        }
        this.chat[uuid] = data
        clearSessionChatMessages(uuid)
        this.recordState()
      }
    },
    clearState() {
      this.history = []
      this.chat = {}
      this.active = null
      this.recordState()
    },
  },
})
