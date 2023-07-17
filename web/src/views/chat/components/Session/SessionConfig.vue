<script lang="ts" setup>
import type { Ref } from 'vue'
import { computed, ref, watch } from 'vue'
import type { FormInst } from 'naive-ui'
import { NForm, NFormItem, NRadio, NRadioGroup, NSlider, NSpace, NSpin, NSwitch } from 'naive-ui'
import { debounce } from 'lodash-es'
import { useChatStore } from '@/store'
import { fetchChatModel } from '@/api'

import {  useQuery } from "@tanstack/vue-query";

const optiomFromModel = (model: any) => {
  return {
    label: model.label,
    value: model.name,
  }
}
const props = defineProps<{
  uuid: string
}>()


const { data, isLoading } = useQuery({
  queryKey: ['chat_models'],
  queryFn: fetchChatModel,
  staleTime: 10 * 60 * 1000,
})

const chatModelOptions = computed(() => 
  data?.value ? data.value.map(optiomFromModel) : []
)

const chatStore = useChatStore()

const session = computed(() => chatStore.getChatSessionByUuid(props.uuid))

interface ModelType {
  chatModel: string
  contextCount: number
  temperature: number
  maxTokens: number
  topP: number
  n: number
  debug: boolean
  summarizeMode: boolean
}

const modelRef: Ref<ModelType> = ref({
  chatModel: session.value?.model ?? 'gpt-3.5-turbo',
  summarizeMode: session.value?.summarizeMode ?? false,
  contextCount: session.value?.maxLength ?? 4,
  temperature: session.value?.temperature ?? 1.0,
  maxTokens: session.value?.maxTokens ?? 2048,
  topP: session.value?.topP ?? 1.0,
  n: session.value?.n ?? 1,
  debug: session.value?.debug ?? false,
})

const formRef = ref<FormInst | null>(null)

const debouneUpdate = debounce(async (model: ModelType) => {
  chatStore.updateChatSession(props.uuid, {
    maxLength: model.contextCount,
    temperature: model.temperature,
    maxTokens: model.maxTokens,
    topP: model.topP,
    n: model.n,
    debug: model.debug,
    model: model.chatModel,
  })
}, 200)

// why watch not work?, missed the deep = true option
watch(modelRef, async (modelValue: ModelType) => {
  debouneUpdate(modelValue)
}, { deep: true })



const tokenUpperLimit = computed(() => {
  if (modelRef.value.chatModel === 'gpt-4')
    return Math.floor(1024 * 8)
  else if (modelRef.value.chatModel === 'gpt-4-32k')
    return Math.floor(32 * 1024)
  else if (modelRef.value.chatModel === 'gpt-3.5-turbo')
    return Math.floor(4 * 1024)
  else if (modelRef.value.chatModel === 'text-davinci-003')
    return Math.floor(4 * 1024)
  else if (modelRef.value.chatModel === 'claude-2')
    return Math.floor(100 * 1024)
  else if (modelRef.value.chatModel === 'gpt-3.5-turbo-16k')
    return Math.floor(16 * 1024)
  else
    return Math.floor(1024 * 2)
})
// 1. how to fix the NSelect error?
</script>

<template>
  <!-- https://platform.openai.com/playground?mode=chat -->
  <div>
    <NForm ref="formRef" :model="modelRef" size="small" label-placement="top" :label-width="20">
      <NFormItem :label="$t('chat.model')" path="chatModel">
        <div v-if="isLoading"><NSpin size="medium" /></div>
        <NRadioGroup v-model:value="modelRef.chatModel">
          <NSpace>
            <NRadio v-for="model in chatModelOptions" :key="model.value" :value="model.value">
              {{ model.label }}
            </NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>
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
        <NSlider v-model:value="modelRef.maxTokens" :min="256" :max="tokenUpperLimit" :step="16" :tooltip="false" />
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
    </NForm>
    <!--
                                        <div class="center">
                                          <pre>{{ JSON.stringify(modelRef, null, 2) }} </pre>
                                        </div>
                                        -->
  </div>
</template>
