<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import { NButton, NDataTable, NDescriptions, NDescriptionsItem, NEmpty, NFormItem, NInput, NInputNumber, NModal, NSpace, NTabPane, NTabs, NTag, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import copy from 'copy-to-clipboard'
import { createApiKey, getApiKeyRequest, getApiKeyUsage, listApiKeyRequests, listApiKeys, revokeApiKey } from '@/api/generated_client'
import type {
  GatewayRequestDetailHttpResponse,
  GetApiKeyUsageResponse,
  ListApiKeyRequestsResponse,
  ListApiKeysResponse,
} from '@/api/generated_client'
import { useAuthStore } from '@/store'

type VirtualApiKey = ListApiKeysResponse[number]
type ApiKeyUsage = GetApiKeyUsageResponse[number]
type GatewayRequestSummary = ListApiKeyRequestsResponse[number]
type CapturedSample = GatewayRequestDetailHttpResponse['requestCapture']

const authStore = useAuthStore()
const message = useMessage()
const dialog = useDialog()
const keys = ref<VirtualApiKey[]>([])
const loading = ref(false)
const showCreate = ref(false)
const showSecret = ref(false)
const showUsage = ref(false)
const showRequests = ref(false)
const showRequestDetail = ref(false)
const requestLoading = ref(false)
const detailLoading = ref(false)
const selectedKey = ref<VirtualApiKey | null>(null)
const requests = ref<GatewayRequestSummary[]>([])
const requestDetail = ref<GatewayRequestDetailHttpResponse | null>(null)
const createdSecret = ref('')
const usage = ref<ApiKeyUsage[]>([])
const form = ref({ name: '', requestsPerMinute: 60 })
const search = ref('')
const gatewayBaseUrl = `${window.location.origin}/v1`
const filteredKeys = computed(() => {
  const term = search.value.trim().toLowerCase()
  if (!term)
    return keys.value
  return keys.value.filter(key => key.name.toLowerCase().includes(term) || key.keyPrefix.toLowerCase().includes(term))
})

const formatDate = (value: string | null) => value ? new Date(value).toLocaleString() : 'Never'

async function loadKeys() {
  loading.value = true
  try {
    keys.value = await listApiKeys()
  }
  finally {
    loading.value = false
  }
}

async function createKey() {
  if (!form.value.name.trim()) {
    message.warning('Enter a name for this key')
    return
  }
  try {
    const result = await createApiKey({
      body: { name: form.value.name.trim(), requestsPerMinute: form.value.requestsPerMinute, expiresAt: '' },
    })
    createdSecret.value = result.key || ''
    showCreate.value = false
    showSecret.value = true
    form.value = { name: '', requestsPerMinute: 60 }
    await loadKeys()
  }
  catch {
    message.error('Could not create API key')
  }
}

function confirmRevoke(key: VirtualApiKey) {
  dialog.warning({
    title: 'Revoke API key',
    content: `Revoke “${key.name}”? Applications using it will stop working immediately.`,
    positiveText: 'Revoke',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      await revokeApiKey({ path: { id: key.id } })
      message.success('API key revoked')
      await loadKeys()
    },
  })
}

async function openUsage(key: VirtualApiKey) {
  usage.value = await getApiKeyUsage({ path: { id: key.id } })
  showUsage.value = true
}

async function openRequests(key: VirtualApiKey) {
  selectedKey.value = key
  showRequests.value = true
  requestLoading.value = true
  try {
    requests.value = await listApiKeyRequests({ path: { id: key.id } })
  }
  catch {
    message.error('Could not load gateway requests')
  }
  finally {
    requestLoading.value = false
  }
}

async function inspectRequest(row: GatewayRequestSummary) {
  if (!selectedKey.value)
    return
  showRequestDetail.value = true
  detailLoading.value = true
  requestDetail.value = null
  try {
    requestDetail.value = await getApiKeyRequest({
      path: { id: selectedKey.value.id, requestId: row.id },
    })
  }
  catch {
    message.error('Could not load request details')
  }
  finally {
    detailLoading.value = false
  }
}

function displayCapture(capture?: CapturedSample): string {
  if (!capture)
    return 'Not captured or retention period expired.'
  if (capture.encoding === 'base64')
    return capture.base64 ? `[base64]\n${capture.base64}` : 'Not captured or retention period expired.'
  if (!capture.text)
    return 'Not captured or retention period expired.'
  try {
    return JSON.stringify(JSON.parse(capture.text), null, 2)
  }
  catch {
    return capture.text
  }
}

const displayJSON = (value?: unknown) => JSON.stringify(value ?? {}, null, 2)

