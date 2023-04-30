<script setup lang="ts">
import { ref } from 'vue'
import { NInput, NList, NListItem } from 'naive-ui'
import { debounce } from 'lodash-es'
import { chatSnapshotSearch } from '@/api'

interface SearchRecord {
  Uuid: string
  Title: string
  Rank: number
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
    <NListItem v-for="result in results" :key="result.Uuid">
      <a :href="`/static/#/snapshot/${result.Uuid}`" target="_blank">{{ result.Title }}</a>
    </NListItem>
  </NList>
</template>
