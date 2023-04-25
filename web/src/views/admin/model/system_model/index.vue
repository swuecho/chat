<script setup lang="ts">
import { h, onMounted, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NDataTable, NInput, NModal, NSwitch, useMessage } from 'naive-ui'
import AddModelForm from './AddModelForm.vue'
import { deleteChatModel, fetchChatModel, updateChatModel } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { t } from '@/locales'

const ms_ui = useMessage()

interface RowData {
  ApiAuthHeader: string
  ApiAuthKey: string
  ID: number
  IsDefault: boolean
  Label: string
  Name: string
  Url: string
  EnablePerModeRatelimit: boolean
}

const data = ref<RowData[]>([])
const dialogVisible = ref(false)

onMounted(async () => {
  refreshData()
})

async function refreshData() {
  data.value = await fetchChatModel()
}

function UpdateRow(row: RowData) {
  updateChatModel(row.ID, row)
}
function createColumns(): DataTableColumns<RowData> {
  const nameField = {
    title: t('admin.chat_model.name'),
    key: 'Name',
    width: 200,
    render(row: RowData, index: number) {
      return h(NInput, {
        value: row.Name,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].Name = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const labelField = {
    title: t('admin.chat_model.label'),
    key: 'Label',
    width: 250,
    render(row: RowData, index: number) {
      return h(NInput, {
        value: row.Label,
        onUpdateValue(v: string) {
          // assuming that `data` is an array of FormData objects
          data.value[index].Label = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const urlField = {
    title: t('admin.chat_model.url'),
    key: 'Url',
    resizable: true,
    minWidth: 200,
    render(row: RowData, index: number) {
      return h(NInput, {
        value: row.Url,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].Url = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const apiAuthKeyField = {
    title: t('admin.chat_model.apiAuthKey'),
    key: 'ApiAuthKey',
    width: 200,
    render(row: RowData, index: number) {
      return h(NInput, {
        value: row.ApiAuthKey,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].ApiAuthKey = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const apiAuthHeaderField = {
    title: t('admin.chat_model.apiAuthHeader'),
    key: 'ApiAuthHeader',
    width: 200,
    render(row: RowData, index: number) {
      return h(NInput, {
        value: row.ApiAuthHeader,
        width: 50,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].ApiAuthHeader = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const isDefaultField = {
    title: t('admin.chat_model.isDefault'),
    key: 'IsDefault',
    render(row: RowData, index: number) {
      return h(NSwitch, {
        value: row.IsDefault,
        onUpdateValue(v: boolean) {
          // Assuming `data` is an array of FormData objects
          const has_default = checkNoRowIsDefaultTrue(v)
          if (has_default)
            return
          data.value[index].IsDefault = v
          UpdateRow(data.value[index])
        },
      })
    },
  }
  const perModelLimit = {
    title: t('admin.chat_model.EnablePerModeRatelimit'),
    key: 'EnablePerModeRatelimit',
    render(row: RowData, index: number) {
      return h(NSwitch, {
        value: row.EnablePerModeRatelimit,
        onUpdateValue(v: boolean) {
          // Assuming `data` is an array of FormData objects
          data.value[index].EnablePerModeRatelimit = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const actionField = {
    title: t('admin.chat_model.actions'),
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
    nameField,
    labelField,
    urlField,
    apiAuthKeyField,
    apiAuthHeaderField,
    isDefaultField,
    perModelLimit,
    actionField,
  ])
}

const columns = createColumns()

async function addRow() {
  await refreshData()
}

async function deleteRow(row: any) {
  await deleteChatModel(row.ID)
  await refreshData()
}

function checkNoRowIsDefaultTrue(v: boolean) {
  if (v === false)
    return
  const defaultTrueRows = data.value.filter((row: RowData) => row.IsDefault === true)
  if (defaultTrueRows.length > 0) {
    // 'Only one row can be default, please disable existing default model first.'
    ms_ui.error(t('admin.model_one_default_only'))
    return true
  }
  return false
}
</script>

<template>
  <div class="mx-5">
    <NModal v-model:show="dialogVisible" :title="$t('admin.add_user_model_rate_limit')" preset="dialog">
      <AddModelForm @new-row-added="addRow" />
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
