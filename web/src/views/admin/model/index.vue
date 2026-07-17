<script setup lang="ts">
import { computed, ref, toRaw, watch } from 'vue'
import { NModal, NSelect, useMessage } from 'naive-ui'
import AddModelForm from './AddModelForm.vue'
import { fetchChatModel, fetchTitleChatModel, updateTitleChatModel } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import ModelCard from '@/components/admin/ModelCard.vue'

const ms_ui = useMessage()
const dialogVisible = ref(false)
const queryClient = useQueryClient()

const modelQuery = useQuery({
  queryKey: ['chat_models'],
  queryFn: fetchChatModel,
})

const isLoading = modelQuery.isPending
const data = ref<Chat.ChatModel[]>(toRaw(modelQuery.data.value))

const titleModelQuery = useQuery({
  queryKey: ['title_chat_model'],
  queryFn: fetchTitleChatModel,
  retry: false,
})

const titleModelMutation = useMutation({
  mutationFn: updateTitleChatModel,
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['title_chat_model'] })
    queryClient.invalidateQueries({ queryKey: ['chat_models'] })
    ms_ui.success(t('admin.chat_model.title_model_update_success'))
  },
  onError: () => ms_ui.error(t('admin.chat_model.title_model_update_failed')),
})

const titleModelOptions = computed(() => (data.value || [])
  .filter(model => model.isEnable && model.apiType === 'gemini')
  .map(model => ({ label: model.label || model.name, value: model.id })))

function handleTitleModelChange(modelId: number) {
  titleModelMutation.mutate(modelId)
}

watch(modelQuery.data, () => {
  data.value = toRaw(modelQuery.data.value)
})

async function newRowEventHandle() {
  dialogVisible.value = false
}
</script>

<template>
  <div class="flex flex-col gap-4 mb-4 sm:flex-row sm:items-end sm:justify-between">
    <div>
      <h1 class="text-xl font-semibold text-gray-900 dark:text-white mb-3">
        {{ t('admin.model') }}
      </h1>
      <label class="block text-sm text-gray-600 dark:text-gray-300 mb-1">
        {{ t('admin.chat_model.title_model') }}
      </label>
      <NSelect
        class="w-full sm:w-80"
        :value="titleModelQuery.data.value?.id"
        :options="titleModelOptions"
        :loading="titleModelQuery.isPending.value || titleModelMutation.isPending.value"
        :placeholder="t('admin.chat_model.title_model_placeholder')"
        @update:value="handleTitleModelChange"
      />
      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
        {{ t('admin.chat_model.title_model_description') }}
      </p>
    </div>
    <HoverButton @click="dialogVisible = true">
      <span class="text-xl"><SvgIcon icon="material-symbols:library-add-rounded" /></span>
    </HoverButton>
  </div>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" v-if="!isLoading">
    <ModelCard 
      v-for="model in data" 
      :key="model.id" 
      :model="model" 
    />
  </div>
  <NModal v-model:show="dialogVisible" :title="$t('admin.add_model')" preset="dialog">
    <AddModelForm @new-row-added="newRowEventHandle" />
  </NModal>
</template>
