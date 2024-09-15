<template>
        <div class="flex flex-row w-full">
                <div class="flex-1 m-5">
                        <n-input v-model:value="svgContent" type="textarea" :autosize="{ minRows: 20, maxRows: 60 }"
                                placeholder="Paste your SVG code here" class="w-full" />
                </div>
                <div class="flex-1 m-5">
                        <div>
                        <n-button @click="clearSvgContent">Clear</n-button> <!-- {{ edit_3 }} -->
                        <n-button @click="saveToPng">Save to PNG</n-button> <!-- {{ edit_1 }} -->
</div>
                        <div v-html="sanitizedSvgContent" class="svg-container w-full h-full">
                        </div>
                </div>
        </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue' // {{ edit_2 }}
import { NInput, NDivider, NButton } from 'naive-ui' // {{ edit_2 }}
import DOMPurify from 'dompurify'

const svgContent = ref(localStorage.getItem('svgContent') || '') // {{ edit_1 }}

watch(svgContent, (newValue) => { // {{ edit_2 }}
    localStorage.setItem('svgContent', newValue);
});

const sanitizedSvgContent = computed(() => {
        return DOMPurify.sanitize(svgContent.value)
})

const saveToPng = () => { // {{ edit_4 }}
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

const clearSvgContent = () => { // {{ edit_3 }}
        svgContent.value = '';
}

</script>

<style scoped>
.svg-container {
        max-width: 50%;
        overflow: auto;
}
</style>