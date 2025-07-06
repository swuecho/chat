<script lang="ts" setup>
import { ref, watch, computed, h } from 'vue'
import { NModal, NCard, NTabs, NTabPane, NSpin, NStatistic, NProgress, NDataTable, useMessage, NButton } from 'naive-ui'
import { getUserAnalysis, getUserSessionHistory } from '@/api'
import SessionSnapshotModal from './SessionSnapshotModal.vue'
import { t } from '@/locales'

interface Props {
  visible: boolean
  userEmail: string
}

interface UserAnalysisData {
  userInfo: {
    email: string
    totalMessages: number
    totalTokens: number
    totalSessions: number
    messages3Days: number
    tokens3Days: number
    rateLimit: number
  }
  modelUsage: Array<{
    model: string
    messageCount: number
    tokenCount: number
    percentage: number
    lastUsed: string
  }>
  recentActivity: Array<{
    date: string
    messages: number
    tokens: number
    sessions: number
  }>
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const message = useMessage()
const loading = ref(false)
const sessionLoading = ref(false)
const analysisData = ref<UserAnalysisData | null>(null)
const sessionHistoryData = ref<any[]>([])
const showSessionSnapshot = ref(false)
const selectedSessionId = ref('')
const selectedSessionModel = ref('')
const sessionPagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  onChange: (page: number) => {
    sessionPagination.value.page = page
    fetchSessionHistory()
  },
  onUpdatePageSize: (pageSize: number) => {
    sessionPagination.value.pageSize = pageSize
    sessionPagination.value.page = 1
    fetchSessionHistory()
  }
})

const show = computed({
  get: () => props.visible,
  set: (visible: boolean) => emit('update:visible', visible)
})

// Watch for when modal opens to fetch data
watch(() => props.visible, (newVal) => {
  if (newVal && props.userEmail) {
    fetchUserAnalysis()
    // Don't fetch session history immediately - let it load when tab is accessed
  }
})

async function fetchUserAnalysis() {
  loading.value = true
  try {
    const response = await getUserAnalysis(props.userEmail)
    analysisData.value = response
  } catch (error: any) {
    message.error(error.message || t('common.fetchFailed'))
  } finally {
    loading.value = false
  }
}

async function fetchSessionHistory() {
  sessionLoading.value = true
  try {
    const response = await getUserSessionHistory(
      props.userEmail, 
      sessionPagination.value.page, 
      sessionPagination.value.pageSize
    )
    sessionHistoryData.value = response.data
    sessionPagination.value.itemCount = response.total
  } catch (error: any) {
    message.error(error.message || t('common.fetchFailed'))
  } finally {
    sessionLoading.value = false
  }
}

// Handle tab change to load session history when needed
function handleTabChange(value: string) {
  if (value === 'sessions') {
    fetchSessionHistory()
  }
}

// Handle session ID click to show snapshot
function handleSessionClick(sessionId: string, model: string) {
  selectedSessionId.value = sessionId
  selectedSessionModel.value = model
  showSessionSnapshot.value = true
}

const modelUsageColumns = [
  { title: t('admin.model'), key: 'model', width: 120 },
  { title: t('admin.messages'), key: 'messageCount', width: 100 },
  { title: t('admin.tokens'), key: 'tokenCount', width: 100 },
  { 
    title: t('admin.usage'), 
    key: 'percentage', 
    width: 100,
    render: (row: any) => `${row.percentage.toFixed(2)}%`
  },
  { title: t('admin.lastUsed'), key: 'lastUsed', width: 120 }
]

const activityColumns = [
  { title: t('admin.date'), key: 'date', width: 120 },
  { title: t('admin.messages'), key: 'messages', width: 100 },
  { title: t('admin.tokens'), key: 'tokens', width: 100 },
  { title: t('admin.sessions'), key: 'sessions', width: 100 }
]

const sessionColumns = [
  { 
    title: t('admin.sessionId'), 
    key: 'sessionId', 
    width: 120,
    render: (row: any) => {
      return h(NButton, {
        text: true,
        type: 'primary',
        size: 'small',
        onClick: () => handleSessionClick(row.sessionId, row.model)
      }, {
        default: () => row.sessionId.slice(0, 8) + '...'
      })
    }
  },
  { title: t('admin.model'), key: 'model', width: 120 },
  { title: t('admin.messages'), key: 'messageCount', width: 100 },
  { title: t('admin.tokens'), key: 'tokenCount', width: 100 },
  { title: t('admin.created'), key: 'createdAt', width: 150 },
  { title: t('admin.updated'), key: 'updatedAt', width: 150 }
]

