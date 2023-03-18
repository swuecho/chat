export async function selectUserByEmail(pool, email: string) {
        const query = {
                text: 'SELECT id, email FROM auth_user WHERE email = $1',
                values: [email],
        };

        const result = await pool.query(query);

        if (result.rows.length === 0) {
                throw new Error(`User with email ${email} not found`);
        }

        // Assuming there's only one user with the given email
        return result.rows[0];
}