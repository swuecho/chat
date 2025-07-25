<script setup lang="ts">
import { computed, ref, h } from 'vue'
import {
  NModal,
  NCard,
  NButton,
  NSpace,
  NInput,
  NGrid,
  NGridItem,
  NEmpty,
  NIcon,
  NSwitch,
  useMessage
} from 'naive-ui'
import { SvgIcon } from '@/components/common'
import { useChatStore } from '@/store'
import { t } from '@/locales'
import WorkspaceCard from './WorkspaceCard.vue'
import WorkspaceModal from './WorkspaceModal.vue'

interface Props {
  visible: boolean
}

interface Emits {
  (e: 'update:visible', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const chatStore = useChatStore()
const message = useMessage()

const searchQuery = ref('')
const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingWorkspace = ref<Chat.Workspace | null>(null)
const dragMode = ref(false)
const draggedWorkspace = ref<Chat.Workspace | null>(null)
const dragOverIndex = ref<number | null>(null)

const isVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const workspaces = computed(() => chatStore.workspaces)

const filteredWorkspaces = computed(() => {
  if (!searchQuery.value.trim()) {
    return workspaces.value
  }
  
  const query = searchQuery.value.toLowerCase()
  return workspaces.value.filter(workspace => 
    workspace.name.toLowerCase().includes(query) ||
    workspace.description?.toLowerCase().includes(query)
  )
})

function handleClose() {
  isVisible.value = false
  searchQuery.value = ''
}

function handleCreateWorkspace() {
  showCreateModal.value = true
}

function handleEditWorkspace(workspace: Chat.Workspace) {
  editingWorkspace.value = workspace
  showEditModal.value = true
}

function handleDeleteWorkspace(workspace: Chat.Workspace) {
  // TODO: Implement delete functionality
  message.info(`Delete ${workspace.name} - Feature coming soon!`)
}

function handleDuplicateWorkspace(workspace: Chat.Workspace) {
  // TODO: Implement duplicate functionality
  message.info(`Duplicate ${workspace.name} - Feature coming soon!`)
}

function handleSetDefaultWorkspace(workspace: Chat.Workspace) {
  // TODO: Implement set default functionality
  message.info(`Set ${workspace.name} as default - Feature coming soon!`)
}

async function handleWorkspaceCreated(workspace: Chat.Workspace) {
  message.success(`Created workspace: ${workspace.name}`)
  showCreateModal.value = false
}

function handleWorkspaceUpdated(workspace: Chat.Workspace) {
  message.success(`Updated workspace: ${workspace.name}`)
  showEditModal.value = false
  editingWorkspace.value = null
}

// Drag and drop handlers
function handleDragStart(event: DragEvent, workspace: Chat.Workspace, index: number) {
  if (!dragMode.value) return
  
  draggedWorkspace.value = workspace
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', workspace.uuid)
  }
}

function handleDragOver(event: DragEvent, index: number) {
  if (!dragMode.value) return
  
  event.preventDefault()
  dragOverIndex.value = index
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
}

function handleDragLeave() {
  if (!dragMode.value) return
  dragOverIndex.value = null
}

function handleDrop(event: DragEvent, targetIndex: number) {
  if (!dragMode.value || !draggedWorkspace.value) return
  
  event.preventDefault()
  
  const draggedIndex = filteredWorkspaces.value.findIndex(w => w.uuid === draggedWorkspace.value!.uuid)
  
  if (draggedIndex !== -1 && draggedIndex !== targetIndex) {
    // Reorder workspaces
    reorderWorkspaces(draggedIndex, targetIndex)
  }
  
  // Reset drag state
  draggedWorkspace.value = null
  dragOverIndex.value = null
}

function handleDragEnd() {
  draggedWorkspace.value = null
  dragOverIndex.value = null
}

async function reorderWorkspaces(fromIndex: number, toIndex: number) {
  try {
    console.log(`🔄 Reordering workspace from ${fromIndex} to ${toIndex}`)
    
    // Create a new array with reordered workspaces
    const reorderedWorkspaces = [...filteredWorkspaces.value]
    const [draggedItem] = reorderedWorkspaces.splice(fromIndex, 1)
    reorderedWorkspaces.splice(toIndex, 0, draggedItem)
    
    console.log('📋 New order:', reorderedWorkspaces.map((w, i) => `${i}: ${w.name}`))
    
    // Update order positions for each workspace
    const updatePromises = reorderedWorkspaces.map(async (workspace, index) => {
      const currentOrder = workspace.orderPosition || 0
      if (currentOrder !== index) {
        console.log(`📝 Updating ${workspace.name}: ${currentOrder} → ${index}`)
        try {
          await chatStore.updateWorkspaceOrder(workspace.uuid, index)
        } catch (error) {
          console.error(`❌ Failed to update order for workspace ${workspace.name}:`, error)
          throw error
        }
      }
    })
    
    await Promise.all(updatePromises)
    
    // Refresh workspaces to get the updated order from backend
    await chatStore.syncWorkspaces()
    
    message.success(t('workspace.reorderSuccess'))
    console.log('✅ Workspace reordering completed')
  } catch (error) {
    console.error('❌ Failed to reorder workspaces:', error)
    message.error(t('workspace.reorderError'))
  }
}
</script>

<template>
  <NModal v-model:show="isVisible" :mask-closable="false" class="workspace-management-modal">
    <NCard 
      :title="t('workspace.manage')" 
      class="w-full max-w-5xl" 
      :bordered="false" 
      size="small" 
      role="dialog" 
      aria-modal="true"
    >
      <template #header-extra>
        <NButton quaternary circle @click="handleClose">
          <template #icon>
            <SvgIcon icon="material-symbols:close" />
          </template>
        </NButton>
      </template>