function copySecret() {
  copy(createdSecret.value)
  message.success('API key copied')
}

const columns: DataTableColumns<VirtualApiKey> = [
  { title: 'Name', key: 'name' },
  { title: 'Key', key: 'keyPrefix', render: row => `${row.keyPrefix}…` },
  { title: 'Status', key: 'status', render: row => h(NTag, { type: row.status === 'active' ? 'success' : 'default', size: 'small' }, { default: () => row.status }) },
  { title: 'Limit', key: 'requestsPerMinute', render: row => `${row.requestsPerMinute}/min` },
  { title: 'Last used', key: 'lastUsedAt', render: row => formatDate(row.lastUsedAt) },
  { title: 'Created', key: 'createdAt', render: row => formatDate(row.createdAt) },
  {
    title: 'Actions',
    key: 'actions',
    render: row => h(NSpace, null, {
      default: () => [
        h(NButton, { size: 'small', type: 'primary', secondary: true, onClick: () => openRequests(row) }, { default: () => 'Requests' }),
        h(NButton, { size: 'small', onClick: () => openUsage(row) }, { default: () => 'Usage' }),
        row.status === 'active' ? h(NButton, { size: 'small', type: 'error', secondary: true, onClick: () => confirmRevoke(row) }, { default: () => 'Revoke' }) : null,
      ],
    }),
  },
]

const usageColumns: DataTableColumns<ApiKeyUsage> = [
  { title: 'Model', key: 'requestedModel' }, { title: 'Requests', key: 'requestCount' },
  { title: 'Input tokens', key: 'promptTokens' }, { title: 'Output tokens', key: 'completionTokens' }, { title: 'Total tokens', key: 'totalTokens' },
]

const requestColumns: DataTableColumns<GatewayRequestSummary> = [
  { title: 'Time', key: 'createdAt', render: row => formatDate(row.createdAt), width: 170 },
  { title: 'Model', key: 'requestedModel', ellipsis: { tooltip: true } },
  { title: 'Status', key: 'status', render: row => h(NTag, { type: row.status === 'succeeded' ? 'success' : row.status === 'started' ? 'warning' : 'error', size: 'small' }, { default: () => row.status }) },
  { title: 'Mode', key: 'stream', render: row => row.stream ? 'Stream' : 'JSON' },
  { title: 'Latency', key: 'latencyMs', render: row => `${row.latencyMs} ms` },
  { title: 'Bytes', key: 'bytes', render: row => `${row.requestBytes} → ${row.responseBytes}` },
  { title: 'Tokens', key: 'totalTokens' },
  { title: '', key: 'inspect', render: row => h(NButton, { size: 'small', onClick: () => inspectRequest(row) }, { default: () => 'Inspect' }) },
]

onMounted(async () => {
  await authStore.initializeAuth()
  if (authStore.isValid)
    await loadKeys()
})
</script>

