<script setup lang="ts">
import { defineEmits, defineProps, ref, watch } from 'vue'
import { NCard, NInputNumber, NSlider } from 'naive-ui'

const props = defineProps({
  sliderValue: {
    type: Number,
    required: true,
  },
})

const emits = defineEmits(['update:sliderValue'])

const value = ref(props.sliderValue)

watch(value, (newValue) => {
  emits('update:sliderValue', newValue)
})
</script>

<template>
  <NCard style="width: 600px" title="会话设置" :bordered="false" size="huge" role="dialog" aria-modal="true">
    <div>{{ $t('chat.contextLength') }}</div>
    <NSlider v-model:value="value" :min="1" :max="20" :tooltip="false" />
    <NInputNumber v-model:value="value" size="small" />
  </NCard>
</template>
