<script lang='ts' setup>
import { computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import Conversation from './components/Conversation.vue'
import { useWorkspaceStore } from '@/store/modules/workspace'
import { useSessionStore } from '@/store/modules/session'

interface Props {
  workspaceUuid?: string
  uuid?: string
}

const props = defineProps<Props>()
const route = useRoute()
const workspaceStore = useWorkspaceStore()
const sessionStore = useSessionStore()

// Get parameters from either props (new routing) or route params (legacy)
const workspaceUuid = computed(() => {
  return props.workspaceUuid || (route.params.workspaceUuid as string)
})

const sessionUuid = computed(() => {
  // First try to get sessionUuid from props or route params
  const urlSessionUuid = props.uuid || (route.params.uuid as string)
  if (urlSessionUuid) {
    return urlSessionUuid
  }

  // If no session in URL, use the active session from session store
  return sessionStore.activeSessionUuid || ''
})

// Set active workspace when workspace is specified in URL
watch(workspaceUuid, (newWorkspaceUuid) => {
  if (newWorkspaceUuid && newWorkspaceUuid !== workspaceStore.activeWorkspace?.uuid) {
    console.log('Setting active workspace from URL:', newWorkspaceUuid)
    workspaceStore.setActiveWorkspace(newWorkspaceUuid)
  }
}, { immediate: true })

// Watch for pending session restores and handle them
watch(() => workspaceStore.pendingSessionRestore, (pending) => {
  if (pending) {
    workspaceStore.restoreActiveSession()
  }
})

// Handle initial workspace setting on mount
onMounted(() => {
  if (workspaceUuid.value && workspaceUuid.value !== workspaceStore.activeWorkspace?.uuid) {
    workspaceStore.setActiveWorkspace(workspaceUuid.value)
  }
})
</script>

<template>
  <div class="h-full flex">
    <Conversation v-if="sessionUuid" :session-uuid="sessionUuid" />
    <div v-else class="h-full w-full flex items-center justify-center text-gray-500">
      Loading...
    </div>
  </div>
</template>