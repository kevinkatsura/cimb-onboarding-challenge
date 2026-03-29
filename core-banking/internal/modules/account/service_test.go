package account_test

// import (
// 	"context"
// 	"core-banking/internal/modules/account"
// 	"core-banking/mocks"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// func TestAccountService_GetAccount(t *testing.T) {
// 	type Mocker struct {
// 		repo *mocks.RepositoryInterface
// 		txm  *mocks.TxManager
// 	}

// 	testCases := []struct {
// 		desc      string
// 		inputID   string
// 		mockSetup func(m *Mocker)
// 		wantErr   bool
// 		expected  *account.Account
// 	}{
// 		{
// 			desc:    "SUCCESS: get account",
// 			inputID: "acc-1",
// 			mockSetup: func(m *Mocker) {
// 				m.repo.On("GetByID", "acc-1").
// 					Return(&account.Account{ID: "acc-1"}, nil)
// 			},
// 			expected: &account.Account{ID: "acc-1"},
// 		},
// 		{
// 			desc:    "ERROR: repository failure",
// 			inputID: "acc-1",
// 			mockSetup: func(m *Mocker) {
// 				m.repo.On("GetByID", "acc-1").
// 					Return(nil, assert.AnError)
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tC := range testCases {
// 		t.Run(tC.desc, func(t *testing.T) {
// 			m := &Mocker{
// 				repo: mocks.NewRepositoryInterface(t),
// 				txm:  mocks.NewTxManager(t),
// 			}

// 			tC.mockSetup(m)

// 			svc := account.NewService(m.repo)

// 			res, err := svc.GetAccount(context.Background(), tC.inputID)

// 			if tC.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tC.expected, res)
// 			}

// 			m.repo.AssertExpectations(t)
// 		})
// 	}
// }

// func TestAccountService_ListAccounts(t *testing.T) {
// 	type Mocker struct {
// 		repo *mocks.RepositoryInterface
// 		txm  *mocks.TxManager
// 	}

// 	testCases := []struct {
// 		desc      string
// 		filter    account.ListFilter
// 		mockSetup func(m *Mocker)
// 		wantErr   bool
// 	}{
// 		{
// 			desc: "SUCCESS: list accounts, with overridden limit, direction, and pagination",
// 			filter: account.ListFilter{
// 				Limit: -1,
// 			},
// 			mockSetup: func(m *Mocker) {
// 				m.repo.On("List", mock.Anything, mock.Anything).
// 					Return([]account.Account{}, 0, nil, nil, nil)
// 			},
// 		},
// 		{
// 			desc: "ERROR: repository failure",
// 			filter: account.ListFilter{
// 				Limit: 10,
// 			},
// 			mockSetup: func(m *Mocker) {
// 				m.repo.On("List", mock.Anything, mock.Anything).
// 					Return(nil, 0, nil, nil, assert.AnError)
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tC := range testCases {
// 		t.Run(tC.desc, func(t *testing.T) {
// 			m := &Mocker{
// 				repo: mocks.NewRepositoryInterface(t),
// 				txm:  mocks.NewTxManager(t),
// 			}

// 			tC.mockSetup(m)

// 			svc := account.NewService(m.repo, m.txm)

// 			_, _, _, _, err := svc.ListAccounts(context.Background(), tC.filter)

// 			if tC.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}

// 			m.repo.AssertExpectations(t)
// 		})
// 	}
// }
