import request from "@/utils/request/axios"

export async function fetchBotAnswerHistory(botUuid: string, page: number, pageSize: number) {
  const { data } = await request.get<{
    items: Bot.BotAnswerHistory[],
    totalPages: number,
    totalCount: number
  }>(`/bot_answer_history/bot/${botUuid}`, {
    params: {
      limit: pageSize,
      offset: (page - 1) * pageSize
    }
  })
  return data
}
