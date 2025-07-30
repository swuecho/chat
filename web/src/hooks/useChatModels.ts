import { useQuery, useQueryClient, useMutation } from '@tanstack/vue-query'
import { computed } from 'vue'
import { useAuthStore } from '@/store'
import { fetchChatModel, updateChatModel, deleteChatModel, createChatModel, fetchDefaultChatModel } from '@/api/chat_model'

export const useChatModels = () => {
  const authStore = useAuthStore()
  const queryClient = useQueryClient()

  const useChatModelsQuery = () => {
    return useQuery({
      queryKey: ['chat_models'],
      queryFn: fetchChatModel,
      staleTime: 10 * 60 * 1000, // 10 minutes
      enabled: computed(() => authStore.isInitialized && !authStore.isInitializing && authStore.isValid),
    })
  }

  const useDefaultChatModelQuery = () => {
    return useQuery({
      queryKey: ['chat_models', 'default'],
      queryFn: fetchDefaultChatModel,
      staleTime: 10 * 60 * 1000, // 10 minutes
      enabled: computed(() => authStore.isInitialized && !authStore.isInitializing && authStore.isValid),
    })
  }

  const useUpdateChatModelMutation = () => {
    return useMutation({
      mutationFn: ({ id, data }: { id: number; data: any }) => updateChatModel(id, data),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['chat_models'] })
      },
    })
  }

  const useDeleteChatModelMutation = () => {
    return useMutation({
      mutationFn: (id: number) => deleteChatModel(id),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['chat_models'] })
      },
    })
  }

  const useCreateChatModelMutation = () => {
    return useMutation({
      mutationFn: (data: any) => createChatModel(data),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['chat_models'] })
      },
    })
  }

  return {
    useChatModelsQuery,
    useDefaultChatModelQuery,
    useUpdateChatModelMutation,
    useDeleteChatModelMutation,
    useCreateChatModelMutation,
  }
}