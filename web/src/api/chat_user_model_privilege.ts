import request from '@/utils/request/axios'

export const ListUserChatModelPrivilege = async () => {
  try {
    const response = await request.get('/admin/user_chat_model_privilege')
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
export const CreateUserChatModelPrivilege = async (data: any) => {
  try {
    const response = await request.post('/admin/user_chat_model_privilege', data)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const UpdateUserChatModelPrivilege = async (id: string, data: any) => {
  try {
    const response = await request.put(`/admin/user_chat_model_privilege/${id}`, data)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}

export const DeleteUserChatModelPrivilege = async (id: string) => {
  try {
    const response = await request.delete(`/admin/user_chat_model_privilege/${id}`)
    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
