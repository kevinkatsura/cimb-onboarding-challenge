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
	AccountNumber string `json:"accountNumber" example:"80202604160001"`
}

type GetAccountResponse struct {
	AccountID        string `json:"accountId" example:"uuid-acc-123"`
	CustomerID       string `json:"customerId" example:"uuid-cust-456"`
	AccountNumber    string `json:"accountNumber" example:"80202604160001"`
	ProductCode      string `json:"productCode" example:"savings"`
	Currency         string `json:"currency" example:"IDR"`
	Status           string `json:"status" example:"active"`
	AvailableBalance int64  `json:"availableBalance" example:"1500000"`
}

func (s *AccountServiceServer) GetAccount(ctx context.Context, req *GetAccountRequest) (*GetAccountResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPC.GetAccount")
	defer span.End()
	span.SetAttributes(attribute.String("account_number", req.AccountNumber))

	logging.Ctx(ctx).Infow("gRPC GetAccount called", "account_number", req.AccountNumber)

	acc, bal, err := s.svc.GetAccountByNumber(ctx, req.AccountNumber)
	if err != nil {
		logging.Ctx(ctx).Warnw("gRPC GetAccount: account not found in DB", "account_number", req.AccountNumber)
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
