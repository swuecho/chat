import {
  createApiKey as createApiKeyRequest,
  getApiKeyRequest,
  getApiKeyUsage,
  listApiKeyRequests,
  listApiKeys,
  revokeApiKey as revokeApiKeyRequest,
} from '@/api/generated_client'

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

export const fetchApiKeys = async (): Promise<VirtualApiKey[]> =>
  await listApiKeys() as VirtualApiKey[]

export const createApiKey = async (data: { name: string; requestsPerMinute: number; expiresAt?: string }): Promise<VirtualApiKey> =>
  await createApiKeyRequest({ body: { ...data, expiresAt: data.expiresAt ?? '' } }) as VirtualApiKey

export const revokeApiKey = async (id: number): Promise<void> => {
  await revokeApiKeyRequest({ path: { id } })
}

export const fetchApiKeyUsage = async (id: number): Promise<ApiKeyUsage[]> =>
  await getApiKeyUsage({ path: { id } }) as ApiKeyUsage[]

export const fetchGatewayRequests = async (keyId: number): Promise<GatewayRequestSummary[]> =>
  await listApiKeyRequests({ path: { id: keyId } }) as GatewayRequestSummary[]

export const fetchGatewayRequest = async (keyId: number, requestId: number): Promise<GatewayRequestDetail> =>
  await getApiKeyRequest({ path: { id: keyId, requestId } }) as GatewayRequestDetail
