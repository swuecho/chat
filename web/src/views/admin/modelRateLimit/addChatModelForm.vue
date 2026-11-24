<script setup lang="ts">
import { NButton, NForm, NFormItem, NInput, NSelect } from 'naive-ui'
import { computed, ref } from 'vue'
import { createUserChatModelPrivilege } from '@/api/generated_client'
import { useChatModels } from '@/hooks/useChatModels'

interface ChatModelPrivilege {
  chatModelName: string
  userEmail: string
  rateLimit: string
}

const emit = defineEmits<Emit>()

const form = ref<ChatModelPrivilege>({
  chatModelName: '',
  userEmail: '',
  rateLimit: '',
})

interface Emit {
  (e: 'newRowAdded'): void
}

async function submitForm() {
  await addRow(form.value)
  emit('newRowAdded')
}

async function addRow(form: ChatModelPrivilege) {
  // create a new chat model, the name is randon string
  const chatModel = await createUserChatModelPrivilege({
    body: {
      userEmail: form.userEmail,
      chatModelName: form.chatModelName,
      rateLimit: parseInt(form.rateLimit, 10),
    },
  })
  // add it to the data array
  return chatModel
}

const modelsQuery = useChatModels().useChatModelsQuery()
const limitEnabledModels = computed<SelectOption[]>(() => (modelsQuery.data.value ?? [])
  .filter(model => model.enablePerModeRatelimit)
  .map(model => ({ value: model.name, label: model.label })))
</script>

<template>
  <div>
    <NForm :model="form">
      <NFormItem path="userEmail" :label="$t('common.email')">
        <NInput v-model:value="form.userEmail" :placeholder="$t('common.email_placeholder')" />
      </NFormItem>
      <NFormItem path="chatModelName" :label="$t('admin.chat_model_name')">
        <NSelect v-model:value="form.chatModelName" :options="limitEnabledModels" />
      </NFormItem>
      <NFormItem path="rateLimit" :label="$t('admin.rate_limit')">
        <NInput v-model:value="form.rateLimit" />
      </NFormItem>
    </NForm>
    <NButton type="primary" block secondary strong @click="submitForm">
      {{ $t('common.confirm') }}
    </NButton>
  </div>
</template>
