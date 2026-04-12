package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"core-banking/internal/snap"
	"core-banking/pkg/apperror"
	"core-banking/pkg/lock"
	"core-banking/pkg/pagination"
	"core-banking/pkg/telemetry"
	"core-banking/pkg/util"

	"github.com/google/uuid"
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

type transferResult struct {
	response *IntrabankTransferResponse
	err      error
}

type Service struct {
	repo         Repository
	lockManager  lock.LockManager
	auditService *AuditService
	delay        func()
}

func NewService(repo Repository, lockManager lock.LockManager, audit *AuditService) *Service {
	return &Service{
		repo:         repo,
		lockManager:  lockManager,
		auditService: audit,
		delay: func() {
			time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
		},
	}
}

func (s *Service) Transfer(ctx context.Context, req IntrabankTransferRequest) (*IntrabankTransferResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "transactionService.Transfer")
	defer span.End()

	parsedAmount, err := util.ParseSNAPAmount(req.Amount)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid amount format")
		return nil, apperror.NewBadRequest(err.Error())
	}

	span.SetAttributes(telemetry.ServiceAttrsWithIdempotency("Transfer", "transaction", "fund_transfer", req.PartnerReferenceNo, 0)...)

	// Idempotency with Response Caching
	ik, err := s.repo.GetIdempotency(ctx, req.PartnerReferenceNo)
	if err == nil && ik != nil {
		span.SetAttributes(attribute.Bool("idempotency.hit", true))
		var cachedResp IntrabankTransferResponse
		if err := json.Unmarshal(ik.ResponseBody, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	// Atomic Transaction Integrity (Service-led Transaction)
	dbTx, err := s.repo.BeginTx(ctx)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to start transaction", err)
	}
	if dbTx != nil {
		defer dbTx.Rollback()
	}

	// Obtain transactional repository
	txRepo := s.repo.WithTx(dbTx)

	// 1. Lock sender and check balance
	sender, err := txRepo.GetSenderForUpdate(ctx, req.SourceAccountNo)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewNotFound("source account not found")
	}

	if sender.Balance < parsedAmount {
		return nil, apperror.New(apperror.ErrInsufficientFunds, "Insufficient funds")
	}

	// 2. Lock receiver
	if err := txRepo.LockReceiver(ctx, req.BeneficiaryAccountNo); err != nil {
		span.RecordError(err)
		return nil, apperror.NewNotFound("beneficiary account not found")
	}

	// 3. Insert Transaction with SNAP metadata
	var transactionDate *time.Time
	if req.TransactionDate != "" {
		t, err := time.Parse(time.RFC3339, req.TransactionDate)
		if err == nil {
			transactionDate = &t
		}
	}

	// 2. Validate FeeType
	if err := validateFeeType(req.FeeType); err != nil {
		return nil, err
	}

	const (
		StandardFee        = 2500
		SystemFeeAccountID = "00000000-0000-0000-0000-000000000009"
	)

	var senderDebit, beneficiaryCredit int64
	switch req.FeeType {
	case "OUR":
		senderDebit = parsedAmount + StandardFee
		beneficiaryCredit = parsedAmount
	case "SHA":
		senderDebit = parsedAmount + 1000
		beneficiaryCredit = parsedAmount - 1500
	case "BEN":
		senderDebit = parsedAmount
		beneficiaryCredit = parsedAmount - StandardFee
	default:
		senderDebit = parsedAmount
		beneficiaryCredit = parsedAmount
	}

	if sender.Balance < senderDebit {
		return nil, apperror.New(apperror.ErrInsufficientFunds, fmt.Sprintf("insufficient funds for transfer and fee: need %d", senderDebit))
	}

	var remark string
	if req.Remark != nil {
		remark = *req.Remark
	}

	var customerRef string
	if req.CustomerReference != nil {
		customerRef = *req.CustomerReference
	}

	originatorInfos, _ := json.Marshal(req.OriginatorInfos)
	additionalInfo, _ := json.Marshal(req.AdditionalInfo)

	txID, err := txRepo.InsertTransaction(ctx, InsertTransactionParams{
		PartnerReferenceNo:   req.PartnerReferenceNo,
		Amount:               parsedAmount,
		Currency:             req.Amount.Currency,
		SourceAccountNo:      req.SourceAccountNo,
		BeneficiaryAccountNo: req.BeneficiaryAccountNo,
		CustomerReference:    customerRef,
		FeeType:              req.FeeType,
		TransactionDate:      transactionDate,
		Remark:               remark,
		OriginatorInfos:      originatorInfos,
		AdditionalInfo:       additionalInfo,
	})
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to record transaction", err)
	}

	// 4. Insert Journal
	journalID, err := txRepo.InsertJournal(ctx, txID)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to record journal", err)
	}

	// 5. Insert Ledger Entries
	journalUUID, _ := uuid.Parse(journalID)
	err = txRepo.InsertLedger(ctx, InsertLedgerParams{
		JournalID: journalUUID,
		Entries: []LedgerEntryParam{
			{AccountID: req.SourceAccountNo, EntryType: "debit", Amount: senderDebit, Currency: req.Amount.Currency},
			{AccountID: req.BeneficiaryAccountNo, EntryType: "credit", Amount: beneficiaryCredit, Currency: req.Amount.Currency},
			// {AccountID: SystemFeeAccountID, EntryType: "credit", Amount: StandardFee, Currency: req.Amount.Currency},
		},
	})
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to record ledger", err)
	}

	// 6. Update Balances
	if err := txRepo.DebitAccount(ctx, req.SourceAccountNo, senderDebit); err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to debit sender account", err)
	}
	if err := txRepo.CreditAccount(ctx, req.BeneficiaryAccountNo, beneficiaryCredit); err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to credit beneficiary account", err)
	}
	if err := txRepo.CreditAccount(ctx, SystemFeeAccountID, StandardFee); err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to credit fee income account", err)
	}

	// 7. Complete Transaction
	if err := txRepo.CompleteTransaction(ctx, txID); err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to complete transaction", err)
	}

	// 8. Commit Transaction
	if dbTx != nil {
		if err := dbTx.Commit(); err != nil {
			span.RecordError(err)
			return nil, apperror.NewInternal("failed to commit transaction", err)
		}
	}

	// 9. Prepare Response
	originators := []OriginatorInfo{}
	if req.OriginatorInfos != nil {
		originators = *req.OriginatorInfos
	}

	result := &IntrabankTransferResponse{
		ResponseCode:         "2001700",
		ResponseMessage:      "Successful",
		Amount:               req.Amount,
		BeneficiaryAccountNo: req.BeneficiaryAccountNo,
		OriginatorInfos:      originators,

		ReferenceNo:        &txID,
		PartnerReferenceNo: &req.PartnerReferenceNo,
		SourceAccountNo:    &req.SourceAccountNo,
		TransactionDate:    req.TransactionDate,
		CustomerReference:  req.CustomerReference,
		AdditionalInfo:     req.AdditionalInfo,
		TraceNo:            &txID,
	}

	// Internal Audit Logging (Improvement 3)
	s.auditService.Log(ctx, "transfer_success", "transaction", txID, nil)

	// Cache For Idempotency (Success only)
	bodyBytes, _ := json.Marshal(result)
	_ = s.repo.SaveIdempotency(ctx, &IdempotencyKey{
		ID:              uuid.New(),
		Key:             req.PartnerReferenceNo,
		ResponseCode:    result.ResponseCode,
		ResponseMessage: result.ResponseMessage,
		ResponseBody:    bodyBytes,
		CreatedAt:       time.Now(),
	})

	transferAmountCounter.Add(ctx, parsedAmount, metric.WithAttributes(attribute.String("currency", req.Amount.Currency)))
	transferCountCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("currency", req.Amount.Currency)))

	span.SetStatus(codes.Ok, "transfer executed successfully")
	return result, nil
}

