<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useDialog, useMessage } from 'naive-ui'
import { fetchSnapshotAll, fetchSnapshotDelete } from '@/api'
import { displayLocaleDate, formatYearMonth } from '@/utils/date'
import { SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { post_url } from '@/utils/url'

const dialog = useDialog()
const nui_msg = useMessage()

interface PostLink {
  uuid: string
  date: string
  title: string
}

const postsByYearMonth = ref<Record<string, PostLink[]>>({})

function postsByYearMonthTransform(posts: PostLink[]) {
  const init: Record<string, PostLink[]> = {}
  return posts.reduce((acc, post) => {
    const yearMonth = formatYearMonth(new Date(post.date))
    if (!acc[yearMonth])
      acc[yearMonth] = []

    acc[yearMonth].push(post)
    return acc
  }, init)
}

async function getPostLinks() {
  const rawPosts = await fetchSnapshotAll()
  const rawPostsFormated = rawPosts.map((post: any) => {
    return {
      uuid: post.Uuid,
      date: displayLocaleDate(post.CreatedAt),
      title: post.Title,
    }
  })
  return postsByYearMonthTransform(rawPostsFormated)
}

onMounted(async () => {
  postsByYearMonth.value = await getPostLinks()
})

function handleDelete(post: PostLink) {
  const dialogBox = dialog.warning({
    title: t('chat_snapshot.deletePost'),
    content: t('chat_snaport.deletePostConfirm'),
    positiveText: t('common.yes'),
    negativeText: t('common.no'),
    onPositiveClick: async () => {
      try {
        dialogBox.loading = true
        await fetchSnapshotDelete(post.uuid)
        postsByYearMonth.value = await getPostLinks()
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
</script>

<template>
  <div class="flex flex-col w-full h-full">
    <header class="flex items-center justify-between p-4">
      <div class="flex items-center">
        <svg
          class="w-8 h-8 mr-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
        <h1 class="text-2xl font-semibold text-gray-900">
          {{ $t('chat_snapshot.title') }}
        </h1>
      </div>
    </header>
    <div id="scrollRef" ref="scrollRef" class="h-full overflow-hidden overflow-y-auto">
      <div class="max-w-screen-xl px-4 py-8 mx-auto">
        <div v-for="[yearMonth, postsOfYearMonth] in Object.entries(postsByYearMonth)" :key="yearMonth" class="flex mb-2">
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
                <a
                  :href="post_url(post.uuid)" :title="post.title"
                  class="block text-xl font-semibold text-gray-900 hover:text-blue-600 mb-2"
                >{{ post.title }}</a>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>
