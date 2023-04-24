<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NDataTable, NInput, NSwitch, useMessage } from 'naive-ui'
import { createChatModel, deleteChatModel, fetchChatModel, updateChatModel } from '@/api'
import { generateRandomString } from '@/utils/rand'
import { HoverButton, SvgIcon } from '@/components/common'
import { t } from '@/locales'


interface RowData {
  Model: string
  UserEmail: string
  Ratelimit: string
}

const data = ref<RowData[]>([])

onMounted(async () => {
  refreshData()
})

async function refreshData() {
  data.value = await fetchPerModelRateLimit()
}

function UpdateRow(row: RowData) {
  updatePerModelRatelimit(row.UserEmail, row)
}

function createColumns(): DataTableColumns<RowData> {
  const userEmailField = {
    title: t('admin.per_model_ratelimit.UserEmail'),
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
    title: t('admin.per_model_ratelimit.Model'),
    key: 'Model',
    width: 250,
    render(row: RowData, index: number) {
      return h(NInput, {
        value: row.Model,
        onUpdateValue(v: string) {
          // assuming that `data` is an array of FormData objects
          data.value[index].Model = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const ratelimitField = {
    title: t('admin.per_model_ratelimit.Ratelimit'),
    key: 'Ratelimit',
    width: 250,
    render(row: RowData, index: number) {
      return h(NInput, {
        value: row.Ratelimit,
        onUpdateValue(v: string) {
          // assuming that `data` is an array of FormData objects
          data.value[index].Ratelimit = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const actionField = {
    title: t('admin.per_model_ratelimit.actions'),
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

async function addRow() {
  // create a new chat model, the name is randon string
  const randModelName = generateRandomString(10)
  const chatModel = await createPerModelRatelimit({
    UserName: randModelName,
    Model: '',
    Ratelimit: 0
  })
  // add it to the data array
  data.value.push(chatModel)
}

async function deleteRow(row: any) {
  await deletePerModelRatelimit(row.ID)
  await refreshData()
}
</script>

<template>
  <div class="ml-5">
    <div class="flex justify-end">
      <HoverButton @click="addRow">
        <span class="text-xl">
          <SvgIcon icon="material-symbols:library-add-rounded" />
        </span>
      </HoverButton>
    </div>
    <NDataTable :columns="columns" :data="data" />
  </div>
</template>
