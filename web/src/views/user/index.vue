<script setup lang="ts">
import { NDataTable } from 'naive-ui'

const tableData = [
  {
    user_email: 'example1@example.com',
    total_session: 10,
    total_message: 30,
    total_sessions_3days: 5,
    total_messages_3days: 15,
    ratelimit: '1msg/s',
  },
  {
    user_email: 'example2@example.com',
    total_session: 5,
    total_message: 20,
    total_sessions_3days: 2,
    total_messages_3days: 7,
    ratelimit: '10msg/s',
  },
  {
    user_email: 'example3@example.com',
    total_session: 20,
    total_message: 100,
    total_sessions_3days: 15,
    total_messages_3days: 50,
    ratelimit: '1000msg/s',
  },
]

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
  },
]

const pagination = {
  pageSize: 2,
  pageSizes: [10, 20, 50],
  showSizeChanger: true,
  showQuickJumper: true,
  total: tableData.length,
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
