export async function selectModels(pool) {
        const query = {
                text: 'SELECT name, label, is_default, url, api_auth_header, api_auth_key FROM chat_model order by id',
        };

        const result = await pool.query(query);
        return result.rows;
}