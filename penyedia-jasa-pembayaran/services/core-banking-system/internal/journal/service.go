package journal

import (
	"context"
	"fmt"
	"time"

	"core-banking-system/pkg/apperror"
	"core-banking-system/pkg/logging"
	"core-banking-system/pkg/telemetry"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateEntry validates double-entry invariant and persists a balanced journal entry.
func (s *Service) CreateEntry(ctx context.Context, params CreateEntryParams) (string, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "JournalService.CreateEntry")
	defer span.End()
	span.SetAttributes(attribute.String("transaction_ref", params.TransactionRef))

	// Validate double-entry: sum(debit) must equal sum(credit)
	var totalDebit, totalCredit int64
	for _, l := range params.Lines {
		totalDebit += l.Debit
		totalCredit += l.Credit
	}
	if totalDebit != totalCredit {
		return "", apperror.NewBadRequest(
			fmt.Sprintf("unbalanced entry: total debit %d != total credit %d", totalDebit, totalCredit))
	}
	if totalDebit == 0 {
		return "", apperror.NewBadRequest("journal entry must have non-zero amounts")
	}

	entryID := uuid.New()
	entry := &JournalEntry{
		ID:             entryID,
		TransactionRef: params.TransactionRef,
		Description:    params.Description,
		EntryDate:      time.Now(),
		CreatedAt:      time.Now(),
	}

	// Build lines with balance_after calculation
	var lines []JournalLine
	for _, lp := range params.Lines {
		// Get current balance for this account
		bal, err := s.repo.GetBalance(ctx, lp.AccountID)
		currentBalance := int64(0)
		if err == nil && bal != nil {
			currentBalance = bal.CurrentBalance
		}

		// Calculate delta: credits increase balance, debits decrease
		delta := lp.Credit - lp.Debit
		newBalance := currentBalance + delta

		lines = append(lines, JournalLine{
			ID:             uuid.New(),
			JournalEntryID: entryID,
			AccountID:      lp.AccountID,
			Debit:          lp.Debit,
			Credit:         lp.Credit,
			Currency:       lp.Currency,
			BalanceAfter:   newBalance,
			CreatedAt:      time.Now(),
		})
	}

	// Persist entry + lines
	if err := s.repo.CreateEntry(ctx, entry, lines); err != nil {
		return "", apperror.NewInternal("failed to create journal entry", err)
	}

	// Update materialized balances
	for _, lp := range params.Lines {
		delta := lp.Credit - lp.Debit
		if err := s.repo.UpdateBalance(ctx, lp.AccountID, delta, lp.Currency, entryID); err != nil {
			logging.Ctx(ctx).Errorw("failed to update materialized balance",
				"account_id", lp.AccountID, "error", err)
		}
	}

	logging.Ctx(ctx).Infow("journal entry created successfully",
		"entry_id", entryID.String(), "transaction_ref", params.TransactionRef,
		"total_debit", totalDebit, "total_credit", totalCredit)
	return entryID.String(), nil
}

// GetBalance returns the materialized balance for an account.
func (s *Service) GetBalance(ctx context.Context, accountID string) (int64, string, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "JournalService.GetBalance")
	defer span.End()
	span.SetAttributes(attribute.String("account_id", accountID))

	bal, err := s.repo.GetBalance(ctx, accountID)
	if err != nil {
		return 0, "IDR", nil // Return zero if not found
	}
	return bal.CurrentBalance, bal.Currency, nil
}

// InitializeAccount creates a zero-balance ledger entry for a new account.
func (s *Service) InitializeAccount(ctx context.Context, accountID, currency string, initialBalance int64) error {
	ctx, span := telemetry.Tracer.Start(ctx, "JournalService.InitializeAccount")
	defer span.End()
	span.SetAttributes(attribute.String("account_id", accountID))

	return s.repo.InitializeBalance(ctx, accountID, currency, initialBalance)
}
