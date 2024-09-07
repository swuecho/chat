<script lang="ts" setup>
import { computed, ref, onMounted } from 'vue'
import { NButton, NInput, useMessage } from 'naive-ui'
import type { Theme } from '@/store/modules/app/helper'
import { SvgIcon } from '@/components/common'
import { useAppStore, useUserStore } from '@/store'
import type { UserInfo } from '@/store/modules/user/helper'
import { t } from '@/locales'
import request from '@/utils/request/axios'


const appStore = useAppStore()
const userStore = useUserStore()

const ms = useMessage()

const theme = computed(() => appStore.theme)

const userInfo = computed(() => userStore.userInfo)

const name = ref(userInfo.value.name || t('setting.defaultName'))

const description = ref(userInfo.value.description || t('setting.defaultDesc'))

const themeOptions: { label: string; key: Theme; icon: string }[] = [
  {
    label: 'Auto',
    key: 'auto',
    icon: 'ri:contrast-line',
  },
  {
    label: 'Light',
    key: 'light',
    icon: 'ri:sun-foggy-line',
  },
  {
    label: 'Dark',
    key: 'dark',
    icon: 'ri:moon-foggy-line',
  },
]

const apiToken = ref('')

function copyToClipboard() {
  if (apiToken.value) {
    navigator.clipboard.writeText(apiToken.value)
      .then(() => ms.success(t('setting.apiTokenCopied')))
      .catch(() => ms.error(t('setting.apiTokenCopyFailed')))
  }
}

onMounted(async () => {
  try {
    const response = await request.get('/token_10years')
    console.log(response)
    if (response.status=== 200) {
      apiToken.value = response.data.accessToken
    }
    else {
      ms.error('Failed to fetch API token')
    }
  } catch (error) {
    ms.error('Error fetching API token')
  }
})

function updateUserInfo(options: Partial<UserInfo>) {
  userStore.updateUserInfo(options)
  ms.success(t('common.success'))
}
</script>

<template>
  <div class="p-4 space-y-5 min-h-[200px]">
    <div class="space-y-6">
      <div class="flex items-center space-x-4">
        <span class="flex-shrink-0 w-[100px]">{{ $t('setting.name') }}</span>
        <div class="w-[200px]">
          <NInput v-model:value="name" placeholder="" />
        </div>
        <NButton size="tiny" text type="primary" @click="updateUserInfo({ name })">
          {{ $t('common.save') }}
        </NButton>
      </div>
      <div class="flex items-center space-x-4">
        <span class="flex-shrink-0 w-[100px]">{{ $t('setting.description') }}</span>
        <div class="flex-1">
          <NInput v-model:value="description" placeholder="" />
        </div>
        <NButton size="tiny" text type="primary" @click="updateUserInfo({ description })">
          {{ $t('common.save') }}
        </NButton>
      </div>
      <div class="flex items-center space-x-4">
        <span class="flex-shrink-0 w-[100px]">{{ $t('setting.theme') }}</span>
        <div class="flex flex-wrap items-center gap-4">
          <template v-for="item of themeOptions" :key="item.key">
            <NButton size="small" :type="item.key === theme ? 'primary' : undefined"
              @click="appStore.setTheme(item.key)">
              <template #icon>
                <SvgIcon :icon="item.icon" />
              </template>
            </NButton>
          </template>
        </div>
      </div>
      <div class="flex items-center space-x-4">
        <span class="flex-shrink-0 w-[100px]">{{ $t('setting.snapshotLink') }}</span>
        <div class="w-[200px]">
          <a href="/#/snapshot_all" target="_blank" class="text-blue-500"> 点击打开 </a>
        </div>
      </div>
      <div class="flex items-center space-x-4">
        <span class="flex-shrink-0 w-[100px]">{{ $t('setting.apiToken') }}</span>
        <div class="flex-1">
          <NInput v-model:value="apiToken" readonly @click="copyToClipboard" />
        </div>
      </div>
    </div>
  </div>
</template>
