import { createApp } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'
import App from './App.vue'
import { setupRouter } from './router'

async function bootstrap() {
  const app = createApp(App)

  await setupRouter(app)

  app.use(VueQueryPlugin)
  app.mount('#app')
}

bootstrap()
