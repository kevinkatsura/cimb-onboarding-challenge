package grpc

import (
	"context"
	"fmt"

	"core-banking-system/internal/journal"
	"core-banking-system/pkg/logging"
	"core-banking-system/pkg/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LedgerServiceServer implements the gRPC LedgerService.
// Proto-generated interface is not imported to avoid circular deps; we use
// manual registration via RegisterLedgerServiceServer in cmd/server.
type LedgerServiceServer struct {
	svc *journal.Service
}

func NewLedgerServiceServer(svc *journal.Service) *LedgerServiceServer {
	return &LedgerServiceServer{svc: svc}
}

type CreateJournalEntryRequest struct {
	TransactionRef string               `json:"transactionRef" example:"TX-123456"`
	Description    string               `json:"description" example:"Transfer to Kevin"`
	Lines          []JournalLineRequest `json:"lines"`
}

type JournalLineRequest struct {
	AccountID string `json:"accountId" example:"ACC001"`
	Debit     int64  `json:"debit" example:"100000"`
	Credit    int64  `json:"credit" example:"0"`
	Currency  string `json:"currency" example:"IDR"`
}

type CreateJournalEntryResponse struct {
	JournalEntryID string `json:"journalEntryId" example:"uuid-1234"`
}

type GetBalanceRequest struct {
	AccountID string `json:"accountId" example:"ACC001"`
}

type GetBalanceResponse struct {
	AccountID      string `json:"accountId" example:"ACC001"`
	CurrentBalance int64  `json:"currentBalance" example:"1000000"`
	Currency       string `json:"currency" example:"IDR"`
}

type InitializeAccountRequest struct {
	AccountID      string
	Currency       string
	InitialBalance int64
}

type InitializeAccountResponse struct {
	AccountID string
	Balance   int64
}

func (s *LedgerServiceServer) CreateJournalEntry(ctx context.Context, req *CreateJournalEntryRequest) (*CreateJournalEntryResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPC.CreateJournalEntry")
	defer span.End()
	span.SetAttributes(attribute.String("transaction_ref", req.TransactionRef))

	logging.Ctx(ctx).Infow("gRPC CreateJournalEntry called", "transaction_ref", req.TransactionRef)

	var lines []journal.LineParam
	for _, l := range req.Lines {
		lines = append(lines, journal.LineParam{
			AccountID: l.AccountID,
			Debit:     l.Debit,
			Credit:    l.Credit,
			Currency:  l.Currency,
		})
	}

	entryID, err := s.svc.CreateEntry(ctx, journal.CreateEntryParams{
		TransactionRef: req.TransactionRef,
		Description:    req.Description,
		Lines:          lines,
	})
	if err != nil {
		logging.Ctx(ctx).Errorw("gRPC CreateJournalEntry failed", "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create journal entry: %v", err))
	}

	return &CreateJournalEntryResponse{JournalEntryID: entryID}, nil
}

func (s *LedgerServiceServer) GetBalance(ctx context.Context, req *GetBalanceRequest) (*GetBalanceResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPC.GetBalance")
	defer span.End()
	span.SetAttributes(attribute.String("account_id", req.AccountID))

	balance, currency, err := s.svc.GetBalance(ctx, req.AccountID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "account balance not found")
	}

	return &GetBalanceResponse{
		AccountID:      req.AccountID,
		CurrentBalance: balance,
		Currency:       currency,
	}, nil
}

func (s *LedgerServiceServer) InitializeAccount(ctx context.Context, req *InitializeAccountRequest) (*InitializeAccountResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPC.InitializeAccount")
	defer span.End()
	span.SetAttributes(attribute.String("account_id", req.AccountID))

	logging.Ctx(ctx).Infow("gRPC InitializeAccount called", "account_id", req.AccountID)

	if err := s.svc.InitializeAccount(ctx, req.AccountID, req.Currency, req.InitialBalance); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initialize account: %v", err))
	}

	return &InitializeAccountResponse{
		AccountID: req.AccountID,
		Balance:   req.InitialBalance,
	}, nil
}
