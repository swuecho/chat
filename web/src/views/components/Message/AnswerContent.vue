<script lang="ts" setup>
import { computed } from 'vue'
import MarkdownIt from 'markdown-it'
import mdKatex from '@vscode/markdown-it-katex'
import hljs from 'highlight.js'
import { useBasicLayout } from '@/hooks/useBasicLayout'
import { t } from '@/locales'
import { escapeBrackets, escapeDollarNumber } from '@/utils/string'

interface Props {
  content: string
  inversion?: boolean
  isMarkdown?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  inversion: false,
  isMarkdown: true
})

const { isMobile } = useBasicLayout()

const mdi = new MarkdownIt({
  html: false,
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
  ]
})

const renderedContent = computed(() => {
  if (!props.isMarkdown || props.inversion) {
    return props.content
  }
  
  const escapedText = escapeBrackets(escapeDollarNumber(props.content))
  return mdi.render(escapedText)
})

function highlightBlock(str: string, lang?: string) {
  return `<pre class="code-block-wrapper"><div class="code-block-header"><span class="code-block-header__lang">${lang}</span><span class="code-block-header__copy">${t('chat.copyCode')}</span></div><code class="hljs code-block-body ${lang}">${str}</code></pre>`
}
</script>

<template>
  <div class="text-black leading-relaxed break-words" :class="wrapClass">
    <div v-if="isMarkdown && !inversion" class="markdown-body" v-html="renderedContent" />
    <div v-else class="whitespace-pre-wrap" v-text="renderedContent" />
  </div>
</template>

<style lang="less">
@import url('./style.less');
</style>