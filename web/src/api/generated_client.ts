import { client } from './generated/client.gen'
import { useAuthStore } from '@/store'

async function accessToken(): Promise<string | undefined> {
  const authStore = useAuthStore()

  if (authStore.isInitializing)
    await authStore.waitForInitialization()

  if (!authStore.isValid) {
    await authStore.refreshToken()
  }
  else if (authStore.needsRefresh && !authStore.isRefreshing) {
    // Match the existing Axios behavior: proactive refresh failure does not
    // discard an access token that is still valid.
    try {
      await authStore.refreshToken()
    }
    catch {
      // Continue with the current token.
    }
  }

  return authStore.getToken ?? undefined
}

async function authenticatedFetch(request: Request): Promise<Response> {
  // Keep an unused clone because the body of the request passed to fetch may
  // no longer be reusable when a 401 response arrives.
  const retryRequest = request.clone()
  const response = await globalThis.fetch(request)
  if (response.status !== 401 || request.url.includes('/api/auth/'))
    return response

  const authStore = useAuthStore()
  await authStore.refreshToken()
  const token = authStore.getToken
  if (!token)
    return response

  const headers = new Headers(retryRequest.headers)
  headers.set('Authorization', `Bearer ${token}`)
  return globalThis.fetch(new Request(retryRequest, { headers }))
}

client.setConfig({
  auth: accessToken,
  credentials: 'include',
  fetch: authenticatedFetch,
})

export { client }
export * from './generated'
