<script setup lang="ts">
import { ref, toRaw, watch } from 'vue'
import { NModal, useMessage } from 'naive-ui'
import AddModelForm from './AddModelForm.vue'
import { fetchChatModel } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { useQuery } from '@tanstack/vue-query'
import ModelCard from '@/components/admin/ModelCard.vue'

const ms_ui = useMessage()
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
  <div class="flex items-center justify-end h-14 w-full border-b border-gray-200">
    <HoverButton @click="dialogVisible = true" class="mr-10">
      <span class="text-xl">
        <SvgIcon icon="material-symbols:library-add-rounded" />
      </span>
    </HoverButton>
  </div>
  <div class="m-5" v-if="!isLoading">
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <ModelCard 
        v-for="model in data" 
        :key="model.id" 
        :model="model" 
      />
    </div>
  </div>
  <NModal v-model:show="dialogVisible" :title="$t('admin.add_model')" preset="dialog">
    <AddModelForm @new-row-added="newRowEventHandle" />
  </NModal>
</template>
