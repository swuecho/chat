<script setup lang="ts">
import type { CSSProperties, Component, Ref } from 'vue'
import { computed, h, onMounted, reactive, ref, watch } from 'vue'
import { NIcon, NLayout, NLayoutSider, NMenu } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { KeyOutline, PulseOutline, ShieldCheckmarkOutline, TextOutline } from '@vicons/ionicons5'
import { RouterLink, useRoute } from 'vue-router'
import Permission from '@/views/components/Permission.vue'
import { t } from '@/locales'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { HoverButton, SvgIcon } from '@/components/common'
import { useAuthStore } from '@/store'

const { isMobile } = useBasicLayout()

const authStore = useAuthStore()

// Initialize auth state on component mount (async)
onMounted(async () => {
  await authStore.initializeAuth()
})

// login modal will appear when there is no token and auth is initialized (but not during initialization)
const currentRoute = useRoute()
const USER_ROUTE = 'AdminUser'
const MODEL_ROUTE = 'AdminModel'
const TITLE_MODEL_ROUTE = 'TitleModel'
const MODELRATELIMIT_ROUTUE = 'ModelRateLimit'
const API_KEYS_ROUTE = 'AdminApiKeys'

const needPermission = computed(() => authStore.isInitialized && !authStore.isInitializing && !authStore.isValid)

const collapsed: Ref<boolean> = ref(isMobile.value)
const activeKey = ref(currentRoute.name?.toString())

const pageDetails = computed(() => {
  const details: Record<string, { title: string; description: string }> = {
    [USER_ROUTE]: {
      title: t('admin.userMessage'),
      description: t('admin.shell.usersDescription'),
    },
    [API_KEYS_ROUTE]: {
      title: t('admin.shell.apiKeys'),
      description: t('admin.shell.apiKeysDescription'),
    },
    [MODEL_ROUTE]: {
      title: t('admin.model'),
      description: t('admin.shell.modelsDescription'),
    },
    [TITLE_MODEL_ROUTE]: {
      title: t('admin.chat_model.title_model'),
      description: t('admin.shell.titleModelDescription'),
    },
    [MODELRATELIMIT_ROUTUE]: {
      title: t('admin.rateLimit'),
      description: t('admin.shell.rateLimitsDescription'),
    },
  }

  return details[currentRoute.name?.toString() || ''] || {
    title: t('admin.title'),
    description: t('admin.shell.defaultDescription'),
  }
})

watch(() => currentRoute.name, (name) => {
  activeKey.value = name?.toString()
  if (isMobile.value)
    collapsed.value = true
})

const getMobileClass = computed<CSSProperties>(() => {
  if (isMobile.value) {
    return {
      position: 'fixed',
      top: '0',
      left: '0',
      height: '100vh',
      zIndex: 50,
    }
  }
  return {}
})

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuOptions: MenuOption[] = reactive([
  {
    label:
      () =>
        h(
          RouterLink,
          {
            to: {
              name: USER_ROUTE,
            },
          },
          { default: () => t('admin.userMessage') },
        ),
    key: USER_ROUTE,
    icon: renderIcon(PulseOutline),
  },
  {
    label: () => h(
      RouterLink,
      { to: { name: API_KEYS_ROUTE } },
      { default: () => t('admin.shell.apiKeys') },
    ),
    key: API_KEYS_ROUTE,
    icon: renderIcon(KeyOutline),
  },
  {
    label: () => h(
      RouterLink,
      {
        to: {
          name: MODEL_ROUTE,
        },
      },
      { default: () => t('admin.model') },
    ),
    key: MODEL_ROUTE,
    icon: renderIcon(ShieldCheckmarkOutline),
  },
  {
    label: () => h(
      RouterLink,
      { to: { name: TITLE_MODEL_ROUTE } },
      { default: () => t('admin.chat_model.title_model') },
    ),
    key: TITLE_MODEL_ROUTE,
    icon: renderIcon(TextOutline),
  },
  {
    label: () => h(
      RouterLink,
      {
        to: {
          name: MODELRATELIMIT_ROUTUE,
        },
      },
      { default: () => t('admin.rateLimit') },
    ),
    key: MODELRATELIMIT_ROUTUE,
    icon: renderIcon(KeyOutline),
  },
])

function handleUpdateCollapsed() {
  collapsed.value = !collapsed.value
}

