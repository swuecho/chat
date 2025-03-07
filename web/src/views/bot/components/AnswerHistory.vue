<script lang="ts" setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { NSpin, NPagination } from 'naive-ui'
import Message from './Message/index.vue'
import { useQuery } from '@tanstack/vue-query'
import { fetchBotAnswerHistory } from '@/api/bot_answer_history'
import { SvgIcon } from '@/components/common';

const { t } = useI18n()
const props = defineProps<{
  botUuid: string
}>()

const page = ref(1)
const pageSize = ref(10)

const { data: historyData, isLoading: isHistoryLoading } = useQuery({
  queryKey: ['botAnswerHistory', props.botUuid, page.value, pageSize.value],
  queryFn: async () => await fetchBotAnswerHistory(props.botUuid, page.value, pageSize.value),
})

const model = computed(() => '') // This should be passed from parent or fetched
</script>

<template>
  <div>
    <div v-if="isHistoryLoading">
      <NSpin size="large" />
    </div>
    <div v-else>
      <div v-if="historyData && historyData.items && historyData.items.length > 0">
        <div v-for="(item, index) in historyData.items" :key="index" class="mb-6">
          <div class="mb-4 border-l-4 border-neutral-200 dark:border-neutral-700 pl-4">
            <div class="text-sm text-neutral-500 dark:text-neutral-400 mb-2">
              {{ t('bot.runNumber', { number: index + 1 }) }} •
              {{ new Date(item.createdAt).toLocaleString() }}
            </div>
            <!-- User Prompt -->
            <Message 
              :date-time="item.createdAt" 
              :model="model" 
              :text="item.prompt"
              :inversion="true" 
              :index="index" 
            />
            <!-- Bot Answer -->
            <Message 
              :date-time="item.createdAt" 
              :model="model" 
              :text="item.answer"
              :inversion="false" 
              :index="index" 
            />
          </div>
        </div>
      </div>
      <div class="flex justify-center my-4" v-if="historyData?.totalPages && historyData?.totalPages > 1">
        <NPagination
          v-model:page="page"
          :page-count="historyData?.totalPages"
          :page-size="pageSize"
          show-size-picker
          :page-sizes="[10, 20, 50]"
          @update:page="page = $event"
          @update:page-size="pageSize = $event"
        />
      </div>
      <div v-if="historyData?.items?.length === 0" class="flex flex-col items-center justify-center h-64 text-neutral-400">
        <SvgIcon icon="mdi:history" class="w-12 h-12 mb-4" />
        <span>{{ t('bot.noHistory') }}</span>
      </div>
    </div>
  </div>
</template>
