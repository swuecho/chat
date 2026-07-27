package svc

import (
	"context"
	"testing"
)

func TestLegacyDuplicateMessageMigration(t *testing.T) {
	ctx := context.Background()
	tx, err := testDB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx() error = %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DROP INDEX chat_message_session_uuid_unique_idx`); err != nil {
		t.Fatalf("drop test index error = %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO chat_message (
			uuid, chat_session_uuid, role, content, reasoning_content,
			score, user_id, created_by, updated_by
		)
		VALUES
			('legacy-duplicate', 'legacy-session', 'assistant', 'oldest', '', 0, 1, 1, 1),
			('legacy-duplicate', 'legacy-session', 'assistant', 'newer', '', 0, 1, 1, 1)
	`); err != nil {
		t.Fatalf("insert legacy duplicates error = %v", err)
	}

	if _, err := tx.ExecContext(ctx, `
		WITH ranked_duplicate_messages AS (
			SELECT
				id,
				ROW_NUMBER() OVER (
					PARTITION BY chat_session_uuid, uuid
					ORDER BY id ASC
				) AS duplicate_rank
			FROM chat_message
			WHERE is_deleted = false
		)
		UPDATE chat_message AS message
		SET is_deleted = true, updated_at = now()
		FROM ranked_duplicate_messages AS duplicate
		WHERE message.id = duplicate.id
		  AND duplicate.duplicate_rank > 1;

		CREATE UNIQUE INDEX chat_message_session_uuid_unique_idx
		ON chat_message (chat_session_uuid, uuid)
		WHERE is_deleted = false;
	`); err != nil {
		t.Fatalf("legacy reconciliation migration error = %v", err)
	}

	var active, deleted int
	if err := tx.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE is_deleted = false),
			COUNT(*) FILTER (WHERE is_deleted = true)
		FROM chat_message
		WHERE chat_session_uuid = 'legacy-session'
		  AND uuid = 'legacy-duplicate'
	`).Scan(&active, &deleted); err != nil {
		t.Fatalf("count reconciled rows error = %v", err)
	}
	if active != 1 || deleted != 1 {
		t.Fatalf("reconciled rows active=%d deleted=%d, want active=1 deleted=1", active, deleted)
	}
}