const mobileOverlayClass = computed(() => {
  if (isMobile.value && !collapsed.value)
    return 'fixed inset-0 bg-black/20 z-40'

  return 'hidden'
})

function handleChatHome() {
  window.location.href = '/'
}
</script>

<template>
  <div class="admin-shell h-full flex flex-col" :class="getMobileClass">
    <header
      v-if="isMobile"
      class="admin-mobile-header sticky flex flex-shrink-0 items-center justify-between overflow-hidden h-14 z-30"
    >
      <div class="flex items-center gap-3">
        <button class="admin-icon-button flex items-center justify-center ml-4" aria-label="Toggle navigation" @click="handleUpdateCollapsed">
          <SvgIcon v-if="collapsed" class="text-2xl" icon="ri:align-justify" />
          <SvgIcon v-else class="text-2xl" icon="ri:align-right" />
        </button>
        <div>
          <div class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('admin.title') }}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400">
            {{ pageDetails.title }}
          </div>
        </div>
      </div>
      <HoverButton class="mr-4" @click="handleChatHome">
        <span class="text-xl text-[#4f555e] dark:text-white">
          <SvgIcon icon="ic:baseline-home" />
        </span>
      </HoverButton>
    </header>
    <div :class="mobileOverlayClass" @click="collapsed = true" />
    <NLayout has-sider class="admin-layout flex-1 overflow-y-auto">
      <NLayoutSider
        class="admin-sidebar" collapse-mode="width" :width="isMobile ? 256 : 220" :collapsed="collapsed"
        :collapsed-width="isMobile ? 0 : 56" :show-trigger="isMobile ? false : 'arrow-circle'" :style="getMobileClass"
        @collapse="collapsed = true" @expand="collapsed = false"
      >
        <div v-if="!collapsed" class="admin-brand">
          <div class="admin-brand-mark">
            <SvgIcon icon="material-symbols:admin-panel-settings-outline-rounded" />
          </div>
          <div>
            <div class="admin-brand-title">
              {{ t('admin.shell.administration') }}
            </div>
            <div class="admin-brand-caption">
              {{ t('admin.shell.controlCenter') }}
            </div>
          </div>
        </div>
        <div v-else-if="!isMobile" class="admin-brand admin-brand--collapsed">
          <div class="admin-brand-mark">
            <SvgIcon icon="material-symbols:admin-panel-settings-outline-rounded" />
          </div>
        </div>
        <div v-if="!collapsed" class="admin-menu-label">
          {{ t('admin.shell.manage') }}
        </div>
        <NMenu v-model:value="activeKey" class="admin-menu" :collapsed="collapsed" :collapsed-icon-size="20" :options="menuOptions" />
        <button v-if="!collapsed" class="admin-home-link" @click="handleChatHome">
          <SvgIcon class="text-lg" icon="ic:round-arrow-back" />
          <span>{{ t('admin.shell.backToChat') }}</span>
        </button>
      </NLayoutSider>
      <NLayout class="admin-content-layout" :style="isMobile && !collapsed ? 'pointer-events: none' : ''">
        <div class="flex flex-col h-full">
          <div v-if="!isMobile" class="admin-page-header">
            <div>
              <h1>{{ pageDetails.title }}</h1>
              <p>{{ pageDetails.description }}</p>
            </div>
            <button class="admin-chat-button" @click="handleChatHome">
              <SvgIcon icon="ic:round-home" />
              <span>{{ t('admin.shell.openChat') }}</span>
            </button>
          </div>
          <div class="admin-content flex-1">
            <div class="admin-content-inner">
              <router-view />
            </div>
          </div>
        </div>
        <Permission :visible="needPermission" />
      </NLayout>
    </NLayout>
  </div>
</template>

