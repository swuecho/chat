<template>
        <div>
                <n-dynamic-input v-model:value="definitions" :on-create="onCreateDefinition" #default="{ value }">
                        <div class="mb-2">
                                <n-input class="mx-2 mb-1" v-model:value="value.key" placeholder="Name"></n-input>
                                <n-input class="mx-2" type="textarea" v-model:value="value.value" placeholder="Definition" style="flex: 2;" />
                        </div>
                </n-dynamic-input>
        </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { NDynamicInput, NInput } from 'naive-ui'

const props = defineProps(['value'])
const emit = defineEmits(['update:value'])

const definitions = ref(props.value || [])

const onCreateDefinition = () => {
        return {
                key: '',
                value: ''
        }
}

watch(definitions, (newValue) => {
        emit('update:value', newValue)
}, { deep: true })


</script>