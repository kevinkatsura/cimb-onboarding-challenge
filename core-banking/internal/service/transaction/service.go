package transaction

import (
	"context"
	txDomain "core-banking/internal/domain/transaction"
	"core-banking/internal/dto"
	"core-banking/internal/service"
	"core-banking/pkg/apperror"
	"core-banking/pkg/pagination"
	"core-banking/pkg/telemetry"
	"fmt"
	"math/rand"
	"time"

	"core-banking/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

var (
	transferAmountCounter metric.Int64Counter
	transferCountCounter  metric.Int64Counter
)

func init() {
	meter := otel.Meter("core-banking.transaction")
	var err error
	transferAmountCounter, err = meter.Int64Counter("core_banking_transfer_amount_total",
		metric.WithDescription("Total amount transferred successfully"),
	)
	if err != nil {
		panic(err)
	}
	transferCountCounter, err = meter.Int64Counter("core_banking_transfer_count_total",
		metric.WithDescription("Total number of successful transfers"),
	)
	if err != nil {
		panic(err)
	}
}

type DelayFunc func()

type transferResult struct {
	response *dto.TransferResponse
	err      error
}

type Service struct {
	repo        txDomain.Repository
	lockManager service.LockManager
	delay       DelayFunc
}

func NewService(repo txDomain.Repository, lockManager service.LockManager) *Service {
	return &Service{
		repo:        repo,
		lockManager: lockManager,
		delay: func() {
			time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
		},
	}
}

func (s *Service) Transfer(ctx context.Context, req dto.TransferRequest) (*dto.TransferResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "transactionService.Transfer")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrsWithIdempotency("Transfer", "transaction", "fund_transfer", req.ReferenceID, 0)...)
	span.SetAttributes(
		attribute.String("account.from", req.FromAccount),
		attribute.String("account.to", req.ToAccount),
		attribute.Int64("transfer.amount", req.Amount),
		attribute.String("transfer.currency", req.Currency),
	)

	// 1. Idempotency
	exists, err := s.repo.IsTransactionExists(ctx, req.ReferenceID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "idempotency check failed")
		return nil, apperror.NewInternal("failed to check idempotency", err)
	}
	if exists {
		span.SetStatus(codes.Error, "duplicate transaction")
		return nil, apperror.NewConflict("transaction with this reference ID already processed")
	}

	// 2. Lock sender
	sender, err := s.repo.GetSenderForUpdate(ctx, req.FromAccount)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "sender account not found")
		return nil, apperror.NewNotFound("sender account not found")
	}

	// 3. Lock receiver
	if err := s.repo.LockReceiver(ctx, req.ToAccount); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "receiver account not found")
		return nil, apperror.NewNotFound("receiver account not found")
	}

	// 4. Validation
	span.AddEvent("Validating balances")
	if sender.Balance < req.Amount {
		err := fmt.Errorf("insufficient balance: available %d, requested %d", sender.Balance, req.Amount)
		span.RecordError(err)
		span.SetStatus(codes.Error, "insufficient funds")
		return nil, apperror.NewBadRequest("insufficient balance for transfer")
	}

	// 5. Insert transaction
	txID, err := s.repo.InsertTransaction(ctx, domain.InsertTransactionParams{
		ReferenceID: req.ReferenceID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		CustomerID:  sender.CustomerID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "transaction insert failed")
		return nil, apperror.NewInternal("failed to insert transaction record", err)
	}

	// 6. Journal
	journalID, err := s.repo.InsertJournal(ctx, txID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "journal insert failed")
		return nil, apperror.NewInternal("failed to insert journal entry", err)
	}

	// 7. Ledger
	err = s.repo.InsertLedger(ctx, domain.InsertLedgerParams{
		JournalID: journalID,
		FromAcc:   req.FromAccount,
		ToAcc:     req.ToAccount,
		Amount:    req.Amount,
		Currency:  req.Currency,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ledger insert failed")
		return nil, apperror.NewInternal("failed to insert ledger entries", err)
	}

	// 8. Update balances
	if err := s.repo.DebitAccount(ctx, req.FromAccount, req.Amount); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "debit failed")
		return nil, apperror.NewInternal("failed to debit sender account", err)
	}

	if err := s.repo.CreditAccount(ctx, req.ToAccount, req.Amount); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "credit failed")
		return nil, apperror.NewInternal("failed to credit receiver account", err)
	}

	newSenderBalance := sender.Balance - req.Amount

	result := &dto.TransferResponse{
		Status:                  "success",
		TransactionID:           &txID,
		SourceAccount:           sender.AccountNo,
		DestinationAccount:      req.ToAccount,
		Amount:                  req.Amount,
		SourceBalanceAfter:      &newSenderBalance,
		DestinationBalanceAfter: nil,
		Message:                 "transfer completed successfully",
	}

	transferAmountCounter.Add(ctx, req.Amount, metric.WithAttributes(attribute.String("currency", req.Currency)))
	transferCountCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("currency", req.Currency)))

	// 9. Complete transaction
	if err := s.repo.CompleteTransaction(ctx, txID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "transaction completion failed")
		return nil, apperror.NewInternal("failed to complete transaction", err)
	}

	span.SetStatus(codes.Ok, "transfer executed successfully")
	return result, nil
}

func (s *Service) TransferWithLock(ctx context.Context, req dto.TransferRequest) (*dto.TransferResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "transactionService.TransferWithLock")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrsWithIdempotency("TransferWithLock", "transaction", "fund_transfer_locked", req.ReferenceID, 0)...)

	lockKey := req.ToAccount

	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	if err := s.lockManager.Lock(ctx, lockKey); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "lock acquisition failed")
		return nil, apperror.NewConflict("failed to acquire transfer lock")
	}
	defer s.lockManager.Unlock(lockKey)

	resultCh := make(chan transferResult, 1)

	go func() {
		resp, err := s.Transfer(ctx, req)
		resultCh <- transferResult{response: resp, err: err}
	}()

	select {
	case result := <-resultCh:
		if result.err != nil {
			span.SetStatus(codes.Error, "transfer failed")
		} else {
			span.SetStatus(codes.Ok, "transfer with lock completed")
		}
		return result.response, result.err

	case <-ctx.Done():
		timeoutResp := &dto.TransferResponse{
			Status:             "failed",
			TransactionID:      nil,
			SourceAccount:      req.FromAccount,
			DestinationAccount: req.ToAccount,
			Amount:             req.Amount,
			Message:            "transfer timeout (exceeded 4s)",
		}
		span.SetStatus(codes.Error, "transfer timeout")
		return timeoutResp, apperror.NewUnavailable("transfer timeout (exceeded 4s)")
	}
}

func (s *Service) List(ctx context.Context, f domain.TransactionListFilter) ([]dto.TransactionHistoryResponse, int, string, string, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "transactionService.List")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrs("List", "transaction", "inquiry")...)

	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Direction == "" {
		f.Direction = "next"
	}

	data, total, nextC, prevC, err := s.repo.List(ctx, f)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "transaction listing failed")
		return nil, 0, "", "", apperror.NewInternal("failed to list transactions", err)
	}

	var nextCursor, prevCursor string
	if nextC != nil {
		nextCursor, _ = pagination.EncodeCursor(*nextC)
	}
	if prevC != nil {
		prevCursor, _ = pagination.EncodeCursor(*prevC)
	}

	span.SetStatus(codes.Ok, "transactions listed")
	return data, total, nextCursor, prevCursor, nil
}
