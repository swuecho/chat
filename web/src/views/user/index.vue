<script lang="ts" setup>
// create a data table with pagination using naive-ui, with the following columns:
// User Email, Total Sessions, Total Messages, Total Sessions (3 days), Total Messages (3 days), Rate Limit
// The data should be fetched from the backend using api 'GetUserData(page, page_size)'
// The Rate Limit column should be editable, and the value should be updated in the backend using api 'UpdateRateLimit(user_email, rate_limit)'
// vue3 code should be in <script lang="ts" setup> style.
import { onMounted, reactive, ref } from 'vue'
import { NDataTable } from 'naive-ui'
import { GetUserData } from '@/api'

const columns = [
  {
    title: 'User Email',
    key: 'email',

  },
  {
    title: 'Total Messages',
    key: 'totalChatMessages',
  },
  {
    title: 'Total Messages (3 days)',
    key: 'totalChatMessages3Days',
  },
  {
    title: 'Rate Limit',
    key: 'rateLimit',
    // render: (row: any) => {
    //   return {
    //     type: 'input',
    //     value: row.rateLimit,
    //     onUpdate: async (value: any) => {
    //       await UpdateRateLimit(row.email, value)
    //     },
    //   }
    // },
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

const tableData = ref([])

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
