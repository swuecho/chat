<script lang="ts" setup>
import { computed, ref, h } from 'vue'
import { DataTableColumns, NButton, NDataTable } from 'naive-ui'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { usePromptStore } from '@/store/modules'


interface Emit {
        (ev: 'usePrompt', key: string, prompt: string): void
}

const emit = defineEmits<Emit>()



// 移动端自适应相关
const { isMobile } = useBasicLayout()

const promptStore = usePromptStore()
const promptList = ref<any>(promptStore.promptList)

interface DataProps {
        renderKey: string
        renderValue: string
        key: string
        value: string
}
const isASCII = (str: string) => /^[\x00-\x7F]*$/.test(str)

// 移动端自适应相关
const renderTemplate = () => {
  const [keyLimit, valueLimit] = isMobile.value ? [6, 9] : [15, 50]
  return promptList.value.map((item: { key: string; value: string }) => {
    let factor = isASCII(item.key) ? 10 : 1
    return {
      renderKey: item.key.length <= keyLimit ? item.key : `${item.key.substring(0, keyLimit * factor)}...`,
      renderValue: item.value.length <= valueLimit ? item.value : `${item.value.substring(0, valueLimit * factor)}...`,
      key: item.key,
      value: item.value,
    }
  })
}

const actionUsePrompt = (type: string, row: any) => {
        console.log(type, row)
        console.log(row.key, row.value)
        emit('usePrompt', row.key, row.value)
}


const pagination = computed(() => {
        const [pageSize, pageSlot] = isMobile.value ? [10, 5] : [20, 15]
        return {
                pageSize, pageSlot,
        }
})

const maxHeight = computed(() => {
        return isMobile.value ? 400 : 600
})

// table相关
const createColumns = (): DataTableColumns<DataProps> => {
        return [
                {
                        title: '提示词标题',
                        key: 'renderKey',
                        minWidth: 100,
                },
                {
                        title: '内容',
                        key: 'renderValue',
                },

                {
                        title: '操作',
                        key: 'actions',
                        width: 100,
                        align: 'center',
                        render(row) {
                                return h('div', { class: 'flex items-center flex-col gap-2' }, {
                                        default: () => [h(
                                                NButton,
                                                {
                                                        tertiary: true,
                                                        size: 'small',
                                                        type: 'info',
                                                        onClick: () => actionUsePrompt('modify', row),
                                                },
                                                { default: () => '使用' },
                                        ),
                                        ],
                                })
                        },
                },
        ]
}
const columns = createColumns()
</script>
<template>
        <NDataTable class="mt-10" :max-height="maxHeight" :columns="columns" :data="renderTemplate()" :pagination="pagination"
                :bordered="false" />
</template>