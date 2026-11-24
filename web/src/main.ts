import { createApp } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'
import App from './App.vue'
import { setupI18n } from './locales'
import { setupAssets } from './plugins'
import { setupStore } from './store'
import { setupRouter } from './router'
import { queryClient } from './plugins/queryClient'

async function bootstrap() {
  const app = createApp(App)
  setupAssets()

  setupStore(app)

  setupI18n(app)

  await setupRouter(app)

  app.use(VueQueryPlugin, { queryClient })
  app.mount('#app')
}

bootstrap()
