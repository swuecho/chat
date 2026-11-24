import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { computed } from 'vue'
import { useAuthStore } from '@/store'
import {
  createChatModel,
  deleteChatModel,
  getTitleChatModel,
  listChatModels,
  setTitleChatModel,
  updateChatModel,
} from '@/api/generated_client'
import { fetchDefaultChatModel } from '@/api/chat_model'
import type { ChatModelHttpResponse, CreateChatModelRequest } from '@/api/generated_client'
import { queryClient as sharedQueryClient } from '@/plugins/queryClient'

// Every chat-model query lives under ['chat_models'], so invalidating
// chatModelKeys.all refreshes the list, the default model, and the title model.
export const chatModelKeys = {
  all: ['chat_models'] as const,
  default: ['chat_models', 'default'] as const,
  title: ['chat_models', 'title'] as const,
}

const STALE_TIME = 5 * 60 * 1000

// For non-component callers (stores): reads the default model through the
// shared query cache instead of issuing a separate request.
export const fetchDefaultChatModelCached = () =>
  sharedQueryClient.fetchQuery({ queryKey: chatModelKeys.default, queryFn: fetchDefaultChatModel, staleTime: STALE_TIME })

interface MutationCallbacks<T> {
  onSuccess?: (data: T) => void
  onError?: (error: Error) => void
}

export const useChatModels = () => {
  const authStore = useAuthStore()
  const queryClient = useQueryClient()
  const enabled = computed(() => authStore.isInitialized && !authStore.isInitializing && authStore.isValid)
  const invalidateModels = () => queryClient.invalidateQueries({ queryKey: chatModelKeys.all })

  const useChatModelsQuery = () => {
    return useQuery<ChatModelHttpResponse[]>({
      queryKey: chatModelKeys.all,
      queryFn: () => listChatModels(),
      staleTime: STALE_TIME,
      enabled,
    })
  }

  const useDefaultChatModelQuery = () => {
    return useQuery<ChatModelHttpResponse>({
      queryKey: chatModelKeys.default,
      queryFn: fetchDefaultChatModel,
      staleTime: STALE_TIME,
      enabled,
    })
  }

  const useTitleChatModelQuery = () => {
    return useQuery<ChatModelHttpResponse>({
      queryKey: chatModelKeys.title,
      queryFn: () => getTitleChatModel(),
      retry: false,
      enabled,
    })
  }

  const useUpdateChatModelMutation = (callbacks: MutationCallbacks<ChatModelHttpResponse> = {}) => {
    return useMutation<ChatModelHttpResponse, Error, { id: number; data: CreateChatModelRequest }>({
      mutationFn: ({ id, data }) => updateChatModel({ path: { id }, body: data }),
      onSuccess: (data) => {
        invalidateModels()
        callbacks.onSuccess?.(data)
      },
      onError: callbacks.onError,
    })
  }

  const useDeleteChatModelMutation = (callbacks: MutationCallbacks<void> = {}) => {
    return useMutation<void, Error, number>({
      mutationFn: async (id: number) => {
        await deleteChatModel({ path: { id } })
      },
      onSuccess: () => {
        invalidateModels()
        callbacks.onSuccess?.()
      },
      onError: callbacks.onError,
    })
  }

  const useCreateChatModelMutation = (callbacks: MutationCallbacks<ChatModelHttpResponse> = {}) => {
    return useMutation<ChatModelHttpResponse, Error, CreateChatModelRequest>({
      mutationFn: (data: CreateChatModelRequest) => createChatModel({ body: data }),
      onSuccess: (data) => {
        invalidateModels()
        callbacks.onSuccess?.(data)
      },
      onError: callbacks.onError,
    })
  }

  const useSetTitleChatModelMutation = (callbacks: MutationCallbacks<ChatModelHttpResponse> = {}) => {
    return useMutation<ChatModelHttpResponse, Error, number>({
      mutationFn: (modelId: number) => setTitleChatModel({ body: { modelId } }),
      onSuccess: (data) => {
        invalidateModels()
        callbacks.onSuccess?.(data)
      },
      onError: callbacks.onError,
    })
  }

  return {
    useChatModelsQuery,
    useDefaultChatModelQuery,
    useTitleChatModelQuery,
    useUpdateChatModelMutation,
    useDeleteChatModelMutation,
    useCreateChatModelMutation,
    useSetTitleChatModelMutation,
  }
}
