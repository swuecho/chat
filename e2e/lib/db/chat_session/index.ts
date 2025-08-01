export async function selectChatSessionByUserId(pool, userId: number) {
        const query = {
                text: 'SELECT id, uuid, topic, created_at, updated_at, active, max_length, temperature, n, debug, model, workspace_id FROM chat_session WHERE user_id = $1 order by id',
                values: [userId],
        };

        const result = await pool.query(query);

        return result.rows;
}