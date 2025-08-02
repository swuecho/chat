<script setup lang="ts">
import { computed, ref, h, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NDropdown, NIcon, NText, NTooltip, useMessage } from 'naive-ui'
import type { DropdownOption } from 'naive-ui'
import { SvgIcon } from '@/components/common'
import { useWorkspaceStore } from '@/store/modules/workspace'
import { useSessionStore } from '@/store/modules/session'
import { t } from '@/locales'
import WorkspaceModal from './WorkspaceModal.vue'
import WorkspaceManagementModal from './WorkspaceManagementModal.vue'

const router = useRouter()

const workspaceStore = useWorkspaceStore()
const sessionStore = useSessionStore()
const message = useMessage()

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showManagementModal = ref(false)
const editingWorkspace = ref<Chat.Workspace | null>(null)
const hasTriedAutoLoad = ref(false)

const activeWorkspace = computed(() => workspaceStore.activeWorkspace)
const workspaces = computed(() => workspaceStore.workspaces)

// Watch for when we have an active workspace but few total workspaces - trigger auto-load
watch([activeWorkspace, workspaces], async ([active, spaces]) => {
  if (active && spaces.length === 1 && !hasTriedAutoLoad.value) {
    hasTriedAutoLoad.value = true
    await nextTick()
    try {
      await workspaceStore.loadAllWorkspaces()
    } catch (error) {
      console.error('Failed to auto-load workspaces:', error)
    }
  }
}, { immediate: true })

// Load all workspaces on component mount
onMounted(async () => {
  // Wait a bit for the store to be fully initialized
  await new Promise(resolve => setTimeout(resolve, 100))

  try {
    await workspaceStore.loadAllWorkspaces()
  } catch (error) {
    console.error('Failed to load workspaces on mount:', error)
  }
})

// Load all workspaces when dropdown is opened
async function handleDropdownVisibilityChange(visible: boolean) {
  if (visible && workspaces.value.length <= 1) {
    try {
      await workspaceStore.loadAllWorkspaces()
    } catch (error) {
      console.error('Failed to load workspaces on dropdown open:', error)
    }
  }
}

// Additional trigger for when dropdown is about to show
async function handleBeforeShow() {
  if (workspaces.value.length <= 1) {
    try {
      await workspaceStore.loadAllWorkspaces()
    } catch (error) {
      console.error('Failed to load workspaces before show:', error)
    }
  }
}


// Icon mapping - convert icon value to full icon string
const getWorkspaceIconString = (iconValue: string) => {
  // If already has prefix, return as is
  if (iconValue.includes(':')) {
    return iconValue
  }
  // Otherwise add material-symbols prefix
  return `material-symbols:${iconValue}`
}

const dropdownOptions = computed((): DropdownOption[] => {
  const options = [
    ...workspaces.value.map(workspace => ({
      key: workspace.uuid,
      label: workspace.name,
      icon: () => h(SvgIcon, { icon: getWorkspaceIconString(workspace.icon), style: { color: workspace.color } }),
    })),
    { type: 'divider', key: 'divider1' },
    { key: 'create-workspace', label: t('workspace.create'), icon: () => h(SvgIcon, { icon: 'material-symbols:add' }) },
    { key: 'manage-workspaces', label: t('workspace.manage'), icon: () => h(SvgIcon, { icon: 'material-symbols:settings' }) }
  ]
  return options
})

async function handleDropdownSelect(key: string) {
  console.log('🔄 Dropdown select triggered, key:', key)
  if (key === 'create-workspace') {
    showCreateModal.value = true
  } else if (key === 'manage-workspaces') {
    showManagementModal.value = true
  } else {
    // Switch to selected workspace
    console.log('🔄 Switching to workspace:', key)
    try {
      console.log('🔄 Calling setActiveWorkspace...')
      await workspaceStore.setActiveWorkspace(key)
      console.log('✅ setActiveWorkspace completed')

      console.log('🔄 Navigating to route:', `/workspace/${key}/chat`)
      await router.push(`/workspace/${key}/chat`)
      console.log('✅ Navigation completed')

      message.success('Workspace switched successfully')
    } catch (error) {
      console.error('❌ Error switching workspace:', error)
      message.error('Failed to switch workspace')
    }
  }
}

async function handleWorkspaceCreated(workspace: Chat.Workspace) {
  await workspaceStore.setActiveWorkspace(workspace.uuid)
  await router.push(`/workspace/${workspace.uuid}/chat`)
  message.success(`Created and switched to ${workspace.name}`)
}

function handleWorkspaceUpdated(workspace: Chat.Workspace) {
  message.success(`Updated ${workspace.name}`)
}
</script>

<template>
  <div class="workspace-selector">
    <NDropdown :options="dropdownOptions" trigger="click" placement="bottom-start" @select="handleDropdownSelect"
      class="workspace-dropdown" :width="'trigger'" @update:visible="handleDropdownVisibilityChange"
      @before-show="handleBeforeShow">
      <div class="workspace-button">
        <div class="workspace-icon" :style="{ color: activeWorkspace?.color || '#6366f1' }">
          <SvgIcon v-if="activeWorkspace" :icon="getWorkspaceIconString(activeWorkspace.icon)" />
          <SvgIcon v-else icon="material-symbols:folder" />
        </div>
        <div class="workspace-content">
          <span v-if="activeWorkspace" class="workspace-name">
            {{ activeWorkspace.name }}
          </span>
          <span v-else class="workspace-loading">
            {{ t('workspace.loading') }}
          </span>
        </div>
        <div class="workspace-arrow">
          <SvgIcon icon="material-symbols:expand-more" />
        </div>
      </div>
    </NDropdown>

    <!-- Create Workspace Modal -->
    <WorkspaceModal v-model:visible="showCreateModal" mode="create" @workspace-created="handleWorkspaceCreated" />

    <!-- Edit Workspace Modal -->
    <WorkspaceModal v-model:visible="showEditModal" mode="edit" :workspace="editingWorkspace"
      @workspace-updated="handleWorkspaceUpdated" />

    <!-- Workspace Management Modal -->
    <WorkspaceManagementModal v-model:visible="showManagementModal" />
  </div>
</template>

<style scoped>
.workspace-selector {
  width: 100%;
}

.workspace-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border: 1px solid #e5e5e5;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
}

.workspace-button:hover {
  background-color: rgb(245 245 245);
}

.workspace-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.workspace-content {
  flex: 1;
  overflow: hidden;
}

.workspace-name {
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.workspace-loading {
  color: #a3a3a3;
  font-style: italic;
}

.workspace-arrow {
  font-size: 16px;
  color: #a3a3a3;
  flex-shrink: 0;
}

/* Dark mode */
@media (prefers-color-scheme: dark) {
  .workspace-button {
    border-color: #404040;
  }

  .workspace-button:hover {
    background-color: #24272e;
  }

  .workspace-loading {
    color: #737373;
  }

  .workspace-arrow {
    color: #737373;
  }
}

/* Dropdown styling to match button */
:deep(.n-dropdown-menu) {
  border: 1px solid #e5e5e5;
  border-radius: 4px;
}

:deep(.n-dropdown-option) {
  padding: 8px;
  font-size: 14px;
  gap: 8px;
}

:deep(.n-dropdown-option .n-dropdown-option-body__prefix) {
  font-size: 16px;
}

:deep(.n-dropdown-option:hover) {
  background-color: rgb(245 245 245);
}

@media (prefers-color-scheme: dark) {
  :deep(.n-dropdown-menu) {
    border-color: #404040;
  }

  :deep(.n-dropdown-option:hover) {
    background-color: #24272e;
  }
}
</style>