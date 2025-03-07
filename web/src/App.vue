<script setup lang="ts">
import { NConfigProvider } from 'naive-ui'
import { VueQueryDevtools } from '@tanstack/vue-query-devtools'


const { theme, themeOverrides } = useTheme()
const { language } = useLanguage()
</script>

<template>
  <NConfigProvider
    class="h-full"
    :theme="theme"
    :theme-overrides="themeOverrides"
    :locale="language"
  >
      <RouterView />
  </NConfigProvider>
</template>
