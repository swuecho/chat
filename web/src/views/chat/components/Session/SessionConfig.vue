<script lang="ts" setup>
import type { Ref } from 'vue'
import { computed, ref, watch, h, nextTick } from 'vue'
import type { FormInst } from 'naive-ui'
import { NForm, NFormItem, NInput, NRadio, NRadioGroup, NSlider, NSpace, NSpin, NSwitch } from 'naive-ui'
import { debounce, isEqual } from 'lodash-es'
import { useSessionStore, useAppStore } from '@/store'
import { fetchChatInstructions, fetchChatModel } from '@/api'

import { useQuery } from "@tanstack/vue-query";
import { formatDistanceToNow, differenceInDays } from 'date-fns'
import { API_TYPE_DISPLAY_NAMES, API_TYPES } from '@/constants/apiTypes'
import type { ChatModel } from '@/types/chat-models'



// format timestamp 2025-02-04T08:17:16.711644Z (string) as  to show time relative to now
const formatTimestamp = (timestamp: string) => {
  const date = new Date(timestamp)
  const days = differenceInDays(new Date(), date)
  if (days > 30) {
    return 'a month ago'
  }
  return formatDistanceToNow(date, { addSuffix: true })
}

const props = defineProps<{
  uuid: string
}>()


const { data, isLoading } = useQuery({
  queryKey: ['chat_models'],
  queryFn: fetchChatModel,
  staleTime: 10 * 60 * 1000,
})

const { data: instructionData, isLoading: isInstructionLoading } = useQuery({
  queryKey: ['chat_instructions'],
  queryFn: fetchChatInstructions,
  staleTime: 10 * 60 * 1000,
})

// Group models by API type/provider
const chatModelOptionsByProvider = computed(() => {
  if (!data?.value) return []

  const enabledModels = data.value.filter((x: ChatModel) => x.isEnable)

  // Group models by apiType
  const modelsByApiType = enabledModels.reduce((acc, model) => {
    const apiType = model.apiType || 'unknown'
    if (!acc[apiType]) {
      acc[apiType] = []
    }
    acc[apiType].push(model)
    return acc
  }, {} as Record<string, ChatModel[]>)

  // Define provider order and display names
  const apiTypeConfig = {
    [API_TYPES.OPENAI]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.OPENAI], order: 1 },
    [API_TYPES.CLAUDE]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.CLAUDE], order: 2 },
    [API_TYPES.GEMINI]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.GEMINI], order: 3 },
    [API_TYPES.OLLAMA]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.OLLAMA], order: 4 },
    [API_TYPES.CUSTOM]: { name: API_TYPE_DISPLAY_NAMES[API_TYPES.CUSTOM], order: 5 },
  }

  // Sort API types by order
  const sortedApiTypes = Object.keys(modelsByApiType).sort((a, b) => {
    const orderA = apiTypeConfig[a as keyof typeof apiTypeConfig]?.order || 999
    const orderB = apiTypeConfig[b as keyof typeof apiTypeConfig]?.order || 999
    return orderA - orderB
  })

  // Create grouped structure
  return sortedApiTypes.map(apiType => {
    const models = modelsByApiType[apiType]
    const apiTypeName = apiTypeConfig[apiType as keyof typeof apiTypeConfig]?.name
      || apiType.charAt(0).toUpperCase() + apiType.slice(1)

    // Sort models within each group by orderNumber
    models.sort((a, b) => (a.orderNumber || 0) - (b.orderNumber || 0))

    return {
      type: apiType,
      label: apiTypeName,
      models: models,
    }
  })
})

const sessionStore = useSessionStore()
const appStore = useAppStore()

const session = computed(() => sessionStore.getChatSessionByUuid(props.uuid))

interface ModelType {
  chatModel: string
  contextCount: number
  temperature: number
  maxTokens: number
  topP: number
  n: number
  debug: boolean
  summarizeMode: boolean
  codeRunnerEnabled: boolean
  artifactEnabled: boolean
  exploreMode: boolean
}

const defaultModel = computed(() => {
  if (!data?.value) return 'gpt-3.5-turbo'
  const defaultModels = data.value.filter((x: any) => x.isDefault && x.isEnable)
  if (defaultModels.length === 0) return 'gpt-3.5-turbo'
  // Sort by order_number to ensure deterministic selection
  defaultModels.sort((a: any, b: any) => (a.orderNumber || 0) - (b.orderNumber || 0))
  return defaultModels[0]?.name || 'gpt-3.5-turbo'
})

const modelRef: Ref<ModelType> = ref({
  chatModel: session.value?.model ?? defaultModel.value,
  summarizeMode: session.value?.summarizeMode ?? false,
  contextCount: session.value?.maxLength ?? 10,
  temperature: session.value?.temperature ?? 1.0,
  maxTokens: session.value?.maxTokens ?? 2048,
  topP: session.value?.topP ?? 1.0,
  n: session.value?.n ?? 1,
  debug: session.value?.debug ?? false,
  exploreMode: session.value?.exploreMode ?? false,
  codeRunnerEnabled: session.value?.codeRunnerEnabled ?? false,
  artifactEnabled: session.value?.artifactEnabled ?? false,
})

