<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import { NButton, NCard, NDataTable, NFormItem, NInput, NInputNumber, NModal, NSpace, NTag, useDialog, useMessage } from 'naive-ui'
import type { DataTableColumns } from 'naive-ui'
import copy from 'copy-to-clipboard'
import { createApiKey, fetchApiKeys, fetchApiKeyUsage, revokeApiKey } from '@/api/api_keys'
import type { ApiKeyUsage, VirtualApiKey } from '@/api/api_keys'
import Permission from '@/views/components/Permission.vue'
import { useAuthStore } from '@/store'

const authStore = useAuthStore()
const message = useMessage()
const dialog = useDialog()
const keys = ref<VirtualApiKey[]>([])
const loading = ref(false)
const showCreate = ref(false)
const showSecret = ref(false)
const showUsage = ref(false)
const createdSecret = ref('')
const usage = ref<ApiKeyUsage[]>([])
const form = ref({ name: '', requestsPerMinute: 60 })
const needPermission = computed(() => authStore.isInitialized && !authStore.isValid)
const gatewayBaseUrl = `${window.location.origin}/v1`

const formatDate = (value: string | null) => value ? new Date(value).toLocaleString() : 'Never'

async function loadKeys() {
  loading.value = true
  try { keys.value = await fetchApiKeys() }
  finally { loading.value = false }
}

async function createKey() {
  if (!form.value.name.trim()) { message.warning('Enter a name for this key'); return }
  try {
    const result = await createApiKey({ name: form.value.name.trim(), requestsPerMinute: form.value.requestsPerMinute })
    createdSecret.value = result.key || ''
    showCreate.value = false
    showSecret.value = true
    form.value = { name: '', requestsPerMinute: 60 }
    await loadKeys()
  }
  catch { message.error('Could not create API key') }
}

function confirmRevoke(key: VirtualApiKey) {
  dialog.warning({
    title: 'Revoke API key',
    content: `Revoke “${key.name}”? Applications using it will stop working immediately.`,
    positiveText: 'Revoke', negativeText: 'Cancel',
    onPositiveClick: async () => { await revokeApiKey(key.id); message.success('API key revoked'); await loadKeys() },
  })
}

async function openUsage(key: VirtualApiKey) {
  usage.value = await fetchApiKeyUsage(key.id)
  showUsage.value = true
}

function copySecret() { copy(createdSecret.value); message.success('API key copied') }

const columns: DataTableColumns<VirtualApiKey> = [
  { title: 'Name', key: 'name' },
  { title: 'Key', key: 'keyPrefix', render: row => `${row.keyPrefix}…` },
  { title: 'Status', key: 'status', render: row => h(NTag, { type: row.status === 'active' ? 'success' : 'default', size: 'small' }, { default: () => row.status }) },
  { title: 'Limit', key: 'requestsPerMinute', render: row => `${row.requestsPerMinute}/min` },
  { title: 'Last used', key: 'lastUsedAt', render: row => formatDate(row.lastUsedAt) },
  { title: 'Created', key: 'createdAt', render: row => formatDate(row.createdAt) },
  { title: 'Actions', key: 'actions', render: row => h(NSpace, null, { default: () => [
    h(NButton, { size: 'small', onClick: () => openUsage(row) }, { default: () => 'Usage' }),
    row.status === 'active' ? h(NButton, { size: 'small', type: 'error', secondary: true, onClick: () => confirmRevoke(row) }, { default: () => 'Revoke' }) : null,
  ] }) },
]

const usageColumns: DataTableColumns<ApiKeyUsage> = [
  { title: 'Model', key: 'requestedModel' }, { title: 'Requests', key: 'requestCount' },
  { title: 'Input tokens', key: 'promptTokens' }, { title: 'Output tokens', key: 'completionTokens' }, { title: 'Total tokens', key: 'totalTokens' },
]

onMounted(async () => { await authStore.initializeAuth(); if (authStore.isValid) await loadKeys() })
</script>

<template>
  <div>
    <NCard title="API keys">
      <template #header-extra><NButton type="primary" @click="showCreate = true">Create API key</NButton></template>
      <p class="mb-5 text-gray-500">Use virtual keys with the OpenAI-compatible base URL <code>{{ gatewayBaseUrl }}</code>.</p>
      <NDataTable :columns="columns" :data="keys" :loading="loading" :row-key="row => row.id" />
    </NCard>

    <NModal v-model:show="showCreate" preset="card" title="Create API key" class="max-w-lg">
      <NFormItem label="Name"><NInput v-model:value="form.name" placeholder="My application" /></NFormItem>
      <NFormItem label="Requests per minute"><NInputNumber v-model:value="form.requestsPerMinute" :min="1" :max="10000" /></NFormItem>
      <div class="flex justify-end gap-2"><NButton @click="showCreate = false">Cancel</NButton><NButton type="primary" @click="createKey">Create</NButton></div>
    </NModal>

    <NModal v-model:show="showSecret" preset="card" title="Copy your API key" class="max-w-xl" :mask-closable="false">
      <p class="mb-3 text-amber-600">This key is shown only once. Store it securely.</p>
      <NInput :value="createdSecret" readonly type="textarea" :autosize="{ minRows: 2 }" />
      <div class="flex justify-end gap-2 mt-4"><NButton type="primary" @click="copySecret">Copy</NButton><NButton @click="showSecret = false">Done</NButton></div>
    </NModal>

    <NModal v-model:show="showUsage" preset="card" title="API usage" class="max-w-4xl"><NDataTable :columns="usageColumns" :data="usage" /></NModal>
    <Permission :visible="needPermission" />
  </div>
</template>
