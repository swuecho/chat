export async function selectChatMessagesBySessionUUID(pool, sessionUUID: string) {
        const query = {
                text: 'SELECT id, uuid, role, content, score, user_id, created_at, updated_at, created_by, updated_by FROM chat_message WHERE chat_session_uuid = $1 and is_deleted = false order by id',
                values: [sessionUUID],
        };

        const result = await pool.query(query);
        return result.rows;
}