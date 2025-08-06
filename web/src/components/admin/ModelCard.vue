<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NButton, NCard, NModal, NForm, NFormItem, NInput, NSwitch, NSelect, useMessage, NBadge, useDialog, NSpin } from 'naive-ui'
import { t } from '@/locales'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { updateChatModel, deleteChatModel } from '@/api'
import copy from 'copy-to-clipboard'
import { API_TYPE_OPTIONS, API_TYPE_DISPLAY_NAMES } from '@/constants/apiTypes'

const props = defineProps<{
  model: Chat.ChatModel
}>()

const queryClient = useQueryClient()
const ms_ui = useMessage()
const dialog = useDialog()
const dialogVisible = ref(false)
const editData = ref({ ...props.model })

// Watch for prop changes to keep editData in sync
watch(() => props.model, (newModel) => {
  editData.value = { ...newModel }
}, { deep: true })

// Computed properties for better performance
const isDefaultModel = computed(() => props.model.isDefault)
const cardClasses = computed(() => ({
  'border-2 border-green-500 bg-green-50 dark:bg-green-900/20': isDefaultModel.value,
  'hover:shadow-lg': true,
  'transition-all': true,
  'duration-200': true
}))

const apiTypeDisplay = computed(() => {
  const apiType = props.model.apiType
  return API_TYPE_DISPLAY_NAMES[apiType as keyof typeof API_TYPE_DISPLAY_NAMES] || apiType || 'Unknown'
})

// API Type options (imported from constants)
const apiTypeOptions = API_TYPE_OPTIONS

const chatModelMutation = useMutation({
  mutationFn: (variables: { id: number, data: any }) => updateChatModel(variables.id, variables.data),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['chat_models'] })
    ms_ui.success(t('admin.chat_model.update_success'))
  },
  onError: (error) => {
    console.error('Failed to update model:', error)
    ms_ui.error(t('admin.chat_model.update_failed'))
  }
})

const deteteModelMutation = useMutation({
  mutationFn: (id: number) => deleteChatModel(id),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['chat_models'] })
    ms_ui.success(t('admin.chat_model.delete_success'))
  },
  onError: (error) => {
    console.error('Failed to delete model:', error)
    ms_ui.error(t('admin.chat_model.delete_failed'))
  }
})

function handleUpdate() {
  if (editData.value.id) {
    const updatedData = {
      id: editData.value.id,
      data: {
        ...editData.value,
        orderNumber: parseInt(editData.value.orderNumber?.toString() || '0'),
        defaultToken: parseInt(editData.value.defaultToken || '0'),
        maxToken: parseInt(editData.value.maxToken || '0'),
      }
    }
    chatModelMutation.mutate(updatedData)
    dialogVisible.value = false
  }
}

function handleEnableToggle(enabled: boolean) {
  if (editData.value.id) {
    const updatedData = {
      id: editData.value.id,
      data: {
        ...editData.value,
        isEnable: enabled
      }
    }
    chatModelMutation.mutate(updatedData)
  }
}

function handleDelete() {
  if (editData.value.id) {
    dialog.warning({
      title: t('common.warning'),
      content: t('admin.chat_model.deleteModelConfirm', { name: editData.value.name }),
      positiveText: t('common.confirm'),
      negativeText: t('common.cancel'),
      onPositiveClick: () => {
        deteteModelMutation.mutate(editData.value.id ?? 0)
      }
    })
  }
}


function copyJson() {
  // Create a clean copy without Vue reactivity
  const dataToCopy = {
    name: editData.value.name,
    label: editData.value.label,
    apiType: editData.value.apiType,
    url: editData.value.url,
    apiAuthHeader: editData.value.apiAuthHeader,
    apiAuthKey: editData.value.apiAuthKey,
    isDefault: editData.value.isDefault,
    enablePerModeRatelimit: editData.value.enablePerModeRatelimit,
    isEnable: editData.value.isEnable,
    orderNumber: editData.value.orderNumber,
    defaultToken: editData.value.defaultToken,
    maxToken: editData.value.maxToken
  }

  const text = JSON.stringify(dataToCopy, null, 2)
  const success = copy(text)

  if (success) {
    ms_ui.success(t('admin.chat_model.copy_success'))
  } else {
    ms_ui.error(t('admin.chat_model.copy_failed'))
  }
}
</script>

