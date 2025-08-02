import { useQuery, useQueryClient, useMutation } from '@tanstack/vue-query'
import { computed } from 'vue'
import { useAuthStore } from '@/store'
import { fetchChatModel, updateChatModel, deleteChatModel, createChatModel, fetchDefaultChatModel } from '@/api/chat_model'
import type { ChatModel, CreateChatModelRequest, UpdateChatModelRequest } from '@/types/chat-models'

export const useChatModels = () => {
  const authStore = useAuthStore()
  const queryClient = useQueryClient()

  const useChatModelsQuery = () => {
    return useQuery<ChatModel[]>({
      queryKey: ['chat_models'],
      queryFn: fetchChatModel,
      staleTime: 5 * 60 * 1000, // 5 minutes - reduced for better responsiveness
      enabled: computed(() => authStore.isInitialized && !authStore.isInitializing && authStore.isValid),
    })
  }

  const useDefaultChatModelQuery = () => {
    return useQuery<ChatModel>({
      queryKey: ['chat_models', 'default'],
      queryFn: fetchDefaultChatModel,
      staleTime: 5 * 60 * 1000, // 5 minutes - reduced for better responsiveness
      enabled: computed(() => authStore.isInitialized && !authStore.isInitializing && authStore.isValid),
    })
  }

  const useUpdateChatModelMutation = () => {
    return useMutation<ChatModel, Error, { id: number; data: UpdateChatModelRequest }>({
      mutationFn: ({ id, data }) => updateChatModel(id, data),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['chat_models'] })
        queryClient.invalidateQueries({ queryKey: ['chat_models', 'default'] })
      },
    })
  }

  const useDeleteChatModelMutation = () => {
    return useMutation<void, Error, number>({
      mutationFn: (id: number) => deleteChatModel(id),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['chat_models'] })
        queryClient.invalidateQueries({ queryKey: ['chat_models', 'default'] })
      },
    })
  }

  const useCreateChatModelMutation = () => {
    return useMutation<ChatModel, Error, CreateChatModelRequest>({
      mutationFn: (data: CreateChatModelRequest) => createChatModel(data),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['chat_models'] })
        queryClient.invalidateQueries({ queryKey: ['chat_models', 'default'] })
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