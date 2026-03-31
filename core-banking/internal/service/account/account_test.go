package account_test

import (
	"context"
	"core-banking/internal/dto"
	account "core-banking/internal/service/account"
	"core-banking/mocks"
	"core-banking/pkg/pagination"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"core-banking/internal/domain"

	"github.com/google/uuid"
)

type MockerForService struct {
	repo      *mocks.AccountRepositoryInterface
	accNumGen *mocks.AccountNumberGenerator
}

func TestAccountService_GetAccount(t *testing.T) {
	testCases := []struct {
		desc      string
		inputID   string
		mockSetup func(m *MockerForService)
		wantErr   bool
		expected  *domain.Account
	}{
		{
			desc:    "SUCCESS: get account",
			inputID: "00000000-0000-0000-0000-000000000001",
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetByID", "00000000-0000-0000-0000-000000000001").
					Return(&domain.Account{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")}, nil)
			},
			expected: &domain.Account{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
		},
		{
			desc:    "ERROR: repository failure",
			inputID: "00000000-0000-0000-0000-000000000001",
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetByID", "00000000-0000-0000-0000-000000000001").
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m := &MockerForService{
				repo: mocks.NewAccountRepositoryInterface(t),
			}

			tC.mockSetup(m)

			svc := account.NewService(m.repo, m.accNumGen)

			res, err := svc.GetAccount(context.Background(), tC.inputID)

			if tC.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tC.expected, res)
			}

			m.repo.AssertExpectations(t)
		})
	}
}

func TestAccountService_ListAccounts(t *testing.T) {
	testCases := []struct {
		desc      string
		filter    domain.ListFilter
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "SUCCESS: list accounts, with overridden limit, direction, and pagination",
			filter: domain.ListFilter{
				Limit: -1,
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("List", mock.Anything, mock.Anything).
					Return([]domain.Account{}, 0,
						&pagination.Cursor{ID: "next"},
						&pagination.Cursor{ID: "prev"}, nil)
			},
		},
		{
			desc: "ERROR: repository failure",
			filter: domain.ListFilter{
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
				repo:      &mocks.AccountRepositoryInterface{},
				accNumGen: &mocks.AccountNumberGenerator{},
			}

			tC.mockSetup(m)

			svc := account.NewService(m.repo, m.accNumGen)

			_, _, _, _, err := svc.ListAccounts(context.Background(), tC.filter)

			if tC.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			m.repo.AssertExpectations(t)
		})
	}
}

func TestAccountService_CreateAccount(t *testing.T) {
	type args struct {
		req dto.CreateAccountRequest
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
				req: dto.CreateAccountRequest{
					CustomerID:  "00000000-0000-0000-0000-000000000002",
					AccountType: "savings",
					Currency:    "IDR",
				},
			},
			mockSetup: func(m *MockerForService) {
				m.accNumGen.On("Generate").
					Return("100000000001", nil)

				m.repo.On("Create", mock.MatchedBy(func(acc *domain.Account) bool {
					return acc.AccountNumber == "100000000001" &&
						acc.CustomerID == uuid.MustParse("00000000-0000-0000-0000-000000000002")
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			desc: "generator error",
			args: args{
				req: dto.CreateAccountRequest{
					CustomerID: "00000000-0000-0000-0000-000000000002",
				},
			},
			mockSetup: func(m *MockerForService) {
				m.accNumGen.On("Generate").
					Return("", errors.New("gen error"))
			},
			wantErr: true,
		},
		{
			desc: "repo create error",
			args: args{
				req: dto.CreateAccountRequest{
					CustomerID: "00000000-0000-0000-0000-000000000002",
				},
			},
			mockSetup: func(m *MockerForService) {
				m.accNumGen.On("Generate").
					Return("100000000001", nil)

				m.repo.On("Create", mock.Anything).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForService{
				repo:      &mocks.AccountRepositoryInterface{},
				accNumGen: &mocks.AccountNumberGenerator{},
			}

			if tt.mockSetup != nil {
				tt.mockSetup(m)
			}

			svc := account.NewService(m.repo, m.accNumGen)

			_, err := svc.CreateAccount(context.Background(), tt.args.req)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}

			m.repo.AssertExpectations(t)
			m.accNumGen.AssertExpectations(t)
		})
	}
}

func TestAccountService_UpdateStatus(t *testing.T) {
	type args struct {
		id     string
		status string
	}

	tests := []struct {
		desc      string
		args      args
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc:    "invalid status",
			args:    args{"1", "invalid"},
			wantErr: true,
		},
		{
			desc: "success",
			args: args{"1", "active"},
			mockSetup: func(m *MockerForService) {
				m.repo.On("UpdateStatus", "1", "active").
					Return(nil)
			},
			wantErr: false,
		},
		{
			desc: "repo error",
			args: args{"1", "closed"},
			mockSetup: func(m *MockerForService) {
				m.repo.On("UpdateStatus", "1", "closed").
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForService{
				repo:      &mocks.AccountRepositoryInterface{},
				accNumGen: &mocks.AccountNumberGenerator{},
			}

			if tt.mockSetup != nil {
				tt.mockSetup(m)
			}

			svc := account.NewService(m.repo, nil)

			err := svc.UpdateStatus(context.Background(), tt.args.id, tt.args.status)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}

			m.repo.AssertExpectations(t)
		})
	}
}

func TestAccountService_DeleteAccount(t *testing.T) {
	type args struct {
		id string
	}

	tests := []struct {
		desc      string
		args      args
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "get error",
			args: args{"1"},
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetByID", "1").
					Return(&domain.Account{}, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			desc: "non-zero balance",
			args: args{"1"},
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetByID", "1").
					Return(&domain.Account{AvailableBalance: 10, Status: "closed"}, nil)
			},
			wantErr: true,
		},
		{
			desc: "not closed",
			args: args{"1"},
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetByID", "1").
					Return(&domain.Account{AvailableBalance: 0, Status: "active"}, nil)
			},
			wantErr: true,
		},
		{
			desc: "success",
			args: args{"1"},
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetByID", "1").
					Return(&domain.Account{AvailableBalance: 0, Status: "closed"}, nil)

				m.repo.On("SoftDelete", "1").
					Return(nil)
			},
			wantErr: false,
		},
		{
			desc: "delete error",
			args: args{"1"},
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetByID", "1").
					Return(&domain.Account{AvailableBalance: 0, Status: "closed"}, nil)

				m.repo.On("SoftDelete", "1").
					Return(errors.New("delete failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForService{
				repo:      &mocks.AccountRepositoryInterface{},
				accNumGen: &mocks.AccountNumberGenerator{},
			}

			if tt.mockSetup != nil {
				tt.mockSetup(m)
			}

			svc := account.NewService(m.repo, nil)

			err := svc.DeleteAccount(context.Background(), tt.args.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}

			m.repo.AssertExpectations(t)
		})
	}
}
