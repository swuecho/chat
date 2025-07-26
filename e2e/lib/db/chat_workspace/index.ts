import { Pool } from 'pg';

export interface ChatWorkspace {
    id: number;
    uuid: string;
    user_id: number;
    name: string;
    description: string;
    color: string;
    icon: string;
    created_at: Date;
    updated_at: Date;
    is_default: boolean;
    order_position: number;
}

export async function selectWorkspacesByUserId(pool: Pool, userId: number): Promise<ChatWorkspace[]> {
    const client = await pool.connect();
    try {
        const result = await client.query(
            'SELECT * FROM chat_workspace WHERE user_id = $1 ORDER BY order_position',
            [userId]
        );
        return result.rows;
    } finally {
        client.release();
    }
}

export async function selectWorkspaceByUuid(pool: Pool, uuid: string): Promise<ChatWorkspace | null> {
    const client = await pool.connect();
    try {
        const result = await client.query(
            'SELECT * FROM chat_workspace WHERE uuid = $1',
            [uuid]
        );
        return result.rows[0] || null;
    } finally {
        client.release();
    }
}

export async function insertWorkspace(pool: Pool, workspace: Omit<ChatWorkspace, 'id' | 'created_at' | 'updated_at'>): Promise<ChatWorkspace> {
    const client = await pool.connect();
    try {
        const result = await client.query(
            `INSERT INTO chat_workspace (uuid, user_id, name, description, color, icon, is_default, order_position)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
             RETURNING *`,
            [workspace.uuid, workspace.user_id, workspace.name, workspace.description, workspace.color, workspace.icon, workspace.is_default, workspace.order_position]
        );
        return result.rows[0];
    } finally {
        client.release();
    }
}

export async function countSessionsInWorkspace(pool: Pool, workspaceId: number): Promise<number> {
    const client = await pool.connect();
    try {
        const result = await client.query(
            'SELECT COUNT(*) as count FROM chat_session WHERE workspace_id = $1',
            [workspaceId]
        );
        return parseInt(result.rows[0].count);
    } finally {
        client.release();
    }
}