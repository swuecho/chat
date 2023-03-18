export const db_config = {
        user: process.env.PG_USER,
        host: process.env.PG_HOST,
        database: process.env.PG_DB,
        password: process.env.PG_PASS,
        port: 5432, // default PostgreSQL port
}