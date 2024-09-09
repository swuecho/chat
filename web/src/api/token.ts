import request from '@/utils/request/axios'

export async function fetchAPIToken() {
        try {
                const response = await request.get('/token_10years')
                return response.data
        } catch (error) {
                throw error
        }
}