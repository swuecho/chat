<script setup lang='ts'>
import { computed, reactive, ref } from 'vue'
import { NButton, NForm, NFormItemRow, NInput, NModal, NTabPane, NTabs, useMessage } from 'naive-ui'
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
const LoginData = reactive({
  email: '',
  password: '',
})
const RegisterData = reactive({
  email: '',
  password: '',
  repwd: '',
})

function check_input(data: object) {
  return !Object.values(data).every(({ length }) => length > 6) || loading.value
}

const login_not_filled = computed(() => check_input(LoginData))
const register_not_filled = computed(() => check_input(RegisterData))

async function handleLogin() {
  const user_email_v = LoginData.email.trim()
  const user_password_v = LoginData.password.trim()

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
    console.log(error)
    const response = error.response
    if (response.status >= 400)
      ms.error(t(response.data.message))
    authStore.removeToken()
    authStore.removeExpiresIn()
  }
  finally {
    loading.value = false
  }
}

async function handleSignup() {
  const user_email_v = RegisterData.email.trim()
  const user_password_v = RegisterData.password.trim()
  const user_repwd_v = RegisterData.repwd.trim()

  if (!user_email_v || !user_password_v || !user_repwd_v)
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

  if (user_password_v !== user_repwd_v) {
    ms.error(t('error.invalidRepwd'))
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
  <NModal :show="visible" style="width: 90%; max-width: 400px">
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
        <NTabs
          class="card-tabs" default-value="signin" size="large" animated
        >
          <NTabPane name="signin" :tab="t('common.login')" :tab-props="{ title: 'signintab' }">
            <NForm :show-label="false">
              <NFormItemRow label="邮箱">
                <NInput
                  v-model:value="LoginData.email" data-testid="email" type="text" :minlength="6"
                  :placeholder="$t('common.email_placeholder')"
                />
              </NFormItemRow>
              <NFormItemRow label="密码">
                <NInput
                  v-model:value="LoginData.password" data-testid="password" type="password" :minlength="6" show-password-on="click"
                  :placeholder="$t('common.password_placeholder')"
                />
              </NFormItemRow>
            </NForm>
            <div class="flex justify-between">
              <NButton
                type="primary" block secondary strong data-testid="login" :disabled="login_not_filled"
                :loading="loading" @click="handleLogin"
              >
                {{ $t('common.login') }}
              </NButton>
            </div>
          </NTabPane>
          <NTabPane name="signup" :tab="t('common.signup')" :tab-props="{ title: 'signuptab' }">
            <NForm :show-label="false">
              <NFormItemRow label="邮箱">
                <NInput
                  v-model:value="RegisterData.email" data-testid="signup_email" type="text" :minlength="6"
                  :placeholder="$t('common.email_placeholder')"
                />
              </NFormItemRow>
              <NFormItemRow label="密码">
                <NInput
                  v-model:value="RegisterData.password" data-testid="signup_password" type="password" :minlength="6" show-password-on="click"
                  :placeholder="$t('common.password_placeholder')"
                />
              </NFormItemRow>
              <NFormItemRow label="确认密码">
                <NInput
                  v-model:value="RegisterData.repwd" data-testid="repwd" type="password" :minlength="6" show-password-on="click"
                  :placeholder="$t('common.password_placeholder')"
                />
              </NFormItemRow>
            </NForm>
            <div class="flex justify-between">
              <NButton
                type="primary" block secondary strong data-testid="signup" :disabled="register_not_filled"
                :loading="loading" @click="handleSignup"
              >
                {{ $t('common.signup') }}
              </NButton>
            </div>
          </NTabPane>
        </NTabs>
      </div>
    </div>
  </NModal>
</template>
