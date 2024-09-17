<template>
        <n-space vertical size="large">
                <n-card title="Create Prompt">
                        <n-form ref="formRef" :model="formValue" :rules="rules" label-placement="left"
                                label-width="auto" require-mark-placement="right-hanging" size="medium">
                                <n-form-item label="Role" path="role">
                                        <n-input v-model:value="formValue.role" placeholder="Enter role" />
                                </n-form-item>
                                <n-form-item label="Role Characteristics" path="characteristics">
                                        <n-dynamic-input v-model:value="formValue.characteristics" type="textarea"
                                                placeholder="Enter role characteristics" />
                                </n-form-item>
                                <n-form-item label="Requirements" path="requirements">
                                        <n-dynamic-input v-model:value="formValue.requirements"
                                                placeholder="Enter a requirement">
                                                <template #create-button-default>
                                                        Add Requirement
                                                </template>
                                        </n-dynamic-input>
                                </n-form-item>
                                <n-form-item label="Definitions" path="definitions">
                                        <Definitions v-model:value="formValue.definitions" />
                                </n-form-item>
                                <n-form-item label="Step by Step" path="process">
                                        <PromptProcess v-model:value="formValue.process" />
                                </n-form-item>
                        </n-form>
                        <n-space justify="end">
                                <n-button @click="handleSubmit" type="primary">Generate XML</n-button>
                                <n-button @click="copyToClipboard" type="primary">Copy XML</n-button>
                        </n-space>
                </n-card>
                <n-card v-if="xmlOutput" title="Generated XML">
                        <pre>{{ xmlOutput }}</pre>
                </n-card>
        </n-space>
</template>

<script setup>
import { ref } from 'vue'
import { NSpace, NCard, NForm, NFormItem, NInput, NDynamicInput, NButton } from 'naive-ui'
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
                trigger: 'blur'
        },
        characteristics: {
                type: 'array',
                min: 1,
                message: 'Please enter role characteristics',
                trigger: 'blur'
        },
        requirements: {
                type: 'array',
                min: 1,
                message: 'Please add at least one requirement',
                trigger: 'change'
        },
        process: {
                type: 'array',
                min: 1,
                message: 'Please add at least one process step',
                trigger: 'change'
        }
}

const handleSubmit = (e) => {
        e.preventDefault()
        formRef.value?.validate((errors) => {
                if (!errors) {
                        generateXML()
                } else {
                        console.log(errors)
                }
        })
}

const generateXML = () => {
        let xml = ''
        xml += `  <role>${formValue.value.role}</role>\n`
        xml += generateCharacteristics()
        xml += '  <requirements>\n'
        formValue.value.requirements.forEach(req => {
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
        formValue.value.characteristics.forEach(req => {
                xml += `    <characteristic>${req}</characteristics>\n`
        })
        xml += '  </characteristics>\n'
        return xml
}

const generateDefinition = () => {
        if (formValue.value.definitions.length > 0) {
                let xml = ''
                xml += '  <definitions>\n'
                formValue.value.definitions.forEach(def => {
                        xml += `    <definition>\n      <name>${def.key}</name>\n      <value>${def.value}</value>\n    </definition>\n`
                })
                xml += '  </definitions>\n'
                return xml
        } else {
                return ''
        }

}

const generateProcessXML = (steps, indent) => {
        let xml = ''
        steps.forEach(step => {
                xml += ' '.repeat(indent * 2) + `<step>\n`
                xml += ' '.repeat((indent + 1) * 2) + `<description>${step.description}</description>\n`
                if (step.children && step.children.length > 0) {
                        xml += ' '.repeat((indent + 1) * 2) + `<substeps>\n`
                        xml += generateProcessXML(step.children, indent + 2)
                        xml += ' '.repeat((indent + 1) * 2) + `</substeps>\n`
                }
                xml += ' '.repeat(indent * 2) + `</step>\n`
        })
        return xml
}

const copyToClipboard = () => {
        navigator.clipboard.writeText(xmlOutput.value)
                .then(() => {
                        console.log('XML copied to clipboard')
                })
                .catch(err => {
                        console.error('Failed to copy XML: ', err)
                })
}
</script>