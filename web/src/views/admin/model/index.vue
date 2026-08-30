<script setup lang="ts">
import { ref, toRaw, watch } from 'vue'
import { NButton, NEmpty, NModal, NSkeleton } from 'naive-ui'
import { useQuery } from '@tanstack/vue-query'
import AddModelForm from './AddModelForm.vue'
import { fetchChatModel } from '@/api'
import { SvgIcon } from '@/components/common'
import { t } from '@/locales'
import ModelCard from '@/components/admin/ModelCard.vue'

const dialogVisible = ref(false)

const modelQuery = useQuery({
  queryKey: ['chat_models'],
  queryFn: fetchChatModel,
})

const isLoading = modelQuery.isPending
const data = ref<Chat.ChatModel[]>(toRaw(modelQuery.data.value))

watch(modelQuery.data, () => {
  data.value = toRaw(modelQuery.data.value)
})

async function newRowEventHandle() {
  dialogVisible.value = false
}
</script>

<template>
  <div class="flex items-center justify-end mb-3">
    <NButton type="primary" @click="dialogVisible = true">
      <template #icon>
        <SvgIcon icon="material-symbols:add-rounded" />
      </template>
      {{ t('admin.add_model') }}
    </NButton>
  </div>
  <div v-if="isLoading" class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
    <NSkeleton v-for="index in 3" :key="index" height="122px" :sharp="false" />
  </div>
  <NEmpty v-else-if="!data?.length" :description="t('admin.noModels')" class="py-12" />
  <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
    <ModelCard
      v-for="model in data"
      :key="model.id"
      :model="model"
    />
  </div>
  <NModal v-model:show="dialogVisible" :title="$t('admin.add_model')" preset="card" class="w-full max-w-2xl">
    <AddModelForm @new-row-added="newRowEventHandle" />
  </NModal>
</template>
