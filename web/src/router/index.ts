import type { App } from 'vue'
import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHashHistory } from 'vue-router'
import { setupPageGuard } from './permission'
import { ChatLayout } from '@/views/chat/layout'

const routes: RouteRecordRaw[] = [
  {
    path: '/snapshot',
    name: 'Snapshot',
    component: () => import('@/views/snapshot/index.vue'),
    children: [
      {
        path: ':uuid?',
        name: 'Snapshot',
        component: () => import('@/views/snapshot/index.vue'),
      },
    ],
  },
  {
    path: '/prompt',
    name: 'Prompt',
    component: () => import('@/views/prompt/creator.vue')
  },
  {
    path: '/bot',
    name: 'Bot',
    component: () => import('@/views/bot/index.vue'),
    children: [
      {
        path: ':uuid?',
        name: 'Bot',
        component: () => import('@/views/bot/index.vue'),
      },
    ],
  },
  {
    path: '/snapshot_all',
    name: 'SnapshotAll',
    component: () => import('@/views/snapshot/all.vue'),
  },
  {
    path: '/bot_all',
    name: 'BotAll',
    component: () => import('@/views/bot/all.vue'),
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('@/views/admin/index.vue'),
    children: [
      {
        path: 'user',
        name: 'AdminUser',
        component: () => import('@/views/admin/user/index.vue'),
      },
      {
        path: 'model',
        name: 'AdminModel',
        component: () => import('@/views/admin/model/index.vue'),
      },
      {
        path: 'model_rate_limit',
        name: 'ModelRateLimit',
        component: () => import('@/views/admin/modelRateLimit/index.vue'),
      }
    ],
  },
  {
    path: '/',
    name: 'Root',
    component: ChatLayout,
    redirect: '/chat',
    children: [
      {
        path: '/chat/:uuid?',
        name: 'Chat',
        component: () => import('@/views/chat/index.vue'),
      },
    ],
  },
  {
    path: '/404',
    name: '404',
    component: () => import('@/views/exception/404/index.vue'),
  },
  {
    path: '/500',
    name: '500',
    component: () => import('@/views/exception/500/index.vue'),
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'notFound',
    redirect: '/404',
  },
]

// !!!
// https://router.vuejs.org/guide/essentials/history-mode.html
// createWebHashHistory
// It uses a hash character (#) before the actual URL that is internally passed.
// Because this section of the URL is never sent to the server,
// it doesn't require any special treatment on the server level.
// It does however have a bad impact in SEO. If that's a concern for you, use the HTML5 history mode.

// this is crazy, router in frontend is a nightmare

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
  scrollBehavior: () => ({ left: 0, top: 0 }),
})

setupPageGuard(router)

export async function setupRouter(app: App) {
  app.use(router)
  await router.isReady()
}
