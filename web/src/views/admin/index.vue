<script setup lang="ts">
import type { CSSProperties, Component, Ref } from 'vue'
import { computed, h, reactive, ref, onMounted } from 'vue'
import { NIcon, NLayout, NLayoutSider, NMenu } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { PulseOutline, ShieldCheckmarkOutline, KeyOutline } from '@vicons/ionicons5'
import { RouterLink, useRoute } from 'vue-router'
import Permission from '@/views/components/Permission.vue'
import { t } from '@/locales'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { SvgIcon, HoverButton } from '@/components/common'
import { useAuthStore } from '@/store'

const { isMobile } = useBasicLayout()

// Initialize auth state on component mount (async)
onMounted(async () => {
  console.log('🔄 Admin layout mounted, initializing auth...')
  await authStore.initializeAuth()
  console.log('✅ Auth initialization completed in Admin layout')
})

// login modal will appear when there is no token and auth is initialized (but not during initialization)
const authStore = useAuthStore()
const currentRoute = useRoute()
const USER_ROUTE = 'AdminUser'
const MODEL_ROUTE = 'AdminModel'
const MODELRATELIMIT_ROUTUE = 'ModelRateLimit'

const needPermission = computed(() => authStore.isInitialized && !authStore.isInitializing && !authStore.isValid)

const collapsed: Ref<boolean> = ref(isMobile.value)
const activeKey = ref(currentRoute.name?.toString())

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
  if (isMobile.value && !collapsed.value) {
    return 'fixed inset-0 bg-black/20 z-40'
  }
  return 'hidden'
})

function handleChatHome() {
  window.open('/#/chat/', '_blank')
}


</script>

<template>
  <div class="h-full flex flex-col" :class="getMobileClass">
    <header v-if="isMobile"
      class="sticky flex flex-shrink-0 items-center justify-between overflow-hidden h-14 z-30 border-b dark:border-neutral-800 bg-white/80 dark:bg-black/20 backdrop-blur">
      <div class="flex items-center">
        <button class="flex items-center justify-center ml-4" @click="handleUpdateCollapsed">
          <SvgIcon v-if="collapsed" class="text-2xl" icon="ri:align-justify" />
          <SvgIcon v-else class="text-2xl" icon="ri:align-right" />
        </button>
      </div>
      <div class="flex-1"></div>
      <HoverButton @click="handleChatHome" class="mr-5">
        <span class="text-xl text-[#4f555e] dark:text-white">
          <SvgIcon icon="ic:baseline-home" />
        </span>
      </HoverButton>
    </header>
    <div :class="mobileOverlayClass" @click="collapsed = true"></div>
    <NLayout has-sider class="flex-1 overflow-y-auto">
      <NLayoutSider bordered collapse-mode="width" :width="isMobile ? 280 : 240" :collapsed="collapsed"
        :collapsed-width="isMobile ? 0 : 64" :show-trigger="isMobile ? false : 'arrow-circle'" :style="getMobileClass"
        @collapse="collapsed = true" @expand="collapsed = false">
        <NMenu v-model:value="activeKey" :collapsed="collapsed" :collapsed-icon-size="22" :options="menuOptions" />
      </NLayoutSider>
      <NLayout :style="isMobile && !collapsed ? 'pointer-events: none' : ''">
        <div class="flex flex-col h-full">
          <div class="flex items-center justify-between px-4 md:px-6 lg:px-8 py-3 border-b dark:border-neutral-800" v-if="!isMobile">
            <nav class="flex items-center space-x-2 text-sm">
              <button 
                @click="handleChatHome" 
                class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-200 transition-colors"
              >
                Home
              </button>
              <span class="text-gray-400 dark:text-gray-500">/</span>
              <span class="text-gray-600 dark:text-gray-300">Admin Dashboard</span>
            </nav>
          </div>
          <div class="flex-1 p-4">
            <router-view />
          </div>
        </div>
        <Permission :visible="needPermission" />
      </NLayout>
    </NLayout>
  </div>
</template>
