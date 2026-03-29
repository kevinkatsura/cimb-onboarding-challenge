package transaction_test

import (
	"core-banking/internal/modules/transaction"
	"core-banking/mocks"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockerForHandler struct {
	service *mocks.TransactionServiceInterface
}

func TestTransactionHandler_Transfer(t *testing.T) {
	type args struct {
		body string
	}

	tests := []struct {
		desc      string
		args      args
		mockSetup func(m *MockerForHandler)
		wantErr   bool
	}{
		{
			desc: "invalid json",
			args: args{
				body: `invalid-json`,
			},
			mockSetup: func(m *MockerForHandler) {},
			wantErr:   true,
		},
		{
			desc: "service error",
			args: args{
				body: `{"reference_id":"ref"}`,
			},
			mockSetup: func(m *MockerForHandler) {
				m.service.On("Transfer", mock.Anything, mock.Anything).
					Return(nil, errors.New("svc error"))
			},
			wantErr: true,
		},
		{
			desc: "success",
			args: args{
				body: `{"reference_id":"ref"}`,
			},
			mockSetup: func(m *MockerForHandler) {
				resp := &transaction.TransferResponse{Status: "success"}
				m.service.On("Transfer", mock.Anything, mock.Anything).
					Return(resp, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForHandler{
				service: mocks.NewTransactionServiceInterface(t),
			}
			tt.mockSetup(m)
			h := transaction.NewHandler(m.service)

			req := httptest.NewRequest(http.MethodPost, "/transfer", strings.NewReader(tt.args.body))
			rec := httptest.NewRecorder()

			h.Transfer(rec, req)

			if tt.wantErr {
				assert.NotEqual(t, http.StatusOK, rec.Code)
			} else {
				assert.Equal(t, http.StatusOK, rec.Code)
			}

			m.service.AssertExpectations(t)
		})
	}
}

func TestTransactionHandler_TransferWithLock(t *testing.T) {
	type args struct {
		body string
	}

	tests := []struct {
		desc      string
		args      args
		mockSetup func(m *MockerForHandler)
		wantErr   bool
	}{
		{
			desc: "invalid json",
			args: args{
				body: `bad-json`,
			},
			mockSetup: func(m *MockerForHandler) {},
			wantErr:   true,
		},
		{
			desc: "service error",
			args: args{
				body: `{"reference_id":"ref"}`,
			},
			mockSetup: func(m *MockerForHandler) {
				m.service.On("TransferWithLock", mock.Anything, mock.Anything).
					Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
		{
			desc: "success",
			args: args{
				body: `{"reference_id":"ref"}`,
			},
			mockSetup: func(m *MockerForHandler) {
				resp := &transaction.TransferResponse{Status: "success"}
				m.service.On("TransferWithLock", mock.Anything, mock.Anything).
					Return(resp, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForHandler{
				service: mocks.NewTransactionServiceInterface(t),
			}
			tt.mockSetup(m)
			h := transaction.NewHandler(m.service)

			req := httptest.NewRequest(http.MethodPost, "/transfer-lock", strings.NewReader(tt.args.body))
			rec := httptest.NewRecorder()

			h.TransferWithLock(rec, req)

			if tt.wantErr {
				assert.NotEqual(t, http.StatusOK, rec.Code)
			} else {
				assert.Equal(t, http.StatusOK, rec.Code)
			}

			m.service.AssertExpectations(t)
		})
	}
}

// func TestTransactionHandler_List(t *testing.T) {
// 	type args struct {
// 		url string
// 	}

// 	tests := []struct {
// 		desc      string
// 		args      args
// 		mockSetup func(m *MockerForHandler)
// 		wantErr   bool
// 	}{
// 		{
// 			desc: "invalid cursor",
// 			args: args{
// 				url: "/accounts/1?cursor=invalid",
// 			},
// 			mockSetup: func(m *MockerForHandler) {},
// 			wantErr:   true,
// 		},
// 		{
// 			desc: "service error",
// 			args: args{
// 				url: "/accounts/1",
// 			},
// 			mockSetup: func(m *MockerForHandler) {
// 				m.service.On("List", mock.Anything, mock.Anything).
// 					Return(nil, 0, "", "", errors.New("error"))
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			desc: "success with filters",
// 			args: args{
// 				url: "/accounts/1?limit=10&type=debit&status=success",
// 			},
// 			mockSetup: func(m *MockerForHandler) {
// 				m.service.On("List", mock.Anything, mock.MatchedBy(func(f transaction.ListFilter) bool {
// 					return f.Limit == 10 &&
// 						f.Type != nil &&
// 						f.Status != nil &&
// 						*f.Type == "debit" &&
// 						*f.Status == "success"
// 				})).Return([]interface{}{}, 1, "next", "prev", nil)
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.desc, func(t *testing.T) {
// 			m := &MockerForHandler{
// 				service: mocks.NewTransactionServiceInterface(t),
// 			}
// 			tt.mockSetup(m)
// 			h := transaction.NewHandler(m.service)

// 			req := httptest.NewRequest(http.MethodGet, tt.args.url, nil)

// 			// inject path param
// 			req.SetPathValue("id", "1")

// 			rec := httptest.NewRecorder()

// 			h.ListByAccount(rec, req)

// 			if tt.wantErr {
// 				assert.NotEqual(t, http.StatusOK, rec.Code)
// 			} else {
// 				assert.Equal(t, http.StatusOK, rec.Code)
// 			}

// 			m.service.AssertExpectations(t)
// 		})
// 	}
// }

// func TestHandler_ListAll(t *testing.T) {
// 	type args struct {
// 		url string
// 	}

// 	tests := []struct {
// 		desc      string
// 		args      args
// 		mockSetup func(m *MockerForHandler)
// 		wantErr   bool
// 	}{
// 		{
// 			desc: "success default limit",
// 			args: args{
// 				url: "/transactions",
// 			},
// 			mockSetup: func(m *MockerForHandler) {
// 				m.service.On("List", mock.Anything, mock.MatchedBy(func(f transaction.ListFilter) bool {
// 					return f.AccountID == nil && f.Limit == 20
// 				})).Return([]interface{}{}, 0, "", "", nil)
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.desc, func(t *testing.T) {
// 			m := &MockerForHandler{
// 				service: mocks.NewTransactionServiceInterface(t),
// 			}
// 			tt.mockSetup(m)
// 			h := transaction.NewHandler(m.service)

// 			req := httptest.NewRequest(http.MethodGet, tt.args.url, nil)
// 			rec := httptest.NewRecorder()

// 			h.ListAll(rec, req)

// 			if tt.wantErr {
// 				assert.NotEqual(t, http.StatusOK, rec.Code)
// 			} else {
// 				assert.Equal(t, http.StatusOK, rec.Code)
// 			}

// 			m.service.AssertExpectations(t)
// 		})
// 	}
// }
