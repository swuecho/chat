<script setup lang='ts'>
import { NTabPane, NTabs } from 'naive-ui'
import UserStat from './UserStat.vue'
</script>

<template>
  <div>
    <h1 style="color: #2c3e50; font-size: 36px; font-weight: bold; margin-bottom: 24px;">User management</h1>
    <NTabs
      class="card-tabs" default-value="user_stat" size="large" animated style="margin: 0 -4px"
      pane-style="padding-left: 4px; padding-right: 4px; box-sizing: border-box;"
    >
      <NTabPane label="User Stat" name="user_stat">
        <UserStat />
      </NTabPane>
      <NTabPane name="tab2" label="TODO">
        Content of Tab Pane 2
      </NTabPane>
    </NTabs>
  </div>
</template>

<style scoped>
.card-tabs {
  background-color: #f5f5f5;
  border-radius: 4px;
  padding: 16px;
}
</style>