// Helper function to get consistent model colors
function getModelColor(modelName: string): string {
  const colorMap: Record<string, string> = {
    'GPT-4': '#10b981',
    'GPT-3.5': '#06b6d4',
    'Claude-3-Sonnet': '#3b82f6',
    'Claude-3-Haiku': '#8b5cf6',
    'Claude-3-Opus': '#ec4899',
    'Gemini': '#f59e0b',
    'Llama': '#ef4444'
  }
  
  // Find matching color by checking if model name contains any key
  for (const [key, color] of Object.entries(colorMap)) {
    if (modelName.toLowerCase().includes(key.toLowerCase())) {
      return color
    }
  }
  
  // Default color for unknown models
  return '#6b7280'
}
</script>

<template>
  <SessionSnapshotModal 
    v-model:visible="showSessionSnapshot"
    :session-id="selectedSessionId"
    :session-model="selectedSessionModel"
    :user-email="userEmail"
  />
  <NModal v-model:show="show" :style="{ width: ['95vw', '1400px'] }" class="elegant-modal">
    <NCard 
      role="dialog" 
      aria-modal="true" 
      :title="`${t('admin.userAnalysis')} - ${userEmail}`"
      :bordered="false" 
      size="huge"
      class="elegant-card"
    >
      <template #header>
        <div class="flex items-center gap-3">
          <div class="w-2 h-8 bg-gradient-to-b from-blue-500 to-purple-600 rounded-full"></div>
          <div>
            <h2 class="text-xl font-bold text-gray-800 dark:text-gray-200">{{ t('admin.userAnalysis') }}</h2>
            <p class="text-sm text-gray-500 font-mono">{{ userEmail }}</p>
          </div>
        </div>
      </template>
      
      <NSpin :show="loading">
        <div v-if="analysisData" class="space-y-6">
          <NTabs type="line" animated @update:value="handleTabChange" class="elegant-tabs">
            <!-- Overview Tab -->
            <NTabPane name="overview" :tab="t('admin.overview')">
              <!-- Key Metrics Cards -->
              <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                <div class="metric-card bg-gradient-to-br from-blue-50 to-blue-100 dark:from-blue-900/20 dark:to-blue-800/20 rounded-xl p-6 border border-blue-200 dark:border-blue-800">
                  <div class="flex items-center justify-between mb-3">
                    <div class="p-2 bg-blue-500 rounded-lg">
                      <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"></path>
                      </svg>
                    </div>
                    <span class="text-2xl font-bold text-blue-700 dark:text-blue-300">{{ analysisData.userInfo.totalMessages.toLocaleString() }}</span>
                  </div>
                  <p class="text-sm font-medium text-blue-600 dark:text-blue-400">{{ t('admin.totalMessages') }}</p>
                </div>
                
                <div class="metric-card bg-gradient-to-br from-emerald-50 to-emerald-100 dark:from-emerald-900/20 dark:to-emerald-800/20 rounded-xl p-6 border border-emerald-200 dark:border-emerald-800">
                  <div class="flex items-center justify-between mb-3">
                    <div class="p-2 bg-emerald-500 rounded-lg">
                      <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
                      </svg>
                    </div>
                    <span class="text-2xl font-bold text-emerald-700 dark:text-emerald-300">{{ analysisData.userInfo.totalTokens.toLocaleString() }}</span>
                  </div>
                  <p class="text-sm font-medium text-emerald-600 dark:text-emerald-400">{{ t('admin.totalTokens') }}</p>
                </div>
                
                <div class="metric-card bg-gradient-to-br from-purple-50 to-purple-100 dark:from-purple-900/20 dark:to-purple-800/20 rounded-xl p-6 border border-purple-200 dark:border-purple-800">
                  <div class="flex items-center justify-between mb-3">
                    <div class="p-2 bg-purple-500 rounded-lg">
                      <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
                      </svg>
                    </div>
                    <span class="text-2xl font-bold text-purple-700 dark:text-purple-300">{{ analysisData.userInfo.totalSessions.toLocaleString() }}</span>
                  </div>
                  <p class="text-sm font-medium text-purple-600 dark:text-purple-400">{{ t('admin.totalSessions') }}</p>
                </div>
                
                <div class="metric-card bg-gradient-to-br from-amber-50 to-amber-100 dark:from-amber-900/20 dark:to-amber-800/20 rounded-xl p-6 border border-amber-200 dark:border-amber-800">
                  <div class="flex items-center justify-between mb-3">
                    <div class="p-2 bg-amber-500 rounded-lg">
                      <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                      </svg>
                    </div>
                    <span class="text-2xl font-bold text-amber-700 dark:text-amber-300">{{ analysisData.userInfo.rateLimit }}/10min</span>
                  </div>
                  <p class="text-sm font-medium text-amber-600 dark:text-amber-400">{{ t('admin.rateLimit') }}</p>
                </div>
              </div>
              
              <!-- Recent Activity Section -->
              <div class="bg-gray-50 dark:bg-gray-800/50 rounded-xl p-6 mb-8">
                <div class="flex items-center gap-3 mb-6">
                  <div class="p-2 bg-indigo-500 rounded-lg">
                    <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path>
                    </svg>
                  </div>
                  <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-200">{{ t('admin.recent3Days') }}</h3>
                </div>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div class="bg-white dark:bg-gray-700 rounded-lg p-4 border border-gray-200 dark:border-gray-600">
                    <div class="flex items-center justify-between">
                      <span class="text-sm font-medium text-gray-600 dark:text-gray-400">{{ t('admin.messages3Days') }}</span>
                      <span class="text-xl font-bold text-indigo-600 dark:text-indigo-400">{{ analysisData.userInfo.messages3Days.toLocaleString() }}</span>
                    </div>
                  </div>
                  <div class="bg-white dark:bg-gray-700 rounded-lg p-4 border border-gray-200 dark:border-gray-600">
                    <div class="flex items-center justify-between">
                      <span class="text-sm font-medium text-gray-600 dark:text-gray-400">{{ t('admin.tokens3Days') }}</span>
                      <span class="text-xl font-bold text-indigo-600 dark:text-indigo-400">{{ analysisData.userInfo.tokens3Days.toLocaleString() }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Model Usage Distribution -->
              <div class="bg-white dark:bg-gray-800 rounded-xl p-6 border border-gray-200 dark:border-gray-700">
                <div class="flex items-center gap-3 mb-6">
                  <div class="p-2 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg">
                    <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
                    </svg>
                  </div>
                  <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-200">{{ t('admin.modelUsageDistribution') }}</h3>
                </div>
                <div class="space-y-4">
                  <div v-for="model in analysisData.modelUsage" :key="model.model" class="model-usage-item group">
                    <div class="flex items-center gap-4 p-4 rounded-lg bg-gray-50 dark:bg-gray-700/50 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                      <div class="flex items-center gap-3 min-w-0 flex-1">
                        <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: getModelColor(model.model) }"></div>
                        <span class="font-medium text-gray-800 dark:text-gray-200 truncate">{{ model.model }}</span>
                      </div>
                      <div class="flex-1 mx-4">
                        <NProgress 
                          :percentage="model.percentage" 
                          :show-indicator="false"
                          :color="getModelColor(model.model)"
                          :height="8"
                          class="model-progress"
                        />
                      </div>
                      <div class="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400">
                        <span class="font-mono">{{ model.messageCount.toLocaleString() }} msg</span>
                        <span class="font-mono">{{ model.tokenCount.toLocaleString() }} tok</span>
                        <span class="font-bold text-gray-800 dark:text-gray-200 min-w-[3rem] text-right">{{ model.percentage.toFixed(2) }}%</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </NTabPane>

            <!-- Model Usage Tab -->
            <NTabPane name="models" :tab="t('admin.modelUsage')">
              <div class="bg-white dark:bg-gray-800 rounded-xl overflow-hidden border border-gray-200 dark:border-gray-700">
                <NDataTable 
                  :data="analysisData.modelUsage" 
                  :columns="modelUsageColumns"
                  :pagination="false"
                  size="medium"
                  :bordered="false"
                  class="elegant-table"
                />
              </div>
            </NTabPane>

            <!-- Activity History Tab -->
            <NTabPane name="activity" :tab="t('admin.activityHistory')">
              <div class="bg-white dark:bg-gray-800 rounded-xl overflow-hidden border border-gray-200 dark:border-gray-700">
                <NDataTable 
                  :data="analysisData.recentActivity" 
                  :columns="activityColumns"
                  :pagination="{ pageSize: 10 }"
                  size="medium"
                  :bordered="false"
                  class="elegant-table"
                />
              </div>
            </NTabPane>

            <!-- Session History Tab -->
            <NTabPane name="sessions" :tab="t('admin.sessionHistory')">
              <div class="bg-white dark:bg-gray-800 rounded-xl overflow-hidden border border-gray-200 dark:border-gray-700">
                <NSpin :show="sessionLoading">
                  <NDataTable 
                    :data="sessionHistoryData" 
                    :columns="sessionColumns"
                    :pagination="sessionPagination"
                    :remote="true"
                    size="medium"
                    :bordered="false"
                    class="elegant-table"
                  />
                </NSpin>
              </div>
            </NTabPane>
          </NTabs>
        </div>
      </NSpin>
    </NCard>
  </NModal>
