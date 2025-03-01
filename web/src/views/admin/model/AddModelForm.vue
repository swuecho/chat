<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NForm, NFormItem, NInput, NSwitch, useMessage } from 'naive-ui'
import { createChatModel } from '@/api'
import { useMutation, useQueryClient } from '@tanstack/vue-query'

const queryClient = useQueryClient()

const emit = defineEmits<Emit>()

const ms_ui = useMessage()
const jsonInput = ref('')
const defaultFormData = {
  name: '',
  label: '',
  url: '',
  isDefault: false,
  apiAuthHeader: '',
  apiAuthKey: '',
  enablePerModeRatelimit: false,
  isEnable: true,
  orderNumber: 0,
  defaultToken: 0,
  maxToken: 0
}

const formData = ref<Chat.ChatModel>({ ...defaultFormData })

function clearForm() {
  formData.value = { ...defaultFormData }
  jsonInput.value = ''
  ms_ui.success('Form cleared successfully')
}

function populateFromJson() {
  try {
    if (!jsonInput.value.trim()) {
      throw new Error('Please paste JSON configuration')
    }

    const jsonData = JSON.parse(jsonInput.value)
    
    // Validate required fields
    const requiredFields = ['name', 'label', 'url']
    const missingFields = requiredFields.filter(field => !jsonData[field])
    
    if (missingFields.length > 0) {
      throw new Error(`Missing required fields: ${missingFields.join(', ')}`)
    }

    // Validate number fields
    const numberFields = ['orderNumber', 'defaultToken', 'maxToken']
    numberFields.forEach(field => {
      if (jsonData[field] && isNaN(jsonData[field])) {
        throw new Error(`${field} must be a number`)
      }
    })

    // Update form data with validation
    formData.value = {
      ...defaultFormData, // Reset to defaults first
      ...jsonData,        // Override with JSON values
      orderNumber: jsonData.orderNumber || 0,
      defaultToken: jsonData.defaultToken || 0,
      maxToken: jsonData.maxToken || 0
    }
    
    ms_ui.success('Form populated successfully from JSON')
  } catch (error) {
    ms_ui.error(`Error: ${error.message}`)
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
      <NInput
        v-model:value="jsonInput"
        type="textarea"
        :placeholder="$t('admin.chat_model.paste_json_placeholder')"
        :rows="5"
      />
    </NFormItem>

    <NFormItem v-if="jsonInput" label="JSON Preview">
      <pre class="p-2 bg-gray-100 rounded">{{ JSON.stringify(JSON.parse(jsonInput), null, 2) }}</pre>
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
        type="warning" 
        secondary 
        strong 
        @click="clearForm"
        class="flex-1"
      >
        {{ $t('admin.chat_model.clear_form') }}
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
