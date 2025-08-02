<script lang="ts" setup>
import { computed, ref } from 'vue'
import { NTabs, NTabPane } from 'naive-ui'
import { usePromptStore } from '@/store/modules'
import { fetchChatbotAll, fetchChatSnapshot } from '@/api'
import { useQuery } from '@tanstack/vue-query'
import PromptCards from './PromptCards.vue'
import { SvgIcon } from '@/components/common'
import { t } from '@/locales'


interface Emit {
        (ev: 'usePrompt', key: string, prompt: string): void
}

const emit = defineEmits<Emit>()
const promptStore = usePromptStore()

// Fetch bots data
const { data: bots } = useQuery({
        queryKey: ['bots'],
        queryFn: async () => await fetchChatbotAll(),
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

const activeTab = ref('prompts')

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
        <NTabs default-value="prompts" type="line" animated>
                <NTabPane name="prompts">
                        <template #tab>
                                <div class="flex items-center gap-1">
                                        <SvgIcon icon="ri:lightbulb-line" class="w-4 h-4" />
                                        <span> {{ t('prompt.store') }}</span>
                                </div>
                        </template>
                        <div class="mt-4">
                                <PromptCards :prompts="promptStore.promptList" @usePrompt="handleUsePrompt" />
                        </div>
                </NTabPane>
                <NTabPane v-if="botPrompts.length > 0" name="bots">
                        <template #tab>
                                <div class="flex items-center gap-1">
                                        <SvgIcon icon="majesticons:robot-line" class="w-4 h-4" />
                                        <span>{{ t('bot.list') }}</span>
                                </div>
                        </template>
                        <div class="mt-4">
                                <PromptCards :prompts="botPrompts" @usePrompt="handleUsePrompt" />
                        </div>
                </NTabPane>
              
        </NTabs>
</template>
