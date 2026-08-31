import {
  createUserChatModelPrivilege,
  deleteUserChatModelPrivilege,
  listUserChatModelPrivileges,
  updateUserChatModelPrivilege,
} from '@/api/generated_client'
import type { CreateUserChatModelPrivilegeData } from '@/api/generated/types.gen'

type ModelPrivilegeInput = CreateUserChatModelPrivilegeData['body']

export const ListUserChatModelPrivilege = () => listUserChatModelPrivileges()

export const CreateUserChatModelPrivilege = (data: ModelPrivilegeInput) =>
  createUserChatModelPrivilege({ body: data })

export const UpdateUserChatModelPrivilege = (id: string, data: ModelPrivilegeInput) =>
  updateUserChatModelPrivilege({ path: { id: Number(id) }, body: data })

export const DeleteUserChatModelPrivilege = (id: string) =>
  deleteUserChatModelPrivilege({ path: { id: Number(id) } })
