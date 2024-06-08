
import request from '@/utils/request/axios'

// /chat_file/{uuid}/list

export async function getChatFilesList(uuid: string) {
        try {
                if (!uuid) return []
                const response = await request.get(`/chat_file/${uuid}/list`)
                console.log(response.data)
                return response.data.map((item: any) => {
                        return {
                                ...item,
                                status: 'finished',
                                url: `/download/${item.id}`,
                                percentage: 100
                        }
                })
        }
        catch (error) {
                console.error(error)
                throw error
        }
}
