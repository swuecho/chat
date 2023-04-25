<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NForm, NFormItem, NInput, NSwitch } from 'naive-ui'
import { createChatModel } from '@/api'

interface RowData {
  ApiAuthHeader: string
  ApiAuthKey: string
  IsDefault: boolean
  Label: string
  Name: string
  Url: string
  EnablePerModelRatelimit: boolean
}

const emit = defineEmits<Emit>()

const formData = ref<RowData>({
  Name: '',
  Label: '',
  Url: '',
  IsDefault: false,
  ApiAuthHeader: '',
  ApiAuthKey: '',
  EnablePerModelRatelimit: false,
})

interface Emit {
  (e: 'newRowAdded'): void
}

async function addRow() {
  // create a new chat model, the name is randon string
  await createChatModel(formData.value)
  // add it to the data array
  emit('newRowAdded')
}
</script>

<template>
  <div>
    <NForm :model="formData">
      <NFormItem path="Label" :label="$t('admin.chat_model.label')">
        <NInput v-model:value="formData.Label" />
      </NFormItem>
      <NFormItem path="Name" :label="$t('admin.chat_model.name')">
        <NInput v-model:value="formData.Name" />
      </NFormItem>
      <NFormItem path="Url" :label="$t('admin.chat_model.url')">
        <NInput v-model:value="formData.Url" />
      </NFormItem>
      <NFormItem path="ApiAuthHeader" :label="$t('admin.chat_model.api_auth_header')">
        <NInput
          v-model:value="formData.ApiAuthHeader"
        />
      </NFormItem>
      <NFormItem path="ApiAuthKey" :label="$t('admin.chat_model.api_auth_key')">
        <NInput
          v-model:value="formData.ApiAuthKey"
        />
      </NFormItem>
      <NFormItem path="IsDefault" :label="$t('admin.chat_model.is_default')">
        <NSwitch v-model:value="formData.IsDefault" />
      </NFormItem>
      <NFormItem path="EnablePerModelRatelimit" :label="$t('admin.chat_model.enable_per_model_rate_limit')">
        <NSwitch v-model:value="formData.EnablePerModelRatelimit" />
      </NFormItem>
    </NForm>

    <NButton type="primary" block secondary strong @click="addRow">
      {{ $t('common.confirm') }}
    </NButton>
  </div>
</template>
