<script setup>
import { ref } from 'vue'
import { NButton, NCard, NDynamicInput, NForm, NFormItem, NInput, NSpace } from 'naive-ui'
import PromptProcess from './PromptProcess.vue'
import Definitions from './Definitions.vue'

const formRef = ref(null)
const xmlOutput = ref('')

const formValue = ref({
  role: '',
  characteristics: [],
  requirements: [],
  process: [],
  definitions: [],
})

const rules = {
  role: {
    required: true,
    message: 'Please enter a role',
    trigger: 'blur',
  },
  characteristics: {
    type: 'array',
    min: 1,
    message: 'Please enter role characteristics',
    trigger: 'blur',
  },
  requirements: {
    type: 'array',
    min: 1,
    message: 'Please add at least one requirement',
    trigger: 'change',
  },
  process: {
    type: 'array',
    min: 1,
    message: 'Please add at least one process step',
    trigger: 'change',
  },
}

const handleSubmit = (e) => {
  e.preventDefault()
  formRef.value?.validate((errors) => {
    if (!errors)
      generateXML()
    else
      console.log(errors)
  })
}

const generateXML = () => {
  let xml = ''
  xml += `  <role>${formValue.value.role}</role>\n`
  xml += generateCharacteristics()
  xml += '  <requirements>\n'
  formValue.value.requirements.forEach((req) => {
    xml += `    <requirement>${req}</requirement>\n`
  })
  xml += '  </requirements>\n'
  xml += generateDefinition()
  xml += '  <process>\n'
  xml += generateProcessXML(formValue.value.process, 2)
  xml += '  </process>\n'
  xmlOutput.value = xml
}

const generateCharacteristics = () => {
  let xml = ''
  xml += '  <characteristics>\n'
  formValue.value.characteristics.forEach((req) => {
    xml += `    <characteristic>${req}</characteristics>\n`
  })
  xml += '  </characteristics>\n'
  return xml
}

const generateDefinition = () => {
  if (formValue.value.definitions.length > 0) {
    let xml = ''
    xml += '  <definitions>\n'
    formValue.value.definitions.forEach((def) => {
      xml += `    <definition>\n      <name>${def.key}</name>\n      <value>${def.value}</value>\n    </definition>\n`
    })
    xml += '  </definitions>\n'
    return xml
  }
  else {
    return ''
  }
}

const generateProcessXML = (steps, indent) => {
  let xml = ''
  steps.forEach((step) => {
    xml += `${' '.repeat(indent * 2)}<step>\n`
    xml += `${' '.repeat((indent + 1) * 2)}<description>${step.description}</description>\n`
    if (step.children && step.children.length > 0) {
      xml += `${' '.repeat((indent + 1) * 2)}<substeps>\n`
      xml += generateProcessXML(step.children, indent + 2)
      xml += `${' '.repeat((indent + 1) * 2)}</substeps>\n`
    }
    xml += `${' '.repeat(indent * 2)}</step>\n`
  })
  return xml
}

const copyToClipboard = () => {
  navigator.clipboard.writeText(xmlOutput.value)
    .then(() => {
      console.log('XML copied to clipboard')
    })
    .catch((err) => {
      console.error('Failed to copy XML: ', err)
    })
}
</script>

<template>
  <NSpace vertical size="large">
    <NCard title="Create Prompt">
      <NForm
        ref="formRef" :model="formValue" :rules="rules" label-placement="left"
        label-width="auto" require-mark-placement="right-hanging" size="medium"
      >
        <NFormItem label="Role" path="role">
          <NInput v-model:value="formValue.role" placeholder="Enter role" />
        </NFormItem>
        <NFormItem label="Role Characteristics" path="characteristics">
          <NDynamicInput
            v-model:value="formValue.characteristics" type="textarea"
            placeholder="Enter role characteristics"
          />
        </NFormItem>
        <NFormItem label="Requirements" path="requirements">
          <NDynamicInput
            v-model:value="formValue.requirements"
            placeholder="Enter a requirement"
          >
            <template #create-button-default>
              Add Requirement
            </template>
          </NDynamicInput>
        </NFormItem>
        <NFormItem label="Definitions" path="definitions">
          <Definitions v-model:value="formValue.definitions" />
        </NFormItem>
        <NFormItem label="Step by Step" path="process">
          <PromptProcess v-model:value="formValue.process" />
        </NFormItem>
      </NForm>
      <NSpace justify="end">
        <NButton type="primary" @click="handleSubmit">
          Generate XML
        </NButton>
        <NButton type="primary" @click="copyToClipboard">
          Copy XML
        </NButton>
      </NSpace>
    </NCard>
    <NCard v-if="xmlOutput" title="Generated XML">
      <pre>{{ xmlOutput }}</pre>
    </NCard>
  </NSpace>
</template>
