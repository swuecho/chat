import { countBotAnswerHistory, listBotAnswerHistory } from '@/api/generated_client'

export async function fetchBotAnswerHistory(botUuid: string, page: number, pageSize: number) {
  const data = await listBotAnswerHistory({
    path: { bot_uuid: botUuid },
    query: { limit: pageSize, offset: (page - 1) * pageSize },
  })
  return {
    items: data.items,
    totalCount: data.total,
    totalPages: Math.ceil(data.total / pageSize),
  }
}

export async function fetchBotRunCount(botUuid: string) {
  const data = await countBotAnswerHistory({ path: { bot_uuid: botUuid } })
  return data.count
}
