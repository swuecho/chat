<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { NButton, NInput } from 'naive-ui'

const results = ref<any[]>(
  [
    {
      url: 'https://oaidalleapiprodscus.blob.core.windows.net/private/org-bOYQvh4o4xsvbnEMtJ9SvHXC/user-MN62B56mPml5Mj2xeVaC3uxS/img-MQPI4E6ytIW3iQBaus9HXSaD.png?st=2023-04-26T06%3A45%3A17Z&se=2023-04-26T08%3A45%3A17Z&sp=r&sv=2021-08-06&sr=b&rscd=inline&rsct=image/png&skoid=6aaadede-4fb3-4698-a8f6-684d7786b067&sktid=a48cca56-e6da-484e-a814-9c849652bcb3&skt=2023-04-26T07%3A43%3A23Z&ske=2023-04-27T07%3A43%3A23Z&sks=b&skv=2021-08-06&sig=QWhBTFGZvp47JyM205ryVj80GLDVAInSVGbjm847OuA%3D',
    },
    {
      url: 'https://oaidalleapiprodscus.blob.core.windows.net/private/org-bOYQvh4o4xsvbnEMtJ9SvHXC/user-MN62B56mPml5Mj2xeVaC3uxS/img-kKQwnTfJc6MJfwxtlDfdVokR.png?st=2023-04-26T06%3A45%3A17Z&se=2023-04-26T08%3A45%3A17Z&sp=r&sv=2021-08-06&sr=b&rscd=inline&rsct=image/png&skoid=6aaadede-4fb3-4698-a8f6-684d7786b067&sktid=a48cca56-e6da-484e-a814-9c849652bcb3&skt=2023-04-26T07%3A43%3A23Z&ske=2023-04-27T07%3A43%3A23Z&sks=b&skv=2021-08-06&sig=G4IiH917yQKODcSUE1uGzon/4yVbU3q7UvLGGrJ0WTw%3D',
    },
  ],
)
const prompt = ref('')
const loading = ref(false)

const url = ref<string>('https://api.openai.com/v1/images/generations')

// todo sync the token to localstorage

const token = ref('')

onMounted(() => {
  token.value = localStorage.getItem('token') || ''
  prompt.value = localStorage.getItem('prompt') || ''
  url.value = localStorage.getItem('url') || ''
})

watch(token, (newVal) => {
  localStorage.setItem('token', newVal)
})
watch(prompt, (newVal) => {
  localStorage.setItem('prompt', newVal)
})
watch(url, (newVal) => {
  localStorage.setItem('url', newVal)
})

async function search() {
  loading.value = true
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token.value}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      prompt: prompt.value,
      n: 2,
      size: '256x256',
    }),
  })
  results.value = (await res.json()).data
  loading.value = false
}

onMounted(() => {
  // search()
})
</script>

<template>
  <div class="flex flex-col w-full h-full">
    <div class="flex mt-10 mx-20">
      <NInput v-model:value="prompt" :autosize="{ minRows: 5, maxRows: 8 }" type="textarea" placeholder="Enter prompt" />
    </div>
    <div class="m-auto">
      <NButton @click="search">
        开始生成
      </NButton>
    </div>
    <div id="image-wrapper" class="m-auto">
      <div v-if="loading">
        Loading...
      </div>
      <div v-for="result in results" :key="result.id">
        <div class="m-5">
          <img :src="result.url">
        </div>
      </div>
    </div>
    <div clas="flex items-center">
      <div>
        <NInput v-model:value="token" placeholder="token" />
      </div>
      <div>
        <NInput v-model:value="url" placeholder="url" />
      </div>
    </div>
  </div>
</template>
