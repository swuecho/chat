<script lang="ts" setup>
// create a data table with pagination using naive-ui, with the following columns:
// User Email, Total Sessions, Total Messages, Total Sessions (3 days), Total Messages (3 days), Rate Limit
// The data should be fetched from the backend using api 'GetUserData(page, page_size)'
// The Rate Limit column should be editable, and the value should be updated in the backend using api 'UpdateRateLimit(user_email, rate_limit)'
// vue3 code should be in <script lang="ts" setup> style.
import { h, onMounted, reactive, ref } from 'vue'
import { NDataTable, NInput, useMessage } from 'naive-ui'
import { GetUserData, UpdateRateLimit } from '@/api'
import { t } from '@/locales'
import HoverButton from '@/components/common/HoverButton/index.vue'

const ms_ui = useMessage()

interface UserData {
  email: string
  totalChatMessages: number
  totalChatMessagesTokenCount: number
  totalChatMessages3Days: number
  totalChatMessages3DaysTokenCount: number
  totalChatMessages3DaysAvgTokenCount: number
  rateLimit: number
}
const tableData = ref<UserData[]>([])

const columns = [
  {
    title: t('admin.userEmail'),
    key: 'email',
    width: 400,

  },
  {
    title: t('admin.totalChatMessages'),
    key: 'totalChatMessages',
    width: 100,
  },
  {
    title: t('admin.totalChatMessagesTokenCount'),
    key: 'totalChatMessagesTokenCount',
    width: 100,
  },
  {
    title: t('admin.totalChatMessages3Days'),
    key: 'totalChatMessages3Days',
    width: 100,
  },
  {
    title: t('admin.totalChatMessages3DaysTokenCount'),
    key: 'totalChatMessages3DaysTokenCount',
    width: 100,
  },
  {
    title: t('admin.totalChatMessages3DaysAvgTokenCount'),
    key: 'avgChatMessages3DaysTokenCount',
    width: 100,
  },
  {
    title: t('admin.rateLimit10Min'),
    key: 'rateLimit',
    width: 100,
    render: (row: any, index: number) => {
      return h(NInput, {
        value: row.rateLimit,
        width: 50,
        async onUpdateValue(v) {
          tableData.value[index].rateLimit = parseInt(v)
          await UpdateRateLimit(row.email, parseInt(v))
        },
      })
    },
  },
]

const pagination = reactive({
  page: 1,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  pageSize: 10,
  itemCount: 10,
  onChange: async (page: number) => {
    pagination.page = page
    await fetchData()
  },
  onUpdatePageSize: async (pageSize: number) => {
    pagination.pageSize = pageSize
    pagination.page = 1
    await fetchData()
  },
})

async function fetchData() {
  try {
    const { data, total } = await GetUserData(pagination.page, pagination.pageSize)
    tableData.value = data
    pagination.itemCount = total
  }
  catch (err: any) {
    if (err.response.status === 401)
      ms_ui.error(t(err.response.data.message))
    else
      ms_ui.error(t(err.response.data.message))
  }
}

onMounted(() => {
  fetchData()
})

async function handleRefresh() {
  await fetchData()
}
</script>

<template>
  <div>
    <div class="flex justify-end z-300">
      <HoverButton :tooltip="$t('admin.refresh')" @click="handleRefresh">
        <span class="text-xl text-[#4f555e] dark:text-white">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path fill="currentColor" d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"/></svg>
        </span>
      </HoverButton>
    </div>
    <NDataTable remote :data="tableData" :columns="columns" :pagination="pagination" />
  </div>
</template>
