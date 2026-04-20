package transfer

import (
	"context"
	"github.com/google/uuid"
)

type Repository interface {
	CreateTransaction(ctx context.Context, tx *Transaction) error
	CreateTransferDetail(ctx context.Context, detail *TransferDetail) error
	GetTransactionByRef(ctx context.Context, refNo string) (*Transaction, error)
	UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status string) error
}