<template>
  <div>
    <NCard hoverable class="mb-4 cursor-pointer relative overflow-hidden" @click="dialogVisible = true"
      :class="cardClasses">
      <!-- Loading overlay -->
      <div v-if="chatModelMutation.isPending.value"
        class="absolute inset-0 bg-white/80 dark:bg-black/80 flex items-center justify-center z-10">
        <NSpin size="medium" />
      </div>

      <div class="flex justify-between items-start gap-4">
        <div class="flex-1 min-w-0">
          <!-- Header with model name and badges -->
          <div class="flex items-start gap-2 mb-2">
            <NBadge :value="model.orderNumber?.toString() || '0'" show-zero type="success" class="flex-shrink-0">
              <h3 class="font-semibold text-lg truncate max-w-[120px] sm:max-w-[150px] md:max-w-[180px]"
                :class="{ 'text-green-700 dark:text-green-300': isDefaultModel }" :title="model.name">
                {{ model.name }}
              </h3>
            </NBadge>
            <NBadge v-if="isDefaultModel" type="success" :value="t('admin.chat_model.default')" size="small"
              class="flex-shrink-0 mt-1" />
          </div>

          <!-- Model label/description -->
          <p class="text-sm mb-3 truncate"
            :class="{ 'text-green-600 dark:text-green-400': isDefaultModel, 'text-gray-600 dark:text-gray-400': !isDefaultModel }"
            :title="model.label">
            {{ model.label }}
          </p>

          <!-- API Type and Status -->
          <div class="flex items-center gap-2 flex-wrap">
            <NBadge type="info" :value="apiTypeDisplay" size="small" />
            <NBadge v-if="!model.isEnable" type="warning" :value="t('admin.chat_model.disabled')" size="small" />
          </div>
        </div>

        <!-- Enable/Disable Toggle -->
        <div class="flex-shrink-0 flex flex-col items-center gap-2" @click.stop>
          <NSwitch :value="model.isEnable" @update:value="handleEnableToggle"
            :loading="chatModelMutation.isPending.value" size="medium" />
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ model.isEnable ? t('common.enabled') : t('common.disabled') }}
          </span>
        </div>
      </div>
    </NCard>

    <NModal v-model:show="dialogVisible" preset="dialog" :title="t('admin.chat_model.edit_model')"
      class="w-full max-w-2xl">
      <NCard :bordered="false">
        <NSpin :show="chatModelMutation.isPending.value || deteteModelMutation.isPending.value">
          <NForm label-placement="top" require-mark-placement="right-hanging">
            <!-- Basic Information -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
              <NFormItem :label="t('admin.chat_model.name')" required>
                <NInput v-model:value="editData.name" placeholder="e.g., gpt-4"
                  :disabled="chatModelMutation.isPending.value" />
              </NFormItem>
              <NFormItem :label="t('admin.chat_model.label')" required>
                <NInput v-model:value="editData.label" placeholder="e.g., GPT-4 Turbo"
                  :disabled="chatModelMutation.isPending.value" />
              </NFormItem>
            </div>

            <!-- API Configuration -->
            <div class="space-y-4 mb-6">
              <NFormItem :label="t('admin.chat_model.apiType')" required>
                <NSelect v-model:value="editData.apiType" :options="apiTypeOptions" placeholder="Select API Type"
                  :disabled="chatModelMutation.isPending.value" />
              </NFormItem>
              <NFormItem :label="t('admin.chat_model.url')">
                <NInput v-model:value="editData.url" placeholder="API endpoint URL"
                  :disabled="chatModelMutation.isPending.value" />
              </NFormItem>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <NFormItem :label="t('admin.chat_model.apiAuthHeader')">
                  <NInput v-model:value="editData.apiAuthHeader" placeholder="Authorization"
                    :disabled="chatModelMutation.isPending.value" />
                </NFormItem>
                <NFormItem :label="t('admin.chat_model.apiAuthKey')">
                  <NInput v-model:value="editData.apiAuthKey" type="password" show-password-on="click"
                    placeholder="API Key" :disabled="chatModelMutation.isPending.value" />
                </NFormItem>
              </div>
            </div>

            <!-- Settings -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
              <div class="space-y-4">
                <NFormItem :label="t('admin.chat_model.isDefault')">
                  <NSwitch v-model:value="editData.isDefault" :disabled="chatModelMutation.isPending.value" />
                </NFormItem>
                <NFormItem :label="t('admin.chat_model.enablePerModeRatelimit')">
                  <NSwitch v-model:value="editData.enablePerModeRatelimit"
                    :disabled="chatModelMutation.isPending.value" />
                </NFormItem>
              </div>
              <div class="space-y-4">
                <NFormItem :label="t('admin.chat_model.orderNumber')">
                  <NInput v-model:value="editData.orderNumber" placeholder="0"
                    :disabled="chatModelMutation.isPending.value" />
                </NFormItem>
              </div>
            </div>

            <!-- Token Configuration -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
              <NFormItem :label="t('admin.chat_model.defaultToken')">
                <NInput v-model:value="editData.defaultToken" placeholder="1000"
                  :disabled="chatModelMutation.isPending.value" />
              </NFormItem>
              <NFormItem :label="t('admin.chat_model.maxToken')">
                <NInput v-model:value="editData.maxToken" placeholder="4000"
                  :disabled="chatModelMutation.isPending.value" />
              </NFormItem>
            </div>
          </NForm>
        </NSpin>

        <!-- Action Buttons -->
        <div class="flex justify-between items-center pt-4 border-t border-gray-200 dark:border-gray-700">
          <NButton type="info" @click="copyJson"
            :disabled="chatModelMutation.isPending.value || deteteModelMutation.isPending.value">
            {{ t('admin.chat_model.copy') }}
          </NButton>

          <div class="flex gap-3">
            <NButton type="error" @click="handleDelete" :loading="deteteModelMutation.isPending.value"
              :disabled="chatModelMutation.isPending.value">
              {{ t('common.delete') }}
            </NButton>
            <NButton type="primary" @click="handleUpdate" :loading="chatModelMutation.isPending.value"
              :disabled="deteteModelMutation.isPending.value">
              {{ t('common.save') }}
            </NButton>
          </div>
        </div>
      </NCard>
    </NModal>
  </div>
</template>
