<script lang="ts"  setup>
import { ref } from "vue";
import { HoverButton, SvgIcon } from '@/components/common'
import request from '@/utils/request/axios'

interface Props {
        text: string
}

const props = defineProps<Props>()

const source = ref('')
const soundPlayer = ref();
const isActive = ref(false);
// const speaker_id = ref('')
// const style_wav = ref('')
// const language_id = ref('')





// Add a method called 'playAudio' to handle sending the request to the backend.
async function playAudio() {
        console.log(props.text)
        if (isActive.value) {
                isActive.value = false
        } else {
                let text = encodeURIComponent(props.text)
                try {
                        // Perform the HTTP request to send the request to the backend.
                        const response = await request.get(`/tts?text=${text}`, { responseType: 'blob' });
                        console.log(response)
                        if (response.status == 200) {
                                // If the HTTP response is successful, parse the body into an object and play the sound.
                                const blob = await response.data;
                                source.value = URL.createObjectURL(blob);
                                console.log(source.value);
                                isActive.value = true;
                        } else {
                                console.log("request failed")
                        }
                } catch (error) {
                        console.log(error);
                }
        }
}


</script>


<template>
        <HoverButton :tooltip="$t('chat.playAudio')" @click="playAudio">
                <span class="text-xl text-[#4f555e] dark:text-white">
                        <SvgIcon icon="system-uicons:audio-wave" />
                </span>
        </HoverButton>
        <audio ref="soundPlayer" id="audio" autoplay :src="source" v-if="isActive" controls></audio>
</template>
      