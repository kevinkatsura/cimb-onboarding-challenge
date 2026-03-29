package account_test

import (
	"core-banking/internal/modules/account"
	"core-banking/mocks"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockerForHandler struct {
	service *mocks.ServiceInterface
}

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		desc      string
		body      string
		mockSetup func(m *MockerForHandler)
		wantCode  int
	}{
		{
			desc:      "invalid json",
			body:      `{invalid}`,
			mockSetup: func(m *MockerForHandler) {},
			wantCode:  http.StatusBadRequest,
		},
		{
			desc: "service error",
			body: `{"customer_id":"1"}`,
			mockSetup: func(m *MockerForHandler) {
				m.service.On("CreateAccount", mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("err"))
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			desc: "success",
			body: `{"customer_id":"1"}`,
			mockSetup: func(m *MockerForHandler) {
				m.service.On("CreateAccount", mock.Anything, mock.Anything).
					Return(&account.Account{}, nil)
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			m := &MockerForHandler{
				service: mocks.NewServiceInterface(t),
			}
			tc.mockSetup(m)
			h := account.NewHandler(m.service)

			req := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			h.Create(w, req)

			assert.Equal(t, tc.wantCode, w.Code)
		})
	}
}

func TestHandler_Get(t *testing.T) {
	tests := []struct {
		desc      string
		id        string
		mockSetup func(m *MockerForHandler)
		wantCode  int
	}{
		{
			desc: "not found",
			id:   "1",
			mockSetup: func(m *MockerForHandler) {
				m.service.On("GetAccount", mock.Anything, "1").
					Return(nil, fmt.Errorf("not found"))
			},
			wantCode: http.StatusNotFound,
		},
		{
			desc: "success",
			id:   "1",
			mockSetup: func(m *MockerForHandler) {
				m.service.On("GetAccount", mock.Anything, "1").
					Return(&account.Account{}, nil)
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			m := &MockerForHandler{
				service: mocks.NewServiceInterface(t),
			}
			tc.mockSetup(m)
			h := account.NewHandler(m.service)

			req := httptest.NewRequest(http.MethodGet, "/accounts/"+tc.id, nil)
			req.SetPathValue("id", tc.id)

			w := httptest.NewRecorder()

			h.Get(w, req)

			assert.Equal(t, tc.wantCode, w.Code)
		})
	}
}

func TestHandler_List(t *testing.T) {
	tests := []struct {
		desc      string
		query     string
		mockSetup func(m *MockerForHandler)
		wantCode  int
	}{
		{
			desc:      "invalid cursor",
			query:     "?cursor=invalid",
			mockSetup: func(m *MockerForHandler) {},
			wantCode:  http.StatusBadRequest,
		},
		{
			desc:  "service error",
			query: "?limit=10",
			mockSetup: func(m *MockerForHandler) {
				m.service.On("ListAccounts", mock.Anything, mock.Anything).
					Return([]account.Account{}, 0, "", "", fmt.Errorf("err"))
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			desc:  "success with filters",
			query: "?limit=10&customer_id=1&account_type=saving&status=active&currency=USD",
			mockSetup: func(m *MockerForHandler) {
				m.service.On("ListAccounts", mock.Anything, mock.Anything).
					Return([]account.Account{}, 10, "next", "prev", nil)
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForHandler{
				service: mocks.NewServiceInterface(t),
			}
			tt.mockSetup(m)
			h := account.NewHandler(m.service)

			req := httptest.NewRequest(http.MethodGet, "/accounts"+tt.query, nil)
			w := httptest.NewRecorder()

			h.List(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestHandler_UpdateStatus(t *testing.T) {
	tests := []struct {
		desc      string
		body      string
		mockSetup func(m *MockerForHandler)
		wantCode  int
	}{
		{
			desc:      "invalid json",
			body:      `{invalid}`,
			mockSetup: func(m *MockerForHandler) {},
			wantCode:  http.StatusBadRequest,
		},
		{
			desc: "service error",
			body: `{"status":"active"}`,
			mockSetup: func(m *MockerForHandler) {
				m.service.On("UpdateStatus", mock.Anything, "1", "active").
					Return(fmt.Errorf("err"))
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			desc: "success",
			body: `{"status":"active"}`,
			mockSetup: func(m *MockerForHandler) {
				m.service.On("UpdateStatus", mock.Anything, "1", "active").
					Return(nil)
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForHandler{
				service: mocks.NewServiceInterface(t),
			}
			tt.mockSetup(m)
			h := account.NewHandler(m.service)

			req := httptest.NewRequest(http.MethodPatch, "/accounts/1", strings.NewReader(tt.body))
			req.SetPathValue("id", "1")

			w := httptest.NewRecorder()

			h.UpdateStatus(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	tests := []struct {
		desc      string
		mockSetup func(m *MockerForHandler)
		wantCode  int
	}{
		{
			desc: "service error",
			mockSetup: func(m *MockerForHandler) {
				m.service.On("DeleteAccount", mock.Anything, "1").
					Return(fmt.Errorf("err"))
			},
			wantCode: http.StatusBadRequest,
		},
		{
			desc: "success",
			mockSetup: func(m *MockerForHandler) {
				m.service.On("DeleteAccount", mock.Anything, "1").
					Return(nil)
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			m := &MockerForHandler{
				service: mocks.NewServiceInterface(t),
			}
			tt.mockSetup(m)
			h := account.NewHandler(m.service)

			req := httptest.NewRequest(http.MethodDelete, "/accounts/1", nil)
			req.SetPathValue("id", "1")

			w := httptest.NewRecorder()

			h.Delete(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}
