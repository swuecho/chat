// src/composables/useSlashToFocus.ts
import { onMounted, onBeforeUnmount, type Ref, toRaw } from 'vue';

/**
 * Custom composable to intercept the '/' key globally and focus a target input.
 *
 * @param targetInputRef - A Vue ref pointing to the input element to focus.
 */
export function useSlashToFocus(targetInputRef: Ref<HTMLInputElement | null>): void {

        const handleGlobalKeyPress = (event: KeyboardEvent): void => {

                // Ensure the ref is pointing to an element
                if (!targetInputRef.value) {
                        return;
                }

                const activeElement = document.activeElement; // This is already a raw element

                // Check if the pressed key is '/'
                if (event.key === '/') {
                        // If the target input is already focused, allow the '/' to be typed
                        // Compare the raw target element with the active element
                        const isTypingInInput = activeElement && (
                                activeElement.tagName === 'INPUT' ||
                                activeElement.tagName === 'TEXTAREA');

                        if (isTypingInInput) {
                                return;
                        }

                        // Prevent default behavior (e.g., typing '/' into another focused input or browser's quick find)
                        event.preventDefault();

                        // Focus the target input (using the original ref.value which is fine for DOM methods)
                        targetInputRef.value.focus();
                }
        };


        onMounted(() => {
                window.addEventListener('keydown', handleGlobalKeyPress);
        });

        onBeforeUnmount(() => {
                window.removeEventListener('keydown', handleGlobalKeyPress);
        });
}
