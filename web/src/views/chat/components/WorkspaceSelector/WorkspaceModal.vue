<script setup lang="ts">
import { computed, ref, watch } from 'vue'
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
import { useChatStore } from '@/store'
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

const chatStore = useChatStore()
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
  { label: 'Folder', value: 'folder', icon: 'material-symbols:folder' },
  { label: 'Work', value: 'work', icon: 'material-symbols:work' },
  { label: 'Home', value: 'home', icon: 'material-symbols:home' },
  { label: 'School', value: 'school', icon: 'material-symbols:school' },
  { label: 'Star', value: 'star', icon: 'material-symbols:star' },
  { label: 'Heart', value: 'heart', icon: 'material-symbols:favorite' },
  { label: 'Code', value: 'code', icon: 'material-symbols:code' },
  { label: 'Research', value: 'research', icon: 'material-symbols:science' },
  { label: 'Game', value: 'game', icon: 'material-symbols:sports-esports' },
  { label: 'Music', value: 'music', icon: 'material-symbols:music-note' },
  { label: 'Travel', value: 'travel', icon: 'material-symbols:flight' },
  { label: 'Shopping', value: 'shopping', icon: 'material-symbols:shopping-cart' }
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

function handleClose() {
  isVisible.value = false
}

async function handleSubmit() {
  if (!formData.value.name.trim()) {
    message.error(t('workspace.nameRequired'))
    return
  }

  loading.value = true
  
  try {
    if (props.mode === 'create') {
      const createData: CreateWorkspaceRequest = {
        name: formData.value.name.trim(),
        description: formData.value.description.trim(),
        color: formData.value.color,
        icon: formData.value.icon
      }
      
      const workspace = await chatStore.createNewWorkspace(createData)
      emit('workspace-created', workspace)
    } else if (props.mode === 'edit' && props.workspace) {
      const updateData: UpdateWorkspaceRequest = {
        name: formData.value.name.trim(),
        description: formData.value.description.trim(),
        color: formData.value.color,
        icon: formData.value.icon
      }
      
      const workspace = await chatStore.updateWorkspaceData(props.workspace.uuid, updateData)
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

function renderIconOption({ node, option }: any) {
  return h('div', { class: 'flex items-center gap-2' }, [
    h(SvgIcon, { 
      icon: option.icon, 
      style: { fontSize: '18px', color: formData.value.color } 
    }),
    node
  ])
}
</script>

<template>
  <NModal v-model:show="isVisible" :mask-closable="false">
    <NCard
      :title="title"
      class="w-full max-w-md"
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

      <NForm>
        <NFormItem :label="t('workspace.name')" required>
          <NInput
            v-model:value="formData.name"
            :placeholder="t('workspace.namePlaceholder')"
            maxlength="50"
            show-count
          />
        </NFormItem>

        <NFormItem :label="t('workspace.description')">
          <NInput
            v-model:value="formData.description"
            type="textarea"
            :placeholder="t('workspace.descriptionPlaceholder')"
            maxlength="200"
            show-count
            :rows="2"
          />
        </NFormItem>

        <NFormItem :label="t('workspace.icon')">
          <NSelect
            v-model:value="formData.icon"
            :options="iconOptions"
            :render-option="renderIconOption"
          >
            <template #default="{ node, option }">
              <div class="flex items-center gap-2">
                <SvgIcon 
                  :icon="option.icon" 
                  :style="{ fontSize: '18px', color: formData.color }" 
                />
                {{ node }}
              </div>
            </template>
          </NSelect>
        </NFormItem>

        <NFormItem :label="t('workspace.color')">
          <NColorPicker
            v-model:value="formData.color"
            :modes="['hex']"
            :swatches="[
              '#6366f1', '#8b5cf6', '#a855f7', '#d946ef', '#ec4899',
              '#f43f5e', '#ef4444', '#f97316', '#f59e0b', '#eab308',
              '#84cc16', '#22c55e', '#10b981', '#14b8a6', '#06b6d4',
              '#0ea5e9', '#3b82f6', '#6366f1', '#8b5cf6', '#a855f7'
            ]"
          />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="handleClose">
            {{ t('common.cancel') }}
          </NButton>
          <NButton
            type="primary"
            :loading="loading"
            @click="handleSubmit"
          >
            {{ submitButtonText }}
          </NButton>
        </NSpace>
      </template>
    </NCard>
  </NModal>
</template>