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

// Remove or comment out the optionFromModel function
// const optionFromModel = (model: any) => { ... }

const chatModelOptions = computed(() =>
  data?.value ? data.value.filter((x: any) => x.isEnable) : []
)

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
  if (!modelRef.value.artifactEnabled && !modelRef.value.codeRunnerEnabled) {
    return false
  }
  if (isInstructionLoading.value) {
    return true
  }
  return Boolean(
    (modelRef.value.artifactEnabled && artifactInstruction.value) ||
    (modelRef.value.codeRunnerEnabled && toolInstruction.value)
  )
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
        <NRadioGroup v-else v-model:value="modelRef.chatModel" class="model-radio-group">
          <div class="model-cards-container">
            <div
              v-for="model in chatModelOptions"
              :key="model.name"
              class="model-card"
              :class="{ 'model-card--checked': modelRef.chatModel === model.name }"
            >
              <NRadio :value="model.name" class="model-card__radio">
                <div class="model-card__content">
                  <div class="model-card__label">{{ model.label }}</div>
                  <div class="model-card__meta">
                    {{ formatTimestamp(model.lastUsageTime) }}
                  </div>
                </div>
              </NRadio>
            </div>
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
          <div v-if="modelRef.artifactEnabled && artifactInstruction" class="instruction-block">
            <div class="instruction-title">{{ $t('chat.artifactInstructionTitle') }}</div>
            <NInput
              class="instruction-input"
              :value="artifactInstruction"
              type="textarea"
              readonly
              :autosize="{ minRows: 3, maxRows: 10 }"
            />
          </div>
          <div v-if="modelRef.codeRunnerEnabled && toolInstruction" class="instruction-block">
            <div class="instruction-title">{{ $t('chat.toolInstructionTitle') }}</div>
            <NInput
              class="instruction-input"
              :value="toolInstruction"
              type="textarea"
              readonly
              :autosize="{ minRows: 3, maxRows: 10 }"
            />
          </div>
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
.model-radio-group {
  width: 100%;
}

.model-cards-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.model-card {
  border: 2px solid transparent;
  border-radius: 8px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: var(--n-color);
  position: relative;
}

.model-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-color: var(--n-border-color);
}

.model-card--checked {
  border-color: #18a058;
  background: linear-gradient(135deg, rgba(24, 160, 88, 0.05) 0%, rgba(24, 160, 88, 0.02) 100%);
}

.model-card__radio {
  width: 100%;
}

.model-card__radio :deep(.n-radio__dot) {
  align-self: flex-start;
  margin-top: 2px;
}

.model-card__content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.model-card__label {
  font-size: 14px;
  font-weight: 500;
  color: var(--n-text-color);
}

.model-card__meta {
  font-size: 12px;
  color: var(--n-text-color-disabled);
}

/* Dark mode adjustments */
@media (prefers-color-scheme: dark) {
  .model-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .model-card--checked {
    background: linear-gradient(135deg, rgba(24, 160, 88, 0.1) 0%, rgba(24, 160, 88, 0.05) 100%);
  }
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
</style>
