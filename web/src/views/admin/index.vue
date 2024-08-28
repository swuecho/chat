<script setup lang="ts">
import type { CSSProperties, Component, Ref } from 'vue'
import { computed, h, reactive, ref } from 'vue'
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

// login modal will appear when there is no token
const authStore = useAuthStore()
const currentRoute = useRoute()
const USER_ROUTE = 'AdminUser'
const MODEL_ROUTE = 'AdminModel'
const MODELRATELIMIT_ROUTUE = 'ModelRateLimit'

const needPermission = computed(() => !authStore.token) // || (!!authStore.token && authStore.expiresIn < Date.now() / 1000))

const collapsed: Ref<boolean> = ref(false)
const activeKey = ref(currentRoute.name?.toString())

const getMobileClass = computed<CSSProperties>(() => {
    if (isMobile.value) {
    return {
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

function handleChatHome() {
  window.open('/#/chat/', '_blank')
}


</script>

<template>
    <div class="h-full flex flex-col" :class="getMobileClass">
      <header
        class="sticky flex flex-shrink-0 items-center justify-between  overflow-hidden h-14 z-30 border-b dark:border-neutral-800 bg-white/80 dark:bg-black/20 backdrop-blur">
        <h1 v-if="isMobile"
          class="flex-1 px-4 pr-6 overflow-hidden cursor-pointer select-none text-ellipsis whitespace-nowrap">
          Admin
        </h1>
        <div v-if="isMobile" class="flex items-center">
          <button class="flex items-center justify-center mr-5" @click="handleUpdateCollapsed">
            <SvgIcon v-if="collapsed" class="text-2xl" icon="ri:align-justify" />
            <SvgIcon v-else class="text-2xl" icon="ri:align-right" />
          </button>
        </div>
        <h1 v-if="!isMobile"
          class="flex-1 px-4 pr-6 overflow-hidden cursor-pointer select-none text-ellipsis whitespace-nowrap ml-16">
          Admin
        </h1>
        <HoverButton @click="handleChatHome" class="mr-5">
          <span class="text-xl text-[#4f555e] dark:text-white">
            <SvgIcon icon="ic:baseline-home" />
          </span>
        </HoverButton>
      </header>
        <NLayout has-sider class="flex-1 overflow-y-auto">
          <NLayoutSider bordered collapse-mode="width" :width="240"  :collapsed="collapsed"  :collapsed-width="isMobile ? 0 : 64"
            :show-trigger="isMobile ? false : 'arrow-circle'" :style="getMobileClass" @collapse="collapsed = true"
            @expand="collapsed = false">
            <NMenu v-model:value="activeKey" :collapsed="collapsed"  :collapsed-icon-size="22"
              :options="menuOptions" />
          </NLayoutSider>
          <NLayout>
            <router-view />
            <Permission :visible="needPermission" />
          </NLayout>
        </NLayout>
    </div>
</template>
