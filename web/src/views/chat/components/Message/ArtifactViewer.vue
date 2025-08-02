<template>
  <ArtifactViewerBase 
    :artifacts="artifacts"
  />
</template>

<script lang="ts" setup>
import { type Artifact } from '@/typings/chat'
import ArtifactViewerBase from './ArtifactViewerBase.vue'

interface Props {
  artifacts: Artifact[]
  inversion?: boolean
}

defineProps<Props>()
</script>

<style scoped>
/* Import styles from the base component if needed */
</style>