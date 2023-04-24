<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import type { DataTableColumns } from 'naive-ui'
import { NButton, NDataTable, NForm, NFormItem, NInput, NModal, NSelect } from 'naive-ui'
import { CreateUserChatModelPrivilege, DeleteUserChatModelPrivilege, ListUserChatModelPrivilege, UpdateUserChatModelPrivilege, fetchChatModel } from '@/api'
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

interface Option {
  label: string
  value: string
}

const dialogVisible = ref(false)

const form = ref<FormData>({
  ChatModelName: '',
  UserEmail: '',
  RateLimit: '',
})

const data = ref<RowData[]>([])
const limitEnabledModels = ref<Option[]>([])

onMounted(async () => {
  refreshData()
  limitEnabledModels.value = (await fetchChatModel()).filter((x: any) => x.EnablePerModeRatelimit)
    .map((x: any) => {
      return {
        value: x.Name,
        label: x.Label,
      }
    })
})

const defaultModel = computed(
  () => limitEnabledModels.value[0].value,
)

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
  }

  const modelField = {
    title: t('admin.per_model_rate_limit.ChatModelName'),
    key: 'ChatModelName',
    width: 250,
    // render(row: RowData, index: number) {
    //   return h(NSelect, {
    //     options: limitEnabledModels.value,
    //     value: row.ChatModelName,
    //     onUpdateValue(v: string) {
    //       // assuming that `data` is an array of FormData objects
    //       data.value[index].ChatModelName = v
    //       UpdateRow(data.value[index])
    //     },
    //   })
    // },
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
    <NModal v-model:show="dialogVisible" :title="$t('admin.add_user_model_rate_limit')" preset="dialog">
      <NForm :model="form">
        <NFormItem prop="UserEmail" :label="$t('common.email')">
          <NInput v-model:value="form.UserEmail" :placeholder="$t('common.email_placeholder')" />
        </NFormItem>
        <NFormItem prop="ChatModelName" :label="$t('admin.chat_model_name')">
          <NSelect
            v-model:value="form.ChatModelName" :options="limitEnabledModels" :default-value="defaultModel"
            placeholder="Please model name"
          />
        </NFormItem>
        <NFormItem prop="RateLimit" :label="$t('admin.rate_limit')">
          <NInput v-model:value="form.RateLimit" />
        </NFormItem>
      </NForm>
      <NButton type="primary" block secondary strong @click="addRow(form)">
        {{ $t('common.confirm') }}
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