const artifactInstruction = computed(() => instructionData.value?.artifactInstruction ?? '')
const toolInstruction = computed(() => instructionData.value?.toolInstruction ?? '')
const showInstructionPanel = computed(() => {
  // Show panel if either artifact or code runner mode is enabled
  // The individual instruction blocks will handle showing/hiding based on data availability
  return modelRef.value.artifactEnabled || modelRef.value.codeRunnerEnabled
})

const formRef = ref<FormInst | null>(null)

// Flag to prevent circular updates
let isUpdatingFromSession = false

const debouneUpdate = debounce(async (model: ModelType) => {
  // Prevent update if we're currently updating from session
  if (isUpdatingFromSession) {
    return
  }
  
  sessionStore.updateSession(props.uuid, {
    maxLength: model.contextCount,
    temperature: model.temperature,
    maxTokens: model.maxTokens,
    topP: model.topP,
    n: model.n,
    debug: model.debug,
    model: model.chatModel,
    summarizeMode: model.summarizeMode,
    codeRunnerEnabled: model.codeRunnerEnabled,
    artifactEnabled: model.artifactEnabled,
    exploreMode: model.exploreMode,
  })
}, 200)

// Watch modelRef changes from user interaction
watch(modelRef, async (modelValue: ModelType) => {
  debouneUpdate(modelValue)
}, { deep: true })

// Watch for session changes and update modelRef
watch(session, (newSession) => {
  if (newSession) {
    const newModelRef = {
      chatModel: newSession.model ?? defaultModel.value,
      summarizeMode: newSession.summarizeMode ?? false,
      contextCount: newSession.maxLength ?? 10,
      temperature: newSession.temperature ?? 1.0,
      maxTokens: newSession.maxTokens ?? 2048,
      topP: newSession.topP ?? 1.0,
      n: newSession.n ?? 1,
      debug: newSession.debug ?? false,
      exploreMode: newSession.exploreMode ?? false,
      codeRunnerEnabled: newSession.codeRunnerEnabled ?? false,
      artifactEnabled: newSession.artifactEnabled ?? false,
    }
    
    // Only update if the values are actually different
    if (!isEqual(modelRef.value, newModelRef)) {
      isUpdatingFromSession = true
      modelRef.value = newModelRef
      
      // Reset flag after Vue's next tick to allow reactivity to settle
      nextTick(() => {
        isUpdatingFromSession = false
      })
    }
  }
}, { deep: true, immediate: true })



const tokenUpperLimit = computed(() => {
  if (data && data.value) {
    for (let modelConfig of data.value) {
      if (modelConfig.name == modelRef.value.chatModel) {
        return modelConfig.maxToken
      }

    }

  }
  return 1024 * 4
})

const defaultToken = computed(() => {
  if (data && data.value) {
    for (let modelConfig of data.value) {
      if (modelConfig.name == modelRef.value.chatModel) {
        return modelConfig.defaultToken
      }
    }
  }
  return 2048
})
// 1. how to fix the NSelect error?
</script>

<template>
  <!-- https://platform.openai.com/playground?mode=chat -->
  <div>
    <NForm ref="formRef" :model="modelRef" size="small" label-placement="top" :label-width="20">
      <NFormItem :label="$t('chat.model')" path="chatModel">
        <div v-if="isLoading">
          <NSpin size="medium" />
        </div>
        <NRadioGroup v-else v-model:value="modelRef.chatModel">
          <div v-for="providerGroup in chatModelOptionsByProvider" :key="providerGroup.type" class="model-provider-group">
            <div class="provider-label">{{ providerGroup.label }}</div>
            <NSpace vertical>
              <NRadio v-for="model in providerGroup.models" :key="model.name" :value="model.name">
                <div>
                  {{ model.label }}
                  <span style="color: #999; font-size: 0.8rem; margin-left: 4px">
                    - {{ formatTimestamp(model.lastUsageTime) }}
                  </span>
                </div>
              </NRadio>
            </NSpace>
          </div>
        </NRadioGroup>
      </NFormItem>
      <!-- not implemented
      <NFormItem :label="$t('chat.summarize_mode')" path="summarize_mode">
        <NSwitch v-model:value="modelRef.summarizeMode" data-testid="summarize_mode">
          <template #checked>
            {{ $t('chat.is_summarize_mode') }}
          </template>
<template #unchecked>
            {{ $t('chat.no_summarize_mode') }}
          </template>
