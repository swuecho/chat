<script setup lang='ts'>
import { NAvatar } from 'naive-ui'
import defaultAvatar from '@/assets/avatar.jpg'
</script>

<template>
  <div class="flex items-center overflow-hidden">
    <div class="w-10 h-10 overflow-hidden rounded-full shrink-0">
      <NAvatar size="large" round :src="defaultAvatar" />
    </div>
  </div>
</template>
