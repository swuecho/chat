<script lang="ts" setup>
import { computed, ref, watch, h, onUnmounted } from 'vue'
import { NSelect, NForm, useMessage } from 'naive-ui'
import { useSessionStore, useAuthStore } from '@/store'
import { useChatModels } from '@/hooks/useChatModels'
import { formatDistanceToNow, differenceInDays } from 'date-fns'
import type { ChatModel } from '@/types/chat-models'
import { API_TYPE_DISPLAY_NAMES, API_TYPES } from '@/constants/apiTypes'

interface ModelFormData {
        model: string | undefined
}

interface ChatSession {
        uuid: string
        model?: string
}

const sessionStore = useSessionStore()
const authStore = useAuthStore()
const message = useMessage()
const { useChatModelsQuery } = useChatModels()

const props = defineProps<{
        uuid: string
        model: string | undefined
}>()

const chatSession = computed(() => sessionStore.getChatSessionByUuid(props.uuid))

const { data, isLoading, error, isError } = useChatModelsQuery()

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

const chatModelOptions = computed(() => {
        if (!data?.value) return []
        
        const enabledModels = data.value.filter((x: ChatModel) => x.isEnable)
        
        // Group models by api type
        const modelsByApiType = enabledModels.reduce((acc, model) => {
                const apiType = model.apiType || 'unknown'
                if (!acc[apiType]) {
                        acc[apiType] = []
                }
                acc[apiType].push(model)
                return acc
        }, {} as Record<string, ChatModel[]>)
        
        // Create grouped options with api type headers
        const groupedOptions: any[] = []
        
        // Define api type display names and order using shared constants
        const apiTypeConfig = {
                [API_TYPES.OPENAI]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.OPENAI], order: 1 },
                [API_TYPES.CLAUDE]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.CLAUDE], order: 2 },
                [API_TYPES.GEMINI]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.GEMINI], order: 3 },
                [API_TYPES.OLLAMA]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.OLLAMA], order: 4 },
                [API_TYPES.CUSTOM]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.CUSTOM], order: 5 },
        }
        
        // Sort api types by defined order
        const sortedApiTypes = Object.keys(modelsByApiType).sort((a, b) => {
                const orderA = apiTypeConfig[a as keyof typeof apiTypeConfig]?.order || 999
                const orderB = apiTypeConfig[b as keyof typeof apiTypeConfig]?.order || 999
                return orderA - orderB
        })
        
        sortedApiTypes.forEach(apiType => {
                const models = modelsByApiType[apiType]
                const apiTypeName = apiTypeConfig[apiType as keyof typeof apiTypeConfig]?.name || apiType.charAt(0).toUpperCase() + apiType.slice(1)
                
                // Sort models within api type by order number
                models.sort((a, b) => (a.orderNumber || 0) - (b.orderNumber || 0))
                
                // Add api type group with models
                groupedOptions.push({
                        type: 'group',
                        label: apiTypeName,
                        key: apiType,
                        children: models.map(optionFromModel)
                })
        })
        
        return groupedOptions
})


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

// Optimistic updates with error handling
const isUpdating = ref(false)

watch(() => modelRef.value.model, async (newModel, oldModel) => {
        // Only trigger update if this is a user-initiated change (both old and new values are defined)
        if (oldModel !== undefined && newModel !== undefined && newModel !== oldModel && newModel) {
                isUpdating.value = true
                
                try {
                        // Persist to server
                        await sessionStore.updateSession(props.uuid, {
                                model: newModel
                        })
                        
                        message.success(`Model updated to ${newModel}`)
                } catch (error) {
                        console.error('Failed to update session model:', error)
                        message.error('Failed to update model selection')
                        
                        // Revert UI state
                        modelRef.value.model = oldModel
                } finally {
                        isUpdating.value = false
                }
        }
})


</script>

<template>
        <NForm ref="formRef" :model="modelRef">
                <NSelect 
                        v-model:value="modelRef.model" 
                        :options="chatModelOptions" 
                        :loading="isLoading || isUpdating"
                        :disabled="isError || isLoading || isUpdating"
                        size='large' 
                        placeholder="Select a model..."
                        :fallback-option="false"
                />
        </NForm>
</template>