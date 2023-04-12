<script setup lang="ts">
import { NSpin } from 'naive-ui';
</script>

<template>
  <NSpin>
    <div class="p-4 space-y-4">
      <div class="p-2 space-y-2 rounded-md bg-neutral-100 dark:bg-neutral-700">
        <p>
          <a class="text-blue-600 dark:text-blue-500" href="/#/admin/user" target="_blank">
            {{ $t('admin.openPanel') }}
          </a>
        </p>
      </div>
    </div>
  </NSpin>
</template>
