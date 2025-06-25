package dbtx

import (
	"context"
	"database/sql"
	"fmt"
)

// TxFuncWithContext adalah tipe callback transaksi yang menerima ctx dan tx
type TxFuncWithContext func(ctx context.Context, tx *sql.Tx) error

// WithTxContext menjalankan fn di dalam transaksi yang didukung context
// Transaksi akan bergantung pada ctx -- dibatalkan jika ctx dibatalkan
func WithTxContext(ctx context.Context, db *sql.DB, fn TxFuncWithContext) (err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("dbtx: begin tx failed: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("dbtx: commit failed: %w", commitErr)
		}
	}()

	err = fn(ctx, tx)
	return
}
