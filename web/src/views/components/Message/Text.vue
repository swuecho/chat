<script lang="ts" setup>
import { computed, ref } from 'vue'
import MarkdownIt from 'markdown-it'
import mdKatex from '@vscode/markdown-it-katex'
import hljs from 'highlight.js'
import { parseText } from './Util'
import { useThinkingContent } from './useThinkingContent'
import ThinkingRenderer from './ThinkingRenderer.vue'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { t } from '@/locales'
import { escapeBrackets, escapeDollarNumber } from '@/utils/string'
interface Props {
  inversion?: boolean // user message is inversioned (on the right side)
  error?: boolean
  text?: string
  loading?: boolean
  code?: boolean
}

const props = defineProps<Props>()

const { isMobile } = useBasicLayout()

const textRef = ref<HTMLElement>()

// Use the new thinking content composable
const { thinkingContent, hasThinking, toggleExpanded, isExpanded } = useThinkingContent(props.text)

const mdi = new MarkdownIt({
  html: false, // true vs false
  linkify: true,
  highlight(code, language) {
    const validLang = !!(language && hljs.getLanguage(language))
    if (validLang) {
      const lang = language ?? ''
      return highlightBlock(hljs.highlight(lang, code, true).value, lang)
    }
    return highlightBlock(hljs.highlightAuto(code).value, '')
  },
})

mdi.use(mdKatex, { blockClass: 'katexmath-block rounded-md p-[10px]', errorColor: ' #cc0000' })

const wrapClass = computed(() => {
  return [
    'text-wrap',
    'min-w-[20px]',
    'rounded-md',
    isMobile.value ? 'p-2' : 'px-3 py-2',
    props.inversion ? 'bg-[#d2f9d1]' : 'bg-[#f4f6f8]',
    props.inversion ? 'dark:bg-[#a1dc95]' : 'dark:bg-[#1e1e20]',
    { 'text-red-500': props.error },
  ]
})

const text = computed(() => {
  const value = parseText(props.text ?? '').answerPart
  // 对数学公式进行处理，自动添加 $$ 符号
  if (!props.inversion) {
    const escapedText = escapeBrackets(escapeDollarNumber(value))
    return mdi.render(escapedText)
  }
  return value
})

const thinkText = computed(() => {
  if (!props.inversion && hasThinking.value) {
    const escapedText = escapeBrackets(escapeDollarNumber(thinkingContent.value?.content || ''))
    return mdi.render(escapedText)
  }
  return ''
})

function highlightBlock(str: string, lang?: string) {
  return `<pre class="code-block-wrapper"><div class="code-block-header"><span class="code-block-header__lang">${lang}</span><span class="code-block-header__copy">${t('chat.copyCode')}</span></div><code class="hljs code-block-body ${lang}">${str}</code></pre>`
}

defineExpose({ textRef })
</script>

<template>
  <div class="text-black relative" :class="wrapClass">
    <template v-if="loading">
      <span class="dark:text-white w-[4px] h-[20px] block animate-blink" />
    </template>
    <template v-else>
      <div ref="textRef" class="leading-relaxed break-words" tabindex="-1">
        <ThinkingRenderer
          v-if="!inversion && thinkingContent"
          :content="thinkingContent"
          :options="{
            enableMarkdown: true,
            enableCollapsible: true,
            defaultExpanded: isExpanded,
            showBorder: true,
            borderColor: 'border-lime-600',
            maxLines: 20,
            enableCopy: true
          }"
          @toggle="toggleExpanded"
        />
        <div v-if="!inversion" class="markdown-body" v-html="text" />
        <div v-else class="whitespace-pre-wrap" v-text="text" />
      </div>
    </template>
  </div>
</template>

<style lang="less">
@import url(./style.less);
</style>
