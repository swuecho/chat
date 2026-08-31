<script lang="ts"  setup>
import { ref } from 'vue'
import { useErrorHandling } from '../../composables/useErrorHandling'
import { HoverButton, SvgIcon } from '@/components/common'
import { textToSpeech } from '@/api/generated_client'

interface Props {
  text: string
}

const props = defineProps<Props>()

const source = ref('')
const soundPlayer = ref()
const isActive = ref(false)
const { handleApiError } = useErrorHandling()
// const speaker_id = ref('')
// const style_wav = ref('')
// const language_id = ref('')

// Add a method called 'playAudio' to handle sending the request to the backend.
async function playAudio() {
  if (isActive.value) {
    isActive.value = false
  }
  else {
    try {
      const blob = await textToSpeech({ query: { text: props.text } })
      source.value = URL.createObjectURL(blob)
      isActive.value = true
    }
    catch (error) {
      handleApiError(error, 'audio-playback')
    }
  }
}
</script>

<template>
  <div>
    <HoverButton :tooltip="$t('chat.playAudio')" @click="playAudio">
      <span class=" text-[#4f555e] dark:text-white">
        <SvgIcon icon="wpf:audio-wave" />
      </span>
    </HoverButton>
    <audio v-if="isActive" id="audio" ref="soundPlayer" autoplay :src="source" controls />
  </div>
</template>