<template>
  <div>
    <div class="api-key-intro">
      <p>Use virtual keys with the OpenAI-compatible base URL <code>{{ gatewayBaseUrl }}</code>.</p>
      <div class="api-key-actions">
        <NInput v-model:value="search" clearable placeholder="Search API keys" aria-label="Search API keys" />
        <NButton type="primary" @click="showCreate = true">
          Create API key
        </NButton>
      </div>
    </div>
    <NDataTable size="small" :columns="columns" :data="filteredKeys" :loading="loading" :row-key="row => row.id" :scroll-x="980" :single-line="false">
      <template #empty>
        <NEmpty description="No API keys found" />
      </template>
    </NDataTable>

    <NModal v-model:show="showCreate" preset="card" title="Create API key" class="max-w-lg">
      <NFormItem label="Name">
        <NInput v-model:value="form.name" placeholder="My application" />
      </NFormItem>
      <NFormItem label="Requests per minute">
        <NInputNumber v-model:value="form.requestsPerMinute" :min="1" :max="10000" />
      </NFormItem>
      <div class="flex justify-end gap-2">
        <NButton @click="showCreate = false">
          Cancel
        </NButton><NButton type="primary" @click="createKey">
          Create
        </NButton>
      </div>
    </NModal>

    <NModal v-model:show="showSecret" preset="card" title="Copy your API key" class="max-w-xl" :mask-closable="false">
      <p class="mb-3 text-amber-600">
        This key is shown only once. Store it securely.
      </p>
      <NInput :value="createdSecret" readonly type="textarea" :autosize="{ minRows: 2 }" />
      <div class="flex justify-end gap-2 mt-4">
        <NButton type="primary" @click="copySecret">
          Copy
        </NButton><NButton @click="showSecret = false">
          Done
        </NButton>
      </div>
    </NModal>

    <NModal v-model:show="showUsage" preset="card" title="API usage" class="max-w-4xl">
      <NDataTable :columns="usageColumns" :data="usage" />
    </NModal>

    <NModal v-model:show="showRequests" preset="card" :title="`Gateway requests — ${selectedKey?.name ?? ''}`" class="max-w-7xl">
      <p class="mb-4 text-sm text-gray-500">
        Showing the 100 most recent requests. Captured bodies are automatically purged after their retention period.
      </p>
      <NDataTable :columns="requestColumns" :data="requests" :loading="requestLoading" :row-key="row => row.id" :scroll-x="1050" />
    </NModal>

    <NModal v-model:show="showRequestDetail" preset="card" title="Gateway request details" class="max-w-6xl">
      <div v-if="detailLoading" class="py-12 text-center text-gray-500">
        Loading request details…
      </div>
      <template v-else-if="requestDetail">
        <NDescriptions bordered :column="3" label-placement="top" class="mb-5">
          <NDescriptionsItem label="Request ID">
            {{ requestDetail.requestUuid }}
          </NDescriptionsItem>
          <NDescriptionsItem label="Model">
            {{ requestDetail.requestedModel }}
          </NDescriptionsItem>
          <NDescriptionsItem label="Status">
            <NTag :type="requestDetail.status === 'succeeded' ? 'success' : 'error'">
              {{ requestDetail.status }}
            </NTag>
          </NDescriptionsItem>
          <NDescriptionsItem label="Created">
            {{ formatDate(requestDetail.createdAt) }}
          </NDescriptionsItem>
          <NDescriptionsItem label="Latency">
            {{ requestDetail.latencyMs }} ms
          </NDescriptionsItem>
          <NDescriptionsItem label="Tokens">
            {{ requestDetail.promptTokens }} + {{ requestDetail.completionTokens }} = {{ requestDetail.totalTokens }}
          </NDescriptionsItem>
          <NDescriptionsItem label="Request bytes">
            {{ requestDetail.requestBytes }}
          </NDescriptionsItem>
          <NDescriptionsItem label="Response bytes">
            {{ requestDetail.responseBytes }}
          </NDescriptionsItem>
          <NDescriptionsItem label="Content retained until">
            {{ formatDate(requestDetail.retentionUntil) }}
          </NDescriptionsItem>
          <NDescriptionsItem label="Request SHA-256" :span="3">
            <code class="break-all">{{ requestDetail.requestSha256 }}</code>
          </NDescriptionsItem>
          <NDescriptionsItem label="Response SHA-256" :span="3">
            <code class="break-all">{{ requestDetail.responseSha256 || 'Unavailable' }}</code>
          </NDescriptionsItem>
        </NDescriptions>

        <NTabs type="line" animated>
          <NTabPane name="request" tab="Request">
            <NTag v-if="requestDetail.requestTruncated" type="warning" class="mb-3">
              Captured sample is truncated
            </NTag>
            <pre class="capture-panel">{{ displayCapture(requestDetail.requestCapture) }}</pre>
          </NTabPane>
          <NTabPane name="response" tab="Response">
            <NTag v-if="requestDetail.responseTruncated" type="warning" class="mb-3">
              Captured sample is truncated
            </NTag>
            <pre class="capture-panel">{{ displayCapture(requestDetail.responseCapture) }}</pre>
          </NTabPane>
          <NTabPane name="classification" tab="Classification">
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <h3 class="mb-2 font-medium">
                  Request
                </h3><pre class="capture-panel">{{ displayJSON(requestDetail.requestClassification) }}</pre>
              </div>
              <div>
                <h3 class="mb-2 font-medium">
                  Response
                </h3><pre class="capture-panel">{{ displayJSON(requestDetail.responseClassification) }}</pre>
              </div>
            </div>
          </NTabPane>
        </NTabs>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.capture-panel {
  max-height: 28rem;
  overflow: auto;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  border: 1px solid rgba(128, 128, 128, 0.25);
  border-radius: 0.5rem;
  background: rgba(128, 128, 128, 0.08);
  padding: 1rem;
  font-size: 0.8rem;
  line-height: 1.5;
}
.api-key-intro { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin-bottom: 14px; }
.api-key-intro p { color: #737983; font-size: 13px; }
.api-key-actions { display: flex; align-items: center; gap: 8px; min-width: 360px; }
@media (max-width: 760px) {
  .api-key-intro, .api-key-actions { align-items: stretch; flex-direction: column; }
  .api-key-actions { min-width: 0; width: 100%; }
}
</style>
