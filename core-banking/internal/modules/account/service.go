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

func (s *Service) ListAccounts(ctx context.Context, customerID string) ([]Account, error) {
	return s.repo.ListByCustomer(customerID)
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
