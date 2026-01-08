<script lang="ts" setup>
import type { Ref } from 'vue'
import { computed, ref, watch, h, nextTick } from 'vue'
import type { FormInst, CollapseInst } from 'naive-ui'
import {
  NForm,
  NFormItem,
  NInput,
  NRadio,
  NRadioGroup,
  NSlider,
  NSpace,
  NSpin,
  NSwitch,
  NCollapse,
  NCollapseItem,
  NIcon,
  NTooltip
} from 'naive-ui'
import {
  SettingsOutlined,
  PsychologyOutlined,
  TuneOutlined,
  ExtensionOutlined,
  CodeOutlined,
  ExploreOutlined,
  BugReportOutlined,
  KeyboardArrowDownOutlined,
  SpeedOutlined,
  MemoryOutlined
} from '@vicons/material'
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
  showToolDebug: boolean
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
  showToolDebug: session.value?.showToolDebug ?? false,
})

const artifactInstruction = computed(() => instructionData.value?.artifactInstruction ?? '')
const toolInstruction = computed(() => instructionData.value?.toolInstruction ?? '')
const showInstructionPanel = computed(() => {
  // Show panel if either artifact or code runner mode is enabled
  // The individual instruction blocks will handle showing/hiding based on data availability
  return modelRef.value.artifactEnabled || modelRef.value.codeRunnerEnabled
})

const formRef = ref<FormInst | null>(null)
const collapseRef = ref<CollapseInst | null>(null)

// Flag to prevent circular updates
let isUpdatingFromSession = false

