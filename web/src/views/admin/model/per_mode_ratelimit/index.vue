<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NDataTable, NInput, NModal } from 'naive-ui'
import AddChatModelForm from './addChatModelForm.vue'
import { DeleteUserChatModelPrivilege, ListUserChatModelPrivilege, UpdateUserChatModelPrivilege } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { t } from '@/locales'

interface RowData {
  ID: string
  ChatModelName: string
  FullName: string
  UserEmail: string
  RateLimit: string
}

const dialogVisible = ref(false)

const data = ref<RowData[]>([])

onMounted(async () => {
  refreshData()
})

async function refreshData() {
  data.value = await ListUserChatModelPrivilege()
}

function UpdateRow(row: RowData) {
  // @ts-expect-error rateLimit is a number in golang
  row.RateLimit = parseInt(row.RateLimit)
  UpdateUserChatModelPrivilege(row.ID, row)
}

function createColumns(): DataTableColumns<RowData> {
  const userEmailField = {
    title: t('admin.per_model_rate_limit.UserEmail'),
    key: 'UserEmail',
    width: 200,
  }

  const userFullNameField = {
    title: t('admin.per_model_rate_limit.FullName'),
    key: 'FullName',
    width: 200,
  }

  const modelField = {
    title: t('admin.per_model_rate_limit.ChatModelName'),
    key: 'ChatModelName',
    width: 250,
  }

  const ratelimitField = {
    title: t('admin.per_model_rate_limit.RateLimit'),
    key: 'RateLimit',
    width: 250,
    render(row: RowData, index: number) {
      return h(NInput, {
        value: row.RateLimit,
        onUpdateValue(v: string) {
          // assuming that `data` is an array of FormData objects
          data.value[index].RateLimit = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const actionField = {
    title: t('admin.per_model_rate_limit.actions'),
    key: 'actions',
    render(row: any) {
      return h(
        HoverButton,
        {
          tooltip: 'Delete',
          onClick: () => deleteRow(row),
        },
        {
          default: () => {
            return h(SvgIcon, {
              class: 'text-xl',
              icon: 'material-symbols:delete',
            })
          },
        },
      )
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

async function deleteRow(row: any) {
  await DeleteUserChatModelPrivilege(row.ID)
  await refreshData()
}

async function newRowAdded() {
  await refreshData()
}
</script>

<template>
  <div class="mx-5">
    <NModal v-model:show="dialogVisible" :title="$t('admin.add_user_model_rate_limit')" preset="dialog">
      <AddChatModelForm @new-row-added="newRowAdded" />
    </NModal>
    <div class="flex justify-end">
      <HoverButton @click="dialogVisible = true">
        <span class="text-xl">
          <SvgIcon icon="material-symbols:library-add-rounded" />
        </span>
      </HoverButton>
    </div>
    <NDataTable :columns="columns" :data="data" />
  </div>
</template>
