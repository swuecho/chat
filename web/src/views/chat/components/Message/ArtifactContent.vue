<script lang="ts" setup>
import { computed } from 'vue'
import { NButton } from 'naive-ui'
import { Icon } from '@iconify/vue'
import MarkdownIt from 'markdown-it'
import { type Artifact } from '@/typings/chat'
import { sanitizeHtml, sanitizeSvg } from '@/utils/sanitize'

interface Props {
  artifact: Artifact
  isEditing: boolean
  editableContent: string
}

const props = defineProps<Props>()

defineEmits<{
  'toggle-edit': [uuid: string, content: string]
  'save-edit': [uuid: string]
  'cancel-edit': [uuid: string]
  'update-editable-content': [uuid: string, content: string]
}>()

const mdi = new MarkdownIt()

const renderedMarkdown = computed(() => sanitizeHtml(mdi.render(props.artifact.content)))
const sanitizedSvg = computed(() => sanitizeSvg(props.artifact.content))

const formatJson = (jsonString: string) => {
  try {
    return JSON.stringify(JSON.parse(jsonString), null, 2)
  }
  catch {
    return jsonString
  }
}
</script>

<template>
  <div class="artifact-content">
    <div v-if="artifact.type === 'code'" class="code-artifact">
      <div v-if="isEditing" class="code-editor">
        <textarea
          :value="editableContent"
          class="code-textarea"
          :style="{ height: `${Math.max(200, editableContent.split('\n').length * 20)}px` }"
          @input="$emit('update-editable-content', artifact.uuid, $event.target.value)"
        />
        <div class="editor-actions">
          <NButton size="small" type="primary" @click="$emit('save-edit', artifact.uuid)">
            Save
          </NButton>
          <NButton size="small" @click="$emit('cancel-edit', artifact.uuid)">
            Cancel
          </NButton>
        </div>
      </div>
      <div v-else class="code-display">
        <pre><code :class="`language-${artifact.language || 'text'}`">{{ artifact.content }}</code></pre>
        <div class="code-actions">
          <NButton size="small" @click="$emit('toggle-edit', artifact.uuid, artifact.content)">
            <Icon icon="ri:edit-line" />
            Edit
          </NButton>
        </div>
      </div>
    </div>

    <div v-else-if="artifact.type === 'html'" class="html-artifact">
      <iframe :srcdoc="artifact.content" class="html-iframe" sandbox="allow-scripts" />
    </div>

    <div v-else-if="artifact.type === 'svg'" class="svg-artifact">
      <div class="svg-content" v-html="sanitizedSvg" />
    </div>

    <div v-else-if="artifact.type === 'mermaid'" class="mermaid-artifact">
      <div class="mermaid-content">
        {{ artifact.content }}
      </div>
    </div>

    <div v-else-if="artifact.type === 'json'" class="json-artifact">
      <pre><code class="language-json">{{ formatJson(artifact.content) }}</code></pre>
    </div>

    <div v-else-if="artifact.type === 'markdown'" class="markdown-artifact">
      <div class="markdown-content" v-html="renderedMarkdown" />
    </div>
  </div>
</template>

<style scoped>
.artifact-content {
  padding: 1rem;
}

.code-textarea {
  width: 100%;
  font-family: monospace;
  border: 1px solid #d1d5db;
  border-radius: 0.5rem;
  padding: 0.75rem;
  resize: vertical;
}

.editor-actions,
.code-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

.html-iframe {
  width: 100%;
  min-height: 320px;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
}

.svg-content,
.mermaid-content,
.markdown-content,
.json-artifact pre,
.code-display pre {
  overflow: auto;
}
</style>
