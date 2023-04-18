<script lang="ts" setup>
import { NAvatar } from 'naive-ui'
import ModelAvatar from './ModelAvatar.vue'
import defaultAvatar from '@/assets/avatar.jpg'

interface Props {
  inversion?: boolean
  model?: string
}

defineProps<Props>()
</script>

<template>
  <template v-if="inversion">
    <NAvatar round :src="defaultAvatar" />
  </template>
  <span v-else class="text-[28px] dark:text-white">
    <ModelAvatar :model="model" />
  </span>
</template>
