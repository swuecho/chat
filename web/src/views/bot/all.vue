<script setup lang="ts">
import { computed, onMounted, ref, h } from 'vue'
import { NModal, useDialog, useMessage } from 'naive-ui'
import Search from '../snapshot/components/Search.vue'
import { fetchSnapshotAll, fetchSnapshotDelete } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { generateAPIHelper, getBotPostLinks } from '@/service/snapshot'
import { fetchAPIToken } from '@/api/token'
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

onMounted(async () => {
  await refreshSnapshot()
  const data = await fetchAPIToken()
  apiToken.value = data.accessToken
})


async function refreshSnapshot() {
  const bots: Snapshot.Snapshot[] = await fetchSnapshotAll()
  postsByYearMonth.value = getBotPostLinks(bots)
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
    onPositiveClick: async () => {
      try {
        // Try modern clipboard API first
        if (navigator.clipboard) {
          await navigator.clipboard.writeText(code)
          message.success(t('common.success'))
          return
        }
        
        // Fallback to textarea method
        const textarea = document.createElement('textarea')
        textarea.value = code
        textarea.style.position = 'fixed' // Avoid scrolling to bottom
        document.body.appendChild(textarea)
        textarea.select()
        
        try {
          const successful = document.execCommand('copy')
          if (!successful) {
            throw new Error('Fallback copy failed')
          }
          message.success(t('common.success'))
        } finally {
          document.body.removeChild(textarea)
        }
      } catch (error) {
        message.error(t('common.copyFailed'))
        console.error('Failed to copy:', error)
      } finally {
        dialogBox.loading = false
      }
    },
  })
}


function postUrl(uuid: string): string {
  return `#/bot/${uuid}`
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
          <div class="w-full grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div v-for="post in postsOfYearMonth" :key="post.uuid" 
              class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-4 hover:shadow-lg transition-shadow">
              <div class="flex justify-between items-start">
                <div>
                  <time :datetime="post.date" class="text-sm text-gray-500 dark:text-gray-400">
                    {{ post.date }}
                  </time>
                  <a :href="postUrl(post.uuid)" 
                    class="block text-lg font-semibold text-gray-900 dark:text-gray-100 hover:text-blue-600 mt-1">
                    {{ post.title }}
                  </a>
                </div>
                <div class="flex space-x-2">
                  <button @click.stop="handleShowCode(post)" 
                    class="p-1 text-gray-500 hover:text-blue-600 transition-colors"
                    :title="t('bot.showCode')">
                    <SvgIcon icon="ic:outline-code" class="w-5 h-5" />
                  </button>
                  <button @click.stop="handleDelete(post)" 
                    class="p-1 text-gray-500 hover:text-red-600 transition-colors"
                    :title="t('common.delete')">
                    <SvgIcon icon="ic:baseline-delete-forever" class="w-5 h-5" />
                  </button>
                </div>
              </div>
              <div class="mt-2">
                <div class="text-xs text-gray-400 break-all">
                  {{ post.uuid }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
