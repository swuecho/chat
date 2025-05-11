// src/components/JumpToBottomButton.vue
<template>
  <button
    v-if="isVisible && scrollableElement"
    @click="scrollToBottom"
    class="jump-to-bottom-button"
    aria-label="Scroll to bottom of content"
  >
    &darr;
  </button>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue';

const props = defineProps({
  targetSelector: {
    type: String,
    required: true,
  },
  scrollThresholdShow: { // Allow customization of threshold
    type: Number,
    default: 100, // Default to 100px for element scrolling, often less than page
  }
});

const isVisible = ref(false);
const scrollableElement = ref(null);

const scrollToBottom = () => {
  if (!scrollableElement.value) return;
  scrollableElement.value.scrollTo({
    top: scrollableElement.value.scrollHeight,
    behavior: 'smooth',
  });
};

const handleScroll = () => {
        console.log("scroll")
  if (!scrollableElement.value) return;

  const el = scrollableElement.value;
  // Show button if scrolled more than the threshold
  let shouldBeVisible = el.scrollTop > props.scrollThresholdShow;

  // Hide button if already at the very bottom (or very close to it)
  const atBottom = (el.clientHeight + el.scrollTop) >= (el.scrollHeight - 10);

  if (atBottom && el.scrollTop > props.scrollThresholdShow) {
    shouldBeVisible = false;
  } else if (el.scrollTop > props.scrollThresholdShow) {
    shouldBeVisible = true;
  }
  
  isVisible.value = shouldBeVisible;
};

const initializeScrollHandling = () => {
  const element = document.querySelector(props.targetSelector);
  if (element) {
    scrollableElement.value = element;
    scrollableElement.value.addEventListener('scroll', handleScroll);
    handleScroll(); // Check initial scroll position
  } else {
    console.warn(`[JumpToBottomButton] Target element "${props.targetSelector}" not found.`);
    scrollableElement.value = null; // Ensure it's reset if target changes and isn't found
    isVisible.value = false; // Hide button if target not found
  }
};

const cleanupScrollHandling = () => {
  if (scrollableElement.value) {
    scrollableElement.value.removeEventListener('scroll', handleScroll);
  }
};

onMounted(() => {
  initializeScrollHandling();
});

onUnmounted(() => {
  cleanupScrollHandling();
});

// Watch for changes in targetSelector prop, in case it's dynamic
watch(() => props.targetSelector, (newSelector, oldSelector) => {
  if (newSelector !== oldSelector) {
    cleanupScrollHandling(); // Clean up old listener
    initializeScrollHandling(); // Initialize with new selector
  }
});

</script>

<style scoped>
.jump-to-bottom-button {
  position: fixed;
  bottom: 75px;
  left: 50%;
  z-index: 1000;

  padding: 2px 50px;
  background-color: #75c788;
  color: white;
  border: none;
  border-radius: 50px;
  cursor: pointer;
  font-size: 1.2em;
  box-shadow: 0 2px 5px rgba(0,0,0,0.2);
  transition: opacity 0.3s ease-in-out, transform 0.3s ease-in-out;
}

.jump-to-bottom-button:hover {
  background-color: #178430;
}
</style>
