<script lang='ts' setup>
import Conversation from './components/Conversation.vue';
import { useRoute } from 'vue-router';

const route = useRoute()
const { uuid: sessionUuid } = route.params as { uuid: string }
</script>
<template>
  <div class="h-full flex">
    <Conversation :session-uuid="sessionUuid" />
  </div>
</template>