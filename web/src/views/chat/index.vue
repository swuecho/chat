<script lang='ts' setup>
import { computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import Conversation from './components/Conversation.vue'
import { useChatStore } from '@/store'

interface Props {
  workspaceUuid?: string
  uuid?: string
}

const props = defineProps<Props>()
const route = useRoute()
const chatStore = useChatStore()

// Get parameters from either props (new routing) or route params (legacy)
const workspaceUuid = computed(() => {
  return props.workspaceUuid || (route.params.workspaceUuid as string)
})

const sessionUuid = computed(() => {
  return props.uuid || (route.params.uuid as string)
})

// Set active workspace when workspace is specified in URL
watch(workspaceUuid, (newWorkspaceUuid) => {
  if (newWorkspaceUuid && newWorkspaceUuid !== chatStore.activeWorkspace) {
    console.log('Setting active workspace from URL:', newWorkspaceUuid)
    chatStore.setActiveWorkspace(newWorkspaceUuid)
  }
}, { immediate: true })

// Handle initial workspace setting on mount
onMounted(() => {
  if (workspaceUuid.value && workspaceUuid.value !== chatStore.activeWorkspace) {
    chatStore.setActiveWorkspace(workspaceUuid.value)
  }
})
</script>

<template>
  <div class="h-full flex">
    <Conversation :session-uuid="sessionUuid" />
  </div>
</template>