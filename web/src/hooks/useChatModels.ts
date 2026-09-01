import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed } from 'vue'
import { useAuthStore } from '@/store'
import { createChatModel, deleteChatModel, listChatModels, updateChatModel } from '@/api/generated_client'
import { fetchDefaultChatModel } from '@/api/chat_model'
import type { ChatModelHttpResponse, CreateChatModelRequest } from '@/api/generated_client'

export const useChatModels = () => {
  const authStore = useAuthStore()
  const queryClient = useQueryClient()

  const useChatModelsQuery = () => {
    return useQuery<ChatModelHttpResponse[]>({
      queryKey: ['chat_models'],
      queryFn: () => listChatModels(),
      staleTime: 5 * 60 * 1000, // 5 minutes - reduced for better responsiveness
      enabled: computed(() => authStore.isInitialized && !authStore.isInitializing && authStore.isValid),
    })
  }

  const useDefaultChatModelQuery = () => {
    return useQuery<ChatModelHttpResponse>({
      queryKey: ['chat_models', 'default'],
      queryFn: fetchDefaultChatModel,
      staleTime: 5 * 60 * 1000, // 5 minutes - reduced for better responsiveness
      enabled: computed(() => authStore.isInitialized && !authStore.isInitializing && authStore.isValid),
    })
  }

  const useUpdateChatModelMutation = () => {
    return useMutation<ChatModelHttpResponse, Error, { id: number; data: CreateChatModelRequest }>({
      mutationFn: ({ id, data }) => updateChatModel({ path: { id }, body: data }),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['chat_models'] })
        queryClient.invalidateQueries({ queryKey: ['chat_models', 'default'] })
      },
    })
  }

  const useDeleteChatModelMutation = () => {
    return useMutation<void, Error, number>({
      mutationFn: async (id: number) => {
        await deleteChatModel({ path: { id } })
      },
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['chat_models'] })
        queryClient.invalidateQueries({ queryKey: ['chat_models', 'default'] })
      },
    })
  }

  const useCreateChatModelMutation = () => {
    return useMutation<ChatModelHttpResponse, Error, CreateChatModelRequest>({
      mutationFn: (data: CreateChatModelRequest) => createChatModel({ body: data }),
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
