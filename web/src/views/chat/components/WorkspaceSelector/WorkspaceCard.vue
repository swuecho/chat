<script setup lang="ts">
import { computed } from 'vue'
import {
  NCard,
  NButton,
  NDropdown,
  NTooltip,
  NBadge,
  NTag,
  useMessage
} from 'naive-ui'
import type { DropdownOption } from 'naive-ui'
import { SvgIcon } from '@/components/common'
import { useChatStore } from '@/store'
import { t } from '@/locales'

interface Props {
  workspace: Chat.Workspace
  dragMode?: boolean
}

interface Emits {
  (e: 'edit', workspace: Chat.Workspace): void
  (e: 'delete', workspace: Chat.Workspace): void
  (e: 'duplicate', workspace: Chat.Workspace): void
  (e: 'set-default', workspace: Chat.Workspace): void
}

const props = withDefaults(defineProps<Props>(), {
  dragMode: false
})
const emit = defineEmits<Emits>()

const chatStore = useChatStore()
const message = useMessage()

// Icon mapping - convert icon value to full icon string
const getWorkspaceIconString = (iconValue: string) => {
  if (iconValue.includes(':')) {
    return iconValue
  }
  return `material-symbols:${iconValue}`
}

const sessionCount = computed(() => {
  return chatStore.getSessionsByWorkspace(props.workspace.uuid).length
})

const isActive = computed(() => {
  return chatStore.activeWorkspace === props.workspace.uuid
})

const dropdownOptions = computed((): DropdownOption[] => [
  {
    key: 'edit',
    label: t('common.edit'),
    icon: () => h(SvgIcon, { icon: 'material-symbols:edit' })
  },
  {
    key: 'duplicate',
    label: t('workspace.duplicate'),
    icon: () => h(SvgIcon, { icon: 'material-symbols:content-copy' })
  },
  {
    key: 'set-default',
    label: t('workspace.setAsDefault'),
    icon: () => h(SvgIcon, { icon: 'material-symbols:star' }),
    disabled: props.workspace.isDefault
  },
  {
    type: 'divider',
    key: 'divider'
  },
  {
    key: 'delete',
    label: t('common.delete'),
    icon: () => h(SvgIcon, { icon: 'material-symbols:delete' }),
    disabled: props.workspace.isDefault,
    props: {
      style: 'color: #ef4444;'
    }
  }
])

function handleDropdownSelect(key: string) {
  switch (key) {
    case 'edit':
      emit('edit', props.workspace)
      break
    case 'delete':
      if (props.workspace.isDefault) {
        message.warning(t('workspace.cannotDeleteDefault'))
        return
      }
      emit('delete', props.workspace)
      break
    case 'duplicate':
      emit('duplicate', props.workspace)
      break
    case 'set-default':
      emit('set-default', props.workspace)
      break
  }
}

async function handleSwitchToWorkspace() {
  if (isActive.value) return
  
  try {
    await chatStore.switchToWorkspace(props.workspace.uuid)
    message.success(t('workspace.switchedTo', { name: props.workspace.name }))
  } catch (error) {
    console.error('Failed to switch workspace:', error)
    message.error(t('workspace.switchError'))
  }
}

// Import h function for rendering icons in dropdown
import { h } from 'vue'
</script>