</template>

<style scoped>
.elegant-modal :deep(.n-modal) {
  backdrop-filter: blur(8px);
}

.elegant-card {
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  border-radius: 16px;
}

.elegant-tabs :deep(.n-tabs-nav) {
  background: transparent;
  border-bottom: 2px solid #e5e7eb;
  border-radius: 0;
  padding: 0;
  margin-bottom: 24px;
}

.elegant-tabs :deep(.n-tabs-tab) {
  border-radius: 0;
  border-bottom: 3px solid transparent;
  transition: all 0.3s ease;
  padding: 12px 24px;
  margin-right: 8px;
  font-weight: 500;
  color: #6b7280;
}

.elegant-tabs :deep(.n-tabs-tab:hover) {
  color: #374151;
  background: rgba(59, 130, 246, 0.05);
  border-radius: 8px 8px 0 0;
}

.elegant-tabs :deep(.n-tabs-tab--active) {
  background: transparent;
  box-shadow: none;
  color: #3b82f6;
  border-bottom-color: #3b82f6;
  font-weight: 600;
}

.metric-card {
  transition: all 0.3s ease;
  backdrop-filter: blur(10px);
}

.metric-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1);
}

.model-usage-item {
  transition: all 0.2s ease;
}

.model-usage-item:hover {
  transform: translateX(4px);
}

