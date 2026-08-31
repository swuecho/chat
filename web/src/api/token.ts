import { createLongLivedToken } from '@/api/generated_client'

export async function fetchAPIToken() {
  return createLongLivedToken()
}
