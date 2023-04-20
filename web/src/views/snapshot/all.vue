<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { fetchSnapshotAll } from '@/api'
import { displayLocaleDate } from '@/utils/date'

interface PostLink {
  uuid: string
  date: string
  title: string
}

const posts = ref<PostLink[]>()

onMounted(async () => {
  posts.value = (await fetchSnapshotAll()).map((post: any) => {
    return {
      uuid: post.Uuid,
      date: displayLocaleDate(post.CreatedAt),
      title: post.Title,
    }
  })
})

function post_url(uuid: string): string {
  return `#/snapshot/${uuid}`
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
        <ul class="space-y-4">
          <li v-for="post in posts" :key="post.uuid" class="flex justify-between">
            <div>
              <time :datetime="post.date" class="block mb-2 text-sm font-medium text-gray-600">{{
                post.date }}</time>
              <a
                :href="post_url(post.uuid)" :title="post.title"
                class="block text-xl font-semibold text-gray-900 hover:text-blue-600"
              >{{
                post.title }}</a>
            </div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
