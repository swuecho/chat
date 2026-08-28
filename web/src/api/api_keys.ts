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

export interface GatewayRequestSummary {
  id: number
  requestUuid: string
  requestedModel: string
  provider: string
  status: string
  stream: boolean
  promptTokens: number
  completionTokens: number
  totalTokens: number
  latencyMs: number
  requestBytes: number
  responseBytes: number
  requestTruncated: boolean
  responseTruncated: boolean
  createdAt: string
  completedAt: string | null
  retentionUntil: string
  errorCode: string
}

export interface CapturedSample { encoding: 'utf-8' | 'base64'; text?: string; base64?: string }

export interface GatewayRequestDetail extends GatewayRequestSummary {
  providerRequestId: string
  requestSha256: string
  responseSha256: string
  requestClassification: Record<string, unknown>
  responseClassification: Record<string, unknown>
  requestCapture: CapturedSample
  responseCapture: CapturedSample
}

const adminKeysPath = '/admin/api-keys'

export const fetchApiKeys = async (): Promise<VirtualApiKey[]> => (await request.get(adminKeysPath)).data
export const createApiKey = async (data: { name: string; requestsPerMinute: number; expiresAt?: string }): Promise<VirtualApiKey> => (await request.post(adminKeysPath, data)).data
export const revokeApiKey = async (id: number): Promise<void> => { await request.delete(`${adminKeysPath}/${id}`) }
export const fetchApiKeyUsage = async (id: number): Promise<ApiKeyUsage[]> => (await request.get(`${adminKeysPath}/${id}/usage`)).data
export const fetchGatewayRequests = async (keyId: number): Promise<GatewayRequestSummary[]> => (await request.get(`${adminKeysPath}/${keyId}/requests`)).data
export const fetchGatewayRequest = async (keyId: number, requestId: number): Promise<GatewayRequestDetail> => (await request.get(`${adminKeysPath}/${keyId}/requests/${requestId}`)).data
