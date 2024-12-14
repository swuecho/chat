<template>
        <div class="flex flex-row w-full">
                <div class="flex-1 m-5">
                        <div class="flex justify-between mb-4">
                                <n-button type="primary" class="rounded-lg">Raw</n-button>
                                <n-button @click="clearSvgContent" type="error" class="rounded-lg">Clear</n-button>
                        </div>
                        <n-input v-model:value="svgContent" type="textarea" :autosize="{ minRows: 20, maxRows: 60 }"
                                placeholder="Paste your SVG code here" class="w-full border rounded-lg shadow-md" />
                </div>
                <div class="flex-1 m-5">
                        <div class="flex justify-between mb-4">
                                <n-button @click="saveToPng" type="success" class="rounded-lg">Save</n-button>
                        </div>
                        <div v-html="sanitizedSvgContent" class="w-full h-full">
                        </div>
                </div>
        </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { NInput, NDivider, NButton } from 'naive-ui'
import DOMPurify from 'dompurify'

const svgContent = ref(localStorage.getItem('svgContent') || '')

watch(svgContent, (newValue) => {
        localStorage.setItem('svgContent', newValue);
});

const sanitizedSvgContent = computed(() => {
        return DOMPurify.sanitize(svgContent.value)
})

const saveToPng = () => {
        const svgElement = new Blob([sanitizedSvgContent.value], { type: 'image/svg+xml;charset=utf-8' });
        const url = URL.createObjectURL(svgElement);
        const img = new Image();
        img.onload = () => {
                const canvas = document.createElement('canvas');
                canvas.width = img.width;
                canvas.height = img.height;
                const ctx = canvas.getContext('2d');
                ctx.drawImage(img, 0, 0);
                const pngUrl = canvas.toDataURL('image/png');
                const a = document.createElement('a');
                a.href = pngUrl;
                a.download = 'image.png';
                a.click();
                URL.revokeObjectURL(url);
        };
        img.src = url;
}

const clearSvgContent = () => {
        svgContent.value = '';
}

</script>

<style scoped>
.svg-container {
        overflow: auto;
        /* Light background for contrast */
}
</style>