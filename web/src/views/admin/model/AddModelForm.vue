<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NForm, NFormItem, NInput, NSwitch, NSelect, useMessage } from 'naive-ui'
import { createChatModel } from '@/api'
import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { t } from '@/locales'
import { API_TYPE_OPTIONS, API_TYPES, type ApiType } from '@/constants/apiTypes'

const queryClient = useQueryClient()

const emit = defineEmits<Emit>()

interface FormData {
  name: string
  label: string
  url: string
  isDefault: boolean
  apiAuthHeader: string
  apiAuthKey: string
  enablePerModeRatelimit: boolean
  isEnable: boolean
  orderNumber: number
  defaultToken: number
  maxToken: number
  apiType: ApiType
}

const ms_ui = useMessage()
const jsonInput = ref('')
const defaultFormData: FormData = {
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
  maxToken: 0,
  apiType: API_TYPES.OPENAI
}

const formData = ref<FormData>({ ...defaultFormData })

// API Type options (imported from constants)
const apiTypeOptions = API_TYPE_OPTIONS

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
    ms_ui.error(`Error: ${(error as Error).message}`)
    console.error('JSON parse error:', error)
  }
}

interface Emit {
  (e: 'newRowAdded'): void
}


const createChatModelMutation = useMutation({
  mutationFn: (formData: FormData) => createChatModel(formData),
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
      <NFormItem path="name" :label="t('admin.chat_model.name')">
        <NInput v-model:value="formData.name" />
      </NFormItem>
      <NFormItem path="label" :label="t('admin.chat_model.label')">
        <NInput v-model:value="formData.label" />
      </NFormItem>
      <NFormItem path="apiType" :label="t('admin.chat_model.apiType')">
        <NSelect v-model:value="formData.apiType" :options="apiTypeOptions" />
      </NFormItem>
      <NFormItem path="url" :label="t('admin.chat_model.url')">
        <NInput v-model:value="formData.url" />
      </NFormItem>
      <NFormItem path="apiAuthHeader" :label="t('admin.chat_model.apiAuthHeader')">
        <NInput v-model:value="formData.apiAuthHeader" />
      </NFormItem>
      <NFormItem path="apiAuthKey" :label="t('admin.chat_model.apiAuthKey')">
        <NInput v-model:value="formData.apiAuthKey" />
      </NFormItem>
      <div class="flex gap-4">
        <NFormItem path="isDefault" :label="t('admin.chat_model.isDefault')" class="flex-1">
          <NSwitch v-model:value="formData.isDefault" />
        </NFormItem>
        <NFormItem path="enablePerModeRatelimit" :label="t('admin.chat_model.enablePerModeRatelimit')" class="flex-1">
          <NSwitch v-model:value="formData.enablePerModeRatelimit" />
        </NFormItem>
      </div>
    </NForm>

    <NFormItem :label="t('admin.chat_model.paste_json')">
      <NInput
        v-model:value="jsonInput"
        type="textarea"
        :placeholder="t('admin.chat_model.paste_json_placeholder')"
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
        {{ t('admin.chat_model.populate_form') }}
      </NButton>
      <NButton 
        type="warning" 
        secondary 
        strong 
        @click="clearForm"
        class="flex-1"
      >
        {{ t('admin.chat_model.clear_form') }}
      </NButton>
      <NButton 
        type="primary" 
        secondary 
        strong 
        @click="addRow"
        class="flex-1"
      >
        {{ t('common.confirm') }}
      </NButton>
    </div>
  </div>
</template>
