<script setup lang="ts">
// create a data table using naive-ui, with the following columns:
// User Email, Total Sessions, Total Messages, Total Sessions (3 days), Total Messages (3 days), Rate Limit
// The data should be fetched from the backend using api 'GetUserData(page, page_size)'
// The Rate Limit column should be editable, and the value should be updated in the backend using api 'UpdateRateLimit(user_email, rate_limit)'
import { NDataTable, NInput } from 'naive-ui'
import { h } from 'vue'
import { fetchUserData } from '@/api/user'

const columns = [
  {
    title: 'User Email',
    key: 'user_email',
  },
  {
    title: 'Total Sessions',
    key: 'total_session',
  },
  {
    title: 'Total Messages',
    key: 'total_message',
  },
  {
    title: 'Total Sessions (3 days)',
    key: 'total_sessions_3days',
  },
  {
    title: 'Total Messages (3 days)',
    key: 'total_messages_3days',
  },
  {
    title: 'Rate Limit',
    key: 'ratelimit',
    render(row, index) {
      return h(NInput, {
        value: row.age,
        onUpdateValue(v) {
          // data.value[index].age = v
          // TOOD: update the value in the backend
        },
      })
    },
    // render: (row: any) => {
    //   return <NInputNumber v-model={[row.ratelimit, 'value']} />
    // },
  },
]

const tableData = fetchUserData()

const pagination = {
  pageSize: 2,
  pageSizes: [10, 20, 50],
  showSizeChanger: true,
  showQuickJumper: true,
  total: 1000,
}
</script>

<template>
  <div class="table-container">
    <NDataTable :data="tableData" :columns="columns" :pagination="pagination" />
  </div>
</template>

<style scoped>
.table-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}
</style>
