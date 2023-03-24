// create a data table with pagination using naive-ui, with the following columns:
// User Email, Total Sessions, Total Messages, Total Sessions (3 days), Total Messages (3 days), Rate Limit
// The data should be fetched from the backend using api 'GetUserData(page, page_size)'
// The Rate Limit column should be editable, and the value should be updated in the backend using api 'UpdateRateLimit(user_email, rate_limit)'
// vue3 code should be in <script lang="ts" setup> style.
<template>
  <div>
    <n-table :data="tableData" :columns="columns" :pagination="pagination" />
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { NTable } from 'naive-ui'
import { GetUserData, UpdateRateLimit } from '@/api/user'

const columns = [
  {
    title: 'User Email',
    key: 'user_email'
  },
  {
    title: 'Total Sessions',
    key: 'total_sessions'
  },
  {
    title: 'Total Messages',
    key: 'total_messages'
  },
  {
    title: 'Total Sessions (3 days)',
    key: 'total_sessions_3_days'
  },
  {
    title: 'Total Messages (3 days)',
    key: 'total_messages_3_days'
  },
  {
    title: 'Rate Limit',
    key: 'rate_limit',
    render: (row: any) => {
      return {
        type: 'input',
        value: row.rate_limit,
        onUpdate: async (value: any) => {
          await UpdateRateLimit(row.user_email, value)
        }
      }
    }
  }
]

const pagination = {
  current: 1,
  pageSize: 10,
  total: 0,
  async onChange(page: number, pageSize: number) {
    pagination.current = page
    pagination.pageSize = pageSize
    await fetchData()
  }
}

const tableData = ref([])

async function fetchData() {
  const { data, total } = await GetUserData(pagination.current, pagination.pageSize)
  tableData.value = data
  pagination.total = total
}

fetchData()

</script>