// Expand/collapse state - accordion mode, modes section open by default
const expandedNames = ref<string[]>(['modes'])

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
    showToolDebug: model.showToolDebug,
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
      showToolDebug: newSession.showToolDebug ?? false,
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
  <div class="session-config-container">
    <!-- Collapsible Sections - Accordion -->
    <NCollapse v-model:expanded-names="expandedNames" class="config-collapse" accordion>
      <!-- Model Selection Section -->
      <NCollapseItem name="model" class="collapse-item">
        <template #header>
          <div class="collapse-header" data-testid="collapse-model">
            <NIcon :component="PsychologyOutlined" size="18" />
            <span>{{ $t('chat.model') }}</span>
          </div>
        </template>
        <div v-if="isLoading" class="loading-container">
          <NSpin size="medium" />
          <span class="loading-text">{{ $t('chat.loading_models') }}</span>
        </div>
        <NRadioGroup v-else v-model:value="modelRef.chatModel" class="model-radio-group">
          <div v-for="providerGroup in chatModelOptionsByProvider" :key="providerGroup.type" class="provider-card">
            <div class="provider-header">
              <div class="provider-label">{{ providerGroup.label }}</div>
              <div class="provider-count">{{ providerGroup.models.length }} {{ $t('chat.models') }}</div>
            </div>
            <div class="model-grid">
              <div
                v-for="model in providerGroup.models"
                :key="model.name"
                :class="['model-card', { active: modelRef.chatModel === model.name }]"
                @click="modelRef.chatModel = model.name"
              >
                <NRadio :value="model.name" :checked="modelRef.chatModel === model.name" class="model-radio" />
                <div class="model-info">
                  <div class="model-name">{{ model.label }}</div>
                  <div class="model-meta">
                    <span class="model-timestamp">{{ formatTimestamp(model.lastUsageTime) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </NRadioGroup>
      </NCollapseItem>

      <!-- Modes Section -->
      <NCollapseItem name="modes" class="collapse-item">
        <template #header>
          <div class="collapse-header" data-testid="collapse-modes">
            <NIcon :component="ExtensionOutlined" size="18" />
            <span>{{ $t('chat.modes') }}</span>
          </div>
        </template>
        <div class="modes-grid">
          <!-- Artifact Mode -->
          <div
            :class="['mode-card', { enabled: modelRef.artifactEnabled }]"
            @click="modelRef.artifactEnabled = !modelRef.artifactEnabled"
          >
            <div class="mode-header">
              <NIcon :component="ExtensionOutlined" :size="24" class="mode-icon" />
              <div class="mode-info">
                <div class="mode-name">{{ $t('chat.artifactMode') }}</div>
                <div class="mode-description">{{ $t('chat.artifactModeDescription') }}</div>
              </div>
            </div>
            <NSwitch v-model:value="modelRef.artifactEnabled" data-testid="artifact_mode" size="medium" @click.stop />
          </div>

          <!-- Code Runner Mode -->
          <div
            :class="['mode-card', { enabled: modelRef.codeRunnerEnabled }]"
            @click="modelRef.codeRunnerEnabled = !modelRef.codeRunnerEnabled"
          >
            <div class="mode-header">
              <NIcon :component="CodeOutlined" :size="24" class="mode-icon" />
              <div class="mode-info">
                <div class="mode-name">{{ $t('chat.codeRunner') }}</div>
                <div class="mode-description">{{ $t('chat.codeRunnerDescription') }}</div>
              </div>
            </div>
            <NSwitch v-model:value="modelRef.codeRunnerEnabled" data-testid="code_runner_mode" size="medium" @click.stop />
          </div>

          <!-- Explore Mode -->
          <div
            :class="['mode-card', { enabled: modelRef.exploreMode }]"
            @click="modelRef.exploreMode = !modelRef.exploreMode"
          >
            <div class="mode-header">
              <NIcon :component="ExploreOutlined" :size="24" class="mode-icon" />
              <div class="mode-info">
                <div class="mode-name">{{ $t('chat.exploreMode') }}</div>
                <div class="mode-description">{{ $t('chat.exploreModeDescription') }}</div>
              </div>
            </div>
            <NSwitch v-model:value="modelRef.exploreMode" data-testid="explore_mode" size="medium" @click.stop />
          </div>
        </div>

        <!-- Instructions Panel -->
        <div v-if="showInstructionPanel" class="instructions-section">
          <div class="instructions-header">
            <NIcon :component="SettingsOutlined" size="16" />
            <span>{{ $t('chat.promptInstructions') }}</span>
          </div>
          <div v-if="isInstructionLoading" class="instruction-loading">
            <NSpin size="small" />
            <span>{{ $t('chat.loading_instructions') }}</span>
          </div>
          <template v-else>
            <!-- Artifact Instructions -->
            <div v-if="modelRef.artifactEnabled && artifactInstruction" class="instruction-block">
              <div class="instruction-label">{{ $t('chat.artifactInstructionTitle') }}</div>
              <NInput
                class="instruction-input"
                :value="artifactInstruction"
                type="textarea"
                readonly
                :autosize="{ minRows: 3, maxRows: 10 }"
              />
            </div>
            <!-- Tool Instructions -->
            <div v-if="modelRef.codeRunnerEnabled && toolInstruction" class="instruction-block">
              <div class="instruction-label">{{ $t('chat.toolInstructionTitle') }}</div>
              <NInput
                class="instruction-input"
                :value="toolInstruction"
                type="textarea"
                readonly
                :autosize="{ minRows: 3, maxRows: 10 }"
              />
            </div>
          </template>
        </div>
      </NCollapseItem>

      <!-- Advanced Settings Section -->
      <NCollapseItem name="advanced" class="collapse-item">
        <template #header>
          <div class="collapse-header" data-testid="collapse-advanced">
            <NIcon :component="TuneOutlined" size="18" />
            <span>{{ $t('chat.advanced_settings') }}</span>
          </div>
        </template>
        <div class="advanced-settings">
          <!-- Context Count -->
          <div class="slider-control">
            <div class="slider-header">
              <div class="slider-label-group">
                <NIcon :component="MemoryOutlined" size="16" />
                <span class="slider-label">{{ $t('chat.contextCount', { contextCount: modelRef.contextCount }) }}</span>
              </div>
              <div class="slider-value">{{ modelRef.contextCount }}</div>
            </div>
            <NSlider
              v-model:value="modelRef.contextCount"
              :min="1"
              :max="40"
              :step="1"
              :tooltip="false"
              class="config-slider"
            />
          </div>

          <!-- Temperature -->
          <div class="slider-control">
            <div class="slider-header">
              <div class="slider-label-group">
                <NIcon :component="SpeedOutlined" size="16" />
                <span class="slider-label">{{ $t('chat.temperature') }}</span>
              </div>
              <div class="slider-value">{{ modelRef.temperature.toFixed(2) }}</div>
            </div>
            <NSlider
              v-model:value="modelRef.temperature"
              :min="0.1"
              :max="1"
              :step="0.01"
              :tooltip="false"
              class="config-slider"
            />
          </div>

          <!-- Top P -->
          <div class="slider-control">
            <div class="slider-header">
              <div class="slider-label-group">
                <NIcon :component="TuneOutlined" size="16" />
                <span class="slider-label">{{ $t('chat.topP') }}</span>
              </div>
              <div class="slider-value">{{ modelRef.topP.toFixed(2) }}</div>
            </div>
            <NSlider
              v-model:value="modelRef.topP"
              :min="0"
              :max="1"
              :step="0.01"
              :tooltip="false"
              class="config-slider"
            />
          </div>

          <!-- Max Tokens -->
          <div class="slider-control">
            <div class="slider-header">
              <div class="slider-label-group">
                <NIcon :component="MemoryOutlined" size="16" />
                <span class="slider-label">{{ $t('chat.maxTokens') }}</span>
              </div>
              <div class="slider-value">{{ modelRef.maxTokens }}</div>
            </div>
            <NSlider
              v-model:value="modelRef.maxTokens"
              :min="256"
              :max="tokenUpperLimit"
              :default-value="defaultToken"
              :step="16"
              :tooltip="false"
              class="config-slider"
            />
          </div>

          <!-- N (only for GPT models) -->
          <div v-if="modelRef.chatModel.startsWith('gpt') || modelRef.chatModel.includes('davinci')" class="slider-control">
            <div class="slider-header">
              <div class="slider-label-group">
                <NIcon :component="PsychologyOutlined" size="16" />
                <span class="slider-label">{{ $t('chat.N') }}</span>
              </div>
              <div class="slider-value">{{ modelRef.n }}</div>
            </div>
            <NSlider
              v-model:value="modelRef.n"
              :min="1"
              :max="10"
              :step="1"
              :tooltip="false"
              class="config-slider"
            />
          </div>

          <!-- Debug Mode -->
          <div class="debug-control">
            <div class="debug-header">
              <NIcon :component="BugReportOutlined" size="20" />
              <div class="debug-info">
                <div class="debug-label">{{ $t('chat.debug') }}</div>
                <div class="debug-description">{{ $t('chat.debugDescription') }}</div>
              </div>
            </div>
            <NSwitch v-model:value="modelRef.debug" data-testid="debug_mode" size="medium" />
          </div>

          <!-- Tool Debug Mode -->
          <div v-if="modelRef.codeRunnerEnabled" class="debug-control">
            <div class="debug-header">
              <NIcon :component="SpeedOutlined" size="20" />
              <div class="debug-info">
                <div class="debug-label">{{ $t('chat.toolDebug') }}</div>
                <div class="debug-description">{{ $t('chat.toolDebugDescription') }}</div>
              </div>
            </div>
            <NSwitch v-model:value="modelRef.showToolDebug" data-testid="tool_debug_mode" size="medium" />
          </div>
        </div>
      </NCollapseItem>
    </NCollapse>
  </div>
</template>

<style scoped>
/* Container - Compact */
.session-config-container {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 2px;
}

/* Loading State - Compact */
.loading-container {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px;
  justify-content: center;
}

.loading-text {
  font-size: 13px;
  color: var(--n-text-color-3);
}

/* Ensure radio group doesn't constrain width */
.model-radio-group {
  width: 100%;
  display: block;
}

/* Provider Card - Compact */
.provider-card {
  margin-bottom: 8px;
  padding: 8px;
  background: var(--n-color-modal);
  border-radius: 8px;
  border: 1px solid var(--n-border-color);
  transition: all 0.3s ease;
  width: 100%;
  max-width: 100%;
}

.provider-card:last-child {
  margin-bottom: 0;
}

.provider-card:hover {
  border-color: var(--n-border-color-hover);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
}

:deep(.dark) .provider-card:hover {
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.2);
}

.provider-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.provider-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--n-text-color-1);
  letter-spacing: 0.3px;
  text-transform: uppercase;
}

