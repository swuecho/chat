export interface UpdateChatSessionPayload {
  uuid: string
  topic: string
  maxLength: number
  temperature: number
  model: string
  topP: number
  n: number
  maxTokens: number
  debug: boolean
  summarizeMode: boolean
  exploreMode: boolean
  artifactEnabled: boolean
  workspaceUuid?: string
}

// Keep the UI model out of the transport contract. Strict backend decoding
// intentionally rejects view-only fields such as title and isEdit.
export function toUpdateChatSessionPayload(session: Chat.Session): UpdateChatSessionPayload {
  if (!session.model)
    throw new Error('Cannot update a session without a model')
  if (session.maxTokens == null || session.maxTokens < 1)
    throw new Error('Cannot update a session without a positive maxTokens value')

  return {
    uuid: session.uuid,
    topic: session.title,
    maxLength: session.maxLength ?? 10,
    temperature: session.temperature ?? 1,
    model: session.model,
    topP: session.topP ?? 1,
    n: session.n ?? 1,
    maxTokens: session.maxTokens,
    debug: session.debug ?? false,
    summarizeMode: session.summarizeMode ?? false,
    exploreMode: session.exploreMode ?? false,
    artifactEnabled: session.artifactEnabled ?? false,
    ...(session.workspaceUuid ? { workspaceUuid: session.workspaceUuid } : {}),
  }
}
