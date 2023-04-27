<script lang="ts" setup>
import type { Ref } from 'vue'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import type { FormInst } from 'naive-ui'
import { NForm, NFormItem, NRadio, NRadioGroup, NSlider, NSpace, NSwitch } from 'naive-ui'
import { debounce } from 'lodash-es'
import { useChatStore } from '@/store'
import { fetchChatModel } from '@/api'

const props = defineProps<{
  uuid: string
}>()

const chatStore = useChatStore()

const session = computed(() => chatStore.getChatSessionByUuid(props.uuid))

interface ModelType {
  chatModel: string
  contextCount: number
  temperature: number
  maxTokens: number
  topP: number
  debug: boolean
}

const modelRef: Ref<ModelType> = ref({
  chatModel: session.value?.model ?? 'gpt-3.5-turbo',
  contextCount: session.value?.maxLength ?? 5,
  temperature: session.value?.temperature ?? 1.0,
  maxTokens: session.value?.maxTokens ?? 512,
  topP: session.value?.topP ?? 1.0,
  debug: session.value?.debug ?? false,
})

const formRef = ref<FormInst | null>(null)

const debouneUpdate = debounce(async (model: ModelType) => {
  chatStore.updateChatSession(props.uuid, {
    maxLength: model.contextCount,
    temperature: model.temperature,
    maxTokens: model.maxTokens,
    topP: model.topP,
    debug: model.debug,
    model: model.chatModel,
  })
}, 200)

// why watch not work?, missed the deep = true option
watch(modelRef, async (modelValue: ModelType) => {
  debouneUpdate(modelValue)
}, { deep: true })

const chatModelOptions: any[] = reactive([])

onMounted(async () => {
  const models = await fetchChatModel()
  chatModelOptions.push(...models.map((model: any) => {
    return {
      label: model.Label,
      value: model.Name,
    }
  }))
})

const tokenUpperLimit = computed(() => {
  const discount_ratio = 0.8
  if (modelRef.value.chatModel === 'gpt-4')
    return Math.floor(1024 * 8 * discount_ratio)
  else if (modelRef.value.chatModel === 'gpt-4-32k')
    return Math.floor(32 * 1024 * discount_ratio)
  else
    return Math.floor(1024 * 2 * discount_ratio)
})
// 1. how to fix the NSelect error?
</script>

<template>
  <!-- https://platform.openai.com/playground?mode=chat -->
  <div>
    <NForm ref="formRef" :model="modelRef" size="small" label-placement="top" :label-width="20">
      <NFormItem :label="$t('chat.model')" path="chatModel">
        <NRadioGroup v-model:value="modelRef.chatModel">
          <NSpace>
            <NRadio v-for="model in chatModelOptions" :key="model.value" :value="model.value">
              {{ model.label }}
            </NRadio>
          </NSpace>
        </NRadioGroup>
      </NFormItem>
      <NFormItem :label=" $t(modelRef.chatModel === 'text-davici-003' ? 'chat.contextCount' : 'chat.completionsCount', { contextCount: modelRef.contextCount })" path="contextCount">
        <NSlider v-model:value="modelRef.contextCount" :min="1" :max="20" :tooltip="false" show-tooltip />
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
