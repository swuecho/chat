<script setup lang='ts'>
import { computed, watch } from 'vue'
import { NLayout, NLayoutContent } from 'naive-ui'
import { useRouter } from 'vue-router'
import Sider from './sider/index.vue'
import Permission from '@/views/components/Permission.vue'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { useAppStore, useAuthStore, useChatStore } from '@/store'


const router = useRouter()
const appStore = useAppStore()
const chatStore = useChatStore()
const authStore = useAuthStore()

const { isMobile } = useBasicLayout()

const collapsed = computed(() => appStore.siderCollapsed)

// Initialize auth state on component mount
authStore.initializeAuth()

// login modal will appear when there is no token and auth is initialized
const needPermission = computed(() => authStore.isInitialized && !authStore.isValid)

// Set up router after auth is initialized
watch(() => authStore.isInitialized, (initialized) => {
  if (initialized) {
    router.replace({ name: 'Chat', params: { uuid: chatStore.active } })
  }
}, { immediate: true })

// Watch for authentication state changes and sync chat sessions when user logs in
watch(() => authStore.isValid, async (isValid) => {
  console.log('Auth state changed, isValid:', isValid)
  if (isValid && chatStore.history.length === 0) {
    console.log('User is now authenticated and no chat sessions loaded, syncing...')
    try {
      await chatStore.syncChatSessions()
      console.log('Chat sessions synced after auth state change')
    } catch (error) {
      console.error('Failed to sync chat sessions after auth state change:', error)
    }
  }
})

const getMobileClass = computed(() => {
  if (isMobile.value)
    return ['rounded-none', 'shadow-none']
  return ['border', 'rounded-md', 'shadow-md', 'dark:border-neutral-800']
})

const getContainerClass = computed(() => {
  return [
    'h-full',
    'transition-all duration-300 ease-in-out',
    { 'pl-[260px]': !isMobile.value && !collapsed.value },
  ]
})
</script>

<template>
  <div class="h-full dark:bg-[#24272e] transition-all" :class="[isMobile ? 'p-0' : 'p-4']">
    <div class="h-full overflow-hidden" :class="getMobileClass">
      <NLayout class="z-40 transition" :class="getContainerClass" has-sider>
        <Sider />
        <NLayoutContent class="h-full">
          <RouterView v-slot="{ Component, route }">
            <component :is="Component" :key="route.fullPath" />
          </RouterView>
        </NLayoutContent>
      </NLayout>
    </div>
    <Permission :visible="needPermission" />
  </div>
</template>
