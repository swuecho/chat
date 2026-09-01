<script setup>
import { ref, watch } from 'vue'
import { NDynamicInput, NInput } from 'naive-ui'

const props = defineProps(['value'])
const emit = defineEmits(['update:value'])

const definitions = ref(props.value || [])

const onCreateDefinition = () => {
  return {
    key: '',
    value: '',
  }
}

watch(definitions, (newValue) => {
  emit('update:value', newValue)
}, { deep: true })
</script>

<template>
  <div>
    <NDynamicInput v-slot="{ value: definition }" v-model:value="definitions" :on-create="onCreateDefinition">
      <div class="mb-2">
        <NInput v-model:value="definition.key" class="mx-2 mb-1" placeholder="Name" />
        <NInput v-model:value="definition.value" class="mx-2" type="textarea" placeholder="Definition" style="flex: 2;" />
      </div>
    </NDynamicInput>
  </div>
</template>
