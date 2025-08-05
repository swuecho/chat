<script setup lang="ts">
import { computed, onMounted, ref, h } from 'vue'
import { NModal, useDialog, useMessage } from 'naive-ui'
import copy from 'copy-to-clipboard'
import Search from '../snapshot/components/Search.vue'
import { fetchChatbotAll, fetchSnapshotDelete } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { generateAPIHelper, getBotPostLinks } from '@/service/snapshot'
import { fetchAPIToken } from '@/api/token'
import { fetchBotRunCount } from '@/api/bot_answer_history'
import { t } from '@/locales'
import { useAuthStore } from '@/store'
import Permission from '@/views/components/Permission.vue'
const authStore = useAuthStore()

const dialog = useDialog()
const message = useMessage()

const searchVisible = ref(false)
const apiToken = ref('')

const needPermission = computed(() => !authStore.isValid)

const postsByYearMonth = ref<Record<string, Snapshot.PostLink[]>>({})
const botRunCounts = ref<Record<string, number>>({})

onMounted(async () => {
  await refreshSnapshot()
  const data = await fetchAPIToken()
  apiToken.value = data.accessToken
})


async function refreshSnapshot() {
  const bots: Snapshot.Snapshot[] = await fetchChatbotAll()
  postsByYearMonth.value = getBotPostLinks(bots)
  
  // Fetch run counts for all bots
  const runCountPromises = bots.map(async (bot) => {
    try {
      const count = await fetchBotRunCount(bot.uuid)
      return { uuid: bot.uuid, count }
    } catch (error) {
      console.warn(`Failed to fetch run count for bot ${bot.uuid}:`, error)
      return { uuid: bot.uuid, count: 0 }
    }
  })
  
  const runCounts = await Promise.all(runCountPromises)
  botRunCounts.value = runCounts.reduce((acc, { uuid, count }) => {
    acc[uuid] = count
    return acc
  }, {} as Record<string, number>)
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
      }
      catch (error: any) {
        message.error(t('chat_snapshot.deleteFailed'))
      }
    },
  })
}


function handleShowCode(post: Snapshot.PostLink) {
  const code = generateAPIHelper(post.uuid, apiToken.value, window.location.origin)
  const dialogBox = dialog.info({
    title: t('bot.showCode'),
    content: () => h('code', { class: 'whitespace-pre-wrap' }, code),
    positiveText: t('common.copy'),
    onPositiveClick: () => {
      const success = copy(code)
      if (success) {
        message.success(t('common.success'))
      } else {
        message.error(t('common.copyFailed'))
      }
      dialogBox.loading = false
    },
  })
}


function postUrl(uuid: string): string {
  return `#/bot/${uuid}`
}

function copyToClipboard(text: string) {
  const success = copy(text)
  if (success) {
    message.success(t('common.success'))
  } else {
    message.error(t('common.copyFailed'))
  }
}


</script>

<template>
  <div class="flex flex-col w-full h-full dark:text-white">
    <header
      class="flex items-center justify-between h-16 z-30 border-b dark:border-neutral-800 bg-white/80 dark:bg-black/20 dark:text-white backdrop-blur">
      <div class="flex items-center ml-10 gap-2">
        <SvgIcon icon="majesticons:robot-line" class="w-6 h-6" />
        <h1 class="text-xl font-semibold text-gray-900">
          {{ $t('bot.all.title') }}
        </h1>
      </div>
      <div class="mr-10">
        <HoverButton @click="searchVisible = true">
          <SvgIcon icon="ic:round-search" class="text-2xl" />
        </HoverButton>
        <NModal v-model:show="searchVisible" preset="dialog">
          <Search />
        </NModal>
      </div>
    </header>
    <Permission :visible="needPermission" />
    <div id="scrollRef" ref="scrollRef" class="h-full overflow-hidden overflow-y-auto">
      <div class="max-w-screen-xl px-4 py-8 mx-auto">
        <div v-for="[yearMonth, postsOfYearMonth] in Object.entries(postsByYearMonth)" :key="yearMonth"
          class="flex flex-col md:flex-row mb-4">
          <h2 class="flex-none w-28 text-lg font-medium mb-2 md:sticky top-8 self-start">
            {{ yearMonth }}
          </h2>
          <div class="w-full grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <div v-for="post in postsOfYearMonth" :key="post.uuid" 
              class="group bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-5 hover:shadow-md hover:border-gray-300 dark:hover:border-gray-600 transition-all duration-200 cursor-pointer"
              @click="window.open(postUrl(post.uuid), '_blank')">
              
              <!-- Header with date and actions -->
              <div class="flex justify-between items-start mb-3">
                <div class="flex items-center gap-2">
                  <div class="p-1.5 bg-blue-50 dark:bg-blue-900/30 rounded-lg">
                    <SvgIcon icon="majesticons:robot-line" class="w-4 h-4 text-blue-600 dark:text-blue-400" />
                  </div>
                  <time :datetime="post.date" class="text-sm text-gray-500 dark:text-gray-400 font-medium">
                    {{ post.date }}
                  </time>
                </div>
                <div class="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button @click.stop="handleShowCode(post)" 
                    class="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded-lg transition-all"
                    :title="t('bot.showCode')">
                    <SvgIcon icon="ic:outline-code" class="w-4 h-4" />
                  </button>
                  <button @click.stop="handleDelete(post)" 
                    class="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-all"
                    :title="t('common.delete')">
                    <SvgIcon icon="ic:baseline-delete-forever" class="w-4 h-4" />
                  </button>
                </div>
              </div>

              <!-- Bot title -->
              <div class="mb-4">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors" 
                    style="display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;">
                  {{ post.title }}
                </h3>
              </div>

              <!-- Statistics and metadata -->
              <div class="space-y-3">
                <!-- Run count with visual indicator -->
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class="flex items-center gap-1.5 px-2.5 py-1 bg-gray-50 dark:bg-gray-700 rounded-full">
                      <SvgIcon icon="ic:baseline-play-arrow" class="w-3.5 h-3.5 text-green-600 dark:text-green-400" />
                      <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                        {{ botRunCounts[post.uuid] || 0 }}
                      </span>
                      <span class="text-xs text-gray-500 dark:text-gray-400">
                        {{ (botRunCounts[post.uuid] || 0) === 1 ? 'run' : 'runs' }}
                      </span>
                    </div>
                  </div>
                  
                  <!-- Activity indicator -->
                  <div class="flex items-center gap-1">
                    <div :class="[
                      'w-2 h-2 rounded-full',
                      (botRunCounts[post.uuid] || 0) > 10 ? 'bg-green-500' : 
                      (botRunCounts[post.uuid] || 0) > 0 ? 'bg-yellow-500' : 'bg-gray-300 dark:bg-gray-600'
                    ]"></div>
                    <span class="text-xs text-gray-500 dark:text-gray-400">
                      {{ (botRunCounts[post.uuid] || 0) > 10 ? 'High activity' : 
                         (botRunCounts[post.uuid] || 0) > 0 ? 'Active' : 'Inactive' }}
                    </span>
                  </div>
                </div>

                <!-- UUID (shortened) -->
                <div class="flex items-center gap-2 text-xs text-gray-400 dark:text-gray-500">
                  <SvgIcon icon="ic:baseline-fingerprint" class="w-3.5 h-3.5" />
                  <span class="font-mono">{{ post.uuid.slice(0, 8) }}...</span>
                  <button @click.stop="copyToClipboard(post.uuid)" 
                    class="p-0.5 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
                    :title="t('common.copy')">
                    <SvgIcon icon="ic:baseline-content-copy" class="w-3 h-3" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