</NSwitch>
</NFormItem>
-->
      <NFormItem :label="$t('chat.contextCount', { contextCount: modelRef.contextCount })" path="contextCount">
        <NSlider v-model:value="modelRef.contextCount" :min="1" :max="40" :tooltip="false" show-tooltip />
      </NFormItem>
      <NFormItem :label="$t('chat.temperature', { temperature: modelRef.temperature })" path="temperature">
        <NSlider v-model:value="modelRef.temperature" :min="0.1" :max="1" :step="0.01" :tooltip="false" />
      </NFormItem>
      <NFormItem :label="$t('chat.topP', { topP: modelRef.topP })" path="topP">
        <NSlider v-model:value="modelRef.topP" :min="0" :max="1" :step="0.01" :tooltip="false" />
      </NFormItem>
      <NFormItem :label="$t('chat.maxTokens', { maxTokens: modelRef.maxTokens })" path="maxTokens">
        <NSlider v-model:value="modelRef.maxTokens" :min="256" :max="tokenUpperLimit" :default-value="defaultToken"
          :step="16" :tooltip="false" />
      </NFormItem>
      <NFormItem v-if="modelRef.chatModel.startsWith('gpt') || modelRef.chatModel.includes('davinci')"
        :label="$t('chat.N', { n: modelRef.n })" path="n">
        <NSlider v-model:value="modelRef.n" :min="1" :max="10" :step="1" :tooltip="false" />
      </NFormItem>
      <NFormItem :label="$t('chat.artifactMode')" path="artifactEnabled">
        <NSwitch v-model:value="modelRef.artifactEnabled" data-testid="artifact_mode">
          <template #checked>
            {{ $t('chat.enable_artifact') }}
          </template>
          <template #unchecked>
            {{ $t('chat.disable_artifact') }}
          </template>
        </NSwitch>
      </NFormItem>
      <NFormItem :label="$t('chat.codeRunner')" path="codeRunnerEnabled">
        <NSwitch v-model:value="modelRef.codeRunnerEnabled" data-testid="code_runner_mode">
          <template #checked>
            {{ $t('chat.enable_code_runner') }}
          </template>
          <template #unchecked>
            {{ $t('chat.disable_code_runner') }}
          </template>
        </NSwitch>
      </NFormItem>
      <NFormItem v-if="showInstructionPanel" :label="$t('chat.promptInstructions')">
        <div class="instruction-panel">
          <div v-if="isInstructionLoading" class="instruction-loading">
            <NSpin size="small" />
            <span>{{ $t('chat.loading_instructions') }}</span>
          </div>
          <template v-else>
            <!-- Artifact Instructions -->
            <div v-if="modelRef.artifactEnabled" class="instruction-block">
              <div class="instruction-title">{{ $t('chat.artifactInstructionTitle') }}</div>
              <NInput
                v-if="artifactInstruction"
                class="instruction-input"
                :value="artifactInstruction"
                type="textarea"
                readonly
                :autosize="{ minRows: 3, maxRows: 10 }"
              />
              <div v-else class="instruction-empty">
                No artifact instructions available
              </div>
            </div>

            <!-- Tool/Code Runner Instructions -->
            <div v-if="modelRef.codeRunnerEnabled" class="instruction-block">
              <div class="instruction-title">{{ $t('chat.toolInstructionTitle') }}</div>
              <NInput
                v-if="toolInstruction"
                class="instruction-input"
                :value="toolInstruction"
                type="textarea"
                readonly
                :autosize="{ minRows: 3, maxRows: 10 }"
              />
              <div v-else class="instruction-empty">
                No code runner instructions available
              </div>
            </div>
          </template>
        </div>
      </NFormItem>
      <NFormItem :label="$t('chat.exploreMode')" path="exploreMode">
        <NSwitch v-model:value="modelRef.exploreMode" data-testid="explore_mode">
          <template #checked>
            {{ $t('chat.enable_explore') }}
          </template>
          <template #unchecked>
            {{ $t('chat.disable_explore') }}
          </template>
        </NSwitch>
      </NFormItem>
      <NFormItem :label="$t('chat.debug')" path="debug">
        <NSwitch v-model:value="modelRef.debug" data-testid="debug_mode">
          <template #checked>
            {{ $t('chat.enable_debug') }}
          </template>
          <template #unchecked>
            {{ $t('chat.disable_debug') }}
          </template>
        </NSwitch>
      </NFormItem>
    
    </NForm>
    <!--
                                        <div class="center">
                                          <pre>{{ JSON.stringify(modelRef, null, 2) }} </pre>
                                        </div>
                                        -->
  </div>
</template>

<style scoped>
.model-provider-group {
  margin-bottom: 16px;
}

.model-provider-group:last-child {
  margin-bottom: 0;
}

.provider-label {
  font-size: 0.9rem;
  font-weight: 600;
  color: #666;
  margin-bottom: 8px;
  padding-bottom: 4px;
  border-bottom: 1px solid #e0e0e0;
}

:deep(.dark) .provider-label {
  color: #999;
  border-bottom-color: #3a3a3a;
}

.instruction-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.instruction-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #666;
  font-size: 0.85rem;
}

.instruction-title {
  font-size: 0.85rem;
  color: #666;
  margin-bottom: 6px;
}

.instruction-input :deep(textarea) {
  font-family: "SFMono-Regular", Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 0.85rem;
}

.instruction-empty {
  padding: 12px;
  background-color: #f5f5f5;
  border-radius: 4px;
  color: #999;
  font-size: 0.85rem;
  font-style: italic;
  text-align: center;
}

:deep(.dark) .instruction-empty {
  background-color: #2a2a2a;
  color: #888;
}
</style>
