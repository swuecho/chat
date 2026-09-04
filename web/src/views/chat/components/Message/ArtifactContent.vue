<script lang="ts" setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import MarkdownIt from 'markdown-it'
import mermaid from 'mermaid'
import { type Artifact } from '@/typings/chat'
import { sanitizeHtml, sanitizeSvg } from '@/utils/sanitize'

interface Props {
  artifact: Artifact
}

const props = defineProps<Props>()

const mdi = new MarkdownIt()

const renderedMarkdown = computed(() => sanitizeHtml(mdi.render(props.artifact.content)))
const sanitizedHtml = computed(() => sanitizeHtml(props.artifact.content))
const sanitizedSvg = computed(() => sanitizeSvg(props.artifact.content))
const mermaidContent = ref<HTMLElement>()
const mermaidError = ref('')
const mermaidRenderId = `artifact-mermaid-${props.artifact.uuid.replace(/[^a-zA-Z0-9_-]/g, '')}-${Math.random().toString(36).slice(2)}`

mermaid.initialize({ startOnLoad: false, securityLevel: 'strict' })

const renderMermaid = async () => {
  if (props.artifact.type !== 'mermaid' || !mermaidContent.value)
    return
  try {
    mermaidError.value = ''
    const { svg } = await mermaid.render(mermaidRenderId, props.artifact.content)
    mermaidContent.value.innerHTML = sanitizeSvg(svg)
  }
  catch (error) {
    mermaidContent.value.innerHTML = ''
    mermaidError.value = error instanceof Error ? error.message : 'Unable to render diagram'
  }
}

onMounted(renderMermaid)
watch(() => props.artifact.content, async () => {
  await nextTick()
  await renderMermaid()
})

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
      <div class="code-display">
        <pre><code :class="`language-${artifact.language || 'text'}`">{{ artifact.content }}</code></pre>
      </div>
    </div>

    <div v-else-if="artifact.type === 'html'" class="html-artifact">
      <iframe :srcdoc="sanitizedHtml" class="html-iframe" sandbox title="HTML artifact preview" />
    </div>

    <div v-else-if="artifact.type === 'svg'" class="svg-artifact">
      <div class="svg-content" v-html="sanitizedSvg" />
    </div>

    <div v-else-if="artifact.type === 'mermaid'" class="mermaid-artifact">
      <div ref="mermaidContent" class="mermaid-content" />
      <div v-if="mermaidError" class="mermaid-error">
        <p>{{ mermaidError }}</p>
        <pre>{{ artifact.content }}</pre>
      </div>
    </div>

    <div v-else-if="artifact.type === 'json'" class="json-artifact">
      <pre><code class="language-json">{{ formatJson(artifact.content) }}</code></pre>
    </div>

    <div v-else-if="artifact.type === 'markdown'" class="markdown-artifact">
      <div class="markdown-content" v-html="renderedMarkdown" />
    </div>

    <div v-else class="unsupported-artifact">
      <p>This artifact type is not supported yet.</p>
      <pre>{{ artifact.content }}</pre>
    </div>
  </div>
</template>

<style scoped>
.artifact-content {
  padding: 1rem;
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
