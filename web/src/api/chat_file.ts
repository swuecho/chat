
import request from '@/utils/request/axios'

// /chat_file/{uuid}/list

const baseURL = "/api"


export async function getChatFilesList(uuid: string) {
        try {
                const response = await request.get(`/chat_file/${uuid}/list`)
                return response.data.map((item: any) => {
                        return {
                                ...item,
                                status: 'finished',
                                url: `${baseURL}/download/${item.id}`,
                                percentage: 100
                        }
                })
        }
        catch (error) {
                console.error(error)
                throw error
        }
}
