<script lang="ts" setup>
// create a data table with pagination using naive-ui, with the following columns:
// User Email, Total Sessions, Total Messages, Total Sessions (3 days), Total Messages (3 days), Rate Limit
// The data should be fetched from the backend using api 'GetUserData(page, page_size)'
// The Rate Limit column should be editable, and the value should be updated in the backend using api 'UpdateRateLimit(user_email, rate_limit)'
// vue3 code should be in <script lang="ts" setup> style.
import { computed, h, onMounted, reactive, ref } from 'vue'
import { NButton, NCard, NDataTable, NEmpty, NForm, NFormItem, NInput, NModal, useMessage } from 'naive-ui'
import { GetUserData, UpdateRateLimit, updateUserFullName } from '@/api'
import { t } from '@/locales'
import UserAnalysisModal from '@/components/admin/UserAnalysisModal.vue'

const ms_ui = useMessage()

const showEditModal = ref(false)
const editingUser = ref<UserData | null>(null)
const showAnalysisModal = ref(false)
const selectedUserEmail = ref('')

interface UserData {
  email: string
  firstName: string
  lastName: string
  totalChatMessages: number
  totalChatMessagesTokenCount: number
  totalChatMessages3Days: number
  totalChatMessages3DaysTokenCount: number
  totalChatMessages3DaysAvgTokenCount: number
  rateLimit: string
}
const tableData = ref<UserData[]>([])
const loading = ref<boolean>(true)

const columns = [
  {
    title: t('admin.userEmail'),
    key: 'email',
    width: 230,
    render: (row: UserData) => {
      return h('span', {
        class: 'cursor-pointer text-blue-600 hover:text-blue-800 hover:underline',
        onClick: () => {
          selectedUserEmail.value = row.email
          showAnalysisModal.value = true
        },
      }, row.email)
    },
  },
  {
    title: t('admin.name'),
    key: 'name',
    width: 130,
    render: (row: UserData) => {
      return h('span', `${row.lastName}${row.firstName}`)
    },
  },

  {
    title: t('admin.rateLimit10Min'),
    key: 'rateLimit',
    width: 140,
  },
  {
    title: t('admin.totalChatMessages'),
    key: 'totalChatMessages',
    width: 130,
  },
  {
    title: t('admin.totalChatMessages3Days'),
    key: 'totalChatMessages3Days',
    width: 150,
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 100,
    fixed: 'right' as const,
    render: (row: UserData) => h(NButton, {
      size: 'small',
      secondary: true,
      onClick: () => {
        editingUser.value = { ...row }
        showEditModal.value = true
      },
    }, { default: () => t('common.edit') }),
  },
]

const search = ref('')
const filteredData = computed(() => {
  const term = search.value.trim().toLowerCase()
  if (!term)
    return tableData.value
  return tableData.value.filter(user =>
    user.email.toLowerCase().includes(term)
    || `${user.firstName} ${user.lastName}`.toLowerCase().includes(term),
  )
})

const pagination = reactive({
  page: 1,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  pageSize: 10,
  itemCount: 10,
  onChange: async (page: number) => {
    pagination.page = page
    await fetchData()
  },
  onUpdatePageSize: async (pageSize: number) => {
    pagination.pageSize = pageSize
    pagination.page = 1
    await fetchData()
  },
})

async function fetchData() {
  loading.value = true
  try {
    const { data, total } = await GetUserData(pagination.page, pagination.pageSize)
    tableData.value = data
    pagination.itemCount = total
  }
  catch (err: any) {
    if (err.response.status === 401)
      ms_ui.error(t(err.response.data.message))
    else
      ms_ui.error(t(err.response.data.message))
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})

async function handleRefresh() {
  await fetchData()
}

async function handleSave() {
  if (!editingUser.value)
    return

  try {
    await updateUserFullName({
      firstName: editingUser.value.firstName,
      lastName: editingUser.value.lastName,
      email: editingUser.value.email,
    })
    await UpdateRateLimit(editingUser.value.email, parseInt(editingUser.value.rateLimit))
    ms_ui.success(t('common.updateSuccess'))
    showEditModal.value = false
    await fetchData()
  }
  catch (error: any) {
    ms_ui.error(error.message || t('common.updateFailed'))
  }
}
</script>

<template>
  <UserAnalysisModal v-model:visible="showAnalysisModal" :user-email="selectedUserEmail" />
  <NModal v-model:show="showEditModal" preset="card" class="w-full max-w-lg" :title="t('common.editUser')">
    <NCard :bordered="false" embedded>
      <NForm label-placement="top">
        <NFormItem :label="t('admin.lastName')">
          <NInput v-model:value="editingUser!.lastName" />
        </NFormItem>
        <NFormItem :label="t('admin.firstName')">
          <NInput v-model:value="editingUser!.firstName" />
        </NFormItem>
        <NFormItem :label="t('admin.rateLimit10Min')">
          <NInput v-model:value="editingUser!.rateLimit" />
        </NFormItem>
        <div class="flex justify-end gap-2 pt-2">
          <NButton @click="showEditModal = false">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="primary" @click="handleSave">
            {{ t('common.save') }}
          </NButton>
        </div>
      </NForm>
    </NCard>
  </NModal>
  <div class="admin-toolbar">
    <NInput v-model:value="search" clearable class="max-w-xs" :placeholder="t('admin.searchUsers')" aria-label="Search users" />
    <NButton :loading="loading" secondary @click="handleRefresh">
      {{ t('admin.refresh') }}
    </NButton>
  </div>
  <NDataTable
    size="small" :loading="loading" remote :data="filteredData" :columns="columns"
    :pagination="pagination" :scroll-x="880" :single-line="false"
  >
    <template #empty>
      <NEmpty :description="t('admin.noUsers')" />
    </template>
  </NDataTable>
</template>

<style scoped>
.admin-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
@media (max-width: 640px) { .admin-toolbar { align-items: stretch; flex-direction: column; } }
</style>
