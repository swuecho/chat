<script lang="ts" setup>
import type { Ref } from 'vue'
import { computed, defineProps, ref, watch } from 'vue'
import type { FormInst } from 'naive-ui'
import { NCard, NForm, NFormItem, NSelect, NSlider, NSwitch } from 'naive-ui'
import { debounce } from 'lodash-es'
import { useChatStore } from '@/store'

const props = defineProps<{
  uuid: string
}>()

const chatStore = useChatStore()

const session = computed(() => chatStore.getChatSessionByUuid(props.uuid))

interface ModelType {
  gptModel: string
  contextLength: number
  temperature: number
  maxTokens: number
  topP: number
  debug: boolean
}

const modelRef: Ref<ModelType> = ref({
  gptModel: session.value?.model ?? 'gpt-3.5-turbo',
  contextLength: session.value?.maxLength ?? 10,
  temperature: session.value?.temperature ?? 1.0,
  maxTokens: session.value?.maxTokens ?? 512,
  topP: session.value?.topP ?? 1.0,
  debug: session.value?.debug ?? false,
})

const formRef = ref<FormInst | null>(null)

const debouneUpdate = debounce(async (modelValue: Ref<ModelType>) => {
  chatStore.updateChatSession(props.uuid, {
    maxLength: modelValue.value.contextLength,
    temperature: modelValue.value.temperature,
    maxTokens: modelValue.value.maxTokens,
    topP: modelValue.value.topP,
    debug: modelValue.value.debug,
    model: modelValue.value.gptModel,
  })
}, 200)

watch(modelRef, async (modelValue: Ref<ModelType>, oldValue: Ref<ModelType>) => {
  console.log('watch')
  console.log('modelValue', modelValue.value)
  debouneUpdate(modelValue)
})

const modelOptions = [
  {
    label: 'gpt-3.5-turbo(chatgpt)',
    value: 'gpt-3.5-turbo',
  },
  {
    label: 'claude-v1 (claude)',
    value: 'claude-v1',
  },
]

</script>

<template>
  <!-- https://platform.openai.com/playground?mode=chat -->
  <NCard id="session-config" :title="$t('chat.sessionConfig')" :bordered="false" size="medium">
    <div>
      <NForm ref="formRef" :model="modelRef" size="small" label-placement="left" :label-width="120">
        <NFormItem :label="$t('chat.model')" path="gptModel">
          <NSelect v-model:value="modelRef.gptModel" :options="modelOptions" />
        </NFormItem>
        <NFormItem :label="$t('chat.contextCount')" path="contextLength">
          <NSlider v-model:value="modelRef.contextLength" :min="1" :max="20" :tooltip="false" show-tooltip />
        </NFormItem>
        <NFormItem :label="$t('chat.temperature')" path="temperature">
          <NSlider v-model:value="modelRef.temperature" :min="0.1" :max="1" :step="0.01" :tooltip="false" />
        </NFormItem>
        <NFormItem :label="$t('chat.topP')" path="topP">
          <NSlider v-model:value="modelRef.topP" :min="0" :max="1" :step="0.01" :tooltip="false" />
        </NFormItem>
        <NFormItem :label="$t('chat.maxTokens')" path="maxTokens">
          <NSlider v-model:value="modelRef.maxTokens" :min="256" :max="2048" :step="16" :tooltip="false" />
        </NFormItem>
        <NFormItem :label="$t('chat.debug')" path="debug">
          <NSwitch v-model:value="modelRef.debug">
            <template #checked>
              {{ $t('chat.enable_debug') }}
            </template>
            <template #unchecked>
              {{ $t('chat.disable_debug') }}
            </template>
          </NSwitch>
        </NFormItem>
      </NForm>
      <pre>{{ JSON.stringify(modelRef, null, 2) }} </pre>
    </div>
  </NCard>
</template>

<style>
#session-config {
  width: 600px
}
</style>
