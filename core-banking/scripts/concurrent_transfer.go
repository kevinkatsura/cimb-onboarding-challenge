package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	transferURL   = "http://localhost:8117/v2/transfer"
	accountsURL   = "http://localhost:8117/accounts"
	totalRequests = 100
)

type Account struct {
	ID            string `json:"id"`
	AccountNumber string `json:"account_number"`
	Balance       int64  `json:"available_balance"`
}

type TransferRequest struct {
	ReferenceID string `json:"reference_id"`
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
}

type TransferResponse struct {
	Status string `json:"status"` // success | failed

	TransactionID string `json:"transaction_id,omitempty"`

	SourceAccount      string `json:"source_account"`
	DestinationAccount string `json:"destination_account"`

	Amount int64 `json:"amount"`

	// Success fields
	SourceBalanceAfter      int64 `json:"source_balance_after,omitempty"`
	DestinationBalanceAfter int64 `json:"destination_balance_after,omitempty"`

	// Failure fields
	CurrentBalance int64 `json:"current_balance,omitempty"`

	Message string `json:"message"`
}

type TranferAPIResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    TransferResponse       `json:"data"`
	Error   string                 `json:"error"`
	Meta    map[string]interface{} `json:"meta"`
}

type AccountAPIResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    []Account              `json:"data"`
	Error   interface{}            `json:"error"`
	Meta    map[string]interface{} `json:"meta"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Initialize logger
	config := zap.NewProductionConfig()
	config.Encoding = "json"
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build(zap.AddCaller())
	defer logger.Sync()
	sugar := logger.Sugar()

	client := &http.Client{Timeout: 6 * time.Second}

	// --- Fetch Accounts ---
	accounts, err := fetchAccounts(client)
	if err != nil {
		panic(err)
	}

	if len(accounts) < 2 {
		panic("not enough accounts")
	}

	destAccount := accounts[0]
	sugar.Infow("destination_account_selected",
		"account_id", destAccount.ID,
		"account_number", destAccount.AccountNumber,
	)

	var wg sync.WaitGroup

	var successCount int64
	var businessFailCount int64
	var timeoutCount int64
	var systemErrorCount int64
	var lockTimeout int64

	start := time.Now()

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			src := randomSource(accounts, destAccount.ID)
			amount := randomAmount(i)

			reqBody := TransferRequest{
				ReferenceID: fmt.Sprintf("ref-%d-%d", i, time.Now().UnixNano()),
				FromAccount: src.ID,
				ToAccount:   destAccount.ID,
				Amount:      amount,
				Currency:    "IDR",
			}

			body, _ := json.Marshal(reqBody)

			req, _ := http.NewRequest("POST", transferURL, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				atomic.AddInt64(&timeoutCount, 1)
				sugar.Errorw("transfer_network_error",
					"reference_id", reqBody.ReferenceID,
					"from_account", src.ID,
					"to_account", destAccount.ID,
					"amount", amount,
					"error", err,
				)
				return
			}
			defer resp.Body.Close()

			var apiResp TranferAPIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				atomic.AddInt64(&systemErrorCount, 1)
				sugar.Errorw("transfer_response_decode_error",
					"reference_id", reqBody.ReferenceID,
					"error", err,
				)
				return
			}

			if !apiResp.Success {
				switch apiResp.Error {
				case "transfer timeout (exceeded 4s)":
					atomic.AddInt64(&timeoutCount, 1)
					sugar.Warnw("transfer_timeout",
						"reference_id", reqBody.ReferenceID,
					)
				case "lock timeout":
					atomic.AddInt64(&lockTimeout, 1)
					sugar.Warnw("transfer_lock_timeout",
						"reference_id", reqBody.ReferenceID,
					)
				default:
					atomic.AddInt64(&systemErrorCount, 1)
					sugar.Errorw("transfer_failed_with_error",
						"reference_id", reqBody.ReferenceID,
						"error", apiResp.Error,
					)
				}
				return
			}

			// Business-level evaluation
			switch apiResp.Data.Status {
			case "success":
				atomic.AddInt64(&successCount, 1)
				sugar.Infow("transfer_success",
					"reference_id", reqBody.ReferenceID,
					"source_account", apiResp.Data.SourceAccount,
					"source_balance_before", src.Balance,
					"source_balance_after", apiResp.Data.SourceBalanceAfter,
					"destination_account", apiResp.Data.DestinationAccount,
					"destination_balance_before", destAccount.Balance+1,
					"destination_balance_after", apiResp.Data.DestinationBalanceAfter,
					"amount", amount,
				)
			case "failed":
				// case: insufficient balance
				atomic.AddInt64(&businessFailCount, 1)
				sugar.Warnw("transfer_business_failed",
					"reference_id", reqBody.ReferenceID,
					"source_account", apiResp.Data.SourceAccount,
					"message", apiResp.Data.Message,
				)
			default:
				atomic.AddInt64(&systemErrorCount, 1)
				sugar.Errorw("transfer_unknown_status",
					"reference_id", reqBody.ReferenceID,
					"status", apiResp.Data.Status,
				)
			}

		}(i)
	}

	wg.Wait()

	duration := time.Since(start)

	sugar.Infow("concurrent_transfer_test_results",
		"total_requests", totalRequests,
		"success_count", successCount,
		"business_failed_count", businessFailCount,
		"transfer_timeout_count", timeoutCount,
		"lock_timeout_count", lockTimeout,
		"system_error_count", systemErrorCount,
		"duration_ms", duration.Milliseconds(),
	)
}

// --- Helpers ---
func fetchAccounts(client *http.Client) ([]Account, error) {
	resp, err := client.Get(accountsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data AccountAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	return data.Data, err
}

func randomSource(accounts []Account, dest string) Account {
	for {
		acc := accounts[rand.Intn(len(accounts))]
		if acc.ID != dest {
			return acc
		}
	}
}

// Force mix: success + insufficient balance
func randomAmount(i int) int64 {
	if i%3 == 0 {
		return 1_000_000_000 // trigger failure
	}
	return int64(rand.Intn(1000) + 1)
}
