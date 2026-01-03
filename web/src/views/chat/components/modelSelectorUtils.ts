export const getInitialModelState = (sessionModel?: string, defaultModel?: string) => {
  return {
    initialModel: sessionModel ?? defaultModel,
    shouldCommit: Boolean(sessionModel),
  }
}
