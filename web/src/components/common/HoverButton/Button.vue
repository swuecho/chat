<script setup lang='ts'>
interface Emit {
  (e: 'click'): void
}

defineProps<{
  ariaLabel?: string
}>()
const emit = defineEmits<Emit>()

function handleClick() {
  emit('click')
}
</script>

<template>
  <button
    type="button"
    :aria-label="ariaLabel"
    class="flex items-center justify-center w-10 h-10 transition rounded-full hover:bg-neutral-100 dark:hover:bg-[#414755]"
    @click="handleClick"
  >
    <slot />
  </button>
</template>
