package transaction

import (
	"context"
	"core-banking/internal/pkg/pagination"
	"core-banking/internal/service"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type Service struct {
	repo        *Repository
	lockManager *service.AccountLockManager
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo:        repo,
		lockManager: service.NewAccountLockManager(),
	}
}

func (s *Service) Transfer(ctx context.Context, req TransferRequest) error {
	// 1. Idempotency
	var exists bool
	err := s.repo.DB.Get(&exists, `SELECT EXISTS(SELECT 1 FROM transactions WHERE reference_id=$1)`, req.ReferenceID)
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
	err = s.repo.DB.Get(&senderAccount, `SELECT available_balance, customer_id FROM accounts WHERE id=$1 FOR UPDATE`, req.FromAccount)
	if err != nil {
		return err
	}

	_, err = s.repo.DB.Exec(`SELECT 1 FROM accounts WHERE id=$1 FOR UPDATE`, req.ToAccount)
	if err != nil {
		return err
	}

	// 3. Validate
	if senderAccount.FromBalance < req.Amount {
		return fmt.Errorf("Insufficient balance")
	}

	// 4. Insert transaction
	var txID string
	err = s.repo.DB.Get(&txID, `
			INSERT INTO transactions(reference_id, transaction_type, status, amount, currency, initiated_by) 
			VALUES ($1, 'transfer', 'pending', $2, $3, $4) RETURNING id;`,
		req.ReferenceID, req.Amount, req.Currency, senderAccount.CustomerID)
	if err != nil {
		return err
	}

	// 5. Journal
	var journalID string
	err = s.repo.DB.Get(&journalID, `
			INSERT INTO journal_entries(transaction_id, journal_type)
			VALUES ($1, 'transfer') RETURNING id;`,
		txID)
	if err != nil {
		return err
	}

	// 6. Ledger
	_, err = s.repo.DB.Exec(`
			INSERT INTO ledger_entries(journal_id, account_id, entry_type, amount, currency)
			VALUES 
				($1, $2, 'debit', $3, $4),
				($1, $5, 'credit', $3, $4)`,
		journalID, req.FromAccount, req.Amount, req.Currency, req.ToAccount)
	if err != nil {
		return err
	}

	// 7. Update balances
	_, err = s.repo.DB.Exec(`
			UPDATE accounts SET available_balance = available_balance - $1 WHERE id=$2`,
		req.Amount, req.FromAccount)
	if err != nil {
		return err
	}

	_, err = s.repo.DB.Exec(`
			UPDATE accounts SET available_balance = available_balance + $1 WHERE id=$2`,
		req.Amount, req.ToAccount)
	if err != nil {
		return err
	}

	// 8.Complete
	_, err = s.repo.DB.Exec(`
			UPDATE transactions
			SET status='completed', completed_at=NOW()
			WHERE id=$1`, txID)
	return err
}

func (s *Service) TransferWithLock(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	lockKey := req.ToAccount
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// --- Acquire Lock ---
	if err := s.lockManager.Lock(ctx, lockKey); err != nil {
		return nil, err
	}
	defer s.lockManager.Unlock(lockKey)

	resultCh := make(chan transferResult, 1)

	// --- Run critical section ---
	go func() {
		resp, err := s.transferCriticalSection(req)
		resultCh <- transferResult{
			response: resp,
			err:      err,
		}
	}()

	// --- Wait for result or timeout ---
	select {
	case result := <-resultCh:
		return result.response, result.err

	case <-ctx.Done():
		timeoutResp := &TransferResponse{
			Status:             "failed",
			TransactionID:      nil,
			SourceAccount:      req.FromAccount,
			DestinationAccount: req.ToAccount,
			Amount:             req.Amount,
			Message:            "transfer timeout (exceeded 4s)",
		}

		return timeoutResp, fmt.Errorf("transfer timeout (exceeded 4s)")
	}
}

