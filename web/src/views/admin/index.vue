<script setup lang="ts">
import type { Component, Ref } from 'vue'
import { h, reactive, ref } from 'vue'
import { NIcon, NLayout, NLayoutSider, NMenu, NSpace } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import { PulseOutline, ShieldCheckmarkOutline } from '@vicons/ionicons5'
import { RouterLink } from 'vue-router'
import i18n from '@/locales'

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
          { default: () => i18n.global.t('admin.rateLimit') },
        ),
    key: 'hear-the-wind-sing',
    icon: renderIcon(PulseOutline),
  },
  {
    label: () => h(
      RouterLink,
      {
        to: {
          name: 'AdminSystem',
        },
      },
      { default: () => i18n.global.t('admin.permission') },
    ),
    key: 'a-wild-sheep-chase',
    icon: renderIcon(ShieldCheckmarkOutline),
  },
])

const collapsed: Ref<boolean> = ref(false)
const activeKey: Ref<string | null> = ref(null)
</script>

<template>
  <div>
    <NSpace vertical>
      <NLayout has-sider>
        <NLayoutSider
          bordered collapse-mode="width" :collapsed-width="64" :width="240" :collapsed="collapsed"
          show-trigger @collapse="collapsed = true" @expand="collapsed = false"
        >
          <NMenu
            v-model:value="activeKey" :collapsed="collapsed" :collapsed-width="64" :collapsed-icon-size="22"
            :options="menuOptions"
          />
        </NLayoutSider>
        <NLayout>
          <router-view />
        </NLayout>
      </NLayout>
    </NSpace>
  </div>
</template>
