<script lang="ts" setup>
// create a data table with pagination using naive-ui, with the following columns:
// User Email, Total Sessions, Total Messages, Total Sessions (3 days), Total Messages (3 days), Rate Limit
// The data should be fetched from the backend using api 'GetUserData(page, page_size)'
// The Rate Limit column should be editable, and the value should be updated in the backend using api 'UpdateRateLimit(user_email, rate_limit)'
// vue3 code should be in <script lang="ts" setup> style.
import { h, onMounted, reactive, ref } from 'vue'
import { NDataTable, NInput, useMessage, NButton, NModal, NForm, NFormItem, useDialog, NCard } from 'naive-ui'
import { GetUserData, UpdateRateLimit, updateUserFullName } from '@/api'
import { t } from '@/locales'
import HoverButton from '@/components/common/HoverButton/index.vue'
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
    width: 200,
    render: (row: UserData) => {
      return h('span', {
        class: 'cursor-pointer text-blue-600 hover:text-blue-800 hover:underline',
        onClick: () => {
          selectedUserEmail.value = row.email
          showAnalysisModal.value = true
        }
      }, row.email)
    }
  },
  {
    title: t('admin.name'),
    key: 'name',
    width: 100,
    render: (row: UserData) => {
      return h('span', `${row.lastName}${row.firstName}`)
    }
  },

  {
    title: t('admin.rateLimit10Min'),
    key: 'rateLimit',
    width: 100,
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 100,
    render: (row: UserData) => {
      return h(NButton, {
        size: 'small',
        onClick: () => {
          editingUser.value = { ...row }
          showEditModal.value = true
        }
      }, {
        default: () => t('common.edit')
      })
    }
  },
  {
    title: t('admin.totalChatMessages'),
    key: 'totalChatMessages',
    width: 100,
  },
  {
    title: t('admin.totalChatMessagesTokenCount'),
    key: 'totalChatMessagesTokenCount',
    width: 100,
  },
  {
    title: t('admin.totalChatMessages3Days'),
    key: 'totalChatMessages3Days',
    width: 100,
  },
  {
    title: t('admin.totalChatMessages3DaysTokenCount'),
    key: 'totalChatMessages3DaysTokenCount',
    width: 100,
  },
  {
    title: t('admin.totalChatMessages3DaysAvgTokenCount'),
    key: 'avgChatMessages3DaysTokenCount',
    width: 100,
  },

]

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
  } finally {
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
  if (!editingUser.value) return

  try {
    await updateUserFullName({
      firstName: editingUser.value.firstName,
      lastName: editingUser.value.lastName,
      email: editingUser.value.email
    })
    await UpdateRateLimit(editingUser.value.email, parseInt(editingUser.value.rateLimit))
    ms_ui.success(t('common.updateSuccess'))
    showEditModal.value = false
    await fetchData()
  } catch (error: any) {
    ms_ui.error(error.message || t('common.updateFailed'))
  }
}
</script>

<template>
  <UserAnalysisModal v-model:visible="showAnalysisModal" :user-email="selectedUserEmail" />
  <NModal v-model:show="showEditModal">
    <NCard style="width: 600px" :title="t('common.editUser')" :bordered="false" size="huge">
      <NForm label-placement="left" label-width="auto">
        <NFormItem :label="t('admin.lastName')">
          <NInput v-model:value="editingUser!.lastName" />
        </NFormItem>
        <NFormItem :label="t('admin.firstName')">
          <NInput v-model:value="editingUser!.firstName" />
        </NFormItem>
        <NFormItem :label="t('admin.rateLimit10Min')">
          <NInput v-model:value="editingUser!.rateLimit" />
        </NFormItem>
        <div class="flex justify-end gap-4">
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
  <div class="flex items-center justify-between mb-4">
    <h1 class="text-xl font-semibold text-gray-900 dark:text-white">
      {{ t('admin.userMessage') }}
    </h1>
    <HoverButton :tooltip="t('admin.refresh')" @click="handleRefresh">

      <span class="text-xl text-[#4f555e] dark:text-white">
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
          <path fill="currentColor"
            d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z" />
        </svg>
      </span>
    </HoverButton>
  </div>
  <NDataTable :loading="loading" remote :data="tableData" :columns="columns" :pagination="pagination" />
</template>
