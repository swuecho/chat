<script lang="ts" setup>
import { computed, ref, watch, h } from 'vue'
import { NSelect, NForm } from 'naive-ui'
import { useChatStore } from '@/store'

import { fetchChatModel } from '@/api'

import { useQuery } from "@tanstack/vue-query";
import { formatDistanceToNow, differenceInDays } from 'date-fns'

const chatStore = useChatStore()

const props = defineProps<{
        uuid: string
        model: string | undefined
}>()



const chatSession = computed(() => chatStore.getChatSessionByUuid(props.uuid))

const { data } = useQuery({
        queryKey: ['chat_models'],
        queryFn: fetchChatModel,
        staleTime: 10 * 60 * 1000,
})

// format timestamp 2025-02-04T08:17:16.711644Z (string) as  to show time relative to now
const formatTimestamp = (timestamp: string) => {
        const date = new Date(timestamp)
        const days = differenceInDays(new Date(), date)
        if (days > 30) {
                return 'a month ago'
        }
        return formatDistanceToNow(date, { addSuffix: true })
}

const optionFromModel = (model: any) => {
        return {
                label: () => h('div', {}, [
                        model.label,
                        h('span', { style: 'color: #999; font-size: 0.8rem; margin-left: 4px' },
                                `- ${formatTimestamp(model.lastUsageTime)}`)
                ]),
                value: model.name,
        }
}
const chatModelOptions = computed(() =>
        data?.value ? data.value.filter((x: any) => x.isEnable).map(optionFromModel) : []
)


const defaultModel = computed(() => data?.value ? data.value.find((x: ({ isDefault: boolean, name: string })) => x.isDefault)?.name : undefined)


const modelRef = ref({
        model: chatSession.value?.model ?? defaultModel.value
})

// why watch not work?, missed the deep = true option
watch(modelRef, async (modelValue: any) => {
        await chatStore.updateChatSession(props.uuid, {
                model: modelValue.model
        })

}, { deep: true })

chatStore.$subscribe((mutation, state) => {
        const session = chatStore.getChatSessionByUuid(props.uuid)
        if (modelRef.value.model != session?.model) {
                modelRef.value.model = session?.model
        }
})


</script>

<template>
        <NForm ref="formRef" :model="modelRef">
                <NSelect v-model:value="modelRef.model" :options="chatModelOptions" size='large' />
        </NForm>
</template>