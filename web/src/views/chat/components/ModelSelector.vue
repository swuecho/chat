<script lang="ts" setup>
import { computed, ref, watch, } from 'vue'
import { NSelect, NForm } from 'naive-ui'
import { useChatStore } from '@/store'

import { fetchChatModel } from '@/api'

import { useQuery } from "@tanstack/vue-query";

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

const optionFromModel = (model: any) => {
        return {
                label: model.label,
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
        console.log(modelValue)
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