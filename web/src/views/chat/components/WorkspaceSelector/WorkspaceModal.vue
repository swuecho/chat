<script setup lang="ts">
import { computed, ref, watch, h } from 'vue'
import {
  NModal,
  NCard,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NSpace,
  NColorPicker,
  NSelect,
  useMessage
} from 'naive-ui'
import { SvgIcon } from '@/components/common'
import { useWorkspaceStore } from '@/store/modules/workspace'
import { t } from '@/locales'
import type { CreateWorkspaceRequest, UpdateWorkspaceRequest } from '@/api'

interface Props {
  visible: boolean
  mode: 'create' | 'edit'
  workspace?: Chat.Workspace | null
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'workspace-created', workspace: Chat.Workspace): void
  (e: 'workspace-updated', workspace: Chat.Workspace): void
}

const props = withDefaults(defineProps<Props>(), {
  workspace: null
})

const emit = defineEmits<Emits>()

const workspaceStore = useWorkspaceStore()
const message = useMessage()

const loading = ref(false)

// Form data
const formData = ref({
  name: '',
  description: '',
  color: '#6366f1',
  icon: 'folder'
})

// Available icons
const iconOptions = [
  // General
  { label: 'Folder', value: 'folder', icon: 'material-symbols:folder' },
  { label: 'Star', value: 'star', icon: 'material-symbols:star' },
  { label: 'Heart', value: 'heart', icon: 'material-symbols:favorite' },
  { label: 'Bookmark', value: 'bookmark', icon: 'material-symbols:bookmark' },
  { label: 'Pin', value: 'pin', icon: 'material-symbols:push-pin' },
  
  // Work & Professional
  { label: 'Work', value: 'work', icon: 'material-symbols:work' },
  { label: 'Business', value: 'business', icon: 'material-symbols:business' },
  { label: 'Briefcase', value: 'briefcase', icon: 'material-symbols:work-outline' },
  { label: 'Chart', value: 'chart', icon: 'material-symbols:bar-chart' },
  { label: 'Analytics', value: 'analytics', icon: 'material-symbols:analytics' },
  { label: 'Calendar', value: 'calendar', icon: 'material-symbols:calendar-today' },
  { label: 'Task', value: 'task', icon: 'material-symbols:task' },
  { label: 'Settings', value: 'settings', icon: 'material-symbols:settings' },
  
  // Development & Tech
  { label: 'Code', value: 'code', icon: 'material-symbols:code' },
  { label: 'Terminal', value: 'terminal', icon: 'material-symbols:terminal' },
  { label: 'Bug', value: 'bug', icon: 'material-symbols:bug-report' },
  { label: 'Database', value: 'database', icon: 'material-symbols:database' },
  { label: 'API', value: 'api', icon: 'material-symbols:api' },
  { label: 'Cloud', value: 'cloud', icon: 'material-symbols:cloud' },
  { label: 'Security', value: 'security', icon: 'material-symbols:security' },
  { label: 'Memory', value: 'memory', icon: 'material-symbols:memory' },
  
  // Education & Learning
  { label: 'School', value: 'school', icon: 'material-symbols:school' },
  { label: 'Book', value: 'book', icon: 'material-symbols:book' },
  { label: 'Research', value: 'research', icon: 'material-symbols:science' },
  { label: 'Lightbulb', value: 'lightbulb', icon: 'material-symbols:lightbulb' },
  { label: 'Quiz', value: 'quiz', icon: 'material-symbols:quiz' },
  { label: 'Psychology', value: 'psychology', icon: 'material-symbols:psychology' },
  
  // Creative & Media
  { label: 'Palette', value: 'palette', icon: 'material-symbols:palette' },
  { label: 'Design', value: 'design', icon: 'material-symbols:design-services' },
  { label: 'Photo', value: 'photo', icon: 'material-symbols:photo-camera' },
  { label: 'Video', value: 'video', icon: 'material-symbols:videocam' },
  { label: 'Music', value: 'music', icon: 'material-symbols:music-note' },
  { label: 'Theatre', value: 'theatre', icon: 'material-symbols:theater-comedy' },
  
  // Lifestyle & Personal
  { label: 'Home', value: 'home', icon: 'material-symbols:home' },
  { label: 'Family', value: 'family', icon: 'material-symbols:family-restroom' },
  { label: 'Health', value: 'health', icon: 'material-symbols:health-and-safety' },
  { label: 'Fitness', value: 'fitness', icon: 'material-symbols:fitness-center' },
  { label: 'Food', value: 'food', icon: 'material-symbols:restaurant' },
  { label: 'Coffee', value: 'coffee', icon: 'material-symbols:coffee' },
  { label: 'Travel', value: 'travel', icon: 'material-symbols:flight' },
  { label: 'Car', value: 'car', icon: 'material-symbols:directions-car' },
  
  // Entertainment & Hobbies
  { label: 'Game', value: 'game', icon: 'material-symbols:sports-esports' },
  { label: 'Sports', value: 'sports', icon: 'material-symbols:sports-soccer' },
  { label: 'Pets', value: 'pets', icon: 'material-symbols:pets' },
  { label: 'Nature', value: 'nature', icon: 'material-symbols:nature' },
  { label: 'Camping', value: 'camping', icon: 'material-symbols:outdoor-grill' },
  
  // Finance & Commerce
  { label: 'Money', value: 'money', icon: 'material-symbols:attach-money' },
  { label: 'Shopping', value: 'shopping', icon: 'material-symbols:shopping-cart' },
  { label: 'Store', value: 'store', icon: 'material-symbols:store' },
  { label: 'Investment', value: 'investment', icon: 'material-symbols:trending-up' },
  
  // Communication & Social
  { label: 'Chat', value: 'chat', icon: 'material-symbols:chat' },
  { label: 'Group', value: 'group', icon: 'material-symbols:group' },
  { label: 'Public', value: 'public', icon: 'material-symbols:public' },
  { label: 'Language', value: 'language', icon: 'material-symbols:language' }
]

const isVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const title = computed(() =>
  props.mode === 'create' ? t('workspace.create') : t('workspace.edit')
)

const submitButtonText = computed(() =>
  props.mode === 'create' ? t('common.create') : t('common.update')
)

// Get the selected icon option for display
const selectedIconOption = computed(() => {
  return iconOptions.find(option => option.value === formData.value.icon)
})

// Reset form when modal opens/closes or mode changes
watch([() => props.visible, () => props.mode, () => props.workspace], () => {
  if (props.visible) {
    if (props.mode === 'edit' && props.workspace) {
      formData.value = {
        name: props.workspace.name,
        description: props.workspace.description || '',
        color: props.workspace.color,
        icon: props.workspace.icon
      }
    } else {
      // Reset for create mode
      formData.value = {
        name: '',
        description: '',
        color: '#6366f1',
        icon: 'folder'
      }
    }
  }
})

// Watch color changes and normalize automatically (debounced)
watch(() => formData.value.color, (newColor, oldColor) => {
  // Skip if color hasn't actually changed or is empty
  if (!newColor || newColor === oldColor || newColor.length < 6) {
    return
  }
  
  try {
    const normalized = normalizeColor(newColor)
    if (normalized !== newColor) {
      // Update the color silently to normalized format
      formData.value.color = normalized
    }
  } catch (error) {
    // Invalid color format, let the validation handle it in submit
    console.warn('Invalid color format:', newColor, error)
  }
}, { 
  // Debounce to avoid excessive updates during typing
  flush: 'post' 
})

function handleClose() {
  isVisible.value = false
}

// Helper function to ensure color is 7-character hex format
function normalizeColor(color: string): string {
  if (!color) {
    throw new Error('Color is required')
  }
  
  // Remove any whitespace and convert to lowercase for consistency
  color = color.trim().toLowerCase()
  
  // If it doesn't start with #, add it
  if (!color.startsWith('#')) {
    color = '#' + color
  }
  
  // If it's a 3-character hex, expand to 6 characters
  if (color.length === 4) {
    color = '#' + color[1] + color[1] + color[2] + color[2] + color[3] + color[3]
  }
  
  // Validate hex color format (case-insensitive)
  const hexColorRegex = /^#[0-9a-f]{6}$/
  if (!hexColorRegex.test(color)) {
    throw new Error(`Invalid color format: ${color}. Expected format: #RRGGBB`)
  }
  
  return color
}

