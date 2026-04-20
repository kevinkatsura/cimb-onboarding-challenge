package grpcserver

import (
	"context"
	"time"

	account_informationpb "proto/account_information/v1"

	"account-information-service/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountInformationServer struct {
	account_informationpb.UnimplementedAccountInformationServiceServer
	repo *repository.PostgresDatabase
}

func NewAccountInformationServer(repo *repository.PostgresDatabase) *AccountInformationServer {
	return &AccountInformationServer{repo: repo}
}

func (s *AccountInformationServer) GetBalance(ctx context.Context, req *account_informationpb.AccountRequest) (*account_informationpb.GetBalanceResponse, error) {
	if req.AccountNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "account_number is required")
	}

	balance, currency, err := s.repo.GetBalance(ctx, req.AccountNumber)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "account not found: %v", err)
	}

	return &account_informationpb.GetBalanceResponse{
		AccountNumber: req.AccountNumber,
		Balance:       balance,
		Currency:      currency,
	}, nil
}

func (s *AccountInformationServer) GetLastTransactionAsSource(ctx context.Context, req *account_informationpb.AccountRequest) (*account_informationpb.TransactionResponse, error) {
	tx, err := s.repo.GetLastTransactionAsSource(ctx, req.AccountNumber)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no transactions found for account: %v", err)
	}
	return toTransactionResponse(tx), nil
}

func (s *AccountInformationServer) GetLastTransactionAsBeneficiary(ctx context.Context, req *account_informationpb.AccountRequest) (*account_informationpb.TransactionResponse, error) {
	tx, err := s.repo.GetLastTransactionAsBeneficiary(ctx, req.AccountNumber)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no transactions found for account: %v", err)
	}
	return toTransactionResponse(tx), nil
}

func (s *AccountInformationServer) GetAverageAmountLast30Transactions(ctx context.Context, req *account_informationpb.AccountRequest) (*account_informationpb.AverageAmountResponse, error) {
	avg, currency, err := s.repo.GetAverageAmountLast30Transactions(ctx, req.AccountNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to compute average: %v", err)
	}
	return &account_informationpb.AverageAmountResponse{
		AccountNumber: req.AccountNumber,
		AverageAmount: avg,
		Currency:      currency,
	}, nil
}

func (s *AccountInformationServer) GetAccountInfo(ctx context.Context, req *account_informationpb.AccountRequest) (*account_informationpb.AccountInfoResponse, error) {
	acc, err := s.repo.GetAccountInfo(ctx, req.AccountNumber)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "account not found: %v", err)
	}
	return &account_informationpb.AccountInfoResponse{
		AccountNumber: acc.AccountNumber,
		AccountId:     acc.AccountID,
		CustomerId:    acc.CustomerID,
		ProductCode:   acc.ProductCode,
		Currency:      acc.Currency,
		Status:        acc.Status,
		CreatedAt:     acc.CreatedAt.Format(time.RFC3339),
	}, nil
}

func toTransactionResponse(tx repository.Transaction) *account_informationpb.TransactionResponse {
	return &account_informationpb.TransactionResponse{
		TransactionRef:           tx.TransactionRef,
		SourceAccountNumber:      tx.SourceAccountNumber,
		BeneficiaryAccountNumber: tx.BeneficiaryAccountNumber,
		Amount:                   tx.Amount,
		Currency:                 tx.Currency,
		CreatedAt:                tx.CreatedAt.Format(time.RFC3339),
	}
}
