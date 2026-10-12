package balances

import (
	"context"
	"database/sql"
)

// Apply persists both snapshots and the processed offset in one SQL transaction.
// Replayed offsets cannot overwrite a newer balance from the same partition.
func Apply(ctx context.Context, db *sql.DB, topic string, partition int32, offset int64, event *BalanceUpdated) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT IGNORE INTO consumer_offsets (topic, partition_id, last_offset) VALUES (?, ?, -1)`, topic, partition)
	if err != nil {
		return false, err
	}
	var last int64
	if err := tx.QueryRowContext(ctx, `SELECT last_offset FROM consumer_offsets WHERE topic = ? AND partition_id = ? FOR UPDATE`, topic, partition).Scan(&last); err != nil {
		return false, err
	}
	if offset <= last {
		return false, nil
	}
	const upsert = `INSERT INTO account_balances (account_id, balance) VALUES (?, ?) ON DUPLICATE KEY UPDATE balance = VALUES(balance)`
	if _, err := tx.ExecContext(ctx, upsert, event.Payload.AccountFrom, event.Payload.BalanceFrom.String()); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, upsert, event.Payload.AccountTo, event.Payload.BalanceTo.String()); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE consumer_offsets SET last_offset = ? WHERE topic = ? AND partition_id = ?`, offset, topic, partition); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
