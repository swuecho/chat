import {
  getAdminSessionMessages,
  getUserAnalysis as getUserAnalysisRequest,
  getUserSessionHistory as getUserSessionHistoryRequest,
  getUserStats,
  updateAdminUser,
  updateUserRateLimit,
} from '@/api/generated_client'
import type { UpdateAdminUserData } from '@/api/generated/types.gen'

export const GetUserData = (page: number, size: number) =>
  getUserStats({ body: { page, size } })

export const UpdateRateLimit = (email: string, rateLimit: number) =>
  updateUserRateLimit({ body: { email, rateLimit } })

export const updateUserFullName = (data: UpdateAdminUserData['body']) =>
  updateAdminUser({ body: data })

export const getUserAnalysis = (userEmail: string) =>
  getUserAnalysisRequest({ path: { email: userEmail } })

export const getUserSessionHistory = (userEmail: string, page = 1, size = 10) =>
  getUserSessionHistoryRequest({
    path: { email: userEmail },
    query: { limit: size, offset: (page - 1) * size },
  })

export const getSessionMessagesForAdmin = (sessionUuid: string) =>
  getAdminSessionMessages({ path: { sessionUuid } })
