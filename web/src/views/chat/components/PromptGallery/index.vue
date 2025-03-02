<script lang="ts" setup>
import { ref, computed } from 'vue'
import { NCard, NButton, NSpace, NTabs, NTabPane } from 'naive-ui'
import { usePromptStore } from '@/store/modules'
import { fetchChatSnapshot, fetchSnapshotAll } from '@/api'
import { useQuery } from '@tanstack/vue-query'
import PromptCards from './PromptCards.vue'

interface Emit {
        (ev: 'usePrompt', key: string, prompt: string): void
}

const emit = defineEmits<Emit>()
const promptStore = usePromptStore()

// Fetch bots data
const { data: bots } = useQuery({
        queryKey: ['bots'],
        queryFn: async () => await fetchSnapshotAll(),
})

interface Bot {
        title: string
        uuid: string
        typ: string
}

// Get bot prompts
const botPrompts = computed(() => {
        return (bots.value || [])
                .filter((bot: Bot) => bot.typ === 'chatbot')
                .map((bot: Bot) => ({
                        key: bot.title,
                        uuid: bot.uuid,
                        value: ''
                }))
})

const handleUsePrompt = (key: string, prompt: string, uuid?: string) => {
        if (uuid) {
                fetchChatSnapshot(uuid).then((data) => {
                        emit('usePrompt', key, data.conversation[0].text)
                })
        } else {
                emit('usePrompt', key, prompt)
        }
}

</script>

<template>
        <NTabs type="line" animated>
                <NTabPane v-if="botPrompts.length > 0" name="bots" tab="Bots">
                        <div class="mt-4">
                                <PromptCards :prompts="botPrompts" @usePrompt="handleUsePrompt" />
                        </div>
                </NTabPane>
                <NTabPane name="prompts" tab="Prompts">
                        <div class="mt-4">
                                <PromptCards :prompts="promptStore.promptList" @usePrompt="handleUsePrompt" />
                        </div>
                </NTabPane>
        </NTabs>
</template>
