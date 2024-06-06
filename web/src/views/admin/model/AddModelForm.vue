<script setup lang="ts">
import { ref } from 'vue'
import { NButton, NForm, NFormItem, NInput, NSwitch } from 'naive-ui'
import { createChatModel } from '@/api'
import { useMutation, useQueryClient } from '@tanstack/vue-query'

const emit = defineEmits<Emit>()

const formData = ref<Chat.ChatModel>({
  name: '',
  label: '',
  url: '',
  isDefault: false,
  apiAuthHeader: '',
  apiAuthKey: '',
  enablePerModeRatelimit: false,
})

interface Emit {
  (e: 'newRowAdded'): void
}


const queryClient = useQueryClient()

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
      <NFormItem path="label" :label="$t('admin.chat_model.label')">
        <NInput v-model:value="formData.label" />
      </NFormItem>
      <NFormItem path="name" :label="$t('admin.chat_model.name')">
        <NInput v-model:value="formData.name" />
      </NFormItem>
      <NFormItem path="url" :label="$t('admin.chat_model.url')">
        <NInput v-model:value="formData.url" />
      </NFormItem>
      <NFormItem path="apiAuthHeader" :label="$t('admin.chat_model.api_auth_header')">
        <NInput v-model:value="formData.apiAuthHeader" />
      </NFormItem>
      <NFormItem path="apiAuthKey" :label="$t('admin.chat_model.api_auth_key')">
        <NInput v-model:value="formData.apiAuthKey" />
      </NFormItem>
      <NFormItem path="isDefault" :label="$t('admin.chat_model.is_default')">
        <NSwitch v-model:value="formData.isDefault" />
      </NFormItem>
      <NFormItem path="enablePerModeRatelimit" :label="$t('admin.chat_model.enable_per_model_rate_limit')">
        <NSwitch v-model:value="formData.enablePerModeRatelimit" />
      </NFormItem>
    </NForm>

    <NButton type="primary" block secondary strong @click="addRow">
      {{ $t('common.confirm') }}
    </NButton>
  </div>
</template>
