<script setup lang='ts'>
import { NAvatar } from 'naive-ui'
import { computed } from 'vue'
import defaultAvatar from '@/assets/avatar.jpg'
import { useUserStore } from '@/store'
import { isString } from '@/utils/is'
import { t } from '@/locales'

const userStore = useUserStore()
const userInfo = computed(() => userStore.userInfo)
</script>

<template>
  <div class="flex items-center overflow-hidden">
    <div class="w-10 h-10 overflow-hidden rounded-full shrink-0">
      <NAvatar size="large" round :src="defaultAvatar" />
    </div>
    <div class="flex-1 min-w-0 ml-2">
      <h2 class="overflow-hidden font-bold text-md text-ellipsis whitespace-nowrap">
        {{ userInfo.name || $t('setting.defaultName') }}
      </h2>
      <p class="overflow-hidden text-xs text-gray-500 text-ellipsis whitespace-nowrap">
        <span
          v-if="isString(userInfo.description)"
          v-html="userInfo.description || t('setting.defaultDesc')"
        />
      </p>
    </div>
  </div>
</template>
