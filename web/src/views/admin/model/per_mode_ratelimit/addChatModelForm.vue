<script setup lang="ts">
import { NButton, NForm, NFormItem, NInput, NSelect } from 'naive-ui'
import { onMounted, ref } from 'vue'
import { CreateUserChatModelPrivilege, fetchChatModel } from '@/api'

interface ChatModelFormData {
  ChatModelName: string
  UserEmail: string
  RateLimit: string
}

const emit = defineEmits<Emit>()

const form = ref<ChatModelFormData>({
  ChatModelName: '',
  UserEmail: '',
  RateLimit: '',
})

interface Emit {
  (e: 'newRowAdded'): void
}

async function submitForm() {
  await addRow(form.value)
  emit('newRowAdded')
}

async function addRow(form: ChatModelFormData) {
  // create a new chat model, the name is randon string
  const chatModel = await CreateUserChatModelPrivilege({
    UserEmail: form.UserEmail,
    ChatModelName: form.ChatModelName,
    RateLimit: parseInt(form.RateLimit, 10),
  })
  // add it to the data array
  return chatModel
}

const limitEnabledModels = ref<SelectOption[]>([])
onMounted(async () => {
  limitEnabledModels.value = (await fetchChatModel()).filter((x: any) => x.EnablePerModeRatelimit)
    .map((x: any) => {
      return {
        value: x.Name,
        label: x.Label,
      }
    })
})
</script>

<template>
  <div>
    <NForm :model="form">
      <NFormItem path="UserEmail" :label="$t('common.email')">
        <NInput v-model:value="form.UserEmail" :placeholder="$t('common.email_placeholder')" />
      </NFormItem>
      <NFormItem path="ChatModelName" :label="$t('admin.chat_model_name')">
        <NSelect v-model:value="form.ChatModelName" :options="limitEnabledModels" />
      </NFormItem>
      <NFormItem path="RateLimit" :label="$t('admin.rate_limit')">
        <NInput v-model:value="form.RateLimit" />
      </NFormItem>
    </NForm>
    <NButton type="primary" block secondary strong @click="submitForm">
      {{ $t('common.confirm') }}
    </NButton>
  </div>
</template>
