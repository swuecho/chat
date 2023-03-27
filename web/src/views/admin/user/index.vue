<script lang="ts" setup>
// create a data table with pagination using naive-ui, with the following columns:
// User Email, Total Sessions, Total Messages, Total Sessions (3 days), Total Messages (3 days), Rate Limit
// The data should be fetched from the backend using api 'GetUserData(page, page_size)'
// The Rate Limit column should be editable, and the value should be updated in the backend using api 'UpdateRateLimit(user_email, rate_limit)'
// vue3 code should be in <script lang="ts" setup> style.
import { h, onMounted, reactive, ref } from 'vue'
import { NDataTable, NInput } from 'naive-ui'
import { GetUserData, UpdateRateLimit } from '@/api'
import { t } from '@/locales'

interface UserData {
  email: string
  totalChatMessages: number
  totalChatMessages3Days: number
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
    title: t('admin.totalChatMessages3Days'),
    key: 'totalChatMessages3Days',
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
  const { data, total } = await GetUserData(pagination.page, pagination.pageSize)
  tableData.value = data
  pagination.itemCount = total
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div>
    <NDataTable remote :data="tableData" :columns="columns" :pagination="pagination" />
  </div>
</template>
