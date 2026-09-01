<script lang="ts" setup>
import ArtifactViewerBase from './ArtifactViewerBase.vue'
import { type Artifact } from '@/utils/artifacts'

interface Props {
  artifacts: Artifact[]
  inversion?: boolean
}

defineProps<Props>()
</script>

<template>
  <ArtifactViewerBase
    :artifacts="artifacts"
  />
</template>

<style scoped>
/* Import styles from the base component if needed */
</style>
