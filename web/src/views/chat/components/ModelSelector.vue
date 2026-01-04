<script lang="ts" setup>
import { computed, ref, watch, h } from 'vue'
import { NSelect, NForm, useMessage } from 'naive-ui'
import { useSessionStore } from '@/store'
import { useChatModels } from '@/hooks/useChatModels'
import { formatDistanceToNow, differenceInDays } from 'date-fns'
import type { ChatModel } from '@/types/chat-models'
import { API_TYPE_DISPLAY_NAMES, API_TYPES } from '@/constants/apiTypes'

const sessionStore = useSessionStore()
const message = useMessage()
const { useChatModelsQuery } = useChatModels()

const props = defineProps<{
  uuid: string
  model: string | undefined
}>()

const chatSession = computed(() => sessionStore.getChatSessionByUuid(props.uuid))
const { data, isLoading, isError } = useChatModelsQuery()

const formatTimestamp = (timestamp?: string) => {
  if (!timestamp) {
    return 'Never used'
  }
  const date = new Date(timestamp)
  if (Number.isNaN(date.getTime())) {
    return 'Never used'
  }
  const days = differenceInDays(new Date(), date)
  if (days > 30) {
    return 'a month ago'
  }
  return formatDistanceToNow(date, { addSuffix: true })
}

const optionFromModel = (model: ChatModel) => ({
  label: () => h('div', {}, [
    model.label,
    h('span', { style: 'color: #999; font-size: 0.8rem; margin-left: 4px' },
      `- ${formatTimestamp(model.lastUsageTime)}`),
  ]),
  value: model.name,
})

const chatModelOptions = computed(() => {
  if (!data?.value) return []

  const enabledModels = data.value.filter((x: ChatModel) => x.isEnable)
  const modelsByApiType = enabledModels.reduce((acc, model) => {
    const apiType = model.apiType || 'unknown'
    if (!acc[apiType]) {
      acc[apiType] = []
    }
    acc[apiType].push(model)
    return acc
  }, {} as Record<string, ChatModel[]>)

  const apiTypeConfig = {
    [API_TYPES.OPENAI]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.OPENAI], order: 1 },
    [API_TYPES.CLAUDE]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.CLAUDE], order: 2 },
    [API_TYPES.GEMINI]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.GEMINI], order: 3 },
    [API_TYPES.OLLAMA]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.OLLAMA], order: 4 },
    [API_TYPES.CUSTOM]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.CUSTOM], order: 5 },
  }

  const sortedApiTypes = Object.keys(modelsByApiType).sort((a, b) => {
    const orderA = apiTypeConfig[a as keyof typeof apiTypeConfig]?.order || 999
    const orderB = apiTypeConfig[b as keyof typeof apiTypeConfig]?.order || 999
    return orderA - orderB
  })

  return sortedApiTypes.map(apiType => {
    const models = modelsByApiType[apiType]
    const apiTypeName = apiTypeConfig[apiType as keyof typeof apiTypeConfig]?.name
      || apiType.charAt(0).toUpperCase() + apiType.slice(1)

    models.sort((a, b) => (a.orderNumber || 0) - (b.orderNumber || 0))

    return {
      type: 'group',
      label: apiTypeName,
      key: apiType,
      children: models.map(optionFromModel),
    }
  })
})

const defaultModel = computed(() => {
  if (!data?.value) return undefined
  const defaultModels = data.value.filter((x: ChatModel) => x.isDefault && x.isEnable)
  if (!defaultModels.length) {
    const enabledModels = data.value.filter((x: ChatModel) => x.isEnable)
    if (!enabledModels.length) return undefined
    enabledModels.sort((a, b) => (a.orderNumber || 0) - (b.orderNumber || 0))
    return enabledModels[0]?.name
  }
  defaultModels.sort((a, b) => (a.orderNumber || 0) - (b.orderNumber || 0))
  return defaultModels[0]?.name
})

const selectedModel = ref<string | undefined>(undefined)
const initialized = ref(false)
const isUpdating = ref(false)
const programmaticValue = ref<string | undefined>(undefined)

const setSelectedProgrammatically = (value: string | undefined) => {
  if (selectedModel.value === value) return
  programmaticValue.value = value
  selectedModel.value = value
}

watch([chatSession, defaultModel], ([session, defaultModelValue]) => {
  if (!session && !defaultModelValue) return
  const nextModel = session?.model ?? defaultModelValue
  if (!initialized.value) {
    setSelectedProgrammatically(nextModel)
    initialized.value = true
    return
  }

  if (session?.model && session.model !== selectedModel.value) {
    setSelectedProgrammatically(session.model)
  }
}, { immediate: true })

const handleUserUpdate = async (newModel: string | undefined) => {
  if (newModel === programmaticValue.value) {
    programmaticValue.value = undefined
    return
  }

  if (!initialized.value || !newModel) {
    return
  }

  const currentSessionModel = chatSession.value?.model
  if (currentSessionModel === newModel) {
    return
  }

  const previousModel = currentSessionModel ?? defaultModel.value
  isUpdating.value = true

  try {
    await sessionStore.updateSession(props.uuid, { model: newModel })
    const selected = data?.value?.find((model: ChatModel) => model.name === newModel)
    const displayName = selected?.label || newModel
    message.success(`Model updated to ${displayName}`)
  } catch (error) {
    console.error('Failed to update session model:', error)
    message.error('Failed to update model selection')
    setSelectedProgrammatically(previousModel)
  } finally {
    isUpdating.value = false
  }
}
</script>

<template>
  <NForm :model="{ model: selectedModel }">
    <NSelect
      v-model:value="selectedModel"
      :options="chatModelOptions"
      :loading="isLoading || isUpdating"
      :disabled="isError || isLoading || isUpdating"
      size="large"
      placeholder="Select a model..."
      :fallback-option="false"
      @update:value="handleUserUpdate"
    />
  </NForm>
</template>
