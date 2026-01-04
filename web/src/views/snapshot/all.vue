<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useDialog, useMessage, NModal, NPagination } from 'naive-ui'
import Search from './components/Search.vue'
import { fetchSnapshotAll, fetchSnapshotDelete, fetchSnapshotAllData } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { getSnapshotPostLinks } from '@/service/snapshot'
import { t } from '@/locales'
import { useAuthStore } from '@/store'
import Permission from '@/views/components/Permission.vue'
const dialog = useDialog()
const message = useMessage()
const searchVisible = ref(false)
const postsByYearMonth = ref<Record<string, Snapshot.PostLink[]>>({})
const authStore = useAuthStore()

// Pagination state
const page = ref(1)
const pageSize = ref(20)
const totalCount = ref(0)

const needPermission = authStore.needPermission

onMounted(async() => {
  await authStore.initializeAuth()
  await refreshSnapshot()
})

function postUrl(uuid: string): string {
  return `#/snapshot/${uuid}`
}

async function refreshSnapshot() {
  try {
    const [response, snapshots] = await Promise.all([
      fetchSnapshotAll(page.value, pageSize.value),
      fetchSnapshotAllData(page.value, pageSize.value)
    ])
    postsByYearMonth.value = getSnapshotPostLinks(snapshots)
    // Update total count from response
    totalCount.value = response.total || snapshots.length || 0
  } catch (error) {
    console.error('Failed to fetch snapshots:', error)
    // Error handling can be implemented here
  }
}

function handlePageChange(newPage: number) {
  page.value = newPage
  refreshSnapshot()
}

function handleDelete(post: Snapshot.PostLink) {
  dialog.warning({
    title: t('chat_snapshot.deletePost'),
    content: post.title,
    positiveText: t('common.yes'),
    negativeText: t('common.no'),
    onPositiveClick: async () => {
      try {
        await fetchSnapshotDelete(post.uuid)
        await refreshSnapshot()
        message.success(t('chat_snapshot.deleteSuccess'))
      } catch (error) {
        message.error(t('chat_snapshot.deleteFailed'))
        console.error('Failed to delete snapshot:', error)
      }
    },
  })
}
</script>

<template>
  <div class="flex flex-col w-full h-full dark:text-white">
    <header
      class="flex items-center justify-between h-16 z-30 border-b dark:border-neutral-800 bg-white/80 dark:bg-black/20 dark:text-white backdrop-blur">
      <div class="flex items-center ml-1 md:ml-10 gap-2">
        <svg class="w-6 h-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <h1 class="text-xl font-semibold text-gray-900">
          {{ $t('chat_snapshot.title') }}
        </h1>
      </div>
      <div class="mr-1 md:mr-10">
        <HoverButton @click="searchVisible = true">
          <SvgIcon icon="ic:round-search" class="text-2xl" />
        </HoverButton>

      </div>

    </header>
    <NModal v-model:show="searchVisible" preset="dialog">
      <Search />
    </NModal>
    <div id="scrollRef" ref="scrollRef" class="h-full overflow-hidden overflow-y-auto">
      <Permission :visible="needPermission" />
      <div v-if="!needPermission" class="max-w-screen-xl px-4 py-8 mx-auto">
        <div v-for="[yearMonth, postsOfYearMonth] in Object.entries(postsByYearMonth)" :key="yearMonth"
          class="flex flex-col md:flex-row mb-4 relative">
          <h2 class="flex-none w-28 text-lg font-medium mb-2 md:sticky top-8 self-start">
            {{ yearMonth }}
          </h2>
          <ul class="w-full">
            <li v-for="post in postsOfYearMonth" :key="post.uuid" class="flex justify-between">
              <div>
                <div class="flex items-center">
                  <time :datetime="post.date" class="text-sm font-medium text-gray-600 dark:text-gray-400">{{
                    post.date
                    }}</time>
                  <div class="ml-2 text-sm flex items-center cursor-pointer" @click="handleDelete(post)">
                    <SvgIcon icon="ic:baseline-delete-forever" class="w-5 h-5" />
                  </div>
                </div>
                <a :href="postUrl(post.uuid)" :title="post.title"
                  class="block text-xl font-semibold text-gray-900 dark:text-gray-200 hover:text-blue-600 mb-2">{{
                    post.title }}</a>
              </div>
            </li>
          </ul>
        </div>
        <!-- Pagination Controls -->
        <div v-if="totalCount > 0" class="flex justify-center mt-8 pb-4">
          <NPagination
            v-model:page="page"
            :page-size="pageSize"
            :item-count="totalCount"
            :show-size-picker="true"
            :page-sizes="[10, 20, 30, 50]"
            @update:page="handlePageChange"
            @update:page-size="(newSize: number) => { pageSize = newSize; refreshSnapshot() }"
          />
        </div>
      </div>
    </div>
  </div>
</template>
