<template>
  <div v-if="isExecutable" class="library-management">
    <div class="library-header">
      <span class="library-title">
        <Icon icon="ri:package-line" />
        Libraries
      </span>
      <NButton size="tiny" @click="$emit('toggle-visibility')" type="tertiary">
        <Icon icon="ri:information-line" />
        {{ isVisible ? 'Hide' : 'Show' }} Available
      </NButton>
    </div>
    
    <div v-if="isVisible" class="library-list">
      <div v-if="isPython" class="library-info">
        <div class="library-packages">
          <strong>Available Python packages:</strong> 
          {{ pythonPackages.join(', ') }}
        </div>
        <div class="library-usage">
          Use <code>import packageName</code> or <code>from packageName import ...</code> in your code
        </div>
      </div>
      
      <div v-else class="library-info">
        <div class="library-packages">
          <strong>Available JavaScript libraries:</strong> 
          {{ jsLibraries.join(', ') }}
        </div>
        <div class="library-usage">
          Use <code>// @import libraryName</code> in your code to auto-load libraries
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { NButton } from 'naive-ui'
import { Icon } from '@iconify/vue'
/// <reference path="@/typings/chat.d.ts" />
type Artifact = Chat.Artifact

interface Props {
  artifact: Artifact
  isExecutable: boolean
  isVisible: boolean
}

const props = defineProps<Props>()

defineEmits<{
  'toggle-visibility': []
}>()

const isPython = computed(() => 
  props.artifact.language === 'python' || props.artifact.language === 'py'
)

const pythonPackages = [
  'numpy', 'pandas', 'matplotlib', 'scipy', 'scikit-learn', 
  'requests', 'beautifulsoup4', 'pillow', 'sympy', 'networkx', 
  'seaborn', 'plotly', 'bokeh', 'altair'
]

const jsLibraries = [
  'lodash', 'd3', 'chart.js', 'moment', 'axios', 'rxjs', 
  'p5', 'three', 'fabric'
]
</script>

<style scoped>
.library-management {
  margin-bottom: 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  overflow: hidden;
}

.library-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0.75rem;
  background: #f1f5f9;
  border-bottom: 1px solid #e2e8f0;
}

.library-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: #475569;
}

.library-list {
  padding: 0.75rem;
}

.library-info {
  font-size: 0.875rem;
  line-height: 1.5;
}

.library-packages {
  margin-bottom: 0.5rem;
  color: #374151;
}

.library-usage {
  color: #6b7280;
  font-style: italic;
}

.library-usage code {
  background: #f3f4f6;
  padding: 0.125rem 0.25rem;
  border-radius: 0.25rem;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.8125rem;
}

:deep(.dark) .library-management {
  border-color: #374151;
}

:deep(.dark) .library-header {
  background: #374151;
  border-bottom-color: #4b5563;
}

:deep(.dark) .library-title {
  color: #d1d5db;
}

:deep(.dark) .library-packages {
  color: #f3f4f6;
}

:deep(.dark) .library-usage {
  color: #9ca3af;
}

:deep(.dark) .library-usage code {
  background: #4b5563;
  color: #f3f4f6;
}
</style>