<script setup lang="ts">
import { computed, ref, h } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NDropdown, NIcon, NText, NTooltip, useMessage } from 'naive-ui'
import type { DropdownOption } from 'naive-ui'
import { SvgIcon } from '@/components/common'
import { useChatStore } from '@/store'
import { t } from '@/locales'
import WorkspaceModal from './WorkspaceModal.vue'

const router = useRouter()

const chatStore = useChatStore()
const message = useMessage()

const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingWorkspace = ref<Chat.Workspace | null>(null)

const activeWorkspace = computed(() => chatStore.getWorkspaceByUuid(chatStore.activeWorkspace))
const workspaces = computed(() => chatStore.workspaces)

// Icon mapping - convert icon value to full icon string
const getWorkspaceIconString = (iconValue: string) => {
  // If already has prefix, return as is
  if (iconValue.includes(':')) {
    return iconValue
  }
  // Otherwise add material-symbols prefix
  return `material-symbols:${iconValue}`
}

const dropdownOptions = computed((): DropdownOption[] => [
  ...workspaces.value.map(workspace => ({
    key: workspace.uuid,
    label: workspace.name,
    icon: () => h(SvgIcon, { 
      icon: getWorkspaceIconString(workspace.icon), 
      style: { color: workspace.color } 
    }),
  })),
  {
    type: 'divider',
    key: 'divider1'
  },
  {
    key: 'create-workspace',
    label: t('workspace.create'),
    icon: () => h(SvgIcon, { icon: 'material-symbols:add' }),
  },
  {
    key: 'manage-workspaces',
    label: t('workspace.manage'),
    icon: () => h(SvgIcon, { icon: 'material-symbols:settings' }),
  }
])

async function handleDropdownSelect(key: string) {
  if (key === 'create-workspace') {
    showCreateModal.value = true
    return
  }
  
  if (key === 'manage-workspaces') {
    // TODO: Open workspace management modal
    message.info('Workspace management coming soon!')
    return
  }
  
  // Switch to selected workspace
  if (key !== chatStore.activeWorkspace) {
    const workspace = workspaces.value.find(w => w.uuid === key)
    if (workspace) {
      await chatStore.switchToWorkspace(key)
      message.success(`Switched to ${workspace.name}`)
    }
  }
}

async function handleWorkspaceCreated(workspace: Chat.Workspace) {
  await chatStore.switchToWorkspace(workspace.uuid)
  message.success(`Created and switched to ${workspace.name}`)
}

function handleWorkspaceUpdated(workspace: Chat.Workspace) {
  message.success(`Updated ${workspace.name}`)
}
</script>

<template>
  <div class="workspace-selector">
    <NDropdown
      :options="dropdownOptions"
      trigger="click"
      placement="bottom-start"
      @select="handleDropdownSelect"
    >
      <NButton
        :disabled="!activeWorkspace"
        class="w-full justify-start"
        :style="{ 
          borderColor: activeWorkspace?.color, 
          color: activeWorkspace?.color 
        }"
      >
        <template #icon>
          <NIcon v-if="activeWorkspace" :style="{ color: activeWorkspace.color }">
            <SvgIcon :icon="getWorkspaceIconString(activeWorkspace.icon)" />
          </NIcon>
        </template>
        <div class="flex-1 text-left truncate">
          <NText v-if="activeWorkspace" :style="{ color: activeWorkspace.color }">
            {{ activeWorkspace.name }}
          </NText>
          <NText v-else depth="3">
            {{ t('workspace.loading') }}
          </NText>
        </div>
        <template #suffix>
          <NIcon class="ml-2">
            <SvgIcon icon="material-symbols:expand-more" />
          </NIcon>
        </template>
      </NButton>
    </NDropdown>

    <!-- Create Workspace Modal -->
    <WorkspaceModal
      v-model:visible="showCreateModal"
      mode="create"
      @workspace-created="handleWorkspaceCreated"
    />

    <!-- Edit Workspace Modal -->
    <WorkspaceModal
      v-model:visible="showEditModal"
      mode="edit"
      :workspace="editingWorkspace"
      @workspace-updated="handleWorkspaceUpdated"
    />
  </div>
</template>

<style scoped>
.workspace-selector {
  width: 100%;
}
</style>