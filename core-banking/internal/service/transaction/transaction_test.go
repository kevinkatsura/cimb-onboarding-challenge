package transaction_test

import (
	"context"
	"core-banking/internal/dto"
	transaction "core-banking/internal/service/transaction"
	"core-banking/mocks"
	"core-banking/pkg/pagination"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"core-banking/internal/domain"
)

type MockerForService struct {
	repo        *mocks.TransactionRepositoryInterface
	lockManager *mocks.LockManager
}

func TestTransactionService_Transfer(t *testing.T) {
	type args struct {
		ctx context.Context
		req dto.TransferRequest
	}

	tests := []struct {
		desc      string
		args      args
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "success",
			args: args{
				ctx: context.Background(),
				req: dto.TransferRequest{
					ReferenceID: "ref-1",
					FromAccount: "A",
					ToAccount:   "B",
					Amount:      100,
					Currency:    "IDR",
				},
			},
			mockSetup: func(m *MockerForService) {
				sender := domain.SenderAccount{
					AccountNo:  "A",
					Balance:    200,
					CustomerID: "cust-1",
				}

				m.repo.On("IsTransactionExists", mock.Anything, "ref-1").
					Return(false, nil)

				m.repo.On("GetSenderForUpdate", mock.Anything, "A").
					Return(sender, nil)

				m.repo.On("LockReceiver", mock.Anything, "B").
					Return(nil)

				m.repo.On("InsertTransaction", mock.Anything, mock.Anything).
					Return("tx-1", nil)

				m.repo.On("InsertJournal", mock.Anything, "tx-1").
					Return("journal-1", nil)

				m.repo.On("InsertLedger", mock.Anything, mock.Anything).
					Return(nil)

				m.repo.On("DebitAccount", mock.Anything, "A", int64(100)).
					Return(nil)

				m.repo.On("CreditAccount", mock.Anything, "B", int64(100)).
					Return(nil)

				m.repo.On("CompleteTransaction", mock.Anything, "tx-1").
					Return(nil)
			},
			wantErr: false,
		},
		{
			desc: "idempotency error",
			args: args{
				ctx: context.Background(),
				req: dto.TransferRequest{ReferenceID: "ref-dup"},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("IsTransactionExists", mock.Anything, "ref-dup").
					Return(true, nil)
			},
			wantErr: true,
		},
		{
			desc: "sender fetch error",
			args: args{
				ctx: context.Background(),
				req: dto.TransferRequest{ReferenceID: "ref"},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("IsTransactionExists", mock.Anything, "ref").
					Return(false, nil)

				m.repo.On("GetSenderForUpdate", mock.Anything, mock.Anything).
					Return(domain.SenderAccount{}, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			desc: "receiver lock error",
			args: args{
				ctx: context.Background(),
				req: dto.TransferRequest{
					ReferenceID: "ref",
					FromAccount: "A",
					ToAccount:   "B",
				},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("IsTransactionExists", mock.Anything, "ref").
					Return(false, nil)

				m.repo.On("GetSenderForUpdate", mock.Anything, "A").
					Return(domain.SenderAccount{Balance: 200}, nil)

				m.repo.On("LockReceiver", mock.Anything, "B").
					Return(errors.New("lock error"))
			},
			wantErr: true,
		},
		{
			desc: "insufficient balance",
			args: args{
				ctx: context.Background(),
				req: dto.TransferRequest{
					ReferenceID: "ref",
					FromAccount: "A",
					ToAccount:   "B",
					Amount:      500,
				},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("IsTransactionExists", mock.Anything, "ref").
					Return(false, nil)

				m.repo.On("GetSenderForUpdate", mock.Anything, "A").
					Return(domain.SenderAccount{Balance: 100}, nil)

				m.repo.On("LockReceiver", mock.Anything, "B").
					Return(nil)
			},
			wantErr: true,
		},
		{
			desc: "insert transaction error",
			args: args{
				ctx: context.Background(),
				req: dto.TransferRequest{
					ReferenceID: "ref",
					FromAccount: "A",
					ToAccount:   "B",
					Amount:      50,
				},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("IsTransactionExists", mock.Anything, "ref").
					Return(false, nil)

				m.repo.On("GetSenderForUpdate", mock.Anything, "A").
					Return(domain.SenderAccount{Balance: 100}, nil)

				m.repo.On("LockReceiver", mock.Anything, "B").
					Return(nil)

				m.repo.On("InsertTransaction", mock.Anything, mock.Anything).
					Return("", errors.New("insert error"))
			},
			wantErr: true,
		},
		{
			desc: "complete transaction error",
			args: args{
				ctx: context.Background(),
				req: dto.TransferRequest{
					ReferenceID: "ref",
					FromAccount: "A",
					ToAccount:   "B",
					Amount:      50,
				},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("IsTransactionExists", mock.Anything, "ref").
					Return(false, nil)

				m.repo.On("GetSenderForUpdate", mock.Anything, "A").
					Return(domain.SenderAccount{Balance: 100}, nil)

				m.repo.On("LockReceiver", mock.Anything, "B").
					Return(nil)

				m.repo.On("InsertTransaction", mock.Anything, mock.Anything).
					Return("tx-1", nil)

				m.repo.On("InsertJournal", mock.Anything, "tx-1").
					Return("j-1", nil)

				m.repo.On("InsertLedger", mock.Anything, mock.Anything).
					Return(nil)

				m.repo.On("DebitAccount", mock.Anything, "A", int64(50)).
					Return(nil)

				m.repo.On("CreditAccount", mock.Anything, "B", int64(50)).
					Return(nil)

				m.repo.On("CompleteTransaction", mock.Anything, "tx-1").
					Return(errors.New("commit error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForService{
				repo:        &mocks.TransactionRepositoryInterface{},
				lockManager: &mocks.LockManager{},
			}

			tt.mockSetup(m)

			svc := transaction.NewService(m.repo, m.lockManager)

			_, err := svc.Transfer(tt.args.ctx, tt.args.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			m.repo.AssertExpectations(t)
		})
	}
}

func TestService_TransferWithLock(t *testing.T) {
	type args struct {
		ctx context.Context
		req dto.TransferRequest
	}

	tests := []struct {
		desc      string
		args      args
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "success",
			args: args{
				ctx: context.Background(),
				req: dto.TransferRequest{
					ReferenceID: "ref",
					FromAccount: "A",
					ToAccount:   "B",
					Amount:      10,
				},
			},
			mockSetup: func(m *MockerForService) {
				m.lockManager.On("Lock", mock.Anything, "B").
					Return(nil)

				m.lockManager.On("Unlock", "B").
					Return()

				// reuse Transfer success mocks
				m.repo.On("IsTransactionExists", mock.Anything, "ref").
					Return(false, nil)

				m.repo.On("GetSenderForUpdate", mock.Anything, "A").
					Return(domain.SenderAccount{Balance: 100}, nil)

				m.repo.On("LockReceiver", mock.Anything, "B").
					Return(nil)

				m.repo.On("InsertTransaction", mock.Anything, mock.Anything).
					Return("tx", nil)

				m.repo.On("InsertJournal", mock.Anything, "tx").
					Return("j", nil)

				m.repo.On("InsertLedger", mock.Anything, mock.Anything).
					Return(nil)

				m.repo.On("DebitAccount", mock.Anything, "A", int64(10)).
					Return(nil)

				m.repo.On("CreditAccount", mock.Anything, "B", int64(10)).
					Return(nil)

				m.repo.On("CompleteTransaction", mock.Anything, "tx").
					Return(nil)
			},
			wantErr: false,
		},
		{
			desc: "lock failure",
			args: args{
				ctx: context.Background(),
				req: dto.TransferRequest{ToAccount: "B"},
			},
			mockSetup: func(m *MockerForService) {
				m.lockManager.On("Lock", mock.Anything, "B").
					Return(errors.New("lock failed"))
			},
			wantErr: true,
		},
		{
			desc: "timeout",
			args: args{
				ctx: context.Background(),
				req: dto.TransferRequest{
					ReferenceID: "ref",
					FromAccount: "A",
					ToAccount:   "B",
				},
			},
			mockSetup: func(m *MockerForService) {
				m.lockManager.On("Lock", mock.Anything, "B").
					Return(nil)

				m.lockManager.On("Unlock", "B").
					Return()

				// simulate slow Transfer
				m.repo.On("IsTransactionExists", mock.Anything, "ref").
					Run(func(args mock.Arguments) {
						time.Sleep(5 * time.Second)
					}).
					Return(false, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForService{
				repo:        &mocks.TransactionRepositoryInterface{},
				lockManager: &mocks.LockManager{},
			}

			tt.mockSetup(m)

			svc := transaction.NewService(m.repo, m.lockManager)

			_, err := svc.TransferWithLock(tt.args.ctx, tt.args.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			m.repo.AssertExpectations(t)
			m.lockManager.AssertExpectations(t)
		})
	}
}

func TestTransactionService_List(t *testing.T) {
	testCases := []struct {
		desc      string
		filter    domain.TransactionListFilter
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "SUCCESS: list accounts, with overridden limit, direction, and pagination",
			filter: domain.TransactionListFilter{
				Limit: -1,
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("List", mock.Anything, mock.Anything).
					Return([]dto.TransactionHistoryResponse{}, 0,
						&pagination.Cursor{ID: "next"},
						&pagination.Cursor{ID: "prev"}, nil)
			},
		},
		{
			desc: "ERROR: repository failure",
			filter: domain.TransactionListFilter{
				Limit: 10,
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("List", mock.Anything, mock.Anything).
					Return(nil, 0, nil, nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m := &MockerForService{
				repo:        &mocks.TransactionRepositoryInterface{},
				lockManager: &mocks.LockManager{},
			}

			tC.mockSetup(m)

			svc := transaction.NewService(m.repo, m.lockManager)

			_, _, _, _, err := svc.List(context.Background(), tC.filter)

			if tC.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			m.repo.AssertExpectations(t)
		})
	}
}
