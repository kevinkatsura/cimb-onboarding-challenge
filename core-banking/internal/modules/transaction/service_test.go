package transaction_test

import (
	"context"
	"core-banking/internal/modules/transaction"
	"core-banking/internal/pkg/pagination"
	"core-banking/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockerForService struct {
	repo *mocks.TransactionRepositoryInterface
}

func TestTransactionService_Transfer(t *testing.T) {
	type args struct {
		ctx context.Context
		req transaction.TransferRequest
	}

	tests := []struct {
		desc      string
		args      args
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "idempotent - already exists",
			args: args{
				ctx: context.Background(),
				req: transaction.TransferRequest{ReferenceID: "ref-1"},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("IsTransactionExists", mock.Anything, "ref-1").
					Return(true, nil).Once()
			},
			wantErr: false,
		},
		{
			desc: "insufficient balance",
			args: args{
				ctx: context.Background(),
				req: transaction.TransferRequest{
					ReferenceID: "ref-2",
					FromAccount: "A",
					ToAccount:   "B",
					Amount:      100,
				},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("IsTransactionExists", mock.Anything, "ref-2").
					Return(false, nil)

				m.repo.On("GetSenderForUpdate", mock.Anything, "A").
					Return(transaction.SenderAccount{Balance: 50}, nil)

				m.repo.On("LockReceiver", mock.Anything, "B").
					Return(nil)
			},
			wantErr: true,
		},
		{
			desc: "success",
			args: args{
				ctx: context.Background(),
				req: transaction.TransferRequest{
					ReferenceID: "ref-3",
					FromAccount: "A",
					ToAccount:   "B",
					Amount:      100,
					Currency:    "IDR",
				},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("IsTransactionExists", mock.Anything, "ref-3").
					Return(false, nil)

				m.repo.On("GetSenderForUpdate", mock.Anything, "A").
					Return(transaction.SenderAccount{
						Balance:    200,
						CustomerID: "cust-1",
					}, nil)

				m.repo.On("LockReceiver", mock.Anything, "B").
					Return(nil)

				m.repo.On("InsertTransaction", mock.Anything, mock.Anything).
					Return("tx-1", nil)

				m.repo.On("InsertJournal", mock.Anything, "tx-1").
					Return("j-1", nil)

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
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForService{
				repo: mocks.NewTransactionRepositoryInterface(t),
			}

			tt.mockSetup(m)

			svc := transaction.NewService(m.repo)

			err := svc.Transfer(tt.args.ctx, tt.args.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}

			m.repo.AssertExpectations(t)
		})
	}
}

// func TestTransactionService_TransferWithLock(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 		req transaction.TransferRequest
// 	}

// 	tests := []struct {
// 		desc      string
// 		args      args
// 		mockSetup func(m *MockerForService)
// 		wantErr   bool
// 	}{
// 		{
// 			desc: "timeout",
// 			args: args{
// 				ctx: context.Background(),
// 				req: transaction.TransferRequest{
// 					ToAccount: "B",
// 				},
// 			},
// 			mockSetup: func(m *MockerForService) {},
// 			wantErr:   true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.desc, func(t *testing.T) {
// 			m := &MockerForService{
// 				repo: mocks.NewTransactionRepositoryInterface(t),
// 			}

// 			tt.mockSetup(m)

// 			svc := transaction.NewService(m.repo)

// 			_, err := svc.TransferWithLock(tt.args.ctx, tt.args.req)

// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("err mismatch")
// 			}
// 		})
// 	}
// }

func TestTransactionService_List(t *testing.T) {
	testCases := []struct {
		desc      string
		filter    transaction.ListFilter
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "SUCCESS: list accounts, with overridden limit, direction, and pagination",
			filter: transaction.ListFilter{
				Limit: -1,
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("List", mock.Anything, mock.Anything).
					Return([]transaction.TransactionHistoryDTO{}, 0,
						&pagination.Cursor{ID: "next"},
						&pagination.Cursor{ID: "prev"}, nil)
			},
		},
		{
			desc: "ERROR: repository failure",
			filter: transaction.ListFilter{
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
				repo: &mocks.TransactionRepositoryInterface{},
			}

			tC.mockSetup(m)

			svc := transaction.NewService(m.repo)

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
