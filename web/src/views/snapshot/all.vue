<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useDialog, useMessage, NModal } from 'naive-ui'
import Search from './components/Search.vue'
import { fetchSnapshotAll, fetchSnapshotDelete } from '@/api'
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

const needPermission = computed(() => !authStore.isValid)

onMounted(refreshSnapshot)

function postUrl(uuid: string): string {
  return `#/snapshot/${uuid}`
}

async function refreshSnapshot() {
  try {
    const snapshots = await fetchSnapshotAll()
    postsByYearMonth.value = getSnapshotPostLinks(snapshots)
  } catch (error) {
    console.error('Failed to fetch snapshots:', error)
    // Error handling can be implemented here
  }
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
      <div class="flex items-center ml-10">
        <svg class="w-8 h-8 mr-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
          stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <h1 class="text-xl font-semibold text-gray-900">
          {{ $t('chat_snapshot.title') }}
        </h1>
      </div>
      <div class="mr-10">
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
          class="flex mb-2">
          <h2 class="flex-none w-28 text-lg font-medium">
            {{ yearMonth }}
          </h2>
          <ul>
            <li v-for="post in postsOfYearMonth" :key="post.uuid" class="flex justify-between">
              <div>
                <div class="flex">
                  <time :datetime="post.date" class="mb-1 text-sm font-medium text-gray-600">{{
                    post.date
                  }}</time>
                  <div class="ml-2 text-sm" @click="handleDelete(post)">
                    <SvgIcon icon="ic:baseline-delete-forever" />
                  </div>
                </div>
                <a :href="postUrl(post.uuid)" :title="post.title"
                  class="block text-xl font-semibold text-gray-900 hover:text-blue-600 mb-2">{{ post.title }}</a>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>