<style scoped>
.admin-shell { background: #f6f7f9; }
.admin-layout, .admin-content-layout { background: transparent; }
.admin-sidebar { position: relative; border-right: 1px solid #e8eaee; background: #fff; }
.admin-brand { height: 68px; display: flex; align-items: center; gap: 10px; padding: 0 16px; border-bottom: 1px solid #f0f1f3; }
.admin-brand--collapsed { justify-content: center; padding: 0; }
.admin-brand-mark { width: 32px; height: 32px; display: grid; place-items: center; flex: none; border-radius: 9px; color: #fff; font-size: 19px; background: #262b33; }
.admin-brand-title { color: #17191d; font-size: 13px; font-weight: 650; line-height: 1.3; }
.admin-brand-caption { margin-top: 2px; color: #8b919a; font-size: 11px; }
.admin-menu-label { padding: 15px 18px 5px; color: #9a9fa8; font-size: 9px; font-weight: 700; letter-spacing: .12em; text-transform: uppercase; }
.admin-menu { padding: 0 8px; }
.admin-menu :deep(.n-menu-item-content) { height: 38px; margin: 2px 0; border-radius: 8px; font-size: 13px; }
.admin-menu :deep(.n-menu-item-content--selected) { background: #f0f2f5; }
.admin-menu :deep(.n-menu-item-content--selected::before) { display: none; }
.admin-menu :deep(.n-menu-item-content--selected .n-menu-item-content-header),
.admin-menu :deep(.n-menu-item-content--selected .n-icon) { color: #20242a; font-weight: 600; }
.admin-home-link { position: absolute; right: 12px; bottom: 14px; left: 12px; display: flex; align-items: center; gap: 8px; padding: 8px 10px; border-radius: 8px; color: #666d77; font-size: 12px; transition: background .18s ease, color .18s ease; }
.admin-home-link:hover { color: #20242a; background: #f3f4f6; }
.admin-page-header { min-height: 78px; display: flex; align-items: center; justify-content: space-between; gap: 20px; padding: 15px 24px; border-bottom: 1px solid #e8eaee; background: rgba(255,255,255,.82); backdrop-filter: blur(12px); }
.admin-page-header h1 { margin: 0; color: #17191d; font-size: 19px; font-weight: 650; letter-spacing: -.02em; }
.admin-page-header p { margin: 3px 0 0; color: #7a8089; font-size: 12px; }
.admin-chat-button { display: flex; align-items: center; gap: 6px; padding: 7px 10px; border: 1px solid #dfe2e6; border-radius: 8px; color: #555b64; background: #fff; font-size: 12px; transition: border-color .18s ease, color .18s ease, box-shadow .18s ease; }
.admin-chat-button:hover { color: #1f2329; border-color: #c8ccd2; box-shadow: 0 2px 8px rgba(27,31,36,.06); }
.admin-content { overflow: auto; padding: 18px 24px 28px; }
.admin-content-inner { width: 100%; max-width: 1440px; margin: 0 auto; padding: 16px; border: 1px solid #e8eaee; border-radius: 12px; background: #fff; box-shadow: 0 1px 2px rgba(20,24,28,.03); }
.admin-content-inner :deep(> div > h1),
.admin-content-inner :deep(> div > div:first-child > h1) { display: none; }
.admin-mobile-header { border-bottom: 1px solid #e8eaee; background: rgba(255,255,255,.92); backdrop-filter: blur(12px); }
.admin-icon-button { width: 34px; height: 34px; border-radius: 8px; color: #4d535c; }
.admin-icon-button:hover { background: #f0f2f4; }
.admin-shell :deep(button:focus-visible), .admin-shell :deep(a:focus-visible), .admin-shell :deep(input:focus-visible) { outline: 2px solid #4f7cff; outline-offset: 2px; }

:global(.dark) .admin-shell { background: #111315; }
:global(.dark) .admin-sidebar { border-color: #292d32; background: #191c20; }
:global(.dark) .admin-brand { border-color: #292d32; }
:global(.dark) .admin-brand-mark { color: #20242a; background: #f1f3f5; }
:global(.dark) .admin-brand-title,
:global(.dark) .admin-page-header h1 { color: #f3f4f6; }
:global(.dark) .admin-menu :deep(.n-menu-item-content--selected) { background: #292d32; }
:global(.dark) .admin-menu :deep(.n-menu-item-content--selected .n-menu-item-content-header),
:global(.dark) .admin-menu :deep(.n-menu-item-content--selected .n-icon) { color: #fff; }
:global(.dark) .admin-home-link:hover { color: #fff; background: #25292e; }
:global(.dark) .admin-page-header,
:global(.dark) .admin-mobile-header { border-color: #292d32; background: rgba(25,28,32,.88); }
:global(.dark) .admin-chat-button { color: #c5c9cf; border-color: #34383e; background: #202328; }
:global(.dark) .admin-content-inner { border-color: #292d32; background: #191c20; box-shadow: none; }

@media (max-width: 767px) {
  .admin-content { padding: 10px; }
  .admin-content-inner { padding: 12px; border-radius: 10px; }
}
</style>
