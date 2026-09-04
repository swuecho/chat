<script lang="ts" setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import MarkdownIt from 'markdown-it'
import mermaid from 'mermaid'
import hljs from 'highlight.js'
import { type Artifact } from '@/typings/chat'
import { sanitizeHtml, sanitizeHtmlDocument, sanitizeSvg } from '@/utils/sanitize'
import { t } from '@/locales'

interface Props {
  artifact: Artifact
}

const props = defineProps<Props>()

const mdi = new MarkdownIt()

const renderedMarkdown = computed(() => sanitizeHtml(mdi.render(props.artifact.content)))
const sanitizedHtml = computed(() => sanitizeHtmlDocument(props.artifact.content))
const sanitizedSvg = computed(() => sanitizeSvg(props.artifact.content))
const mermaidContent = ref<HTMLElement>()
const mermaidError = ref('')
const mermaidLoading = ref(false)
const mermaidRenderId = `artifact-mermaid-${props.artifact.uuid.replace(/[^a-zA-Z0-9_-]/g, '')}-${Math.random().toString(36).slice(2)}`

mermaid.initialize({ startOnLoad: false, securityLevel: 'strict' })

const renderMermaid = async () => {
  if (props.artifact.type !== 'mermaid' || !mermaidContent.value)
    return
  try {
    mermaidLoading.value = true
    mermaidError.value = ''
    const { svg } = await mermaid.render(mermaidRenderId, props.artifact.content)
    mermaidContent.value.innerHTML = sanitizeSvg(svg)
  }
  catch (error) {
    mermaidContent.value.innerHTML = ''
    mermaidError.value = error instanceof Error ? error.message : t('artifact.diagramFailed')
  }
  finally {
    mermaidLoading.value = false
  }
}

onMounted(renderMermaid)
watch(() => props.artifact.content, async () => {
  await nextTick()
  await renderMermaid()
})

const formattedJson = computed(() => {
  try {
    return { content: JSON.stringify(JSON.parse(props.artifact.content), null, 2), error: '' }
  }
  catch {
    return { content: props.artifact.content, error: t('artifact.invalidJson') }
  }
})

const highlightedCode = computed(() => {
  const language = props.artifact.language || 'text'
  try {
    return hljs.getLanguage(language)
      ? hljs.highlight(props.artifact.content, { language }).value
      : hljs.highlightAuto(props.artifact.content).value
  }
  catch {
    return hljs.highlightAuto(props.artifact.content).value
  }
})

const highlightedJson = computed(() => hljs.highlight(formattedJson.value.content, { language: 'json' }).value)
</script>

<template>
  <div class="artifact-content">
    <div v-if="artifact.type === 'code'" class="code-artifact">
      <div class="code-display">
        <pre><code :class="`language-${artifact.language || 'text'}`" v-html="highlightedCode" /></pre>
      </div>
    </div>

    <div v-else-if="artifact.type === 'html'" class="html-artifact">
      <iframe :srcdoc="sanitizedHtml" class="html-iframe" sandbox referrerpolicy="no-referrer" :title="$t('artifact.htmlPreview')" />
    </div>

    <div v-else-if="artifact.type === 'svg'" class="svg-artifact">
      <div class="svg-content" v-html="sanitizedSvg" />
    </div>

    <div v-else-if="artifact.type === 'mermaid'" class="mermaid-artifact">
      <div v-if="mermaidLoading" class="preview-status">
        {{ $t('artifact.renderingDiagram') }}
      </div>
      <div ref="mermaidContent" class="mermaid-content" />
      <div v-if="mermaidError" class="mermaid-error">
        <p>{{ mermaidError }}</p>
        <pre>{{ artifact.content }}</pre>
      </div>
    </div>

    <div v-else-if="artifact.type === 'json'" class="json-artifact">
      <div v-if="formattedJson.error" class="preview-error">
        {{ formattedJson.error }}
      </div>
      <pre><code class="language-json" v-html="highlightedJson" /></pre>
    </div>

    <div v-else-if="artifact.type === 'markdown'" class="markdown-artifact">
      <div class="markdown-content" v-html="renderedMarkdown" />
    </div>

    <div v-else class="unsupported-artifact">
      <p>{{ $t('artifact.unsupportedType') }}</p>
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
  background: white;
}

.svg-content,
.mermaid-content,
.markdown-content,
.json-artifact pre,
.code-display pre {
  overflow: auto;
}

.preview-status { color: #6b7280; padding: 0.5rem 0; }
.preview-error, .mermaid-error { color: #b91c1c; margin-bottom: 0.5rem; }

@media (max-width: 640px) {
  .artifact-content { padding: 0.5rem; }
  .html-iframe { min-height: 240px; }
}
</style>
