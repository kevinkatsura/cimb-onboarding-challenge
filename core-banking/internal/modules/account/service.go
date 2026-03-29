package account

import (
	"context"
	"core-banking/internal/pkg/pagination"
	"fmt"
)

type Service struct {
	repo      RepositoryInterface
	accNumGen AccountNumberGenerator
}

func NewService(repo RepositoryInterface, accNumGen AccountNumberGenerator) *Service {
	return &Service{
		repo:      repo,
		accNumGen: accNumGen,
	}
}

func (s *Service) CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error) {
	var acc Account

	// 1. Get account number
	accNumber, err := s.accNumGen.Generate()
	if err != nil {
		return nil, err
	}

	acc = Account{
		CustomerID:     req.CustomerID,
		AccountNumber:  accNumber,
		AccountType:    req.AccountType,
		Currency:       req.Currency,
		OverdraftLimit: req.OverdraftLimit,
	}

	// 2. Create account
	err = s.repo.Create(&acc)
	if err != nil {
		return nil, err
	}

	return &acc, err
}

func (s *Service) GetAccount(ctx context.Context, id string) (*Account, error) {
	return s.repo.GetByID(id)
}

func (s *Service) ListAccounts(ctx context.Context, f ListFilter) ([]Account, int, string, string, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Direction == "" {
		f.Direction = "next"
	}

	accounts, total, nextC, prevC, err := s.repo.List(ctx, f)
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

	return accounts, total, nextCursor, prevCursor, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status string) error {
	// business rule
	if status != "active" && status != "frozen" && status != "closed" {
		return fmt.Errorf("invalid status")
	}

	return s.repo.UpdateStatus(id, status)
}

func (s *Service) DeleteAccount(ctx context.Context, id string) error {
	// 1. Lock account
	acc, err := s.repo.GetByID(id)
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
	return s.repo.SoftDelete(id)
}
