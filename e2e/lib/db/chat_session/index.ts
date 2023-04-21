export async function selectChatSessionByUserId(pool, userId: number) {
        const query = {
                text: 'SELECT id, uuid, topic, created_at, updated_at, active, max_length, prompt_length, temperature, debug FROM chat_session WHERE user_id = $1 order by id',
                values: [userId],
        };

        const result = await pool.query(query);

        return result.rows;
}