import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import { createChatSession, deleteChatSession, getChatSessionsByUser, renameChatSession, updateChatSession } from './chat_session'

// Get QueryClient from the context
const queryClient = useQueryClient()

// queryClient.invalidateQueries({ queryKey: ['sessions'] })

// query a session, when session updated, it will be invalidated
const sessionListQuery = useQuery({
        queryKey: ['sessions'],
        queryFn: getChatSessionsByUser,
})

const createChatSessionQuery = useMutation({
        mutationFn: (variables: { uuid: string, name: string, model?: string }) => createChatSession(variables.uuid, variables.name, variables.model),
        onSuccess: () => {
                queryClient.invalidateQueries({ queryKey: ['sessions'] })
        }
})

const deleteChatSessionQuery = useMutation({
        mutationFn: (uuid: string) => deleteChatSession(uuid),
        onSuccess: () => {
                queryClient.invalidateQueries({ queryKey: ['sessions'] })
        }
})

const renameChatSessionQuery = useMutation({
        mutationFn: (variables: { uuid: string, name: string }) => renameChatSession(variables.uuid, variables.name),
        onSuccess: () => {
                queryClient.invalidateQueries({ queryKey: ['sessions'] })
        }
})

const updateChatSessionQuery = useMutation({
        mutationFn: (variables: { sessionUuid: string, sessionData: Chat.Session }) => updateChatSession(variables.sessionUuid, variables.sessionData),
        onSuccess: () => {
                queryClient.invalidateQueries({ queryKey: ['sessions'] })
        }
})

export {
        sessionListQuery,
        createChatSessionQuery,
        deleteChatSessionQuery,
        renameChatSessionQuery,
        updateChatSessionQuery
}