async function handleSubmit() {
  if (!formData.value.name.trim()) {
    message.error(t('workspace.nameRequired'))
    return
  }

  // Validate and normalize color
  let normalizedColor: string
  try {
    normalizedColor = normalizeColor(formData.value.color)
  } catch (error) {
    message.error(t('workspace.invalidColor'))
    return
  }

  loading.value = true

  try {
    if (props.mode === 'create') {
      const createData: CreateWorkspaceRequest = {
        name: formData.value.name.trim(),
        description: formData.value.description.trim(),
        color: normalizedColor,
        icon: formData.value.icon
      }

      const workspace = await workspaceStore.createWorkspace(formData.value.name.trim(), formData.value.description.trim(), normalizedColor, formData.value.icon)
      emit('workspace-created', workspace)
    } else if (props.mode === 'edit' && props.workspace) {
      const updateData: UpdateWorkspaceRequest = {
        name: formData.value.name.trim(),
        description: formData.value.description.trim(),
        color: normalizedColor,
        icon: formData.value.icon
      }

      const workspace = await workspaceStore.updateWorkspace(props.workspace.uuid, updateData)
      emit('workspace-updated', workspace)
    }

    handleClose()
  } catch (error) {
    console.error('Error saving workspace:', error)
    message.error(t('workspace.saveError'))
  } finally {
    loading.value = false
  }
}

interface IconOption {
  label: string
  value: string
  icon: string
}

function renderIconLabel(option: IconOption) {
  return h('div', { class: 'flex items-center gap-2' }, [
    h(SvgIcon, {
      icon: option.icon,
      style: { fontSize: '18px', color: formData.value.color }
    }),
    h('span', option.label)
  ])
}

</script>

<template>
  <NModal v-model:show="isVisible" :mask-closable="false">
    <NCard :title="title" class="w-full max-w-md" :bordered="false" size="small" role="dialog" aria-modal="true">
      <template #header-extra>
        <NButton quaternary circle @click="handleClose">
          <template #icon>
            <SvgIcon icon="material-symbols:close" />
          </template>
        </NButton>
      </template>

      <NForm>
        <NFormItem :label="t('workspace.name')" required>
          <NInput v-model:value="formData.name" :placeholder="t('workspace.namePlaceholder')" maxlength="50"
            show-count />
        </NFormItem>

       
        

        <NFormItem :label="t('workspace.color')">
          <NColorPicker 
            v-model:value="formData.color" 
            :modes="['hex']" 
            :show-alpha="false"
            :show-preview="true"
            :swatches="[
              '#6366f1', '#8b5cf6', '#a855f7', '#d946ef', '#ec4899',
              '#f43f5e', '#ef4444', '#f97316', '#f59e0b', '#eab308',
              '#84cc16', '#22c55e', '#10b981', '#14b8a6', '#06b6d4',
              '#0ea5e9', '#3b82f6', '#6366f1', '#8b5cf6', '#a855f7'
            ]" 
          />
        </NFormItem>
        <NFormItem :label="t('workspace.icon')">
          <NSelect 
            v-model:value="formData.icon" 
            :options="iconOptions" 
            :render-label="renderIconLabel"
          />
        </NFormItem>

        <NFormItem :label="t('workspace.description')">
          <NInput v-model:value="formData.description" type="textarea"
            :placeholder="t('workspace.descriptionPlaceholder')" maxlength="200" show-count :rows="2" />
        </NFormItem>

      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="handleClose">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="primary" :loading="loading" @click="handleSubmit">
            {{ submitButtonText }}
          </NButton>
        </NSpace>
      </template>
    </NCard>
  </NModal>
</template>