.model-progress :deep(.n-progress-graph) {
  border-radius: 4px;
}

.model-progress :deep(.n-progress-graph-line) {
  transition: all 0.3s ease;
}

.elegant-table :deep(.n-data-table-thead) {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
}

.elegant-table :deep(.n-data-table-th) {
  font-weight: 600;
  color: #374151;
  border-bottom: 2px solid #e5e7eb;
}

.elegant-table :deep(.n-data-table-td) {
  border-bottom: 1px solid #f3f4f6;
  transition: background-color 0.2s ease;
}

.elegant-table :deep(.n-data-table-tr:hover .n-data-table-td) {
  background-color: #f8fafc;
}

.dark .elegant-tabs :deep(.n-tabs-nav) {
  background: transparent;
  border-bottom: 2px solid #4b5563;
}

.dark .elegant-tabs :deep(.n-tabs-tab) {
  color: #9ca3af;
}

.dark .elegant-tabs :deep(.n-tabs-tab:hover) {
  color: #d1d5db;
  background: rgba(59, 130, 246, 0.1);
}

.dark .elegant-tabs :deep(.n-tabs-tab--active) {
  background: transparent;
  box-shadow: none;
  color: #60a5fa;
  border-bottom-color: #60a5fa;
}

.dark .elegant-table :deep(.n-data-table-thead) {
  background: linear-gradient(135deg, #374151 0%, #1f2937 100%);
}

.dark .elegant-table :deep(.n-data-table-th) {
  color: #d1d5db;
  border-bottom: 2px solid #4b5563;
}

.dark .elegant-table :deep(.n-data-table-td) {
  border-bottom: 1px solid #374151;
}

.dark .elegant-table :deep(.n-data-table-tr:hover .n-data-table-td) {
  background-color: #1f2937;
}
</style>