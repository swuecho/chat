<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NButton, NDataTable, NForm, NFormItem, NInput, NModal } from 'naive-ui'
import { CreateUserChatModelPrivilege, DeleteUserChatModelPrivilege, ListUserChatModelPrivilege, UpdateUserChatModelPrivilege } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { t } from '@/locales'

interface RowData {
  ID: string
  ChatModelName: string
  UserEmail: string
  RateLimit: string
}

interface FormData {
  ChatModelName: string
  UserEmail: string
  RateLimit: string

}

const dialogVisible = ref(false)

const form = ref<FormData>({
  ChatModelName: '',
  UserEmail: '',
  RateLimit: '',
})

const data = ref<RowData[]>([])

onMounted(async () => {
  refreshData()
})

async function refreshData() {
  data.value = await ListUserChatModelPrivilege()
}

function UpdateRow(row: RowData) {
  UpdateUserChatModelPrivilege(row.ID, row)
}

function createColumns(): DataTableColumns<RowData> {
  const userEmailField = {
    title: t('admin.per_model_rate_limit.UserEmail'),
    key: 'UserEmail',
    width: 200,
    render(row: RowData, index: number) {
      return h(NInput, {
        value: row.UserEmail,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].UserEmail = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const modelField = {
    title: t('admin.per_model_rate_limit.ChatModelName'),
    key: 'ChatModelName',
    width: 250,
    render(row: RowData, index: number) {
      return h(NInput, {
        value: row.ChatModelName,
        onUpdateValue(v: string) {
          // assuming that `data` is an array of FormData objects
          data.value[index].ChatModelName = v
          UpdateRow(data.value[index])
        },
      })
    },
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
    userEmailField,
    modelField,
    ratelimitField,
    actionField,
  ])
}

const columns = createColumns()

async function addRow(form: FormData) {
  console.log(form)
  // create a new chat model, the name is randon string
  const chatModel = await CreateUserChatModelPrivilege({
    ID: 0,
    UserEmail: form.UserEmail,
    ChatModelName: form.ChatModelName,
    RateLimit: parseInt(form.RateLimit, 10),
  })
  // add it to the data array
  data.value.push(chatModel)
}

async function deleteRow(row: any) {
  await DeleteUserChatModelPrivilege(row.ID)
  await refreshData()
}
</script>

<template>
  <div class="mx-5">
    <NModal v-model:show="dialogVisible" title="Submit Email" preset="dialog">
      <NForm :model="form">
        <NFormItem prop="UserEmail" label="Email">
          <NInput v-model:value="form.UserEmail" placeholder="Please email" />
        </NFormItem>
        <NFormItem prop="ChatModelName" label="ChatModelName">
          <NInput v-model:value="form.ChatModelName" placeholder="Please model name" />
        </NFormItem>
        <NFormItem prop="RateLimit" label="Ratelimit">
          <NInput v-model:value="form.RateLimit" placeholder="Please input rate" />
        </NFormItem>
      </NForm>
      <NButton type="primary" block secondary strong @click="addRow(form)">
        确认
      </NButton>
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
