import request from "@/utils/request/axios"

export async function fetchBotAnswerHistory(botUuid: string) {
  const { data } = await request.get<Bot.BotAnswerHistory[]>(`/bot_answer_history/bot/${botUuid}`)
  return data
}
