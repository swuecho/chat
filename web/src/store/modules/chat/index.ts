import { defineStore } from 'pinia'
import { v4 as uuidv4 } from 'uuid'

import { getLocalState, setLocalState } from './helper'
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
      const index = state.sessions.findIndex(
        item => item.uuid === state.active,
      )
      if (index !== -1)
        return state.sessions[index]
      return null
    },

    getChatSessionByUuid(state: Chat.ChatState) {
      return (uuid?: string) => {
        if (uuid)
          return state.sessions.find(item => item.uuid === uuid)
        return (
          state.sessions.find(item => item.uuid === state.active)
        )
      }
    },

    getChatSessionDataByUuid(state: Chat.ChatState) {
      return (uuid?: string) => {
        if (uuid)
          return state.chat.find(item => item.uuid === uuid)?.data ?? []
        return (
          state.chat.find(item => item.uuid === state.active)?.data ?? []
        )
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
      this.sessions = []
      this.chat = []
      await sessions.forEach(async (r: Chat.Session) => {
        this.sessions.unshift(r)
        const chatData = await getChatSessionHistory(r.uuid)
        // chatData.uuid = chatData?.uuid
        this.chat.unshift({ uuid: r.uuid, data: chatData })
      })

      if (this.sessions.length === 0) {
        const new_chat_text = t('chat.new')
        this.addChatSession(await getChatSessionDefault(new_chat_text))
      }

      let active_session_uuid = this.sessions[0].uuid
      const active_session = await getUserActiveChatSession()

      if (active_session)
        active_session_uuid = active_session.ChatSessionUuid

      this.active = active_session_uuid
      this.reloadRoute(this.active)
    },

    addChatSession(history: Chat.Session, chatData: Chat.Message[] = []) {
      createChatSession(history.uuid, history.title)
      this.sessions.unshift(history)
      this.chat.unshift({ uuid: history.uuid, data: chatData })
      this.active = history.uuid
      this.reloadRoute(history.uuid)
    },

    async updateChatSession(uuid: string, edit: Partial<Chat.Session>) {
      const index = this.sessions.findIndex(item => item.uuid === uuid)
      if (index !== -1) {
        this.sessions[index] = { ...this.sessions[index], ...edit }
        // update chat session
        await fetchUpdateChatByUuid(uuid, this.sessions[index])
        this.recordState()
      }
    },

    deleteChatSession(index: number) {
      deleteChatSession(this.sessions[index].uuid)
      this.sessions.splice(index, 1)
      this.chat.splice(index, 1)

      if (this.sessions.length === 0) {
        this.active = null
        this.reloadRoute()
        return
      }

      if (index > 0 && index <= this.sessions.length) {
        const uuid = this.sessions[index - 1].uuid
        this.active = uuid
        this.reloadRoute(uuid)
        return
      }

      if (index === 0) {
        if (this.sessions.length > 0) {
          const uuid = this.sessions[0].uuid
          this.active = uuid
          this.reloadRoute(uuid)
        }
      }

      if (index > this.sessions.length) {
        const uuid = this.sessions[this.sessions.length - 1].uuid
        this.active = uuid
        this.reloadRoute(uuid)
      }
    },

    async setActive(uuid: string) {
      this.active = uuid
      await createOrUpdateUserActiveChatSession(uuid)
      return await this.reloadRoute(uuid)
    },

    getChatByUuidAndIndex(uuid: string, index: number) {
      if (!uuid) {
        if (this.chat.length)
          return this.chat[0].data[index]
        return null
      }
      const chatIndex = this.chat.findIndex(item => item.uuid === uuid)
      if (chatIndex !== -1)
        return this.chat[chatIndex].data[index]
      return null
    },

    addChatByUuid(uuid: string, chat: Chat.Message) {
      const new_chat_text = t('chat.new')
      if (!uuid) {
        if (this.sessions.length === 0) {
          const uuid = uuidv4()
          createChatSession(uuid, chat.text)
          this.sessions.push({ uuid, title: chat.text, isEdit: false })
          // first chat message is prompt
          this.chat.push({ uuid, data: [{ ...chat, isPrompt: true, isPin: false }] })
          this.active = uuid
          this.recordState()
        }
        else {
          this.chat[0].data.push(chat)
          if (this.sessions[0].title === new_chat_text) {
            this.sessions[0].title = chat.text
            renameChatSession(this.sessions[0].uuid, chat.text.substring(0, 20))
          }
          this.recordState()
        }
      }

      const index = this.chat.findIndex(item => item.uuid === uuid)
      if (index !== -1) {
        if (this.chat[index].data.length === 0)
          this.chat[index].data.push({ ...chat, isPrompt: true, isPin: false })
        else
          this.chat[index].data.push(chat)

        if (this.sessions[0].title === new_chat_text) {
          this.sessions[0].title = chat.text
          renameChatSession(this.sessions[0].uuid, chat.text.substring(0, 20))
        }
        this.recordState()
      }
    },

    async updateChatByUuid(uuid: string, index: number, chat: Chat.Message) {
      // TODO: sync with server
      if (!uuid) {
        if (this.chat.length) {
          this.chat[0].data[index] = chat
          this.recordState()
        }
        return
      }

      const chatIndex = this.chat.findIndex(item => item.uuid === uuid)
      if (chatIndex !== -1) {
        this.chat[chatIndex].data[index] = chat
        this.recordState()
      }
    },

    updateChatPartialByUuid(
      uuid: string,
      index: number,
      chat: Partial<Chat.Message>,
    ) {
      if (!uuid) {
        if (this.chat.length) {
          this.chat[0].data[index] = { ...this.chat[0].data[index], ...chat }
          this.recordState()
        }
        return
      }

      const chatIndex = this.chat.findIndex(item => item.uuid === uuid)
      if (chatIndex !== -1) {
        this.chat[chatIndex].data[index] = {
          ...this.chat[chatIndex].data[index],
          ...chat,
        }
        this.recordState()
      }
    },

    async deleteChatByUuid(uuid: string, index: number) {
      if (!uuid) {
        if (this.chat.length) {
          const chatData = this.chat[0].data
          const chat = chatData[index]
          chatData.splice(index, 1)
          this.recordState()
          if (chat)
            await deleteChatData(chat)
        }
        return
      }

      const chatIndex = this.chat.findIndex(item => item.uuid === uuid)
      if (chatIndex !== -1) {
        const chatData = this.chat[chatIndex].data
        const chat = chatData[index]
        chatData.splice(index, 1)
        this.recordState()
        if (chat)
          await deleteChatData(chat)
      }
    },

    clearChatByUuid(uuid: string) {
      // does this every happen?
      if (!uuid) {
        if (this.chat.length) {
          this.chat[0].data = []
          this.recordState()
        }
        return
      }

      const index = this.chat.findIndex(item => item.uuid === uuid)
      if (index !== -1) {
        const data: Chat.Message[] = []
        for (const chat of this.chat[index].data) {
          if (chat.isPin || chat.isPrompt)
            data.push(chat)
        }
        this.chat[index].data = data
        clearSessionChatMessages(uuid)
        this.recordState()
      }
    },
    clearState() {
      this.sessions = []
      this.chat = []
      this.active = null
      this.recordState()
    },
  },
})
