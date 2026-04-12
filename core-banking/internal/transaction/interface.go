package transaction

import (
	"context"
)

type Interface interface {
	Transfer(ctx context.Context, req IntrabankTransferRequest) (*IntrabankTransferResponse, error)
	TransferWithLock(ctx context.Context, req IntrabankTransferRequest) (*IntrabankTransferResponse, error)
	TransferStatusInquiry(ctx context.Context, req TransferStatusInquiryRequest) (*TransferStatusInquiryResponse, error)
	List(ctx context.Context, f TransactionListFilter) ([]TransactionHistoryResponse, int, string, string, error)
}
