<script setup lang="ts">
import type { CSSProperties, Component, Ref } from 'vue'
import { computed, h, reactive, ref } from 'vue'
import { NIcon, NLayout, NLayoutSider, NMenu } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { PulseOutline, ShieldCheckmarkOutline } from '@vicons/ionicons5'
import { RouterLink } from 'vue-router'
import Permission from '@/views/components/Permission.vue'
import { t } from '@/locales'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { SvgIcon } from '@/components/common'
import { useAuthStore } from '@/store'

const { isMobile } = useBasicLayout()

// login modal will appear when there is no token
const authStore = useAuthStore()

const needPermission = computed(() => !authStore.token) // || (!!authStore.token && authStore.expiresIn < Date.now() / 1000))

const collapsed: Ref<boolean> = ref(true)
const activeKey: Ref<string> = ref('rateLimit')

const getMobileClass = computed<CSSProperties>(() => {
  if (isMobile.value) {
    return {
      position: 'fixed',
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
              name: 'AdminUser',
            },
          },
          { default: () => t('admin.rateLimit') },
        ),
    key: 'rateLimit',
    icon: renderIcon(PulseOutline),
  },
  {
    label: () => h(
      RouterLink,
      {
        to: {
          name: 'AdminModel',
        },
      },
      { default: () => t('admin.model') },
    ),
    key: 'model',
    icon: renderIcon(ShieldCheckmarkOutline),
  },
])

function handleUpdateCollapsed() {
  collapsed.value = !collapsed.value
}
</script>

<template>
  <div>
    <div class="h-full dark:bg-[#24272e] transition-all" :class="[isMobile ? 'p-0' : 'p-4']">
      <div class="h-full overflow-hidden" :class="getMobileClass">
        <header
          class="sticky top-0 left-0 right-0 z-30 border-b dark:border-neutral-800 bg-white/80 dark:bg-black/20 backdrop-blur"
        >
          <div class="relative flex items-center justify-between min-w-0 overflow-hidden h-14">
            <h1 v-if="isMobile" class="flex-1 px-4 pr-6 overflow-hidden cursor-pointer select-none text-ellipsis whitespace-nowrap">
              Admin
            </h1>
            <div class="flex items-center">
              <button class="flex items-center justify-center w-11 h-11" @click="handleUpdateCollapsed">
                <SvgIcon v-if="collapsed" class="text-2xl" icon="ri:align-justify" />
                <SvgIcon v-else class="text-2xl" icon="ri:align-right" />
              </button>
            </div>
            <h1  v-if="!isMobile" class="flex-1 px-4 pr-6 overflow-hidden cursor-pointer select-none text-ellipsis whitespace-nowrap">
              Admin
            </h1>
            <!-- <div class="flex items-center space-x-2">
              <HoverButton>
                <span class="text-xl">
                  <SvgIcon icon="ri:chat-history-line" />
                </span>
              </HoverButton>
              <HoverButton>
                <span class="text-xl text-[#4f555e] dark:text-white">
                  <SvgIcon icon="ri:download-2-line" />
                </span>
              </HoverButton>
            </div> -->
          </div>
        </header>
        <NLayout has-sider>
          <NLayoutSider
            bordered :width="240" :collapsed-width="10" :collapsed="collapsed"
            :show-trigger="isMobile ? false : 'arrow-circle'" collapse-mode="transform" position="absolute"
            :style="getMobileClass" @collapse="collapsed = true" @expand="collapsed = false"
          >
            <NMenu
              v-model:value="activeKey" :collapsed="collapsed" :collapsed-width="64" :collapsed-icon-size="22"
              :options="menuOptions"
            />
          </NLayoutSider>
          <NLayout>
            <router-view />
            <Permission :visible="needPermission" />
          </NLayout>
        </NLayout>
      </div>
    </div>
  </div>
</template>
