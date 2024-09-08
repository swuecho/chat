<template>
        <Message v-for="(item, index) of dataSources" :key="index" :date-time="item.dateTime"
                :model="chatSession?.model" :text="item.text" :inversion="item.inversion" :error="item.error"
                :is-prompt="item.isPrompt" :is-pin="item.isPin" :loading="item.loading" :index="index"
                @regenerate="onRegenerate(index)" @toggle-pin="handleTogglePin(index)" @delete="handleDelete(index)" @after-edit="handleAfterEdit" />
</template>

<script lang='ts' setup>
import Message from './Message/index.vue';
import { computed, ref } from 'vue';
import { useChatStore } from '@/store';
import { useChat } from '../hooks/useChat'
import { updateChatData } from '@/api'
import { useDialog } from 'naive-ui'


import { t } from '@/locales'
const dialog = useDialog()
const { updateChatText, updateChat } = useChat()

const props = defineProps({
        sessionUuid: {
                type: String,
                required: true
        },
        onRegenerate: {
                type: Function,
                required: true
        },
});

const chatStore = useChatStore()
const dataSources = computed(() => chatStore.getChatSessionDataByUuid(props.sessionUuid))
const chatSession = computed(() => chatStore.getChatSessionByUuid(props.sessionUuid))

// The user wants to delete the message with the given index.
// If the message is already being deleted, we ignore the request.
// If the user confirms that they want to delete the message, we call
// the deleteChatByUuid function from the chat store.
function handleDelete(index: number) {
        dialog.warning({
                title: t('chat.deleteMessage'),
                content: t('chat.deleteMessageConfirm'),
                positiveText: t('common.yes'),
                negativeText: t('common.no'),
                onPositiveClick: async () => {
                        chatStore.deleteChatByUuid(props.sessionUuid, index)
                },
        })
}


function handleAfterEdit(index: number, text: string) {
        console.log(index, text)
        updateChatText(
                props.sessionUuid,
                index,
                text,
        )
}

const pining = ref<boolean>(false)

async function handleTogglePin(index: number) {
        if (pining.value)
                return
        const chat = chatStore.getChatByUuidAndIndex(props.sessionUuid, index)
        if (chat == null)
                return

        chat.isPin = !chat.isPin
        try {
                pining.value = true
                await updateChatData(chat)
                updateChat(
                        props.sessionUuid,
                        index,
                        chat,
                )
        }
        finally {
                pining.value = false
        }
}


</script>
