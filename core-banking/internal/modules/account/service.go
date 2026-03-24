package account

import (
	"context"
	"core-banking/internal/database"
	"crypto/rand"
	"fmt"
	"math/big"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func generateAccountNumber() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1e12))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("10%010d", n.Int64()), nil
}

func (s *Service) CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error) {
	var acc Account

	err := database.WithSerializableRetry(ctx, func() error {
		tx, err := database.BeginSerializableTx(ctx, s.repo.DB)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// 1. Get account number
		accNumber, err := generateAccountNumber()
		if err != nil {
			return err
		}

		acc = Account{
			CustomerID:     req.CustomerID,
			AccountNumber:  accNumber,
			AccountType:    req.AccountType,
			Currency:       req.Currency,
			OverdraftLimit: req.OverdraftLimit,
		}

		// 2. Create account
		err = s.repo.Create(tx, &acc)
		if err != nil {
			return err
		}

		return tx.Commit()
	})

	return &acc, err
}

func (s *Service) GetAccount(ctx context.Context, id string) (*Account, error) {
	return s.repo.GetByID(id)
}

func (s *Service) ListAccounts(ctx context.Context, f ListFilter) ([]Account, int, string, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	accounts, total, nextCursor, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, "", err
	}

	var cursorStr string
	if nextCursor != nil {
		cursorStr, _ = EncodeCursor(*nextCursor)
	}

	return accounts, total, cursorStr, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status string) error {
	return database.WithSerializableRetry(ctx, func() error {
		tx, err := database.BeginSerializableTx(ctx, s.repo.DB)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// business rule
		if status != "active" && status != "frozen" && status != "closed" {
			return fmt.Errorf("invalid status")
		}

		err = s.repo.UpdateStatus(tx, id, status)
		if err != nil {
			return err
		}

		return tx.Commit()
	})
}

func (s *Service) DeleteAccount(ctx context.Context, id string) error {
	return database.WithSerializableRetry(ctx, func() error {
		tx, err := database.BeginSerializableTx(ctx, s.repo.DB)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// 1. Lock account
		var acc Account
		err = tx.Get(&acc, `
			SELECT id,
				customer_id,
				account_number,
				account_type,
				currency,
				status,
				available_balance,
				pending_balance,
				overdraft_limit,
				opened_at,
				closed_at,
				created_at,
				updated_at 
			FROM accounts
			WHERE id=$id FOR UPDATE`, id)
		if err != nil {
			return err
		}

		// 2. Business rules (CRITICAL)
		if acc.AvailableBalance != 0 {
			return fmt.Errorf("cannot delete account with non-zero balance")
		}

		if acc.Status != "closed" {
			return fmt.Errorf("account must be closed before deletion")
		}

		// 3. Soft delete
		err = s.repo.SoftDelete(tx, id)
		if err != nil {
			return err
		}

		return tx.Commit()
	})
}
