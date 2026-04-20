package grpc

import (
	"context"
	"fmt"

	"core-banking-system/internal/journal"
	"core-banking-system/pkg/logging"
	"core-banking-system/pkg/telemetry"

	ledgerpb "proto/ledger/v1"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LedgerServiceServer implements the gRPC LedgerService.
type LedgerServiceServer struct {
	ledgerpb.UnimplementedLedgerServiceServer
	svc *journal.Service
}

func NewLedgerServiceServer(svc *journal.Service) *LedgerServiceServer {
	return &LedgerServiceServer{svc: svc}
}

func (s *LedgerServiceServer) CreateJournalEntry(ctx context.Context, req *ledgerpb.CreateJournalEntryRequest) (*ledgerpb.CreateJournalEntryResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPC.CreateJournalEntry")
	defer span.End()
	span.SetAttributes(attribute.String("transaction_ref", req.GetTransactionRef()))

	logging.Ctx(ctx).Infow("gRPC CreateJournalEntry called", "transaction_ref", req.GetTransactionRef())

	var lines []journal.LineParam
	for _, l := range req.GetLines() {
		lines = append(lines, journal.LineParam{
			AccountID: l.GetAccountId(),
			Debit:     l.GetDebit(),
			Credit:    l.GetCredit(),
			Currency:  l.GetCurrency(),
		})
	}

	entryID, err := s.svc.CreateEntry(ctx, journal.CreateEntryParams{
		TransactionRef: req.GetTransactionRef(),
		Description:    req.GetDescription(),
		Lines:          lines,
	})
	if err != nil {
		logging.Ctx(ctx).Errorw("gRPC CreateJournalEntry failed", "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create journal entry: %v", err))
	}

	return &ledgerpb.CreateJournalEntryResponse{JournalEntryId: entryID}, nil
}

func (s *LedgerServiceServer) GetBalance(ctx context.Context, req *ledgerpb.GetBalanceRequest) (*ledgerpb.GetBalanceResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPC.GetBalance")
	defer span.End()
	span.SetAttributes(attribute.String("account_id", req.GetAccountId()))

	balance, currency, err := s.svc.GetBalance(ctx, req.GetAccountId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "account balance not found")
	}

	return &ledgerpb.GetBalanceResponse{
		AccountId:      req.GetAccountId(),
		CurrentBalance: balance,
		Currency:       currency,
	}, nil
}

func (s *LedgerServiceServer) InitializeAccount(ctx context.Context, req *ledgerpb.InitializeAccountRequest) (*ledgerpb.InitializeAccountResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPC.InitializeAccount")
	defer span.End()
	span.SetAttributes(attribute.String("account_id", req.GetAccountId()))

	logging.Ctx(ctx).Infow("gRPC InitializeAccount called", "account_id", req.GetAccountId())

	if err := s.svc.InitializeAccount(ctx, req.GetAccountId(), req.GetCurrency(), req.GetInitialBalance()); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initialize account: %v", err))
	}

	return &ledgerpb.InitializeAccountResponse{
		AccountId: req.GetAccountId(),
		Balance:   req.GetInitialBalance(),
	}, nil
}