func (s *Service) TransferWithLock(ctx context.Context, req IntrabankTransferRequest) (*IntrabankTransferResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "transactionService.TransferWithLock")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrsWithIdempotency("TransferWithLock", "transaction", "fund_transfer_locked", req.PartnerReferenceNo, 0)...)

	lockKey := req.BeneficiaryAccountNo
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
		}
		return result.response, result.err
	case <-ctx.Done():
		span.SetStatus(codes.Error, "transfer timeout")
		return nil, apperror.NewUnavailable("transfer timeout")
	}
}

func (s *Service) TransferStatusInquiry(ctx context.Context, req TransferStatusInquiryRequest) (*TransferStatusInquiryResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "transactionService.TransferStatusInquiry")
	defer span.End()

	tx, err := s.repo.GetTransactionByReferenceID(ctx, req.OriginalPartnerReferenceNo)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewNotFound("transaction not found")
	}

	snapStatus := "03" // pending
	if tx.Status == "completed" {
		snapStatus = "00"
	} else if tx.Status == "failed" {
		snapStatus = "01"
	}

	return &TransferStatusInquiryResponse{
		ResponseCode:               "2001700",
		ResponseMessage:            "Successful",
		OriginalPartnerReferenceNo: req.OriginalPartnerReferenceNo,
		OriginalReferenceNo:        req.OriginalReferenceNo,
		ServiceCode:                req.ServiceCode,
		Amount: snap.SNAPAmount{
			Value:    util.FormatSNAPAmount(tx.Amount),
			Currency: tx.Currency,
		},
		LatestTransactionStatus: snapStatus,
		TransactionStatusDesc:   tx.Status,
		ReferenceNumber:         tx.ID.String(),
	}, nil
}

func (s *Service) List(ctx context.Context, f TransactionListFilter) ([]TransactionHistoryResponse, int, string, string, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "transactionService.List")
	defer span.End()

	data, total, nextC, prevC, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, "", "", apperror.NewInternal("list failed", err)
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
func validateFeeType(feeType string) error {
	switch feeType {
	case "OUR", "BEN", "SHA":
		return nil
	default:
		return apperror.NewBadRequest("invalid feeType, must be 'OUR', 'BEN', or 'SHA'")
	}
}
