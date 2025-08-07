<script setup lang="ts">
interface Props {
  questions: string[]
  loading?: boolean
}

interface Emit {
  (ev: 'useQuestion', question: string): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emit>()

function handleQuestionClick(question: string) {
  emit('useQuestion', question)
}
</script>

<template>
  <div class="suggested-questions mt-3 p-3 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
    <div class="flex items-center mb-2">
      <svg class="w-4 h-4 mr-2 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
      </svg>
      <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ $t('chat.suggestedQuestions') }}</span>
      <!-- Loading spinner inline with title -->
      <div v-if="loading" class="ml-2">
        <div class="animate-spin rounded-full h-3 w-3 border-b-2 border-blue-500"></div>
      </div>
    </div>
    
    <!-- Actual questions -->
    <div v-if="!loading && questions.length > 0" class="space-y-2">
      <button
        v-for="(question, index) in questions"
        :key="index"
        class="w-full text-left p-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 hover:bg-blue-50 dark:hover:bg-gray-600 transition-colors duration-200 text-sm text-gray-800 dark:text-gray-200"
        @click="handleQuestionClick(question)"
      >
        {{ question }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.suggested-questions {
  /* Ensure proper responsive behavior */
  max-width: 100%;
}

.suggested-questions button:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

/* Dark mode hover effect */
.dark .suggested-questions button:hover {
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}
</style>