<script setup lang="ts">
import { NButton, NForm, NInput, NSelect } from 'naive-ui'
import { computed, onMounted, ref } from 'vue'
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

function submitForm() {
  addRow(form.value)
  emit('newRowAdded')
}

async function addRow(form: ChatModelFormData) {
  // create a new chat model, the name is randon string
  const chatModel = await CreateUserChatModelPrivilege({
    ID: 0,
    UserEmail: form.UserEmail,
    ChatModelName: form.ChatModelName,
    RateLimit: parseInt(form.RateLimit, 10),
  })
  // add it to the data array
  return chatModel
}

const limitEnabledModels = ref<SelectOption[]>([])
const defaultModel = ref<string>('gpt-4')

onMounted(async () => {
  limitEnabledModels.value = (await fetchChatModel()).filter((x: any) => x.EnablePerModeRatelimit)
    .map((x: any) => {
      return {
        value: x.Name,
        label: x.Label,
      }
    })
  defaultModel.value = limitEnabledModels.value[0].value
})
</script>

<template>
  <NForm :model="form">
    <NFormItem prop="UserEmail" :label="$t('common.email')">
      <NInput v-model:value="form.UserEmail" :placeholder="$t('common.email_placeholder')" />
    </NFormItem>
    <NFormItem prop="ChatModelName" :label="$t('admin.chat_model_name')">
      <NSelect v-model:value="form.ChatModelName" :options="limitEnabledModels" :default-value="defaultModel"
        placeholder="Please model name" />
    </NFormItem>
    <NFormItem prop="RateLimit" :label="$t('admin.rate_limit')">
      <NInput v-model:value="form.RateLimit" />
    </NFormItem>
    <NButton type="primary" block secondary strong @click="submitForm">
      {{ $t('common.confirm') }}
    </NButton>
  </NForm>
</template>
