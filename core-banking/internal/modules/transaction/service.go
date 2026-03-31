package transaction

import (
	"context"
	"core-banking/internal/pkg/logging"
	"core-banking/internal/pkg/pagination"
	"core-banking/internal/service"
	"fmt"
	"math/rand"
	"time"
)

type DelayFunc func()

type Service struct {
	repo        TransactionRepositoryInterface
	lockManager service.LockManager
	delay       DelayFunc
}

func NewService(repo TransactionRepositoryInterface, lockManager service.LockManager) *Service {
	return &Service{
		repo:        repo,
		lockManager: lockManager,
		delay: func() {
			time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
		},
	}
}

func (s *Service) Transfer(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	logging.Logger().Debugw("transfer_initiated",
		"reference_id", req.ReferenceID,
		"from_account", req.FromAccount,
		"to_account", req.ToAccount,
		"amount", req.Amount,
		"currency", req.Currency,
	)

	// 1. Idempotency
	exists, err := s.repo.IsTransactionExists(ctx, req.ReferenceID)
	if err != nil {
		logging.Logger().Errorw("idempotency_check_failed",
			"reference_id", req.ReferenceID,
			"error", err,
		)
		return nil, err
	}
	if exists {
		logging.Logger().Warnw("transaction_already_processed",
			"reference_id", req.ReferenceID,
		)
		return nil, fmt.Errorf("idempotency check failed")
	}

	// 2. Lock sender
	sender, err := s.repo.GetSenderForUpdate(ctx, req.FromAccount)
	if err != nil {
		logging.Logger().Errorw("sender_account_not_found",
			"account_id", req.FromAccount,
			"error", err,
		)
		return nil, err
	}

	// 3. Lock receiver
	if err := s.repo.LockReceiver(ctx, req.ToAccount); err != nil {
		logging.Logger().Errorw("failed_to_lock_receiver",
			"account_id", req.ToAccount,
			"error", err,
		)
		return nil, err
	}

	// 4. Validation
	if sender.Balance < req.Amount {
		logging.Logger().Warnw("insufficient_balance_for_transfer",
			"from_account", req.FromAccount,
			"to_account", req.ToAccount,
			"current_balance", sender.Balance,
			"requested_amount", req.Amount,
			"reference_id", req.ReferenceID,
		)
		return nil, fmt.Errorf("insufficient balance")
	}

	// 5. Insert transaction
	txID, err := s.repo.InsertTransaction(ctx, InsertTransactionParams{
		ReferenceID: req.ReferenceID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		CustomerID:  sender.CustomerID,
	})
	if err != nil {
		logging.Logger().Errorw("failed_to_insert_transaction",
			"reference_id", req.ReferenceID,
			"error", err,
		)
		return nil, err
	}

	// 6. Journal
	journalID, err := s.repo.InsertJournal(ctx, txID)
	if err != nil {
		logging.Logger().Errorw("failed_to_insert_journal",
			"transaction_id", txID,
			"error", err,
		)
		return nil, err
	}

	// 7. Ledger
	err = s.repo.InsertLedger(ctx, InsertLedgerParams{
		JournalID: journalID,
		FromAcc:   req.FromAccount,
		ToAcc:     req.ToAccount,
		Amount:    req.Amount,
		Currency:  req.Currency,
	})
	if err != nil {
		logging.Logger().Errorw("failed_to_insert_ledger",
			"journal_id", journalID,
			"error", err,
		)
		return nil, err
	}

	// 8. Update balances
	if err := s.repo.DebitAccount(ctx, req.FromAccount, req.Amount); err != nil {
		logging.Logger().Errorw("failed_to_debit_account",
			"from_account", req.FromAccount,
			"amount", req.Amount,
			"error", err,
		)
		return nil, err
	}

	if err := s.repo.CreditAccount(ctx, req.ToAccount, req.Amount); err != nil {
		logging.Logger().Errorw("failed_to_credit_account",
			"to_account", req.ToAccount,
			"amount", req.Amount,
			"error", err,
		)
		return nil, err
	}

	// Compute balances
	newSenderBalance := sender.Balance - req.Amount

	result := &TransferResponse{
		Status:                  "success",
		TransactionID:           &txID,
		SourceAccount:           sender.AccountNo,
		DestinationAccount:      req.ToAccount,
		Amount:                  req.Amount,
		SourceBalanceAfter:      &newSenderBalance,
		DestinationBalanceAfter: nil,
		Message:                 "transfer completed successfully",
	}

	logging.Logger().Infow("transfer_success",
		"transaction_id", txID,
		"reference_id", req.ReferenceID,
		"source_account", result.SourceAccount,
		"source_balance_after", result.SourceBalanceAfter,
		"destination_account", result.DestinationAccount,
		"amount", req.Amount,
		"currency", req.Currency,
	)

	// 9. Complete transaction
	return result, s.repo.CompleteTransaction(ctx, txID)
}

