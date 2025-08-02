import { ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { v4 as uuidv4 } from 'uuid'
import { createChatBot, createChatSnapshot, getChatSessionDefault } from '@/api'
import { useAppStore, useSessionStore, useMessageStore } from '@/store'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { useChat } from '@/views/chat/hooks/useChat'
import { nowISO } from '@/utils/date'
import { extractArtifacts } from '@/utils/artifacts'
import { t } from '@/locales'

export function useChatActions(sessionUuid: string) {
  const dialog = useDialog()
  const nui_msg = useMessage()
  const sessionStore = useSessionStore()
  const messageStore = useMessageStore()
  const appStore = useAppStore()
  const { isMobile } = useBasicLayout()
  const { addChat } = useChat()

  const snapshotLoading = ref<boolean>(false)
  const botLoading = ref<boolean>(false)
  const showUploadModal = ref<boolean>(false)
  const showModal = ref<boolean>(false)
  const showArtifactGallery = ref<boolean>(false)

  async function handleAdd(dataSources: any[]) {
    if (dataSources.length > 0) {
      const new_chat_text = t('chat.new')
      try {
        await sessionStore.createNewSession(new_chat_text)
        if (isMobile.value)
          appStore.setSiderCollapsed(true)
      } catch (error) {
        console.error('Failed to create new session:', error)
      }
    } else {
      nui_msg.warning(t('chat.alreadyInNewChat'))
    }
  }

  async function handleSnapshot() {
    snapshotLoading.value = true
    try {
      const snapshot = await createChatSnapshot(sessionUuid)
      const snapshot_uuid = snapshot.uuid
      window.open(`#/snapshot/${snapshot_uuid}`, '_blank')
      nui_msg.success(t('chat.snapshotSuccess'))
    } catch (error) {
      nui_msg.error(t('chat.snapshotFailed'))
    } finally {
      snapshotLoading.value = false
    }
  }

  async function handleCreateBot() {
    botLoading.value = true
    try {
      const snapshot = await createChatBot(sessionUuid)
      const snapshot_uuid = snapshot.uuid
      window.open(`#/snapshot/${snapshot_uuid}`, '_blank')
      nui_msg.success(t('chat.botSuccess'))
    } catch (error) {
      nui_msg.error(t('chat.botFailed'))
    } finally {
      botLoading.value = false
    }
  }

  function handleClear(loading: any) {
    if (loading.value)
      return

    console.log('🔄 handleClear called with sessionUuid:', sessionUuid)

    dialog.warning({
      title: t('chat.clearChat'),
      content: t('chat.clearChatConfirm'),
      positiveText: t('common.yes'),
      negativeText: t('common.no'),
      onPositiveClick: () => {
        console.log('🔄 Clearing messages for sessionUuid:', sessionUuid)
        messageStore.clearSessionMessages(sessionUuid)
      },
    })
  }

  const toggleArtifactGallery = (): void => {
    showArtifactGallery.value = !showArtifactGallery.value
  }

  const handleVFSFileUploaded = (fileInfo: any) => {
    nui_msg.success(`📁 File uploaded: ${fileInfo.filename}`)
  }

  const handleCodeExampleAdded = async (codeInfo: any, streamResponse: any) => {
    const exampleMessage = `📁 **Files uploaded successfully!**

**Python example:**
\`\`\`python <!-- executable: Python code to use the uploaded files -->
${codeInfo.python}
\`\`\`

**JavaScript example:**
\`\`\`javascript <!-- executable: JavaScript code to use the uploaded files -->
${codeInfo.javascript}
\`\`\`

Your files are now available in the Virtual File System! 🚀`

    const chatUuid = uuidv4()
    addChat(
      sessionUuid,
      {
        uuid: chatUuid,
        dateTime: nowISO(),
        text: exampleMessage,
        inversion: true,
        error: false,
        loading: false,
        artifacts: extractArtifacts(exampleMessage),
      },
    )

    try {
      await streamResponse(chatUuid, exampleMessage)
      nui_msg.success('Files uploaded! Code examples added to chat.')
    } catch (error) {
      console.error('Failed to stream code example response:', error)
    }
  }

  return {
    snapshotLoading,
    botLoading,
    showUploadModal,
    showModal,
    showArtifactGallery,
    handleAdd,
    handleSnapshot,
    handleCreateBot,
    handleClear,
    toggleArtifactGallery,
    handleVFSFileUploaded,
    handleCodeExampleAdded
  }
}