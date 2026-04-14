package account

import (
	"context"
	"testing"

	"core-banking/pkg/pagination"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockerForService struct {
	repo      *MockRepository
	accNumGen *MockAccountNumberGenerator
}

type MockAccountNumberGenerator struct {
	mock.Mock
}

func (m *MockAccountNumberGenerator) Generate() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestAccountService_GetAccount(t *testing.T) {
	testCases := []struct {
		desc      string
		inputID   string
		mockSetup func(m *MockerForService)
		wantErr   bool
		expected  *Account
	}{
		{
			desc:    "SUCCESS: get account",
			inputID: "00000000-0000-0000-0000-000000000001",
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetByID", mock.Anything, "00000000-0000-0000-0000-000000000001").
					Return(&Account{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")}, nil)
			},
			expected: &Account{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001")},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m := &MockerForService{
				repo:      NewMockRepository(t),
				accNumGen: &MockAccountNumberGenerator{},
			}

			tC.mockSetup(m)

			svc := NewService(m.repo, m.accNumGen, nil)

			res, err := svc.GetAccount(context.Background(), tC.inputID)

			if tC.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tC.expected, res)
			}
		})
	}
}

func TestAccountService_ListAccounts(t *testing.T) {
	testCases := []struct {
		desc      string
		filter    ListFilter
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "SUCCESS: list accounts",
			filter: ListFilter{
				Limit: 10,
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("List", mock.Anything, mock.Anything).
					Return([]Account{}, 0,
						&pagination.Cursor{ID: "next"},
						&pagination.Cursor{ID: "prev"}, nil)
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m := &MockerForService{
				repo:      NewMockRepository(t),
				accNumGen: &MockAccountNumberGenerator{},
			}

			tC.mockSetup(m)

			svc := NewService(m.repo, m.accNumGen, nil)

			_, _, _, _, err := svc.ListAccounts(context.Background(), tC.filter)

			if tC.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAccountService_CreateAccount(t *testing.T) {
	type args struct {
		req CreateAccountRequest
	}

	tests := []struct {
		desc      string
		args      args
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "success - existing customer",
			args: args{
				req: CreateAccountRequest{
					CustomerID:  "00000000-0000-0000-0000-000000000002",
					ProductCode: "savings",
					Currency:    "IDR",
				},
			},
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetCustomerByID", mock.Anything, "00000000-0000-0000-0000-000000000002").
					Return(&Customer{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002")}, nil)

				m.repo.On("UpdateCustomer", mock.Anything, mock.Anything).Return(nil)

				m.accNumGen.On("Generate").
					Return("100000000001", nil)

				m.repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForService{
				repo:      NewMockRepository(t),
				accNumGen: &MockAccountNumberGenerator{},
			}

			if tt.mockSetup != nil {
				tt.mockSetup(m)
			}

			svc := NewService(m.repo, m.accNumGen, nil)

			_, err := svc.CreateAccount(context.Background(), tt.args.req)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAccountService_UpdateStatus(t *testing.T) {
	tests := []struct {
		desc      string
		id        string
		status    string
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc:    "invalid status",
			id:      "1",
			status:  "invalid",
			wantErr: true,
		},
		{
			desc:   "success",
			id:     "1",
			status: "active",
			mockSetup: func(m *MockerForService) {
				m.repo.On("UpdateStatus", mock.Anything, "1", "active").Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForService{
				repo: NewMockRepository(t),
			}

			if tt.mockSetup != nil {
				tt.mockSetup(m)
			}

			svc := NewService(m.repo, nil, nil)

			err := svc.UpdateStatus(context.Background(), tt.id, tt.status)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAccountService_DeleteAccount(t *testing.T) {
	tests := []struct {
		desc      string
		id        string
		mockSetup func(m *MockerForService)
		wantErr   bool
	}{
		{
			desc: "success",
			id:   "1",
			mockSetup: func(m *MockerForService) {
				m.repo.On("GetByID", mock.Anything, "1").
					Return(&Account{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Status: "closed"}, nil)

				m.repo.On("GetBalance", mock.Anything, "1").
					Return(&AccountBalance{AvailableBalance: 0}, nil)

				m.repo.On("SoftDelete", mock.Anything, "1").Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForService{
				repo: NewMockRepository(t),
			}

			if tt.mockSetup != nil {
				tt.mockSetup(m)
			}

			svc := NewService(m.repo, nil, nil)

			err := svc.DeleteAccount(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
