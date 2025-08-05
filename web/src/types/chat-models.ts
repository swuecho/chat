export interface ChatModel {
  id: number
  name: string
  label: string
  isEnable: boolean
  isDefault: boolean
  orderNumber: number
  lastUsageTime: string
  apiType: string
  maxTokens?: number
  costPer1kTokens?: number
  description?: string
}

export interface CreateChatModelRequest {
  name: string
  label: string
  apiType: string
  isEnable?: boolean
  isDefault?: boolean
  orderNumber?: number
  maxTokens?: number
  costPer1kTokens?: number
  description?: string
}

export interface UpdateChatModelRequest {
  name?: string
  label?: string
  apiType?: string
  isEnable?: boolean
  isDefault?: boolean
  orderNumber?: number
  maxTokens?: number
  costPer1kTokens?: number
  description?: string
}

export interface ChatModelSelectOption {
  label: string | (() => any)
  value: string
  disabled?: boolean
}

export interface ChatModelsResponse {
  models: ChatModel[]
  total: number
}