import { QueryClient } from '@tanstack/vue-query'

// Shared so non-component code (Pinia stores) reads the same cache as components.
export const queryClient = new QueryClient()
