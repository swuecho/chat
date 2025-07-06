<template>

  <div class="jump-buttons">
    <button
      v-if="showTopButton && scrollableElement"
      @click="scrollToTop"
      class="jump-button jump-top-button"
      aria-label="Scroll to top of content"
    >
      &uarr;
    </button>
    <button
      v-if="showBottomButton && scrollableElement"
      @click="scrollToBottom"
      class="jump-button jump-bottom-button"
      aria-label="Scroll to bottom of content"
    >
      &darr;
    </button>
  </div>

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


const showTopButton = ref(false);
const showBottomButton = ref(false);
const scrollableElement = ref(null);

const scrollToTop = () => {
  if (!scrollableElement.value) return;
  scrollableElement.value.scrollTo({
    top: 0,
    behavior: 'smooth',
  });
};

const scrollToBottom = () => {
  if (!scrollableElement.value) return;
  scrollableElement.value.scrollTo({
    top: scrollableElement.value.scrollHeight,
    behavior: 'smooth',
  });
};

let scrollTimeoutId = null;

const handleScroll = () => {
  if (!scrollableElement.value) return;

  // Throttle scroll events for better performance
  if (scrollTimeoutId) return;
  
  scrollTimeoutId = setTimeout(() => {
    scrollTimeoutId = null;
    
    const el = scrollableElement.value;
    if (!el) return;
    
    const scrollHeight = el.scrollHeight;
    if (scrollHeight < 2000) {
      showTopButton.value = false;
      showBottomButton.value = false;
      return;
    }
    
    const clientHeight = el.clientHeight;
    const scrollTop = el.scrollTop;

    // Show bottom button if scrolled more than the threshold and not at bottom
    const nearBottom = (clientHeight + scrollTop) >= (scrollHeight - 10);
    showBottomButton.value = scrollTop > props.scrollThresholdShow && !nearBottom;

    // Show top button if scrolled more than the threshold and not at top, add not near bottom
    showTopButton.value = scrollTop > props.scrollThresholdShow && !nearBottom;
  }, 16); // ~60fps throttling
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

    showTopButton.value = false;
    showBottomButton.value = false; // Hide buttons if target not found

  }
};

const cleanupScrollHandling = () => {
  if (scrollTimeoutId) {
    clearTimeout(scrollTimeoutId);
    scrollTimeoutId = null;
  }
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

.jump-buttons {
  position: fixed;
  bottom: 75px;
  right: 1%;
  transform: translateX(-50%);
  z-index: 1000;
  display: flex;
  flex-direction: row;
  gap: 10px;
}

.jump-button {
  padding: 8px 16px;
  color: white;
  border: none;
  border-radius: 50px;
  cursor: pointer;
  font-size: 1.2em;
  box-shadow: 0 2px 5px rgba(0,0,0,0.2);
  transition: opacity 0.3s ease-in-out, transform 0.3s ease-in-out;
  will-change: opacity, transform;
  min-width: 40px;
  min-height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}


.jump-top-button {
  background-color: #75c788;
}

.jump-top-button:hover {
  background-color: #178430;
}

.jump-bottom-button {
  background-color: #75c788;
}

.jump-bottom-button:hover {

  background-color: #178430;
}
</style>
