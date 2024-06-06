<script setup lang="ts">
import { h, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NDataTable, NInput, NModal, NSwitch, useMessage, useDialog } from 'naive-ui'
import AddModelForm from './AddModelForm.vue'
import { deleteChatModel, fetchChatModel, updateChatModel } from '@/api'
import { HoverButton, SvgIcon } from '@/components/common'
import { t } from '@/locales'
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'

const ms_ui = useMessage()
const dialog = useDialog()
const dialogVisible = ref(false)
const queryClient = useQueryClient()


// const data = ref<Chat.ChatModel[]>([])
const { data, isLoading } = useQuery({
  queryKey: ['chat_models'],
  queryFn: fetchChatModel,
})

const chatModelMutation = useMutation({
  mutationFn: (variables: { id: number, data: any }) => updateChatModel(variables.id, variables.data),
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ['chat_models'] })
})

const deteteModelMutation = useMutation({
  mutationFn: (id: number) => deleteChatModel(id),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['chat_models'] })
  },
})


const UpdateRow = (row: Chat.ChatModel) => {
  if (row.id) {
    chatModelMutation.mutate({
      id: row.id,
      data: {
        ...row,
        orderNumber: parseInt(row.orderNumber || '0'),
        defaultToken: parseInt(row.defaultToken || '0'),
        maxToken: parseInt(row.maxToken || '0'),
      },
    })
  }
}


function createColumns(): DataTableColumns<Chat.ChatModel> {
  const nameField = {
    title: t('admin.chat_model.name'),
    key: 'name',
    width: 200,
    render(row: Chat.ChatModel, index: number) {
      return h(NInput, {
        value: row.name,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].name = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const labelField = {
    title: t('admin.chat_model.label'),
    key: 'label',
    width: 150,
    render(row: Chat.ChatModel, index: number) {
      return h(NInput, {
        value: row.label,
        onUpdateValue(v: string) {
          // assuming that `data` is an array of FormData objects
          data.value[index].label = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const urlField = {
    title: t('admin.chat_model.url'),
    key: 'url',
    resizable: true,
    minWidth: 200,
    render(row: Chat.ChatModel, index: number) {
      return h(NInput, {
        value: row.url,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].url = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const apiAuthKeyField = {
    title: t('admin.chat_model.apiAuthKey'),
    key: 'apiAuthKey',
    width: 150,
    render(row: Chat.ChatModel, index: number) {
      return h(NInput, {
        value: row.apiAuthKey,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].apiAuthKey = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const apiAuthHeaderField = {
    title: t('admin.chat_model.apiAuthHeader'),
    key: 'apiAuthHeader',
    width: 150,
    render(row: Chat.ChatModel, index: number) {
      return h(NInput, {
        value: row.apiAuthHeader,
        width: 50,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].apiAuthHeader = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const orderNumber = {
    title: t('admin.chat_model.orderNumber'),
    key: 'orderNumber',
    width: 100,
    render(row: Chat.ChatModel, index: number) {
      return h(NInput, {
        value: row.orderNumber?.toString(),
        width: 5,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].orderNumber = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const defaultToken = {
    title: t('admin.chat_model.defaultToken'),
    key: 'defaultToken',
    width: 100,
    render(row: Chat.ChatModel, index: number) {
      return h(NInput, {
        value: row.defaultToken?.toString(),
        width: 5,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].defaultToken = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const maxToken = {
    title: t('admin.chat_model.maxToken'),
    key: 'maxToken',
    width: 100,
    render(row: Chat.ChatModel, index: number) {
      return h(NInput, {
        value: row.maxToken?.toString(),
        width: 5,
        onUpdateValue(v: string) {
          // Assuming `data` is an array of FormData objects
          data.value[index].maxToken = v
          UpdateRow(data.value[index])
        },
      })
    },
  }

  const isDefaultField = {
    title: t('admin.chat_model.isDefault'),
    key: 'isDefault',
    render(row: Chat.ChatModel, index: number) {
      return h(NSwitch, {
        value: row.isDefault,
        onUpdateValue(v: boolean) {
          // Assuming `data` is an array of FormData objects
          const has_default = checkNoRowIsDefaultTrue(v)
          if (has_default)
            return
          data.value[index].isDefault = v
          UpdateRow(data.value[index])
        },
      })
    },
  }
  const perModelLimit = {
    title: t('admin.chat_model.EnablePerModeRatelimit'),
    key: 'enablePerModeRatelimit',
    render(row: Chat.ChatModel, index: number) {
      return h(NSwitch, {
        value: row.enablePerModeRatelimit,
        onUpdateValue(v: boolean) {
          // Assuming `data` is an array of FormData objects
          data.value[index].enablePerModeRatelimit = v
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
    defaultToken,
    maxToken,
    orderNumber,
    actionField,
  ])
}

const columns = createColumns()

async function newRowEventHandle() {
  dialogVisible.value = false
}

async function deleteRow(row: any) {
  dialog.warning({
    title: t('admin.chat_model.deleteModel'),
    content: t('admin.chat_model.deleteModelConfirm'),
    positiveText: t('common.yes'),
    negativeText: t('common.no'),
    onPositiveClick: () => {
      deteteModelMutation.mutate(row.id)
    },
  })
}

function checkNoRowIsDefaultTrue(v: boolean) {
  if (v === false)
    return
  const defaultTrueRows = data.value.filter((row: Chat.ChatModel) => row.isDefault === true)
  if (defaultTrueRows.length > 0) {
    // 'Only one row can be default, please disable existing default model first.'
    ms_ui.error(t('admin.model_one_default_only'))
    return true
  }
  return false
}
</script>

<template>
  <div class="flex items-center justify-end h-14 w-full border-b border-gray-200">
    <HoverButton @click="dialogVisible = true" class="mr-10">
      <span class="text-xl">
        <SvgIcon icon="material-symbols:library-add-rounded" />
      </span>
    </HoverButton>
  </div>
  <div class="m-5">
    <NDataTable :columns="columns" :data="data" :loading="isLoading" />
  </div>
  <NModal v-model:show="dialogVisible" :title="$t('admin.add_user_model_rate_limit')" preset="dialog">
    <AddModelForm @new-row-added="newRowEventHandle" />
  </NModal>
</template>
