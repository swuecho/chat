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
  if (value === 'sessions' && sessionHistoryData.value.length === 0) {
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
    render: (row: any) => `${row.percentage}%`
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
</script>

<template>
  <SessionSnapshotModal 
    v-model:visible="showSessionSnapshot"
    :session-id="selectedSessionId"
    :session-model="selectedSessionModel"
    :user-email="userEmail"
  />
  <NModal v-model:show="show" :style="{ width: ['95vw', '1200px'] }">
    <NCard 
      role="dialog" 
      aria-modal="true" 
      :title="`${t('admin.userAnalysis')} - ${userEmail}`"
      :bordered="false" 
      size="huge"
    >
      <NSpin :show="loading">
        <div v-if="analysisData">
          <NTabs type="card" animated @update:value="handleTabChange">
            <!-- Overview Tab -->
            <NTabPane name="overview" :tab="t('admin.overview')">
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
                <NStatistic :label="t('admin.totalMessages')" :value="analysisData.userInfo.totalMessages" />
                <NStatistic :label="t('admin.totalTokens')" :value="analysisData.userInfo.totalTokens" />
                <NStatistic :label="t('admin.totalSessions')" :value="analysisData.userInfo.totalSessions" />
                <NStatistic :label="t('admin.rateLimit')" :value="`${analysisData.userInfo.rateLimit}/10min`" />
              </div>
              
              <div class="mb-6">
                <h3 class="text-lg font-semibold mb-4">{{ t('admin.recent3Days') }}</h3>
                <div class="grid grid-cols-2 gap-4">
                  <NStatistic :label="t('admin.messages3Days')" :value="analysisData.userInfo.messages3Days" />
                  <NStatistic :label="t('admin.tokens3Days')" :value="analysisData.userInfo.tokens3Days" />
                </div>
              </div>

              <div>
                <h3 class="text-lg font-semibold mb-4">{{ t('admin.modelUsageDistribution') }}</h3>
                <div class="space-y-3">
                  <div v-for="model in analysisData.modelUsage" :key="model.model" class="flex items-center gap-4">
                    <div class="w-24 text-sm">{{ model.model }}</div>
                    <div class="flex-1">
                      <NProgress 
                        :percentage="model.percentage" 
                        :show-indicator="false"
                        :color="model.model === 'GPT-4' ? '#10b981' : model.model === 'Claude-3-Sonnet' ? '#3b82f6' : '#f59e0b'"
                      />
                    </div>
                    <div class="w-16 text-sm text-right">{{ model.percentage }}%</div>
                  </div>
                </div>
              </div>
            </NTabPane>

            <!-- Model Usage Tab -->
            <NTabPane name="models" :tab="t('admin.modelUsage')">
              <NDataTable 
                :data="analysisData.modelUsage" 
                :columns="modelUsageColumns"
                :pagination="false"
                size="small"
              />
            </NTabPane>

            <!-- Activity History Tab -->
            <NTabPane name="activity" :tab="t('admin.activityHistory')">
              <NDataTable 
                :data="analysisData.recentActivity" 
                :columns="activityColumns"
                :pagination="{ pageSize: 10 }"
                size="small"
              />
            </NTabPane>

            <!-- Session History Tab -->
            <NTabPane name="sessions" :tab="t('admin.sessionHistory')">
              <NSpin :show="sessionLoading">
                <NDataTable 
                  :data="sessionHistoryData" 
                  :columns="sessionColumns"
                  :pagination="sessionPagination"
                  :remote="true"
                  size="small"
                />
              </NSpin>
            </NTabPane>
          </NTabs>
        </div>
      </NSpin>
    </NCard>
  </NModal>
</template>