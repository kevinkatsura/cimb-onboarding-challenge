package transaction_test

import (
	"context"
	"core-banking/internal/snap"
	"core-banking/internal/transaction"
	"core-banking/mocks"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockerForService struct {
	repo        *mocks.Repository
	lockManager *mocks.LockManager
}

const (
	StandardFee        = 2500
	SystemFeeAccountID = "00000000-0000-0000-0000-000000000009"
)

func TestTransactionService_Transfer(t *testing.T) {
	type args struct {
		ctx context.Context
		req transaction.IntrabankTransferRequest
	}

	tests := []struct {
		desc      string
		args      args
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "success - full SNAP payload",
			args: args{
				ctx: context.Background(),
				req: transaction.IntrabankTransferRequest{
					PartnerReferenceNo:   "SNAP-001",
					SourceAccountNo:      "S-001",
					BeneficiaryAccountNo: "B-001",
					FeeType:              "OUR",
					Amount:               snap.SNAPAmount{Value: "50000.00", Currency: "IDR"},
					Remark:               ptrStr("test remark"),
					OriginatorInfos: &[]transaction.OriginatorInfo{
						{OriginatorCustomerName: "Originator Name"},
					},
				},
			},
			mockSetup: func(m *MockerForService) {
				senderID := uuid.New()
				beneficiaryID := uuid.New()

				m.repo.On("GetIdempotency", mock.Anything, "SNAP-001").Return(nil, errors.New("not found"))
				m.repo.On("BeginTx", mock.Anything).Return(nil, nil)
				m.repo.On("WithTx", mock.Anything).Return(m.repo)
				m.repo.On("GetSenderForUpdate", mock.Anything, "S-001").
					Return(transaction.SenderAccount{ID: senderID, Balance: 1000000, AccountNo: "S-001"}, nil)
				m.repo.On("LockReceiver", mock.Anything, "B-001").Return(beneficiaryID, nil)
				m.repo.On("InsertTransaction", mock.Anything, mock.MatchedBy(func(p transaction.InsertTransactionParams) bool {
					return p.FeeType == "OUR" && p.Remark == "test remark"
				})).Return("00000000-0000-0000-0000-000000000005", nil)

				txUUID := uuid.MustParse("00000000-0000-0000-0000-000000000005")
				m.repo.On("InsertLedger", mock.Anything, mock.MatchedBy(func(p transaction.InsertLedgerParams) bool {
					return len(p.Entries) == 2 && p.Entries[0].Amount == 52500 && p.TransactionID == txUUID
				})).Return(nil)

				m.repo.On("InsertAccountTransaction", mock.Anything, mock.MatchedBy(func(p transaction.AccountTransaction) bool {
					return p.AccountID == senderID && p.Direction == "out"
				})).Return(nil)
				m.repo.On("InsertAccountTransaction", mock.Anything, mock.MatchedBy(func(p transaction.AccountTransaction) bool {
					return p.AccountID == beneficiaryID && p.Direction == "in"
				})).Return(nil)

				m.repo.On("DebitAccount", mock.Anything, "S-001", int64(52500)).Return(nil)
				m.repo.On("CreditAccount", mock.Anything, "B-001", int64(50000)).Return(nil)
				m.repo.On("CreditAccount", mock.Anything, SystemFeeAccountID, int64(StandardFee)).Return(nil)
				m.repo.On("CompleteTransaction", mock.Anything, "00000000-0000-0000-0000-000000000005").Return(nil)
				m.repo.On("SaveIdempotency", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			desc: "success - new transaction",
			args: args{
				ctx: context.Background(),
				req: transaction.IntrabankTransferRequest{
					PartnerReferenceNo:   "00000000-0000-0000-0000-000000000004",
					SourceAccountNo:      "A",
					BeneficiaryAccountNo: "B",
					Amount:               snap.SNAPAmount{Value: "10000.00", Currency: "IDR"},
					Remark:               ptrStr("test"),
					FeeType:              "BEN",
				},
			},
			mockSetup: func(m *MockerForService) {
				senderID := uuid.New()
				beneficiaryID := uuid.New()

				m.repo.On("GetIdempotency", mock.Anything, "00000000-0000-0000-0000-000000000004").
					Return(nil, errors.New("not found"))

				m.repo.On("BeginTx", mock.Anything).Return(nil, nil)
				m.repo.On("WithTx", mock.Anything).Return(m.repo)

				m.repo.On("GetSenderForUpdate", mock.Anything, "A").
					Return(transaction.SenderAccount{ID: senderID, Balance: 1000000, AccountNo: "A"}, nil)
				m.repo.On("LockReceiver", mock.Anything, "B").Return(beneficiaryID, nil)
				m.repo.On("InsertTransaction", mock.Anything, mock.Anything).
					Return("00000000-0000-0000-0000-000000000002", nil)

				txUUID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
				m.repo.On("InsertLedger", mock.Anything, mock.MatchedBy(func(p transaction.InsertLedgerParams) bool {
					return len(p.Entries) == 2 && p.Entries[1].Amount == 7500 && p.TransactionID == txUUID
				})).Return(nil)

				m.repo.On("InsertAccountTransaction", mock.Anything, mock.Anything).Return(nil)
				m.repo.On("InsertAccountTransaction", mock.Anything, mock.Anything).Return(nil)

				m.repo.On("DebitAccount", mock.Anything, "A", int64(10000)).Return(nil)
				m.repo.On("CreditAccount", mock.Anything, "B", int64(7500)).Return(nil)
				m.repo.On("CreditAccount", mock.Anything, SystemFeeAccountID, int64(StandardFee)).Return(nil)
				m.repo.On("CompleteTransaction", mock.Anything, "00000000-0000-0000-0000-000000000002").
					Return(nil)

				m.repo.On("SaveIdempotency", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			desc: "success - idempotent hit",
			args: args{
				ctx: context.Background(),
				req: transaction.IntrabankTransferRequest{
					PartnerReferenceNo: "ref-dup",
					Amount:             snap.SNAPAmount{Value: "100.00", Currency: "IDR"},
					FeeType:            "OUR",
				},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetIdempotency", mock.Anything, "ref-dup").
					Return(&transaction.IdempotencyKey{
						ResponseCode:    "2001700",
						ResponseMessage: "Successful",
						ResponseBody:    []byte(`{"referenceNo":"00000000-0000-0000-0000-000000000002"}`),
					}, nil)
			},
			wantErr: false,
		},
		{
			desc: "insufficient funds",
			args: args{
				ctx: context.Background(),
				req: transaction.IntrabankTransferRequest{
					PartnerReferenceNo: "ref-fail",
					Amount:             snap.SNAPAmount{Value: "1000.00", Currency: "IDR"},
					FeeType:            "OUR",
				},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetIdempotency", mock.Anything, "ref-fail").
					Return(nil, errors.New("not found"))

				m.repo.On("BeginTx", mock.Anything).Return(nil, nil)
				m.repo.On("WithTx", mock.Anything).Return(m.repo)

				m.repo.On("GetSenderForUpdate", mock.Anything, mock.Anything).
					Return(transaction.SenderAccount{Balance: 10, AccountNo: "A"}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForService{
				repo:        &mocks.Repository{},
				lockManager: &mocks.LockManager{},
			}
			tt.mockSetup(m)

			auditSvc := transaction.NewAuditService(m.repo)
			svc := transaction.NewService(m.repo, m.lockManager, auditSvc)
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
	m := &MockerForService{
		repo:        &mocks.Repository{},
		lockManager: &mocks.LockManager{},
	}

	req := transaction.IntrabankTransferRequest{
		PartnerReferenceNo:   "ref-lock",
		SourceAccountNo:      "A",
		BeneficiaryAccountNo: "B",
		Amount:               snap.SNAPAmount{Value: "10000.00", Currency: "IDR"},
		Remark:               ptrStr("lock test"),
		FeeType:              "SHA",
	}

	m.lockManager.On("Lock", mock.Anything, "B").Return(nil)
	m.lockManager.On("Unlock", "B").Return()

	m.repo.On("GetIdempotency", mock.Anything, "ref-lock").Return(nil, errors.New("not found"))

	m.repo.On("BeginTx", mock.Anything).Return(nil, nil)
	m.repo.On("WithTx", mock.Anything).Return(m.repo)

	senderID := uuid.New()
	beneficiaryID := uuid.New()
	m.repo.On("GetSenderForUpdate", mock.Anything, "A").Return(transaction.SenderAccount{ID: senderID, Balance: 100000, AccountNo: "A"}, nil)
	m.repo.On("LockReceiver", mock.Anything, "B").Return(beneficiaryID, nil)
	m.repo.On("InsertTransaction", mock.Anything, mock.Anything).Return("00000000-0000-0000-0000-000000000001", nil)

	txUUID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	m.repo.On("InsertLedger", mock.Anything, mock.MatchedBy(func(p transaction.InsertLedgerParams) bool {
		return len(p.Entries) == 2 && p.Entries[0].Amount == 11000 && p.TransactionID == txUUID
	})).Return(nil)

	m.repo.On("InsertAccountTransaction", mock.Anything, mock.Anything).Return(nil)
	m.repo.On("InsertAccountTransaction", mock.Anything, mock.Anything).Return(nil)

	m.repo.On("DebitAccount", mock.Anything, "A", int64(11000)).Return(nil)
	m.repo.On("CreditAccount", mock.Anything, "B", int64(8500)).Return(nil)
	m.repo.On("CreditAccount", mock.Anything, SystemFeeAccountID, int64(StandardFee)).Return(nil)
	m.repo.On("CompleteTransaction", mock.Anything, "00000000-0000-0000-0000-000000000001").Return(nil)

	m.repo.On("SaveIdempotency", mock.Anything, mock.Anything).Return(nil)

	auditSvc := transaction.NewAuditService(m.repo)
	svc := transaction.NewService(m.repo, m.lockManager, auditSvc)
	_, err := svc.TransferWithLock(context.Background(), req)
	assert.NoError(t, err)
	m.lockManager.AssertExpectations(t)
}

func ptrStr(s string) *string {
	return &s
}