.provider-count {
  font-size: 11px;
  color: var(--n-text-color-3);
  background: var(--n-color-target);
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}

/* Model Grid - Compact */
.model-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 8px;
  width: 100%;
}

@media (max-width: 480px) {
  .model-grid {
    grid-template-columns: 1fr;
  }
}

@media (min-width: 481px) and (max-width: 768px) {
  .model-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 769px) {
  .model-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1024px) {
  .model-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

/* Model Card - Compact */
.model-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 4px;
  border: 1px solid #e5e7eb;
  background: var(--n-color);
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;
}

:deep(.dark) .model-card {
  border-color: #373737;
}

.model-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: #4b9e5f;
  opacity: 0;
  transition: opacity 0.2s ease;
  pointer-events: none;
}

.model-card:hover {
  border-color: #4b9e5f;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.model-card.active {
  border-color: #4b9e5f;
  background: var(--n-color);
}

.model-card.active::before {
  opacity: 0.08;
}

.model-card.active .model-name {
  color: #4b9e5f;
  font-weight: 600;
}

.model-card.active .model-timestamp {
  color: var(--n-text-color-2);
}

.model-radio {
  pointer-events: none;
}

:deep(.model-radio .n-radio__dot) {
  box-shadow: 0 0 0 2px #e5e7eb;
}

:deep(.dark) .model-radio .n-radio__dot {
  box-shadow: 0 0 0 2px #373737;
}

.model-card.active :deep(.model-radio .n-radio__dot) {
  box-shadow: 0 0 0 2px #4b9e5f;
}

.model-info {
  flex: 1;
  min-width: 0;
}

