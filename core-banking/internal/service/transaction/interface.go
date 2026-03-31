package transaction

import (
	"context"
	"core-banking/internal/dto"

	"core-banking/internal/domain"
)

type Interface interface {
	Transfer(ctx context.Context, req dto.TransferRequest) (*dto.TransferResponse, error)
	TransferWithLock(ctx context.Context, req dto.TransferRequest) (*dto.TransferResponse, error)
	List(ctx context.Context, f domain.TransactionListFilter) ([]dto.TransactionHistoryResponse, int, string, string, error)
}
