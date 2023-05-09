<script setup lang="ts">
import { ref } from 'vue'
import { NInput, NList, NListItem } from 'naive-ui'
import { debounce } from 'lodash-es'
import { chatSnapshotSearch } from '@/api'

interface SearchRecord {
  uuid: string
  title: string
  rank: number
}

const searchText = ref('')
const results = ref<SearchRecord[]>([])

const search = async () => {
  results.value = await chatSnapshotSearch(searchText.value)
}

const debouncedSearch = debounce(search, 200)
</script>

<template>
  <NInput v-model:value="searchText" placeholder="Search ...(support english only)" @keyup="debouncedSearch" />
  <NList>
    <NListItem v-for="result in results" :key="result.uuid">
      <a :href="`/static/#/snapshot/${result.uuid}`" target="_blank">{{ result.title }}</a>
    </NListItem>
  </NList>
</template>
