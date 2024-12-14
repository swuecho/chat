import type { AxiosProgressEvent } from 'axios'
import request from '@/utils/request/axios'



export async function fetchChatStream(
  sessionUuid: string,
  chatUuid: string,
  regenerate: boolean,
  prompt: string,
  onDownloadProgress?: (progressEvent: AxiosProgressEvent) => void,
) {
  try {
    const response = await request.post(
      '/chat_stream',
      { regenerate, prompt, sessionUuid, chatUuid, stream: true },
      { onDownloadProgress },
    )

    return response.data
  }
  catch (error) {
    console.error(error)
    throw error
  }
}
