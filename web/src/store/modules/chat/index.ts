import { defineStore } from 'pinia'
import { getChatKeys, getLocalState } from './helper'
import { router } from '@/router'
import {
  clearSessionChatMessages,
  createChatSession,
  createOrUpdateUserActiveChatSession,
  deleteChatData,
  deleteChatSession,
  updateChatSession as fetchUpdateChatByUuid,
  getChatSessionDefault,
  getChatMessagesBySessionUUID,
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
    async reloadRoute(uuid?: string) {
      await router.push({ name: 'Chat', params: { uuid } })
    },

    async syncChatSessions() {
      this.history = await getChatSessionsByUser()

      if (this.history.length === 0) {
        const new_chat_text = t('chat.new')
        await this.addChatSession(await getChatSessionDefault(new_chat_text))
      }

      let active_session_uuid = this.history[0].uuid

      const active_session = await getUserActiveChatSession()
      if (active_session)
        active_session_uuid = active_session.chatSessionUuid

      this.active = active_session_uuid
      if (router.currentRoute.value.params.uuid !== this.active) {
        await this.reloadRoute(this.active)
      }
    },

    async syncChatMessages(need_uuid: string) {
      if (need_uuid) {
        const messageData = await getChatMessagesBySessionUUID(need_uuid)
        this.chat[need_uuid] = messageData
        await createOrUpdateUserActiveChatSession(need_uuid)
        this.setActiveLocal(need_uuid)
        //await this.reloadRoute(this.active) // !!! this cause cycle
      }
    },

    async addChatSession(history: Chat.Session, chatData: Chat.Message[] = []) {
      await createChatSession(history.uuid, history.title, history.model)
      this.history.unshift(history)
      this.chat[history.uuid] = chatData
      this.active = history.uuid
      this.reloadRoute(this.active)
    },

    async updateChatSession(uuid: string, edit: Partial<Chat.Session>) {
      const index = this.history.findIndex(item => item.uuid === uuid)
      if (index !== -1) {
        this.history[index] = { ...this.history[index], ...edit }
        await fetchUpdateChatByUuid(uuid, this.history[index])
      }
    },

    async updateChatSessionIfEdited(uuid: string, edit: Partial<Chat.Session>) {
      const index = this.history.findIndex(item => item.uuid === uuid)
      if (index !== -1) {
        if (this.history[index].isEdit) {
          this.history[index] = { ...this.history[index], ...edit }
          await fetchUpdateChatByUuid(uuid, this.history[index])
        }
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

    async setActiveLocal(uuid: string) {
      this.active = uuid
    },

    getChatByUuidAndIndex(uuid: string, index: number) {
      const [keys, keys_length] = getChatKeys(this.chat)
      if (!uuid) {
        if (keys_length)
          return this.chat[keys[0]][index]
        return null
      }
      if (keys.includes(uuid))
        return this.chat[uuid][index]
      return null
    },

    async addChatByUuid(uuid: string, chat: Chat.Message) {
      const new_chat_text = t('chat.new')
      const [keys] = getChatKeys(this.chat, false)
      if (!uuid) {
        if (this.history.length === 0) {
          const default_model_parameters = await getChatSessionDefault(new_chat_text)
          const uuid = default_model_parameters.uuid;
          await createChatSession(uuid, chat.text, default_model_parameters.model)
          this.history.push({ uuid, title: chat.text, isEdit: false })
          this.chat[uuid] = [{ ...chat, isPrompt: true, isPin: false }]
          this.active = uuid
        }
        else {
          this.chat[keys[0]].push(chat)
          if (this.history[0].title === new_chat_text) {
            this.history[0].title = chat.text
            renameChatSession(this.history[0].uuid, chat.text.substring(0, 40))
          }
        }
      }

      if (keys.includes(uuid)) {
        if (this.chat[uuid].length === 0)
          this.chat[uuid].push({ ...chat, isPrompt: true, isPin: false })
        else
          this.chat[uuid].push(chat)

        if (this.history[0].title === new_chat_text) {
          this.history[0].title = chat.text
          renameChatSession(this.history[0].uuid, chat.text.substring(0, 40))
        }
      }
    },

    async updateChatByUuid(uuid: string, index: number, chat: Chat.Message) {
      const [keys, keys_length] = getChatKeys(this.chat)
      if (!uuid) {
        if (keys_length) {
          this.chat[keys[0]][index] = chat
        }
        return
      }

      if (keys.includes(uuid)) {
        this.chat[uuid][index] = chat
      }
    },

    updateChatPartialByUuid(
      uuid: string,
      index: number,
      chat: Partial<Chat.Message>,
    ) {
      const [keys, keys_length] = getChatKeys(this.chat)
      if (!uuid) {
        if (keys_length) {
          this.chat[keys[0]][index] = { ...this.chat[keys[0]][index], ...chat }
        }
        return
      }

      if (keys.includes(uuid)) {
        this.chat[uuid][index] = {
          ...this.chat[uuid][index],
          ...chat,
        }
      }
    },

    async deleteChatByUuid(uuid: string, index: number) {
      const [keys, keys_length] = getChatKeys(this.chat)
      if (!uuid) {
        if (keys_length) {
          const chatData = this.chat[keys[0]]
          const chat = chatData[index]
          chatData.splice(index, 1)
          if (chat)
            await deleteChatData(chat)
        }
        return
      }

      if (keys.includes(uuid)) {
        const chatData = this.chat[uuid]
        const chat = chatData[index]
        chatData.splice(index, 1)
        if (chat)
          await deleteChatData(chat)
      }
    },

    clearChatByUuid(uuid: string) {
      const [keys, keys_length] = getChatKeys(this.chat)
      if (!uuid) {
        if (keys_length) {
          this.chat[keys[0]] = []
        }
        return
      }
      if (keys.includes(uuid)) {
        const data: Chat.Message[] = []
        for (const chat of this.chat[uuid]) {
          if (chat.isPin || chat.isPrompt)
            data.push(chat)
        }
        this.chat[uuid] = data
        clearSessionChatMessages(uuid)
      }
    },
    clearState() {
      this.history = []
      this.chat = {}
      this.active = null
    },
  },
})
