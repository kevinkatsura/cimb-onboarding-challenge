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
	// 1. Idempotency
	exists, err := s.repo.IsTransactionExists(ctx, req.ReferenceID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("idempotency check failed")
	}

	// 2. Lock sender
	sender, err := s.repo.GetSenderForUpdate(ctx, req.FromAccount)
	if err != nil {
		return nil, err
	}

	// 3. Lock receiver
	if err := s.repo.LockReceiver(ctx, req.ToAccount); err != nil {
		return nil, err
	}

	// 4. Validation
	if sender.Balance < req.Amount {
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
		return nil, err
	}

	// 6. Journal
	journalID, err := s.repo.InsertJournal(ctx, txID)
	if err != nil {
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
		return nil, err
	}

	// 8. Update balances
	if err := s.repo.DebitAccount(ctx, req.FromAccount, req.Amount); err != nil {
		return nil, err
	}

	if err := s.repo.CreditAccount(ctx, req.ToAccount, req.Amount); err != nil {
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
		"source_account", result.SourceAccount,
		"source_balance_after", result.SourceBalanceAfter,
		"destination_account", result.DestinationAccount,
	)

	// 9. Complete transaction
	return result, s.repo.CompleteTransaction(ctx, txID)
}

func (s *Service) TransferWithLock(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	lockKey := req.ToAccount

	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// Acquire lock
	if err := s.lockManager.Lock(ctx, lockKey); err != nil {
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
