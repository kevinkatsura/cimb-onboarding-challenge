package transaction

import (
	"context"
	"core-banking/internal/database"
	"core-banking/internal/pkg/pagination"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Transfer(ctx context.Context, req TransferRequest) error {
	return database.WithSerializableRetry(ctx, func() error {
		tx, err := database.BeginSerializableTx(ctx, s.repo.DB)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// 1. Idempotency
		var exists bool
		err = tx.Get(&exists, `SELECT EXISTS(SELECT 1 FROM transactions WHERE reference_id=$1)`, req.ReferenceID)
		if err != nil {
			return err
		}
		// duplicate request
		if exists {
			return nil
		}

		// 2. Lock accounts (ordered)
		var senderAccount struct {
			FromBalance int64  `db:"available_balance"`
			CustomerID  string `db:"customer_id"`
		}
		err = tx.Get(&senderAccount, `SELECT available_balance, customer_id FROM accounts WHERE id=$1 FOR UPDATE`, req.FromAccount)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`SELECT 1 FROM accounts WHERE id=$1 FOR UPDATE`, req.ToAccount)
		if err != nil {
			return err
		}

		// 3. Validate
		if senderAccount.FromBalance < req.Amount {
			return fmt.Errorf("Insufficient balance")
		}

		// 4. Insert transaction
		var txID string
		err = tx.Get(&txID, `
			INSERT INTO transactions(reference_id, transaction_type, status, amount, currency, initiated_by) 
			VALUES ($1, 'transfer', 'pending', $2, $3, $4) RETURNING id;`,
			req.ReferenceID, req.Amount, req.Currency, senderAccount.CustomerID)
		if err != nil {
			return err
		}

		// 5. Journal
		var journalID string
		err = tx.Get(&journalID, `
			INSERT INTO journal_entries(transaction_id, journal_type)
			VALUES ($1, 'transfer') RETURNING id;`,
			txID)
		if err != nil {
			return err
		}

		// 6. Ledger
		_, err = tx.Exec(`
			INSERT INTO ledger_entries(journal_id, account_id, entry_type, amount, currency)
			VALUES 
				($1, $2, 'debit', $3, $4),
				($1, $5, 'credit', $3, $4)`,
			journalID, req.FromAccount, req.Amount, req.Currency, req.ToAccount)
		if err != nil {
			return err
		}

		// 7. Update balances
		_, err = tx.Exec(`
			UPDATE accounts SET available_balance = available_balance - $1 WHERE id=$2`,
			req.Amount, req.FromAccount)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			UPDATE accounts SET available_balance = available_balance + $1 WHERE id=$2`,
			req.Amount, req.ToAccount)
		if err != nil {
			return err
		}

		// 8.Complete
		_, err = tx.Exec(`
			UPDATE transactions
			SET status='completed', completed_at=NOW()
			WHERE id=$1`, txID)
		if err != nil {
			return err
		}

		return tx.Commit()
	})
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]TransactionHistoryDTO, int, string, string, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Direction == "" {
		f.Direction = "next"
	}

	data, total, nextC, prevC, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, "", "", err
	}

	var nextCursor, prevCursor string
	if nextC != nil {
		nextCursor, _ = pagination.EncodeCursor(*nextC)
	}
	if prevC != nil {
		prevCursor, _ = pagination.EncodeCursor(*prevC)
	}

	return data, total, nextCursor, prevCursor, nil
}
