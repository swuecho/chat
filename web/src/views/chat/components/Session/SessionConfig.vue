<script lang="ts" setup>
import { computed, defineProps, ref, watch } from 'vue'
import { NCard, NInputNumber, NSlider, NSwitch } from 'naive-ui'
import { debounce } from 'lodash'
import { useChatStore } from '@/store'

const props = defineProps<{
  uuid: string
}>()

const chatStore = useChatStore()

const session = computed(() => chatStore.getChatSessionByUuid(props.uuid))

const slider = ref(session.value?.maxLength ?? 10)

const temperature = ref(session.value?.temperature ?? 1.0)
const maxTokens = ref(session.value?.maxTokens ?? 512)
const topP = ref(session.value?.topP ?? 1.0)
const debug = ref(session.value?.debug ?? false)
// const frequencyPenalty = ref(0)
// const presencePenalty = ref(0)

const debouneUpdate = debounce(async ([newValueSlider, newValueTemperature, newMaxTokens, topP, debug]: Array<any>) => {
  chatStore.updateChatSession(props.uuid, {
    maxLength: newValueSlider,
    temperature: newValueTemperature,
    maxTokens: newMaxTokens,
    topP,
    debug,
  })
}, 200)

watch([slider, temperature, maxTokens, topP, debug], ([newValueSlider, newValueTemperature, newMaxTokens, topP, debug], _) => {
  debouneUpdate([newValueSlider, newValueTemperature, newMaxTokens, topP, debug])
})
</script>

<template>
  <NCard style="width: 600px" title="会话设置" :bordered="false" size="huge" role="dialog" aria-modal="true">
    <div>{{ $t('chat.slider') }}</div>
    <NSlider v-model:value="slider" :min="1" :max="20" :tooltip="false" />
    <NInputNumber v-model:value="slider" size="small" />

    <div>{{ $t('chat.temperature') }}</div>
    <NSlider v-model:value="temperature" :min="0.1" :max="2" :step="0.1" :tooltip="false" />
    <NInputNumber v-model:value="temperature" size="small" />

    <div>{{ $t('chat.topP') }}</div>
    <NSlider v-model:value="topP" :min="0" :max="1" :step="0.1" :tooltip="false" />
    <NInputNumber v-model:value="topP" size="small" />

    <div>{{ $t('chat.maxTokens') }}</div>
    <NSlider v-model:value="maxTokens" :min="256" :max="4096" :step="10" :tooltip="false" />
    <NInputNumber v-model:value="maxTokens" size="small" />
    <div> {{ $t('chat.debug') }}</div>
    <NSwitch v-model:value="debug">
      <template #checked>
        启用
      </template>
      <template #unchecked>
        关闭
      </template>
    </NSwitch>
    <!--
    <div>{{ $t('chat.presencePenalty') }}</div>
    <NSlider v-model:value="presencePenalty" :min="-2" :max="2" :step="0.1" :tooltip="false" />
    <NInputNumber v-model:value="presencePenalty" size="small" />
    <div>{{ $t('chat.frequencyPenalty') }}</div>
    <NSlider v-model:value="frequencyPenalty" :min="-2" :max="2" :step="0.1" :tooltip="false" />
    <NInputNumber v-model:value="frequencyPenalty" size="small" />
     -->
  </NCard>
</template>
