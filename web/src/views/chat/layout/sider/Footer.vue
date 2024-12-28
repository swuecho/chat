<script setup lang='ts'>
import { computed, defineAsyncComponent, h, ref, watch } from 'vue'
import { NDropdown } from 'naive-ui'
import { HoverButton, SvgIcon, UserAvatar } from '@/components/common'
import { useAppStore, useAuthStore, useChatStore, useUserStore } from '@/store/modules'
import { isAdmin } from '@/utils/jwt'
import { t } from '@/locales'
const Setting = defineAsyncComponent(() => import('@/components/common/Setting/index.vue'))

const authStore = useAuthStore()
const userStore = useUserStore()
const chatStore = useChatStore()
const appStore = useAppStore()

const show = ref(false)

const isAdminUser = computed(() => isAdmin(authStore.getToken() ?? ''))

function handleLogout() {
  // clear all stores
  authStore.removeToken()
  userStore.resetUserInfo()
  chatStore.clearState()
}

function handleChangelang() {
  appStore.setNextLanguage()
}

function openAdminPanel() {
  window.open('/#/admin/user', '_blank')
}

function openSnapshotAll() {
  window.open('/#/snapshot_all', '_blank')
}

function handleSetting() {
  show.value = true
}

const renderIcon = (icon: string) => {
  return () => h(SvgIcon, {
    class: 'text-xl',
    icon,
  })
}

function handleSelect(key: string) {
  if (key === 'profile')
    handleSetting()
  else if (key === 'language')
    handleChangelang()
  else if (key === 'logout')
    handleLogout()
}

const options = ref<any>([
  {
    label: t('setting.setting'),
    key: 'profile',
    icon: renderIcon('ph:user-circle-light'),
  },
  {
    label: t('setting.language'),
    key: 'language',
    icon: renderIcon('carbon:ibm-watson-language-translator'),
  },
  {
    label: t('common.logout'),
    key: 'logout',
    icon: renderIcon('ri:logout-circle-r-line'),
  },
])

// refresh after lang change
watch(appStore, () => {
  options.value = [
    {
      label: t('setting.setting'),
      key: 'profile',
      icon: renderIcon('ph:user-circle-light'),
    },
    {
      label: t('setting.language'),
      key: 'language',
      icon: renderIcon('carbon:ibm-watson-language-translator'),
    },
    {
      label: t('common.logout'),
      key: 'logout',
      icon: renderIcon('ri:logout-circle-r-line'),
    },
  ]
})
</script>

<template>
  <footer class="flex items-center justify-between min-w-0 p-4 overflow-hidden border-t dark:border-neutral-800">
    <Setting v-if="show" v-model:visible="show" />
    <div class="flex-1 flex-shrink-0 overflow-hidden">
      <UserAvatar />
    </div>
    <HoverButton v-if="isAdminUser" :tooltip="$t('setting.admin')" @click="openAdminPanel">
      <span class="text-xl text-[#4f555e] dark:text-white">
        <SvgIcon icon="eos-icons:admin-outlined" />
      </span>
    </HoverButton>
    <NDropdown :options="options" @select="handleSelect">
      <HoverButton>
        <SvgIcon icon="lucide:more-vertical" />
      </HoverButton>
    </NDropdown>
  </footer>
</template>
