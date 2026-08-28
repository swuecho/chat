import request from '@/utils/request/axios'

export interface VirtualApiKey {
  id: number
  name: string
  keyPrefix: string
  status: 'active' | 'revoked'
  requestsPerMinute: number
  expiresAt: string | null
  lastUsedAt: string | null
  createdAt: string
  key?: string
}

export interface ApiKeyUsage {
  requestedModel: string
  requestCount: number
  promptTokens: number
  completionTokens: number
  totalTokens: number
  lastUsedAt: string | null
}

export const fetchApiKeys = async (): Promise<VirtualApiKey[]> => (await request.get('/api-keys')).data
export const createApiKey = async (data: { name: string; requestsPerMinute: number; expiresAt?: string }): Promise<VirtualApiKey> => (await request.post('/api-keys', data)).data
export const revokeApiKey = async (id: number): Promise<void> => { await request.delete(`/api-keys/${id}`) }
export const fetchApiKeyUsage = async (id: number): Promise<ApiKeyUsage[]> => (await request.get(`/api-keys/${id}/usage`)).data
