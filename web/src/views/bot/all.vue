<script setup lang="ts">
import { onMounted, ref, h } from 'vue'
import { NModal, useDialog, useMessage } from 'naive-ui'
import Search from '../snapshot/components/Search.vue'
import { fetchSnapshotAll, fetchSnapshotDelete } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { generateAPIHelper, getBotPostLinks } from '@/service/snapshot'
import request from '@/utils/request/axios'
import { t } from '@/locales'
const dialog = useDialog()
const nui_msg = useMessage()

const search_visible = ref(false)
const apiToken = ref('')

const postsByYearMonth = ref<Record<string, Snapshot.PostLink[]>>({})

onMounted(async () => {
  await refreshSnapshot()
  try {
    const response = await request.get('/token_10years')
    if (response.status === 200) {
      apiToken.value = response.data.accessToken
    }
    else {
      nui_msg.error('Failed to fetch API token')
    }
  } catch (error) {
    nui_msg.error('Error fetching API token')
  }
})


async function refreshSnapshot() {
  const bots: Snapshot.Snapshot[] = await fetchSnapshotAll()
  postsByYearMonth.value = getBotPostLinks(bots)
}

function handleDelete(post: Snapshot.PostLink) {
  const dialogBox = dialog.warning({
    title: t('chat_snapshot.deletePost'),
    content: post.title,
    positiveText: t('common.yes'),
    negativeText: t('common.no'),
    onPositiveClick: async () => {
      try {
        dialogBox.loading = true
        await fetchSnapshotDelete(post.uuid)
        await refreshSnapshot()
        dialogBox.loading = false
        nui_msg.success(t('chat_snapshot.deleteSuccess'))
        Promise.resolve()
      }
      catch (error: any) {
        nui_msg.error(t('chat_snapshot.deleteFailed'))
      }
      finally {
        dialogBox.loading = false
      }
    },
  })
}


function handleShowCode(post: Snapshot.PostLink) {
  const dialogBox = dialog.info({
    title: t('bot.showCode'),
    content: () => h('code', { class: 'whitespace-pre-wrap' }, genAPIHelper(post)),
    positiveText: t('common.copy'),
    onPositiveClick: () => {
      // copy to clipboard
      navigator.clipboard.writeText(genAPIHelper(post))
      dialogBox.loading = false
      nui_msg.success(t('common.success'))
    },
  })
}


function post_url(uuid: string): string {
        return `#/bot/${uuid}`
}

function genAPIHelper(post: Snapshot.PostLink) {
  return generateAPIHelper(post.uuid, apiToken.value, window.location.origin)
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
          {{ $t('bot.all.title') }}
        </h1>
      </div>
      <div class="mr-10">
        <HoverButton @click="search_visible = true">
          <SvgIcon icon="ic:round-search" class="text-2xl" />
        </HoverButton>
        <NModal v-model:show="search_visible" preset="dialog">
          <Search />
        </NModal>
      </div>
    </header>
    <div id="scrollRef" ref="scrollRef" class="h-full overflow-hidden overflow-y-auto">
      <div class="max-w-screen-xl px-4 py-8 mx-auto">
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
                  <div class="ml-4 cursor-pointer" @click="handleShowCode(post)">
                    <SvgIcon icon="ic:outline-code" />
                  </div>
                </div>
                <a :href="post_url(post.uuid)" :title="post.title"
                  class="block text-xl font-semibold text-gray-900 hover:text-blue-600 mb-2">{{ post.title }}</a>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>
