package grpc

import (
	"context"

	"account-issuance-service/internal/account"
	"account-issuance-service/pkg/logging"
	"account-issuance-service/pkg/telemetry"

	accountpb "proto/account/v1"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AccountServiceServer serves gRPC calls for account lookups.
type AccountServiceServer struct {
	accountpb.UnimplementedAccountServiceServer
	svc *account.Service
}

func NewAccountServiceServer(svc *account.Service) *AccountServiceServer {
	return &AccountServiceServer{svc: svc}
}

func (s *AccountServiceServer) GetAccount(ctx context.Context, req *accountpb.GetAccountRequest) (*accountpb.GetAccountResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPC.GetAccount")
	defer span.End()
	span.SetAttributes(attribute.String("account_number", req.GetAccountNumber()))

	logging.Ctx(ctx).Infow("gRPC GetAccount called", "account_number", req.GetAccountNumber())

	acc, bal, err := s.svc.GetAccountByNumber(ctx, req.GetAccountNumber())
	if err != nil {
		logging.Ctx(ctx).Warnw("gRPC GetAccount: account not found in DB", "account_number", req.GetAccountNumber())
		return nil, status.Error(codes.NotFound, "account not found")
	}

	resp := &accountpb.GetAccountResponse{
		AccountId:     acc.ID.String(),
		CustomerId:    acc.CustomerID.String(),
		AccountNumber: acc.AccountNumber,
		ProductCode:   acc.ProductCode,
		Currency:      acc.Currency,
		Status:        acc.Status,
	}
	if bal != nil {
		resp.AvailableBalance = bal.Available
	}
	return resp, nil
}
