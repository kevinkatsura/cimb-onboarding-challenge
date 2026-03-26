package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
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
	fmt.Println("Destination account:", destAccount.ID)

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
				fmt.Println("network/timeout error:", err)
				return
			}
			defer resp.Body.Close()

			var apiResp TranferAPIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				atomic.AddInt64(&systemErrorCount, 1)
				fmt.Println("decode error:", err)
				return
			}

			if !apiResp.Success {
				switch apiResp.Error {
				case "transfer timeout (exceeded 4s)":
					atomic.AddInt64(&timeoutCount, 1)
				case "lock timeout":
					atomic.AddInt64(&lockTimeout, 1)
				default:
					atomic.AddInt64(&systemErrorCount, 1)
				}
				return
			}

			// Business-level evaluation
			switch apiResp.Data.Status {
			case "success":
				atomic.AddInt64(&successCount, 1)
				log.Println("transfer_success",
					", source_account: ", apiResp.Data.SourceAccount,
					", source_balane_before: ", src.Balance,
					", source_balance_after: ", apiResp.Data.SourceBalanceAfter,
					", destination_account: ", apiResp.Data.DestinationAccount,
					", destination_balance_before: ", destAccount.Balance+1,
					", destination_balance_after: ", apiResp.Data.DestinationBalanceAfter,
				)
			case "failed":
				// case: insufficient balance
				atomic.AddInt64(&businessFailCount, 1)
			default:
				atomic.AddInt64(&systemErrorCount, 1)
			}

		}(i)
	}

	wg.Wait()

	duration := time.Since(start)

	fmt.Println("\n========== RESULT ==========")
	fmt.Println("Total Requests:", totalRequests)
	fmt.Println("Success:", successCount)
	fmt.Println("Business Failed (insufficient balance):", businessFailCount)
	fmt.Println("Tranfer timeout:", timeoutCount)
	fmt.Println("Lock timeout:", lockTimeout)
	fmt.Println("Rest/System Error:", systemErrorCount)
	fmt.Println("Duration:", duration)
	fmt.Println("============================")
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
