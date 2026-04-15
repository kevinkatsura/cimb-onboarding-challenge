package grpc

import (
	"account-issuance-service/internal/account"
	"account-issuance-service/pkg/logging"
	"account-issuance-service/pkg/telemetry"
	"context"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AccountServiceServer serves gRPC calls for account lookups.
type AccountServiceServer struct {
	svc *account.Service
}

func NewAccountServiceServer(svc *account.Service) *AccountServiceServer {
	return &AccountServiceServer{svc: svc}
}

type GetAccountRequest struct {
	AccountNumber string
}

type GetAccountResponse struct {
	AccountID        string
	CustomerID       string
	AccountNumber    string
	ProductCode      string
	Currency         string
	Status           string
	AvailableBalance int64
}

func (s *AccountServiceServer) GetAccount(ctx context.Context, req *GetAccountRequest) (*GetAccountResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPC.GetAccount")
	defer span.End()
	span.SetAttributes(attribute.String("account_number", req.AccountNumber))

	logging.Ctx(ctx).Infow("gRPC GetAccount called", "account_number", req.AccountNumber)

	acc, bal, err := s.svc.GetAccountByNumber(ctx, req.AccountNumber)
	if err != nil {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	resp := &GetAccountResponse{
		AccountID:     acc.ID.String(),
		CustomerID:    acc.CustomerID.String(),
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
