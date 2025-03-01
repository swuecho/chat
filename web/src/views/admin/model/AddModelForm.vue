<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NForm, NFormItem, NInput, NSwitch, NTextarea, useMessage } from 'naive-ui'
import { createChatModel } from '@/api'
import { useMutation, useQueryClient } from '@tanstack/vue-query'

const queryClient = useQueryClient()

const emit = defineEmits<Emit>()

const ms_ui = useMessage()
const jsonInput = ref('')
const formData = ref<Chat.ChatModel>({
  name: '',
  label: '',
  url: '',
  isDefault: false,
  apiAuthHeader: '',
  apiAuthKey: '',
  enablePerModeRatelimit: false,
  isEnable: true,
})

function populateFromJson() {
  try {
    const jsonData = JSON.parse(jsonInput.value)
    
    // Validate required fields
    if (!jsonData.name || !jsonData.label || !jsonData.url) {
      throw new Error('Missing required fields (name, label, url)')
    }

    // Update form data
    formData.value = {
      ...formData.value, // Keep default values
      ...jsonData       // Override with JSON values
    }
    
    ms_ui.success('Form populated successfully')
  } catch (error) {
    ms_ui.error('Invalid JSON or missing required fields')
    console.error('JSON parse error:', error)
  }
}

interface Emit {
  (e: 'newRowAdded'): void
}


const createChatModelMutation = useMutation({
  mutationFn: (formData: Chat.ChatModel) => createChatModel(formData),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['chat_models'] })
  },
})

async function addRow() {
  // create a new chat model, the name is randon string
  createChatModelMutation.mutate(formData.value)
  // add it to the data array
  emit('newRowAdded')
}
</script>

<template>
  <div>
    <NForm :model="formData">
      <NFormItem path="name" :label="$t('admin.chat_model.name')">
        <NInput v-model:value="formData.name" />
      </NFormItem>
      <NFormItem path="label" :label="$t('admin.chat_model.label')">
        <NInput v-model:value="formData.label" />
      </NFormItem>
      <NFormItem path="url" :label="$t('admin.chat_model.url')">
        <NInput v-model:value="formData.url" />
      </NFormItem>
      <NFormItem path="apiAuthHeader" :label="$t('admin.chat_model.apiAuthHeader')">
        <NInput v-model:value="formData.apiAuthHeader" />
      </NFormItem>
      <NFormItem path="apiAuthKey" :label="$t('admin.chat_model.apiAuthKey')">
        <NInput v-model:value="formData.apiAuthKey" />
      </NFormItem>
      <NFormItem path="isDefault" :label="$t('admin.chat_model.isDefault')">
        <NSwitch v-model:value="formData.isDefault" />
      </NFormItem>
      <NFormItem path="enablePerModeRatelimit" :label="$t('admin.chat_model.enablePerModeRatelimit')">
        <NSwitch v-model:value="formData.enablePerModeRatelimit" />
      </NFormItem>
    </NForm>

    <NFormItem :label="$t('admin.chat_model.paste_json')">
      <NTextarea
        v-model:value="jsonInput"
        :placeholder="$t('admin.chat_model.paste_json_placeholder')"
        :rows="5"
      />
    </NFormItem>
    
    <div class="flex gap-2 mt-4">
      <NButton 
        type="info" 
        secondary 
        strong 
        @click="populateFromJson"
        class="flex-1"
      >
        {{ $t('admin.chat_model.populate_form') }}
      </NButton>
      <NButton 
        type="primary" 
        secondary 
        strong 
        @click="addRow"
        class="flex-1"
      >
        {{ $t('common.confirm') }}
      </NButton>
    </div>
  </div>
</template>