      <!-- Header Actions -->
      <div class="mb-6 space-y-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <NInput
              v-model:value="searchQuery"
              :placeholder="t('workspace.searchPlaceholder')"
              clearable
              class="w-64"
            >
              <template #prefix>
                <SvgIcon icon="material-symbols:search" />
              </template>
            </NInput>
            
            <div class="flex items-center gap-2">
              <SvgIcon icon="material-symbols:drag-indicator" class="text-sm" />
              <span class="text-sm">{{ t('workspace.reorderMode') }}</span>
              <NSwitch 
                v-model:value="dragMode" 
                size="small"
                :round="false"
              />
            </div>
          </div>
          
          <NButton type="primary" @click="handleCreateWorkspace">
            <template #icon>
              <SvgIcon icon="material-symbols:add" />
            </template>
            {{ t('workspace.create') }}
          </NButton>
        </div>

        <!-- Summary Info -->
        <div class="text-sm text-gray-600 dark:text-gray-400">
          {{ t('workspace.totalCount', { count: filteredWorkspaces.length }) }}
          <span v-if="searchQuery.trim()" class="ml-2">
            ({{ t('workspace.filteredResults', { total: workspaces.length }) }})
          </span>
        </div>
      </div>

      <!-- Workspace Grid -->
      <div class="workspace-grid">
        <NEmpty v-if="filteredWorkspaces.length === 0" :description="t('workspace.noWorkspaces')">
          <template #icon>
            <SvgIcon icon="material-symbols:folder-off" class="text-4xl" />
          </template>
          <template #extra>
            <NButton type="primary" @click="handleCreateWorkspace">
              {{ t('workspace.createFirst') }}
            </NButton>
          </template>
        </NEmpty>

        <NGrid v-else :cols="3" :x-gap="16" :y-gap="16" responsive="screen">
          <NGridItem 
            v-for="(workspace, index) in filteredWorkspaces" 
            :key="workspace.uuid"
            class="workspace-grid-item"
            :class="{ 
              'workspace-grid-item--drag-mode': dragMode,
              'workspace-grid-item--drag-over': dragOverIndex === index,
              'workspace-grid-item--dragging': draggedWorkspace?.uuid === workspace.uuid
            }"
            :draggable="dragMode"
            @dragstart="handleDragStart($event, workspace, index)"
            @dragover="handleDragOver($event, index)"
            @dragleave="handleDragLeave"
            @drop="handleDrop($event, index)"
            @dragend="handleDragEnd"
          >
            <WorkspaceCard
              :workspace="workspace"
              :drag-mode="dragMode"
              @edit="handleEditWorkspace"
              @delete="handleDeleteWorkspace"
              @duplicate="handleDuplicateWorkspace"
              @set-default="handleSetDefaultWorkspace"
            />
          </NGridItem>
        </NGrid>
      </div>

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
    </NCard>
  </NModal>
</template>

<style scoped>
.workspace-management-modal {
  --n-bezier: cubic-bezier(0.4, 0, 0.2, 1);
}

.workspace-grid {
  min-height: 400px;
}

/* Drag and drop styles */
.workspace-grid-item {
  transition: all 0.2s ease;
}

.workspace-grid-item--drag-mode {
  cursor: grab;
}

.workspace-grid-item--drag-mode:active {
  cursor: grabbing;
}

.workspace-grid-item--dragging {
  opacity: 0.5;
  transform: scale(0.95);
}

.workspace-grid-item--drag-over {
  transform: scale(1.02);
  background: rgba(24, 160, 88, 0.1);
  border-radius: 8px;
  border: 2px dashed #18a058;
}

/* Responsive grid adjustments */
@media (max-width: 1024px) {
  .workspace-grid :deep(.n-grid) {
    grid-template-columns: repeat(2, 1fr) !important;
  }
}

@media (max-width: 640px) {
  .workspace-grid :deep(.n-grid) {
    grid-template-columns: 1fr !important;
  }
}
</style>