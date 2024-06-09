
import request from '@/utils/request/axios'

// /chat_file/{uuid}/list

const baseURL = import.meta.env.VITE_GLOB_API_URL


export async function getChatFilesList(uuid: string) {
        try {
                if (!uuid) return []
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
