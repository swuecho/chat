<template>
        <Message v-for="(item, index) of dataSources" :key="item.uuid || `message-${index}`" :date-time="item.dateTime"
                :model="item?.model || chatSession?.model" :text="item.text" :inversion="item.inversion" :error="item.error"
                :is-prompt="item.isPrompt" :is-pin="item.isPin" :loading="item.loading" :index="index"
                :artifacts="item.artifacts" :suggested-questions="item.suggestedQuestions" 
                :suggested-questions-loading="item.suggestedQuestionsLoading" 
                :suggested-questions-batches="item.suggestedQuestionsBatches"
                :current-suggested-questions-batch="item.currentSuggestedQuestionsBatch"
                :suggested-questions-generating="item.suggestedQuestionsGenerating"
                :explore-mode="chatSession?.exploreMode"
                @regenerate="onRegenerate(index)" 
                @toggle-pin="handleTogglePin(index)" 
                @delete="handleDelete(index)" 
                @after-edit="handleAfterEdit" 
                @use-question="handleUseQuestion"
                @generate-more-suggestions="handleGenerateMoreSuggestions(index)"
                @previous-suggestions-batch="handlePreviousSuggestionsBatch(index)"
                @next-suggestions-batch="handleNextSuggestionsBatch(index)" />
</template>

<script lang='ts' setup>
import Message from './Message/index.vue';
import { computed, ref } from 'vue';
import { useMessageStore, useSessionStore } from '@/store';
import { useChat } from '@/views/chat/hooks/useChat'
import { updateChatData } from '@/api'
import { useDialog } from 'naive-ui'
import { useCopyCode } from '@/views/chat/hooks/useCopyCode'
import { useErrorHandling } from '../composables/useErrorHandling'


import { t } from '@/locales'
const dialog = useDialog()
const { updateChatText, updateChat } = useChat()
const { handleApiError } = useErrorHandling()

useCopyCode()


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

const emit = defineEmits(['useQuestion']);

const messageStore = useMessageStore()
const sessionStore = useSessionStore()
const dataSources = computed(() => messageStore.getChatSessionDataByUuid(props.sessionUuid))
const chatSession = computed(() => sessionStore.getChatSessionByUuid(props.sessionUuid))

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
                        const message = dataSources.value[index]
                        if (message && message.uuid) {
                                messageStore.removeMessage(props.sessionUuid, message.uuid)
                        }
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

function handleUseQuestion(question: string) {
        emit('useQuestion', question)
}

const pining = ref<boolean>(false)

async function handleTogglePin(index: number) {
        if (pining.value)
                return
        const messages = messageStore.getChatSessionDataByUuid(props.sessionUuid)
        const message = messages && messages[index] ? messages[index] : null
        if (message == null)
                return

        const previousPin = message.isPin
        message.isPin = !message.isPin
        try {
                pining.value = true
                await updateChatData(message)
                updateChat(
                        props.sessionUuid,
                        index,
                        message,
                )
        } catch (error) {
                message.isPin = previousPin
                handleApiError(error, 'toggle-pin')
        }
        finally {
                pining.value = false
        }
}

// Handle suggested questions functionality
async function handleGenerateMoreSuggestions(index: number) {
        const message = dataSources.value[index]
        if (message && message.uuid) {
                try {
                        await messageStore.generateMoreSuggestedQuestions(props.sessionUuid, message.uuid)
                } catch (error) {
                        handleApiError(error, 'generate-more-suggestions')
                }
        }
}

function handlePreviousSuggestionsBatch(index: number) {
        const message = dataSources.value[index]
        if (message && message.uuid) {
                messageStore.previousSuggestedQuestionsBatch(props.sessionUuid, message.uuid)
        }
}

function handleNextSuggestionsBatch(index: number) {
        const message = dataSources.value[index]
        if (message && message.uuid) {
                messageStore.nextSuggestedQuestionsBatch(props.sessionUuid, message.uuid)
        }
}


</script>