func (s *Service) transferCriticalSection(req TransferRequest) (*TransferResponse, error) {
	var result *TransferResponse

	// --- RANDOM DELAY (1–5 seconds) ---
	delay := time.Duration(rand.Intn(5)+1) * time.Second
	time.Sleep(delay)

	// --- Idempotency ---
	var exists bool
	if err := s.repo.DB.Get(&exists,
		`SELECT EXISTS(SELECT 1 FROM transactions WHERE reference_id=$1)`,
		req.ReferenceID); err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("idempotency check failed")
	}

	// --- Lock accounts ---
	var sender struct {
		Balance   int64  `db:"available_balance"`
		AccountNo string `db:"account_number"`
		Customer  string `db:"customer_id"`
	}

	if err := s.repo.DB.Get(&sender,
		`SELECT available_balance, account_number, customer_id
			 FROM accounts WHERE id=$1 FOR UPDATE`,
		req.FromAccount); err != nil {
		return nil, err
	}

	var receiver struct {
		Balance   int64  `db:"available_balance"`
		AccountNo string `db:"account_number"`
	}

	if err := s.repo.DB.Get(&receiver,
		`SELECT available_balance, account_number
			 FROM accounts WHERE id=$1 FOR UPDATE`,
		req.ToAccount); err != nil {
		return nil, err
	}

	// --- Validation ---
	if sender.Balance < req.Amount {
		log.Println("transfer_failed",
			"transaction_id", nil,
			"source_account", sender.AccountNo,
			"destination_account", receiver.AccountNo,
			"amount", req.Amount,
			"current_balance", sender.Balance,
		)

		return &TransferResponse{
			Status:                  "failed",
			TransactionID:           nil,
			SourceAccount:           sender.AccountNo,
			DestinationAccount:      receiver.AccountNo,
			Amount:                  req.Amount,
			SourceBalanceAfter:      &sender.Balance,
			DestinationBalanceAfter: &receiver.Balance,
			Message:                 "insufficient balance",
		}, fmt.Errorf("insufficient balance")

	}

	// --- Insert transaction ---
	var txID string
	if err := s.repo.DB.Get(&txID, `
			INSERT INTO transactions(reference_id, transaction_type, status, amount, currency, initiated_by)
			VALUES ($1, 'transfer', 'pending', $2, $3, $4)
			RETURNING id`,
		req.ReferenceID, req.Amount, req.Currency, sender.Customer); err != nil {
		return nil, err
	}

	// --- Journal ---
	var journalID string
	if err := s.repo.DB.Get(&journalID, `
			INSERT INTO journal_entries(transaction_id, journal_type)
			VALUES ($1, 'transfer') RETURNING id`,
		txID); err != nil {
		return nil, err
	}

	// --- Ledger ---
	if _, err := s.repo.DB.Exec(`
			INSERT INTO ledger_entries(journal_id, account_id, entry_type, amount, currency)
			VALUES
				($1, $2, 'debit', $3, $4),
				($1, $5, 'credit', $3, $4)`,
		journalID, req.FromAccount, req.Amount, req.Currency, req.ToAccount); err != nil {
		return nil, err
	}

	// --- Update balances ---
	if _, err := s.repo.DB.Exec(`
			UPDATE accounts SET available_balance = available_balance - $1 WHERE id=$2`,
		req.Amount, req.FromAccount); err != nil {
		return nil, err
	}

	if _, err := s.repo.DB.Exec(`
			UPDATE accounts SET available_balance = available_balance + $1 WHERE id=$2`,
		req.Amount, req.ToAccount); err != nil {
		return nil, err
	}

	// --- Complete ---
	if _, err := s.repo.DB.Exec(`
			UPDATE transactions
			SET status='completed', completed_at=NOW()
			WHERE id=$1`, txID); err != nil {
		return nil, err
	}

	sender.Balance -= req.Amount
	receiver.Balance += req.Amount

	// --- Success logging ---
	result = &TransferResponse{
		Status:                  "success",
		TransactionID:           &txID,
		SourceAccount:           sender.AccountNo,
		DestinationAccount:      receiver.AccountNo,
		Amount:                  req.Amount,
		SourceBalanceAfter:      &sender.Balance,
		DestinationBalanceAfter: &receiver.Balance,
		Message:                 "transfer completed successfully",
	}

	log.Println("transfer_success",
		"source_account", result.SourceAccount,
		"source_balance_after", result.SourceBalanceAfter,
		"destination_account", result.DestinationAccount,
		"destination_balance_after", result.DestinationBalanceAfter,
	)

	return result, nil
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
