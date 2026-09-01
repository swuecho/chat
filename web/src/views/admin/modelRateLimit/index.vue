<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NButton, NDataTable, NEmpty, NInput, NModal, useDialog, useMessage } from 'naive-ui'
import AddChatModelForm from './addChatModelForm.vue'
import { deleteUserChatModelPrivilege, listUserChatModelPrivileges, updateUserChatModelPrivilege } from '@/api/generated_client'
import { SvgIcon } from '@/components/common'
import { t } from '@/locales'

const dialogVisible = ref(false)
const dialog = useDialog()
const message = useMessage()
const search = ref('')

const data = ref<Chat.ChatModelPrivilege[]>([])
const loading = ref(true)

onMounted(async () => {
  refreshData()
})

async function refreshData() {
  loading.value = true
  try {
    data.value = await listUserChatModelPrivileges()
  }
  finally {
    loading.value = false
  }
}

async function updateRow(row: Chat.ChatModelPrivilege) {
  await updateUserChatModelPrivilege({ path: { id: Number(row.id) }, body: { ...row, rateLimit: parseInt(row.rateLimit) } })
  message.success(t('common.updateSuccess'))
}

const filteredData = computed(() => {
  const term = search.value.trim().toLowerCase()
  if (!term)
    return data.value
  return data.value.filter(row => [row.fullName, row.userEmail, row.chatModelName]
    .some(value => value?.toLowerCase().includes(term)))
})

function createColumns(): DataTableColumns<Chat.ChatModelPrivilege> {
  const userEmailField = {
    title: t('admin.per_model_rate_limit.UserEmail'),
    key: 'userEmail',
    width: 180,
  }

  const userFullNameField = {
    title: t('admin.per_model_rate_limit.FullName'),
    key: 'fullName',
    width: 170,
  }

  const modelField = {
    title: t('admin.per_model_rate_limit.ChatModelName'),
    key: 'chatModelName',
    width: 180,
  }

  const ratelimitField = {
    title: t('admin.per_model_rate_limit.RateLimit'),
    key: 'rateLimit',
    width: 150,
    render(row: Chat.ChatModelPrivilege) {
      return h(NInput, {
        value: row.rateLimit.toString(),
        onUpdateValue(v: string) {
          // assuming that `data` is an array of FormData objects
          row.rateLimit = v
        },
        onBlur: () => updateRow(row),
        inputProps: { 'aria-label': `${t('admin.rateLimit')} ${row.userEmail}` },
      })
    },
  }

  const actionField = {
    title: t('admin.per_model_rate_limit.actions'),
    key: 'actions',
    render(row: any) {
      return h(NButton, { size: 'small', type: 'error', secondary: true, onClick: () => confirmDelete(row) }, {
        default: () => t('common.delete'),
      })
    },
  }

  return ([
    userFullNameField,
    userEmailField,
    modelField,
    ratelimitField,
    actionField,
  ])
}

const columns = createColumns()

function confirmDelete(row: Chat.ChatModelPrivilege) {
  dialog.warning({
    title: t('admin.deleteRateLimit'),
    content: t('admin.deleteRateLimitConfirm', { email: row.userEmail, model: row.chatModelName }),
    positiveText: t('common.delete'),
    negativeText: t('common.cancel'),
    onPositiveClick: () => deleteRow(row),
  })
}

async function deleteRow(row: Chat.ChatModelPrivilege) {
  await deleteUserChatModelPrivilege({ path: { id: Number(row.id) } })
  message.success(t('admin.rateLimitDeleted'))
  await refreshData()
}

async function newRowAdded() {
  await refreshData()
}
</script>

<template>
  <div class="admin-toolbar">
    <NInput v-model:value="search" clearable class="max-w-xs" :placeholder="t('admin.searchRateLimits')" aria-label="Search rate limits" />
    <NButton type="primary" @click="dialogVisible = true">
      <template #icon>
        <SvgIcon icon="material-symbols:add-rounded" />
      </template>
      {{ t('admin.addRateLimit') }}
    </NButton>
  </div>
  <NDataTable size="small" :columns="columns" :data="filteredData" :loading="loading" :scroll-x="850" :single-line="false">
    <template #empty>
      <NEmpty :description="t('admin.noRateLimits')" />
    </template>
  </NDataTable>
  <NModal v-model:show="dialogVisible" :title="$t('admin.add_user_model_rate_limit')" preset="card" class="w-full max-w-lg">
    <AddChatModelForm @new-row-added="newRowAdded" />
  </NModal>
</template>

<style scoped>
.admin-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
@media (max-width: 640px) { .admin-toolbar { align-items: stretch; flex-direction: column; } }
</style>
