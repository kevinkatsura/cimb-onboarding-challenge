package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"payment-initiation-acquiring-service/pkg/apperror"
	"payment-initiation-acquiring-service/pkg/grpcclient"
	"payment-initiation-acquiring-service/pkg/logging"
	"payment-initiation-acquiring-service/pkg/messaging"
	"payment-initiation-acquiring-service/pkg/telemetry"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

const idempotencyTTL = 24 * time.Hour

type Service struct {
	repo          Repository
	accountClient *grpcclient.AccountClient
	ledgerClient  *grpcclient.LedgerClient
	producer      messaging.Producer
	redis         *redis.Client
}

func NewService(
	repo Repository,
	accountClient *grpcclient.AccountClient,
	ledgerClient *grpcclient.LedgerClient,
	producer messaging.Producer,
	redisClient *redis.Client,
) *Service {
	return &Service{
		repo:          repo,
		accountClient: accountClient,
		ledgerClient:  ledgerClient,
		producer:      producer,
		redis:         redisClient,
	}
}

// Transfer executes the intrabank transfer flow.
func (s *Service) Transfer(ctx context.Context, req TransferRequest) (*TransferResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "TransferService.Transfer")
	defer span.End()
	span.SetAttributes(
		attribute.String("partner_reference_no", req.PartnerReferenceNo),
		attribute.String("source_account", req.SourceAccount.AccountNo),
		attribute.String("beneficiary_account", req.BeneficiaryAccount.AccountNo),
	)

	// 1. Idempotency check
	idempKey := fmt.Sprintf("pias:idempotency:%s", req.PartnerReferenceNo)
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, idempKey).Bytes()
		if err == nil {
			var resp TransferResponse
			if json.Unmarshal(cached, &resp) == nil {
				logging.Ctx(ctx).Infow("idempotent response returned", "partner_ref", req.PartnerReferenceNo)
				return &resp, nil
			}
		}
	}

	// 2. Parse amount
	amount, err := strconv.ParseInt(req.Amount.Value, 10, 64)
	if err != nil {
		return nil, apperror.New(apperror.ErrInvalidFieldFormat, "invalid amount value")
	}
	if amount <= 0 {
		return nil, apperror.New(apperror.ErrInvalidFieldFormat, "amount must be positive")
	}

	// 3. Validate accounts via gRPC
	sourceAcc, err := s.accountClient.GetAccount(ctx, req.SourceAccount.AccountNo)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "source account not found")
	}
	if sourceAcc.Status != "active" {
		return nil, apperror.NewBadRequest("source account is not active")
	}

	beneficiaryAcc, err := s.accountClient.GetAccount(ctx, req.BeneficiaryAccount.AccountNo)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "beneficiary account not found")
	}
	if beneficiaryAcc.Status != "active" {
		return nil, apperror.NewBadRequest("beneficiary account is not active")
	}

	// 4. Calculate fee
	feeType := defaultStr(req.FeeType, "OUR")
	feeAmount := CalculateFee(amount)

	// 5. Check sufficient balance (from ledger)
	balance, err := s.ledgerClient.GetBalance(ctx, sourceAcc.AccountID)
	if err != nil {
		logging.Ctx(ctx).Warnw("ledger balance check failed, proceeding", "error", err)
	} else {
		totalDebit := amount
		if feeType == "OUR" || feeType == "SHA" {
			totalDebit += feeAmount
		}
		if balance < totalDebit {
			return nil, apperror.NewBadRequest("insufficient funds")
		}
	}

	// 6. Create transaction record
	txID := uuid.New()
	refNo := fmt.Sprintf("REF%s", txID.String()[:12])
	tx := &Transaction{
		ID:                 txID,
		PartnerReferenceNo: req.PartnerReferenceNo,
		ReferenceNo:        refNo,
		Type:               "intrabank",
		Status:             "pending",
		Amount:             amount,
		Currency:           req.Amount.Currency,
		FeeAmount:          feeAmount,
		FeeType:            feeType,
		Remark:             req.Remark,
	}
	if err := s.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, apperror.NewInternal("failed to create transaction", err)
	}

	detail := &TransferDetail{
		ID:                     uuid.New(),
		TransactionID:          txID,
		SourceAccountNo:        req.SourceAccount.AccountNo,
		SourceAccountName:      sourceAcc.CustomerID,
		BeneficiaryAccountNo:   req.BeneficiaryAccount.AccountNo,
		BeneficiaryAccountName: beneficiaryAcc.CustomerID,
	}
	if err := s.repo.CreateTransferDetail(ctx, detail); err != nil {
		return nil, apperror.NewInternal("failed to create transfer detail", err)
	}

	// 7. Post to ledger (double-entry)
	lines := []grpcclient.JournalLine{
		{AccountID: sourceAcc.AccountID, Debit: amount, Credit: 0, Currency: req.Amount.Currency},
		{AccountID: beneficiaryAcc.AccountID, Debit: 0, Credit: amount, Currency: req.Amount.Currency},
	}
	if feeAmount > 0 {
		lines = append(lines,
			grpcclient.JournalLine{AccountID: sourceAcc.AccountID, Debit: feeAmount, Credit: 0, Currency: req.Amount.Currency},
			grpcclient.JournalLine{AccountID: "FEE_REVENUE", Debit: 0, Credit: feeAmount, Currency: req.Amount.Currency},
		)
	}
	_, err = s.ledgerClient.CreateJournalEntry(ctx, refNo, fmt.Sprintf("Transfer %s -> %s", req.SourceAccount.AccountNo, req.BeneficiaryAccount.AccountNo), lines)
	if err != nil {
		// Mark transaction as failed
		_ = s.repo.UpdateTransactionStatus(ctx, txID, "failed")
		return nil, apperror.NewInternal("ledger posting failed", err)
	}

	// 8. Mark success
	_ = s.repo.UpdateTransactionStatus(ctx, txID, "completed")

	resp := &TransferResponse{
		PartnerReferenceNo: req.PartnerReferenceNo,
		ReferenceNo:        refNo,
		Amount:             req.Amount.Value,
		Currency:           req.Amount.Currency,
		FeeAmount:          fmt.Sprintf("%.2f", float64(feeAmount)),
		FeeType:            feeType,
		SourceAccount:      req.SourceAccount.AccountNo,
		BeneficiaryAccount: req.BeneficiaryAccount.AccountNo,
		Status:             "completed",
	}

	// 9. Cache for idempotency
	if s.redis != nil {
		data, _ := json.Marshal(resp)
		s.redis.Set(ctx, idempKey, data, idempotencyTTL)
	}

	// 10. Publish event
	s.publishTransferCompleted(ctx, tx, req)

	logging.Ctx(ctx).Infow("transfer completed",
		"transaction_id", txID.String(), "ref_no", refNo,
		"amount", amount, "fee", feeAmount)
	return resp, nil
}

