<script setup lang='ts'>
import { NTabPane, NTabs } from 'naive-ui'
import UserStat from './UserStat.vue'
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mt-8 mx-4 ">
      {{ $t('admin.title') }}
    </h1>
    <NTabs
      class="card-tabs" default-value="user_stat" size="large" animated style="margin: 0 -4px"
      pane-style="padding-left: 4px; padding-right: 4px; box-sizing: border-box;"
    >
      <NTabPane :label="$t('admin.permission')" name="tab2">
        Content of Tab Pane 2
      </NTabPane>
      <NTabPane :label="$t('admin.rateLimit')" name="user_stat">
        <UserStat />
      </NTabPane>
    </NTabs>
  </div>
</template>

<style scoped>
.card-tabs {
  background-color: #fff;
  border-radius: 4px;
  padding: 16px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
h1 {
  color: #2c3e50;
}
</style>
