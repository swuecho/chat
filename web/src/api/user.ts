import request from '@/utils/request/axios'
export async function fetchLogin(email: string, password: string) {
  try {
    const response = await request.post('/login', { email, password })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export async function fetchSignUp(email: string, password: string) {
  try {
    const response = await request.post('/signup', { email, password })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
