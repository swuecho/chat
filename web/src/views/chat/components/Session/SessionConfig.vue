<script lang="ts" setup>
import type { Ref } from 'vue'
import { computed, ref, watch, h } from 'vue'
import type { FormInst } from 'naive-ui'
import { NForm, NFormItem, NRadio, NRadioGroup, NSlider, NSpace, NSpin, NSwitch } from 'naive-ui'
import { debounce } from 'lodash-es'
import { useSessionStore, useAppStore } from '@/store'
import { fetchChatModel } from '@/api'

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
})

const formRef = ref<FormInst | null>(null)

const debouneUpdate = debounce(async (model: ModelType) => {
  sessionStore.updateSession(props.uuid, {
    maxLength: model.contextCount,
    temperature: model.temperature,
    maxTokens: model.maxTokens,
    topP: model.topP,
    n: model.n,
    debug: model.debug,
    model: model.chatModel,
    summarizeMode: model.summarizeMode,
    exploreMode: model.exploreMode,
  })
}, 200)

// why watch not work?, missed the deep = true option
watch(modelRef, async (modelValue: ModelType) => {
  debouneUpdate(modelValue)
}, { deep: true })



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
        <NRadioGroup v-model:value="modelRef.chatModel">
          <NSpace>
            <NRadio v-for="model in chatModelOptions" :key="model.name" :value="model.name">
              <div>
                {{ model.label }}
                <span style="color: #999; font-size: 0.8rem; margin-left: 4px">
                  - {{ formatTimestamp(model.lastUsageTime) }}
                </span>
              </div>
            </NRadio>
          </NSpace>
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
    </NForm>
    <!--
                                        <div class="center">
                                          <pre>{{ JSON.stringify(modelRef, null, 2) }} </pre>
                                        </div>
                                        -->
  </div>
</template>
