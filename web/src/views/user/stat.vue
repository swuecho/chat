// create a data table with pagination using naive-ui, with the following columns:
// User Email, Total Sessions, Total Messages, Total Sessions (3 days), Total Messages (3 days), Rate Limit
// The data should be fetched from the backend using api 'GetUserData'
// The Rate Limit column should be editable, and the value should be updated in the backend using api 'UpdateRateLimit(user_email, rate_limit)'
// vue3 code should be in <script lang="ts" setup> style.

<template>
  <n-table :data="tableData" :columns="columns" :pagination="{ pageSize: 10 }">
    <template #cell(rateLimit)="scope">
      <n-input-number v-model:value="scope.row.rateLimit" @change="updateRateLimit(scope.row.userEmail, scope.row.rateLimit)" />
    </template>
  </n-table>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { NTable, NInputNumber } from 'naive-ui'
import { GetUserData, UpdateRateLimit } from '@/api/user'

const tableData = ref([])
const columns = [
  {
    title: 'User Email',
    key: 'userEmail'
  },
  {
    title: 'Total Sessions',
    key: 'totalSessions'
  },
  {
    title: 'Total Messages',
    key: 'totalMessages'
  },
  {
    title: 'Total Sessions (3 days)',
    key: 'totalSessions3Days'
  },
  {
    title: 'Total Messages (3 days)',
    key: 'totalMessages3Days'
  },
  {
    title: 'Rate Limit',
    key: 'rateLimit',
    align: 'center'
  }
]

async function fetchData() {
  const { data } = await GetUserData()
  tableData.value = data
}

async function updateRateLimit(userEmail, rateLimit) {
  await UpdateRateLimit(userEmail, rateLimit)
}

fetchData()

</script>
