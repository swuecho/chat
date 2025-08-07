<script setup lang="ts">
import { defineComponent, h, onMounted, onUnmounted } from 'vue'
import {
  NDialogProvider,
  NLoadingBarProvider,
  NMessageProvider,
  NNotificationProvider,
  useDialog,
  useLoadingBar,
  useMessage,
  useNotification,
} from 'naive-ui'
import { notificationManager } from '@/utils/notificationManager'

function registerNaiveTools() {
  window.$loadingBar = useLoadingBar()
  window.$dialog = useDialog()
  window.$message = useMessage()
  window.$notification = useNotification()
  
  // Initialize notification manager
  notificationManager.setMessageInstance(window.$message)
}

// Handle online/offline status
function handleNetworkStatus() {
  if (!navigator.onLine) {
    window.$message?.error('You are offline. Please check your internet connection.', {
      duration: 0,
      closable: true,
      action: {
        text: 'Retry',
        onClick: () => window.location.reload()
      }
    })
  }
}

onMounted(() => {
  window.addEventListener('online', () => {
    window.$message?.success('You are back online!', { duration: 3000 })
  })
  
  window.addEventListener('offline', handleNetworkStatus)
  
  // Check initial network status
  handleNetworkStatus()
})

onUnmounted(() => {
  window.removeEventListener('online', () => {})
  window.removeEventListener('offline', handleNetworkStatus)
})

const NaiveProviderContent = defineComponent({
  name: 'NaiveProviderContent',
  setup() {
    registerNaiveTools()
  },
  render() {
    return h('div')
  },
})
</script>

<template>
  <NLoadingBarProvider>
    <NDialogProvider>
      <NNotificationProvider :max="5" :placement="'top-right'">
        <NMessageProvider 
          :placement="'top-right'" 
          :max="3"
          :duration="5000"
          :closable="true"
        >
          <slot />
          <NaiveProviderContent />
        </NMessageProvider>
      </NNotificationProvider>
    </NDialogProvider>
  </NLoadingBarProvider>
</template>
