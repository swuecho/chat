import type { App } from 'vue'
import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHashHistory } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/snapshot',
    name: 'Snapshot',
    component: () => import('@/views/snapshot/page.vue'),
    children: [
      {
        path: ':uuid?',
        name: 'Snapshot',
        component: () => import('@/views/snapshot/page.vue'),
      },
    ],
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


export async function setupRouter(app: App) {
  app.use(router)
  await router.isReady()
}
