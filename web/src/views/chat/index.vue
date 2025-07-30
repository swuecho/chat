<script lang='ts' setup>
import { computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import Conversation from './components/Conversation.vue'
import { useWorkspaceStore } from '@/store/modules/workspace'

interface Props {
  workspaceUuid?: string
  uuid?: string
}

const props = defineProps<Props>()
const route = useRoute()
const workspaceStore = useWorkspaceStore()

// Get parameters from either props (new routing) or route params (legacy)
const workspaceUuid = computed(() => {
  return props.workspaceUuid || (route.params.workspaceUuid as string)
})

const sessionUuid = computed(() => {
  return props.uuid || (route.params.uuid as string)
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
    <Conversation :session-uuid="sessionUuid" />
  </div>
</template>