// CalculateFee returns fee based on tiered schedule.
func CalculateFee(amount int64) int64 {
	switch {
	case amount <= 1_000_000:
		return 2500
	case amount <= 10_000_000:
		fee := int64(float64(amount) * 0.001)
		return fee
	default:
		fee := int64(float64(amount) * 0.0005)
		if fee > 25000 {
			return 25000
		}
		return fee
	}
}

func (s *Service) publishTransferCompleted(ctx context.Context, tx *Transaction, req TransferRequest) {
	if s.producer == nil {
		return
	}
	event := TransferCompletedEvent{
		TransactionID:      tx.ID.String(),
		PartnerReferenceNo: tx.PartnerReferenceNo,
		ReferenceNo:        tx.ReferenceNo,
		Amount:             tx.Amount,
		Currency:           tx.Currency,
		FeeAmount:          tx.FeeAmount,
		SourceAccount:      req.SourceAccount.AccountNo,
		BeneficiaryAccount: req.BeneficiaryAccount.AccountNo,
		Status:             "completed",
		CompletedAt:        time.Now().Format(time.RFC3339),
	}
	if err := s.producer.Publish(ctx, "transfer-completed", tx.ID.String(), event); err != nil {
		logging.Ctx(ctx).Errorw("failed to publish transfer-completed event", "error", err)
	}
}

func defaultStr(v, fb string) string {
	if v == "" {
		return fb
	}
	return v
}
