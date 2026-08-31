import { login, signUp } from '@/api/generated_client'
export async function fetchLogin(email: string, password: string) {
  try {
    return await login({ body: { email, password } })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export async function fetchSignUp(email: string, password: string) {
  try {
    return await signUp({ body: { email, password } })
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
