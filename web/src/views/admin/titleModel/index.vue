<script setup lang="ts">
import { computed } from 'vue'
import { NCard, NSelect, useMessage } from 'naive-ui'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { getTitleChatModel, listChatModels, setTitleChatModel } from '@/api/generated_client'
import { t } from '@/locales'

const message = useMessage()
const queryClient = useQueryClient()

const modelQuery = useQuery({ queryKey: ['chat_models'], queryFn: () => listChatModels() })
const titleModelQuery = useQuery({
  queryKey: ['title_chat_model'],
  queryFn: () => getTitleChatModel(),
  retry: false,
})

const updateMutation = useMutation({
  mutationFn: (modelId: number) => setTitleChatModel({ body: { modelId } }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['title_chat_model'] })
    queryClient.invalidateQueries({ queryKey: ['chat_models'] })
    message.success(t('admin.chat_model.title_model_update_success'))
  },
  onError: () => message.error(t('admin.chat_model.title_model_update_failed')),
})

const options = computed(() => (modelQuery.data.value || [])
  .filter((model: Chat.ChatModel) => model.isEnable)
  .map((model: Chat.ChatModel) => ({
    label: `${model.label || model.name} (${model.apiType})`,
    value: model.id,
  })))
</script>

<template>
  <div class="max-w-xl">
    <NCard size="small">
      <p class="text-sm text-gray-600 dark:text-gray-300 mb-4">
        {{ t('admin.chat_model.title_model_description') }}
      </p>
      <NSelect
        :value="titleModelQuery.data.value?.id"
        :options="options"
        :loading="modelQuery.isPending.value || titleModelQuery.isPending.value || updateMutation.isPending.value"
        :placeholder="t('admin.chat_model.title_model_placeholder')"
        @update:value="updateMutation.mutate"
      />
    </NCard>
  </div>
</template>
