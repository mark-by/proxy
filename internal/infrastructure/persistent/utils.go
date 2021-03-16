package persistent

import "github.com/jackc/pgx"

func endTx(tx *pgx.Tx, err error) {
	if tx.Status() == -1 {
		return
	}
	if err != nil {
		_ = tx.Rollback()
		return
	}
	_ = tx.Commit()
}