<template>
  <NCard 
    class="workspace-card" 
    :class="{ 
      'workspace-card--active': isActive,
      'workspace-card--drag-mode': dragMode
    }"
    size="small"
    :hoverable="!dragMode"
  >
    <!-- Header with icon and actions -->
    <div class="workspace-card__header">
      <div class="workspace-card__icon-container">
        <div 
          class="workspace-card__icon"
          :style="{ color: workspace.color }"
        >
          <SvgIcon :icon="getWorkspaceIconString(workspace.icon)" />
        </div>
        <NBadge v-if="isActive" :value="t('workspace.active')" type="success" />
        <NBadge v-else-if="workspace.isDefault" :value="t('workspace.default')" type="info" />
      </div>
      
      <div class="workspace-card__actions">
        <div 
          v-if="dragMode" 
          class="workspace-card__drag-handle"
          :title="t('workspace.dragToReorder')"
        >
          <SvgIcon icon="material-symbols:drag-indicator" />
        </div>
        
        <NDropdown
          v-if="!dragMode"
          :options="dropdownOptions"
          trigger="click"
          placement="bottom-end"
          @select="handleDropdownSelect"
        >
          <NButton quaternary circle size="small" class="workspace-card__menu">
            <template #icon>
              <SvgIcon icon="material-symbols:more-vert" />
            </template>
          </NButton>
        </NDropdown>
      </div>
    </div>

    <!-- Workspace content -->
    <div class="workspace-card__content" @click="handleSwitchToWorkspace">
      <div class="workspace-card__title">
        <h3 class="workspace-card__name">{{ workspace.name }}</h3>
        <div class="workspace-card__meta">
          <span class="workspace-card__session-count">
            {{ t('workspace.sessionCount', { count: sessionCount }) }}
          </span>
        </div>
      </div>

      <div v-if="workspace.description" class="workspace-card__description">
        {{ workspace.description }}
      </div>

      <div class="workspace-card__footer">
        <div class="workspace-card__tags">
          <NTag v-if="workspace.isDefault" size="small" type="primary">
            {{ t('workspace.default') }}
          </NTag>
          <NTag v-if="isActive" size="small" type="success">
            {{ t('workspace.active') }}
          </NTag>
        </div>
        
        <div class="workspace-card__date">
          {{ t('workspace.lastUpdated') }}: {{ new Date(workspace.updatedAt).toLocaleDateString() }}
        </div>
      </div>
    </div>
  </NCard>
</template>

<style scoped>
.workspace-card {
  height: 200px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 2px solid transparent;
  position: relative;
}

.workspace-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.workspace-card--active {
  border-color: #18a058;
  background: linear-gradient(135deg, rgba(24, 160, 88, 0.05) 0%, rgba(24, 160, 88, 0.02) 100%);
}

.workspace-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.workspace-card__icon-container {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
}

.workspace-card__icon {
  font-size: 24px;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.05);
}

.workspace-card__menu {
  opacity: 0;
  transition: opacity 0.2s ease;
}

.workspace-card:hover .workspace-card__menu {
  opacity: 1;
}

.workspace-card--drag-mode {
  cursor: grab;
  user-select: none;
}

.workspace-card--drag-mode:active {
  cursor: grabbing;
}

.workspace-card__actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.workspace-card__drag-handle {
  font-size: 18px;
  color: var(--n-text-color-disabled);
  cursor: grab;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.workspace-card__drag-handle:hover {
  color: var(--n-text-color);
  background: var(--n-color-hover);
}

.workspace-card__drag-handle:active {
  cursor: grabbing;
}

.workspace-card__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: calc(100% - 52px);
}

.workspace-card__title {
  margin-bottom: 8px;
}

.workspace-card__name {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: var(--n-text-color);
  line-height: 1.2;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.workspace-card__meta {
  margin-top: 4px;
}

.workspace-card__session-count {
  font-size: 12px;
  color: var(--n-text-color-disabled);
}

.workspace-card__description {
  font-size: 14px;
  color: var(--n-text-color-disabled);
  line-height: 1.4;
  flex: 1;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 12px;
}

.workspace-card__footer {
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid var(--n-divider-color);
}

.workspace-card__tags {
  display: flex;
  gap: 4px;
  margin-bottom: 4px;
}

.workspace-card__date {
  font-size: 11px;
  color: var(--n-text-color-disabled);
}

/* Dark mode adjustments */
@media (prefers-color-scheme: dark) {
  .workspace-card__icon {
    background: rgba(255, 255, 255, 0.05);
  }
  
  .workspace-card:hover {
    box-shadow: 0 8px 25px rgba(0, 0, 0, 0.3);
  }
  
  .workspace-card--active {
    background: linear-gradient(135deg, rgba(24, 160, 88, 0.1) 0%, rgba(24, 160, 88, 0.05) 100%);
  }
}
</style>