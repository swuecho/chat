<template>
        <main class="h-full flex">
                        <n-card title="Create XML Prompt">
                                <n-form>
                                        <n-form-item label="Role">
                                                <n-input v-model:value="role" placeholder="Enter the role" />
                                        </n-form-item>
                                        <n-form-item label="Role Characteristics">
                                                <n-input v-model:value="roleCharacteristics"
                                                        placeholder="Enter role characteristics" />
                                        </n-form-item>
                                        <n-form-item label="Specific Requirements">
                                                <n-space vertical>
                                                        <n-input v-for="(req, index) in requirements" :key="index"
                                                                v-model:value="requirements[index]"
                                                                placeholder="Enter a requirement" />
                                                        <n-button @click="addRequirement">Add Requirement</n-button>
                                                </n-space>
                                        </n-form-item>
                                        <n-form-item label="Process (Steps)">
                                                <n-space vertical>
                                                        <n-input v-for="(step, index) in steps" :key="index"
                                                                v-model:value="steps[index]"
                                                                placeholder="Enter a step" />
                                                        <n-button @click="addStep">Add Step</n-button>
                                                </n-space>
                                        </n-form-item>
                                        <n-form-item>
                                                <n-button type="primary" @click="generateXML">Generate XML</n-button>
                                        </n-form-item>
                                </n-form>
                        </n-card>
                        <n-card v-if="xmlOutput" title="Generated XML">
                                <pre>{{ xmlOutput }}</pre>
                        </n-card>
        </main>
</template>

<script setup>
import { ref } from 'vue';
import { NConfigProvider, NLayout, NLayoutContent, NCard, NForm, NFormItem, NInput, NButton, NSpace } from 'naive-ui';

const role = ref('');
const roleCharacteristics = ref('');
const requirements = ref(['']);
const steps = ref(['']);
const xmlOutput = ref('');

const addRequirement = () => {
        requirements.value.push('');
};

const addStep = () => {
        steps.value.push('');
};

const generateXML = () => {
        const xml = `
      <prompt>
        <role>${role.value}</role>
        <roleCharacteristics>${roleCharacteristics.value}</roleCharacteristics>
        <requirements>
          ${requirements.value.map(req => `<requirement>${req}</requirement>`).join('\n    ')}
        </requirements>
        <process>
          ${steps.value.map(step => `<step>${step}</step>`).join('\n    ')}
        </process>
      </prompt>
        `.trim();

        xmlOutput.value = xml;
};
</script>

<style scoped>
.n-card {
        margin-bottom: 20px;
}
</style>