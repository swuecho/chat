<script setup lang='ts'>
import { computed, defineAsyncComponent, h, ref, watch } from 'vue'
import { NDropdown } from 'naive-ui'
import { HoverButton, SvgIcon, UserAvatar } from '@/components/common'
import { useAppStore, useAuthStore, useUserStore, useMessageStore, useSessionStore, useWorkspaceStore } from '@/store'
import { isAdmin } from '@/utils/jwt'
import { t } from '@/locales'
const Setting = defineAsyncComponent(() => import('@/components/common/Setting/index.vue'))

const authStore = useAuthStore()
const userStore = useUserStore()
const messageStore = useMessageStore()
const sessionStore = useSessionStore()
const workspaceStore = useWorkspaceStore()
const appStore = useAppStore()

const show = ref(false)

const isAdminUser = computed(() => isAdmin(authStore.getToken ?? ''))

function handleLogout() {
  // clear all stores
  authStore.removeToken()
  userStore.resetUserInfo()
  messageStore.clearAllMessages()
  sessionStore.clearWorkspaceSessions(workspaceStore.activeWorkspace?.uuid || '')
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
  <footer class="flex items-center justify-between min-w-0 p-2 overflow-hidden border-t dark:border-neutral-800">
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
      <HoverButton data-testid="config-button">
        <SvgIcon icon="lucide:more-vertical" />
      </HoverButton>
    </NDropdown>
  </footer>
</template>
