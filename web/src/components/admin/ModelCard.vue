<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NCard, NModal, NForm, NFormItem, NInput, NSwitch, useMessage, NBadge } from 'naive-ui'
import { t } from '@/locales'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { updateChatModel, deleteChatModel } from '@/api'
import { mode } from 'crypto-js'

const props = defineProps<{
  model: Chat.ChatModel
}>()

const queryClient = useQueryClient()
const ms_ui = useMessage()
const dialogVisible = ref(false)
const editData = ref({ ...props.model })

const chatModelMutation = useMutation({
  mutationFn: (variables: { id: number, data: any }) => updateChatModel(variables.id, variables.data),
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ['chat_models'] })
})

const deteteModelMutation = useMutation({
  mutationFn: (id: number) => deleteChatModel(id),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['chat_models'] })
  },
})

function handleUpdate() {
  if (editData.value.id) {
    const updatedData = {
      id: editData.value.id,
      data: {
        ...editData.value,
        orderNumber: parseInt(editData.value.orderNumber.toString() || '0'),
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
    deteteModelMutation.mutate(editData.value.id)
  }
}

async function copyJson() {
  try {
    // Create a clean copy without Vue reactivity
    const dataToCopy = {
      name: editData.value.name,
      label: editData.value.label,
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
    console.log(dataToCopy)
    
    const text = JSON.stringify(dataToCopy, null, 2)
    
    // Use modern clipboard API if available
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(text)
    } else {
      // Fallback for older browsers
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    
    ms_ui.success(t('admin.chat_model.copy_success'))
  } catch (error) {
    console.error('Copy failed:', error)
    ms_ui.error(t('admin.chat_model.copy_failed'))
  }
}
</script>

<template>
  <div>
    <NCard hoverable class="mb-4 cursor-pointer" @click="dialogVisible = true">
      <div class="flex justify-between items-center">
        <div>
          <div class="flex items-center gap-2">
            <NBadge :value="model.orderNumber?.toString() || '0'" show-zero type="success">
              <h3 class="font-bold">{{ model.name }}</h3>
            </NBadge>
          </div>
          <p class="text-gray-500">{{ model.label }}</p>
        </div>
        <NSwitch 
          :value="model.isEnable" 
          @update:value="handleEnableToggle"
          @click.stop   
        />
      </div>
    </NCard>

    <NModal v-model:show="dialogVisible" preset="dialog">
      <NCard>
        <NForm>
          <NFormItem :label="t('admin.chat_model.name')">
            <NInput v-model:value="editData.name" />
          </NFormItem>
          <NFormItem :label="t('admin.chat_model.label')">
            <NInput v-model:value="editData.label" />
          </NFormItem>
          <NFormItem :label="t('admin.chat_model.url')">
            <NInput v-model:value="editData.url" />
          </NFormItem>
          <NFormItem :label="t('admin.chat_model.apiAuthHeader')">
            <NInput v-model:value="editData.apiAuthHeader" />
          </NFormItem>
          <NFormItem :label="t('admin.chat_model.apiAuthKey')">
            <NInput v-model:value="editData.apiAuthKey" />
          </NFormItem>
          <div class="flex gap-4">
            <NFormItem :label="t('admin.chat_model.isDefault')" class="flex-1">
              <NSwitch v-model:value="editData.isDefault" />
            </NFormItem>
            <NFormItem :label="t('admin.chat_model.enablePerModeRatelimit')" class="flex-1">
              <NSwitch v-model:value="editData.enablePerModeRatelimit" />
            </NFormItem>
          </div>
          <div class="flex gap-4">
            <NFormItem :label="t('admin.chat_model.defaultToken')" class="flex-1">
              <NInput v-model:value="editData.defaultToken" />
            </NFormItem>
            <NFormItem :label="t('admin.chat_model.maxToken')" class="flex-1">
              <NInput v-model:value="editData.maxToken" />
            </NFormItem>
          </div>
            <NFormItem :label="t('admin.chat_model.orderNumber')" class="flex-1">
              <NInput v-model:value="editData.orderNumber.toString()" />
            </NFormItem>
        </NForm>

        <div class="flex justify-end gap-2 mt-4">
          <NButton type="info" @click="copyJson">
            {{ t('admin.chat_model.copy') }}
          </NButton>
          <NButton type="error" @click="handleDelete">
            {{ t('common.delete') }}
          </NButton>
          <NButton type="primary" @click="handleUpdate">
            {{ t('common.save') }}
          </NButton>
        </div>
      </NCard>
    </NModal>
  </div>
</template>
