
import request from '@/utils/request/axios'

// /chat_file/{uuid}/list

export async function getChatFilesList(uuid: string) {
        try {
                const response = await request.get(`/chat_file/${uuid}/list`)
                return response.data.map((item: any) => {
                        return {
                                ...item,
                                status: 'finished',
                                url: `/download/${item.id}`
                        }})
        }
        catch (error) {
                console.error(error)
                throw error
        }
}