.model-name {
  font-size: 12px;
  font-weight: 500;
  color: var(--n-text-color-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 1px;
}

.model-meta {
  display: flex;
  align-items: center;
  gap: 6px;
}

.model-timestamp {
  font-size: 10px;
  color: var(--n-text-color-3);
}

/* Collapse - Compact */
.config-collapse {
  border: none;
}

:deep(.config-collapse .n-collapse-item) {
  margin-bottom: 6px;
  border-radius: 10px;
  border: 1px solid var(--n-border-color);
  background: var(--n-color);
  overflow: hidden;
  transition: all 0.3s ease;
}

:deep(.config-collapse .n-collapse-item:hover) {
  border-color: var(--n-border-color-hover);
}

:deep(.config-collapse .n-collapse-item__header) {
  padding: 8px 12px;
  background: var(--n-color-modal);
  border: none;
}

:deep(.config-collapse .n-collapse-item__content-wrapper) {
  padding: 0 !important;
  border: none;
}

:deep(.config-collapse .n-collapse-item .n-collapse-item__content-wrapper .n-collapse-item__content-inner) {
  padding: 6px !important;
}

.collapse-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--n-text-color-1);
}

.collapse-header .n-icon {
  color: var(--n-text-color-2);
}

/* Modes Grid - Compact */
.modes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 10px;
  margin-bottom: 12px;
}

@media (max-width: 768px) {
  .modes-grid {
    grid-template-columns: 1fr;
  }
}

/* Mode Card - Compact */
.mode-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px;
  border-radius: 4px;
  border: 1px solid #e5e7eb;
  background: var(--n-color);
  cursor: pointer;
  transition: all 0.2s ease;
}

:deep(.dark) .mode-card {
  border-color: #373737;
}

.mode-card:hover {
  border-color: #4b9e5f;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.mode-card.enabled {
  border-color: #4b9e5f;
  background: var(--n-color);
}

.mode-header {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  flex: 1;
}

.mode-icon {
  color: var(--n-text-color-2);
  transition: color 0.2s ease;
  flex-shrink: 0;
  font-size: 20px !important;
}

.mode-card.enabled .mode-icon {
  color: #4b9e5f;
}

.mode-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.mode-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--n-text-color-1);
}

.mode-card.enabled .mode-name {
  color: #4b9e5f;
}

.mode-description {
  font-size: 11px;
  color: var(--n-text-color-3);
  line-height: 1.3;
}

/* Instructions Section - Compact */
.instructions-section {
  margin-top: 8px;
  padding: 8px;
  background: var(--n-color-modal);
  border-radius: 8px;
  border: 1px solid var(--n-border-color);
}

.instructions-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  font-size: 12px;
  font-weight: 600;
  color: var(--n-text-color-2);
}

.instruction-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  justify-content: center;
  color: var(--n-text-color-3);
  font-size: 12px;
}

.instruction-block {
  margin-bottom: 8px;
}

.instruction-block:last-child {
  margin-bottom: 0;
}

.instruction-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--n-text-color-2);
  margin-bottom: 6px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.instruction-input :deep(.n-input__textarea) {
  font-family: "SFMono-Regular", Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 11px;
  line-height: 1.5;
}

/* Advanced Settings - Compact */
.advanced-settings {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* Slider Control - Compact */
.slider-control {
  padding: 8px 10px;
  background: var(--n-color-modal);
  border-radius: 8px;
  border: 1px solid var(--n-border-color);
  transition: all 0.2s ease;
}

.slider-control:hover {
  border-color: var(--n-border-color-hover);
}

.slider-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.slider-label-group {
  display: flex;
  align-items: center;
  gap: 6px;
}

.slider-label-group .n-icon {
  color: var(--n-text-color-2);
}

.slider-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--n-text-color-1);
}

.slider-value {
  font-size: 12px;
  font-weight: 600;
  color: var(--n-primary-color);
  background: var(--n-color-target);
  padding: 2px 10px;
  border-radius: 4px;
}

.config-slider {
  margin-top: 2px;
}

:deep(.config-slider .n-slider-rail) {
  height: 4px;
  border-radius: 2px;
}

:deep(.config-slider .n-slider-handle) {
  width: 14px;
  height: 14px;
  border: 2px solid var(--n-primary-color);
}

/* Debug Control - Compact */
.debug-control {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px;
  background: var(--n-color-modal);
  border-radius: 8px;
  border: 1px solid var(--n-border-color);
  transition: all 0.2s ease;
}

.debug-control:hover {
  border-color: var(--n-border-color-hover);
}

.debug-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.debug-header .n-icon {
  color: #f56c6c;
}

.debug-info {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.debug-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--n-text-color-1);
}

.debug-description {
  font-size: 11px;
  color: var(--n-text-color-3);
}
</style>
