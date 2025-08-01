<script lang="ts" setup>
import { computed, ref, watch, h, onUnmounted } from 'vue'
import { NSelect, NForm } from 'naive-ui'
import { useSessionStore, useAuthStore } from '@/store'
import { useChatModels } from '@/hooks/useChatModels'
import { formatDistanceToNow, differenceInDays } from 'date-fns'

// Type definitions
interface ChatModel {
        id: string
        name: string
        label: string
        isEnable: boolean
        isDefault: boolean
        orderNumber: number
        lastUsageTime: string
        provider: string
}

interface ModelFormData {
        model: string
}

interface ChatSession {
        uuid: string
        model?: string
        // Add other session properties as needed
}

const sessionStore = useSessionStore()
const authStore = useAuthStore()
const { useChatModelsQuery } = useChatModels()

const props = defineProps<{
        uuid: string
        model: string | undefined
}>()

const chatSession = computed(() => sessionStore.getChatSessionByUuid(props.uuid))

const { data } = useChatModelsQuery()

// format timestamp 2025-02-04T08:17:16.711644Z (string) as  to show time relative to now
const formatTimestamp = (timestamp: string) => {
        const date = new Date(timestamp)
        const days = differenceInDays(new Date(), date)
        if (days > 30) {
                return 'a month ago'
        }
        return formatDistanceToNow(date, { addSuffix: true })
}

const optionFromModel = (model: ChatModel) => {
        return {
                label: () => h('div', {}, [
                        model.label,
                        h('span', { style: 'color: #999; font-size: 0.8rem; margin-left: 4px' },
                                `- ${formatTimestamp(model.lastUsageTime)}`)
                ]),
                value: model.name,
        }
}
const chatModelOptions = computed(() =>
        data?.value ? data.value.filter((x: ChatModel) => x.isEnable).map(optionFromModel) : []
)


const defaultModel = computed(() => {
        if (!data?.value) return undefined
        const defaultModels = data.value.filter((x: ChatModel) => x.isDefault && x.isEnable)
        if (defaultModels.length === 0) {
                // Fallback to first enabled model if no default is set
                const enabledModels = data.value.filter((x: ChatModel) => x.isEnable)
                if (enabledModels.length > 0) {
                        enabledModels.sort((a: ChatModel, b: ChatModel) => (a.orderNumber || 0) - (b.orderNumber || 0))
                        return enabledModels[0]?.name
                }
                return undefined
        }
        // Sort by order_number to ensure deterministic selection
        defaultModels.sort((a: ChatModel, b: ChatModel) => (a.orderNumber || 0) - (b.orderNumber || 0))
        return defaultModels[0]?.name
})


const modelRef = ref<ModelFormData>({
        model: undefined
})

// Initialize model once both session and default model are available
watch([chatSession, defaultModel], ([session, defaultModelValue]) => {
        if (!modelRef.value.model) {
                // Use session model if available, otherwise use default model
                modelRef.value.model = session?.model ?? defaultModelValue
        }
}, { immediate: true })

// Use computed property instead of manual store subscription for better performance
const sessionModel = computed(() => chatSession.value?.model)

// Watch session model changes to keep form in sync (but only after initialization)
watch(sessionModel, (newSessionModel) => {
        if (modelRef.value.model && newSessionModel && modelRef.value.model !== newSessionModel) {
                modelRef.value.model = newSessionModel
        }
})

// Watch only the model property for user-initiated changes
watch(() => modelRef.value.model, async (newModel, oldModel) => {
        // Only trigger update if this is a user-initiated change (both old and new values are defined)
        if (oldModel !== undefined && newModel !== undefined && newModel !== oldModel && newModel) {
                try {
                        await sessionStore.updateSession(props.uuid, {
                                model: newModel
                        })
                } catch (error) {
                        console.error('Failed to update session model:', error)
                        // Revert the model selection if update failed
                        modelRef.value.model = oldModel
                }
        }
})


</script>

<template>
        <NForm ref="formRef" :model="modelRef">
                <NSelect v-model:value="modelRef.model" :options="chatModelOptions" size='large' />
        </NForm>
</template>