func (s *Service) TransferWithLock(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	lockKey := req.ToAccount

	logging.Logger().Debugw("transfer_with_lock_initiated",
		"reference_id", req.ReferenceID,
		"lock_key", lockKey,
	)

	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// Acquire lock
	if err := s.lockManager.Lock(ctx, lockKey); err != nil {
		logging.Logger().Errorw("failed_to_acquire_transfer_lock",
			"lock_key", lockKey,
			"error", err,
		)
		return nil, err
	}
	defer s.lockManager.Unlock(lockKey)

	resultCh := make(chan transferResult, 1)

	// Run async critical section
	go func() {
		resp, err := s.Transfer(ctx, req)
		resultCh <- transferResult{
			response: resp,
			err:      err,
		}
	}()

	// Wait result or timeout
	select {
	case result := <-resultCh:
		if result.err == nil {
			logging.Logger().Infow("transfer_with_lock_completed",
				"reference_id", req.ReferenceID,
				"status", "success",
			)
		} else {
			logging.Logger().Warnw("transfer_with_lock_failed",
				"reference_id", req.ReferenceID,
				"error", result.err,
			)
		}
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

		logging.Logger().Warnw("transfer_with_lock_timeout",
			"reference_id", req.ReferenceID,
			"lock_key", lockKey,
			"timeout_duration", "4s",
		)

		return timeoutResp, fmt.Errorf("transfer timeout (exceeded 4s)")
	}
}

// func (s *Service) transferCriticalSection(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
// 	// Controlled delay (mockable in test)
// 	s.delay()

// 	// 1. Idempotency
// 	exists, err := s.repo.IsTransactionExists(ctx, req.ReferenceID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if exists {
// 		return nil, fmt.Errorf("idempotency check failed")
// 	}

// 	// 2. Lock sender
// 	sender, err := s.repo.GetSenderForUpdate(ctx, req.FromAccount)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 3. Lock receiver
// 	if err := s.repo.LockReceiver(ctx, req.ToAccount); err != nil {
// 		return nil, err
// 	}

// 	// 4. Validation
// 	if sender.Balance < req.Amount {
// 		log.Println("transfer_failed",
// 			"source_account", sender.AccountNo,
// 			"destination_account", req.ToAccount,
// 			"amount", req.Amount,
// 			"current_balance", sender.Balance,
// 		)

// 		return &TransferResponse{
// 			Status:                  "failed",
// 			TransactionID:           nil,
// 			SourceAccount:           sender.AccountNo,
// 			DestinationAccount:      req.ToAccount,
// 			Amount:                  req.Amount,
// 			SourceBalanceAfter:      &sender.Balance,
// 			DestinationBalanceAfter: nil,
// 			Message:                 "insufficient balance",
// 		}, fmt.Errorf("insufficient balance")
// 	}

// 	// 5. Insert transaction
// 	txID, err := s.repo.InsertTransaction(ctx, InsertTransactionParams{
// 		ReferenceID: req.ReferenceID,
// 		Amount:      req.Amount,
// 		Currency:    req.Currency,
// 		CustomerID:  sender.CustomerID,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 6. Journal
// 	journalID, err := s.repo.InsertJournal(ctx, txID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 7. Ledger
// 	if err := s.repo.InsertLedger(ctx, InsertLedgerParams{
// 		JournalID: journalID,
// 		FromAcc:   req.FromAccount,
// 		ToAcc:     req.ToAccount,
// 		Amount:    req.Amount,
// 		Currency:  req.Currency,
// 	}); err != nil {
// 		return nil, err
// 	}

// 	// 8. Update balances
// 	if err := s.repo.DebitAccount(ctx, req.FromAccount, req.Amount); err != nil {
// 		return nil, err
// 	}

// 	if err := s.repo.CreditAccount(ctx, req.ToAccount, req.Amount); err != nil {
// 		return nil, err
// 	}

// 	// 9. Complete
// 	if err := s.repo.CompleteTransaction(ctx, txID); err != nil {
// 		return nil, err
// 	}

// 	// Compute balances
// 	newSenderBalance := sender.Balance - req.Amount

// 	result := &TransferResponse{
// 		Status:                  "success",
// 		TransactionID:           &txID,
// 		SourceAccount:           sender.AccountNo,
// 		DestinationAccount:      req.ToAccount,
// 		Amount:                  req.Amount,
// 		SourceBalanceAfter:      &newSenderBalance,
// 		DestinationBalanceAfter: nil,
// 		Message:                 "transfer completed successfully",
// 	}

// 	log.Println("transfer_success",
// 		"source_account", result.SourceAccount,
// 		"source_balance_after", result.SourceBalanceAfter,
// 		"destination_account", result.DestinationAccount,
// 	)

// 	return result, nil
// }

func (s *Service) List(ctx context.Context, f ListFilter) ([]TransactionHistoryDTO, int, string, string, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Direction == "" {
		f.Direction = "next"
	}

	logging.Logger().Debugw("transaction_list_requested",
		"limit", f.Limit,
		"direction", f.Direction,
		"account_id", f.AccountID,
	)

	data, total, nextC, prevC, err := s.repo.List(ctx, f)
	if err != nil {
		logging.Logger().Errorw("transaction_list_failed",
			"limit", f.Limit,
			"account_id", f.AccountID,
			"error", err,
		)
		return nil, 0, "", "", err
	}

	var nextCursor, prevCursor string
	if nextC != nil {
		nextCursor, _ = pagination.EncodeCursor(*nextC)
	}
	if prevC != nil {
		prevCursor, _ = pagination.EncodeCursor(*prevC)
	}

	logging.Logger().Debugw("transaction_list_retrieved",
		"limit", f.Limit,
		"total_count", total,
		"returned_count", len(data),
	)

	return data, total, nextCursor, prevCursor, nil
}
