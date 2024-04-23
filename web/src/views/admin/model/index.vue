<script setup lang="ts">
import { NTabPane, NTabs } from 'naive-ui'
import { ref } from 'vue'
import PerModelRatelimit from './per_mode_ratelimit/index.vue'
import SystemModel from './system_model/index.vue'
import { SvgIcon } from '@/components/common'
const active = ref('System')
</script>

<template>
  <NTabs v-model:value="active" type="line" animated class="ml-5">
    <NTabPane name="System" tab="System">
      <template #tab>
        <SvgIcon class="text-lg" icon="ri:file-user-line" />
        <span class="ml-2">{{ $t('admin.system_model_tab_title') }}</span>
      </template>
      <div class="min-h-[100px]">
        <SystemModel />
      </div>
    </NTabPane>
    <NTabPane name="Admin" tab="Admin">
      <template #tab>
        <SvgIcon class="text-lg" icon="ri:list-settings-line" />
        <span class="ml-2">{{ $t('admin.per_model_rate_limit_title') }}</span>
      </template>
      <PerModelRatelimit />
    </NTabPane>
  </NTabs>
</template>
