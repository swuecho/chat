<script setup lang='ts'>
import { computed, ref } from 'vue'
import { NButton, NInput, NModal, useMessage } from 'naive-ui'
import { fetchLogin, fetchSignUp } from '@/api'
import { t } from '@/locales'
import { useAuthStore } from '@/store'
import Icon403 from '@/icons/403.vue'

interface Props {
  visible: boolean
}

defineProps<Props>()

const authStore = useAuthStore()

const ms = useMessage()

const loading = ref(false)
const user_email = ref('')
const user_password = ref('')

const user_pass_not_filled = computed(() => !user_email.value.trim() || !user_password.value.trim() || loading.value)

async function handleLogin() {
  const user_email_v = user_email.value.trim()
  const user_password_v = user_password.value.trim()

  if (!user_email_v || !user_password_v)
    return

  // check user_email_v  is valid email
  if (!user_email_v.match(/^[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$/)) {
    ms.error(t('error.invalidEmail'))
    return
  }
  // check password is length >=6 and include a number, a lowercase letter, an uppercase letter, and a special character
  if (!user_password_v.match(/^(?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?=.*[a-zA-Z]).{6,}$/)) {
    // ms.error(t('error.invalidPassword'))
    ms.error(t('error.invalidPassword'))
    return
  }

  loading.value = true
  try {
    const { accessToken, expiresIn } = await fetchLogin(user_email_v, user_password_v)
    authStore.setToken(accessToken)
    authStore.setExpiresIn(expiresIn)
    ms.success(t('common.loginSuccess'))
    window.location.reload()
  }
  catch (error: any) {
    if (error.response?.status === 401 && error.response?.data === 'invalid email or password: sql: no rows in result set\n')
      ms.error(t('common.please_register'))
    else
      ms.error(error.message ?? 'error')
    authStore.removeToken()
    authStore.removeExpiresIn()
  }
  finally {
    loading.value = false
  }
}

async function handleSignup() {
  const user_email_v = user_email.value.trim()
  const user_password_v = user_password.value.trim()

  if (!user_email_v || !user_password_v)
    return

  if (!user_email_v || !user_password_v)
    return

  // check user_email_v  is valid email
  if (!user_email_v.match(/^[\w-]+(\.[\w-]+)*@[\w-]+(\.[\w-]+)+$/)) {
    ms.error(t('error.invalidEmail'))
    return
  }
  // check password is length >=6 and include a number, a lowercase letter, an uppercase letter, and a special character
  if (!user_password_v.match(/^(?=.*\d)(?=.*[a-z])(?=.*[A-Z])(?=.*[a-zA-Z]).{6,}$/)) {
    ms.error(t('error.invalidPassword'))
    return
  }
  loading.value = true
  try {
    const { accessToken, expiresIn } = await fetchSignUp(user_email_v, user_password_v)
    authStore.setToken(accessToken)
    authStore.setExpiresIn(expiresIn)
    ms.success('success')
    window.location.reload()
  }
  catch (error: any) {
    ms.error(error.message ?? 'error')
    authStore.removeToken()
  }
  finally {
    loading.value = false
  }
}

// function handlePress(event: KeyboardEvent) {
//   if (event.key === 'Enter' && !event.shiftKey) {
//     event.preventDefault()
//     handleLogin()
//   }
// }
</script>

<template>
  <NModal :show="visible" style="width: 90%; max-width: 640px">
    <div class="p-10 bg-white rounded dark:bg-slate-800">
      <div class="space-y-4">
        <header class="space-y-2">
          <h2 class="text-2xl font-bold text-center text-slate-800 dark:text-neutral-200">
            欢迎
          </h2>
          <p class="text-base text-center text-slate-500 dark:text-slate-500">
            {{ $t('common.unauthorizedTips') }}
          </p>
          <Icon403 class="w-[200px] m-auto" />
        </header>
        <NInput v-model:value="user_email" data-testid="email" type="text" :minlength="6"
          :placeholder="$t('common.email_placeholder')" />
        <NInput v-model:value="user_password" data-testid="password" type="text" :minlength="6"
          :placeholder="$t('common.password_placeholder')" />
        <div class="flex justify-between">
          <NButton type="primary" data-testid="signup" :disabled="user_pass_not_filled" :loading="loading"
            @click="handleSignup">
            {{ $t('common.signup') }}
          </NButton>
          <NButton type="primary" data-testid="login" :disabled="user_pass_not_filled" :loading="loading"
            @click="handleLogin">
            {{ $t('common.login') }}
          </NButton>
        </div>
      </div>
    </div>
  </NModal>
</template>
