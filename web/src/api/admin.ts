import request from '@/utils/request/axios'

export const GetUserData = async (page: number, size: number) => {
  try {
    const response = await request.post('/admin/user_stats', {
      page,
      size,
    })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const UpdateRateLimit = async (email: string, rateLimit: number) => {
  try {
    const response = await request.post('/admin/rate_limit', {
      email,
      rateLimit,
    })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const updateUserFullName = async (data: any): Promise<any> => {
  try {
    const response = await request.put('/admin/users', data)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const getUserAnalysis = async (userEmail: string) => {
  try {
    const response = await request.get(`/admin/user_analysis/${encodeURIComponent(userEmail)}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const getUserSessionHistory = async (userEmail: string, page: number = 1, size: number = 10) => {
  try {
    const response = await request.get(`/admin/user_session_history/${encodeURIComponent(userEmail)}`, {
      params: { page, size }
    })
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const getSessionMessagesForAdmin = async (sessionUuid: string) => {
  try {
    const response = await request.get(`/admin/session_messages/${encodeURIComponent(sessionUuid)}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
