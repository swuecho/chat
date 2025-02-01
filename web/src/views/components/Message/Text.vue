<script lang="ts" setup>
import { computed, ref } from 'vue'
import MarkdownIt from 'markdown-it'
import mdKatex from '@vscode/markdown-it-katex'
import hljs from 'highlight.js'
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

function processTextForCollapsibles(text: string): string {
  return text.replace(/<think>(.*?)<\/think>/gs, (_, content) => {
    return `**reason**

${content.trim()}

**answer**
`
  })
}

const text = computed(() => {
  const value = props.text ?? ''
  // 对数学公式进行处理，自动添加 $$ 符号
  if (!props.inversion) {
    // prcessing <think></think> tag
    const thinked = processTextForCollapsibles(value)
    const escapedText = escapeBrackets(escapeDollarNumber(thinked))
    return mdi.render(escapedText)
  }
  return value
})

function highlightBlock(str: string, lang?: string) {
  return `<pre class="code-block-wrapper"><div class="code-block-header"><span class="code-block-header__lang">${lang}</span><span class="code-block-header__copy">${t('chat.copyCode')}</span></div><code class="hljs code-block-body ${lang}">${str}</code></pre>`
}

defineExpose({ textRef })

</script>

<template>
  <div class="text-black" :class="wrapClass">
    <template v-if="loading">
      <span class="dark:text-white w-[4px] h-[20px] block animate-blink" />
    </template>
    <template v-else>
      <div ref="textRef" class="leading-relaxed break-words" tabindex="-1">
        <div v-if="!inversion" class="markdown-body" v-html="text" />
        <div v-else class="whitespace-pre-wrap" v-text="text" />
      </div>
    </template>
  </div>
</template>

<style lang="less">
@import url(./style.less);
</style>
