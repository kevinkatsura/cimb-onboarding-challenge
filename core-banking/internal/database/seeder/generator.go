package seeder

import (
	"core-banking/internal/account"
	"core-banking/internal/transaction"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type SeedData struct {
	Customers           []account.Customer
	Documents           []account.CustomerDocument
	Products            []account.Product
	Accounts            []account.Account
	AccountBalances     []account.AccountBalance
	AccountTransactions []account.AccountTransaction
	Transactions        []transaction.Transaction
	TransferDetails     []transaction.TransferDetail
	Ledgers             []transaction.LedgerEntry
	AuditLogs           []transaction.AuditLog
	IdempotencyKeys     []transaction.IdempotencyKey
	FXRates             []transaction.FXRate
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func ptrDate(t time.Time) *time.Time {
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return &d
}

func GenerateAll(n int) SeedData {
	rand.Seed(time.Now().UnixNano())

	var data SeedData

	// 1. PRODUCTS
	products := []account.Product{
		{Code: "savings", Name: "Savings Account", Currency: "IDR", MinBalance: 50000, OverdraftLimit: 0, CreatedAt: time.Now()},
		{Code: "checking", Name: "Checking Account", Currency: "IDR", MinBalance: 0, OverdraftLimit: 1000000, CreatedAt: time.Now()},
		{Code: "premium", Name: "Premium Account", Currency: "IDR", MinBalance: 10000000, OverdraftLimit: 5000000, CreatedAt: time.Now()},
	}
	data.Products = products

	for i := 0; i < n; i++ {
		customerID := uuid.New()
		accountID := uuid.New()
		txID := uuid.New()

		amount := int64(rand.Intn(1_000_000) + 1000)
		prod := products[rand.Intn(len(products))]

		// CUSTOMER
		customer := account.Customer{
			ID:            customerID,
			FullName:      fmt.Sprintf("Customer %d", i),
			DateOfBirth:   time.Now().AddDate(-20-rand.Intn(30), 0, 0),
			Nationality:   "ID",
			Email:         fmt.Sprintf("user%d@mail.com", i),
			PhoneNumber:   fmt.Sprintf("08123%06d", i),
			KYCStatus:     "verified",
			KYCVerifiedAt: ptrTime(time.Now()),
			RiskLevel:     "low",
			PEPFlag:       false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		data.Customers = append(data.Customers, customer)

		// DOCUMENT
		data.Documents = append(data.Documents, account.CustomerDocument{
			ID:             uuid.New(),
			CustomerID:     customerID,
			DocumentType:   "KTP",
			DocumentNumber: fmt.Sprintf("DOC%06d", i),
			IssuingCountry: "ID",
			ExpiresAt:      ptrDate(time.Now().AddDate(5, 0, 0)),
			CreatedAt:      time.Now(),
		})

		// ACCOUNT
		acc := account.Account{
			ID:            accountID,
			CustomerID:    customerID,
			AccountNumber: fmt.Sprintf("ACC%06d", i),
			ProductCode:   prod.Code,
			Currency:      "IDR",
			Status:        "active",
			OpenedAt:      time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		data.Accounts = append(data.Accounts, acc)

		// BALANCE (Simplified for seeder: set to amount deposited)
		data.AccountBalances = append(data.AccountBalances, account.AccountBalance{
			AccountID:        accountID,
			AvailableBalance: amount,
			PendingBalance:   0,
			LastUpdated:      time.Now(),
		})

		// TRANSACTION
		tx := transaction.Transaction{
			ID:                 txID,
			PartnerReferenceNo: fmt.Sprintf("REF%06d", i),
			TransactionType:    "deposit",
			Status:             "completed",
			Amount:             amount,
			Currency:           "IDR",
			CreatedAt:          time.Now(),
			CompletedAt:        ptrTime(time.Now()),
		}
		data.Transactions = append(data.Transactions, tx)

		// ACCOUNT TRANSACTION (History link)
		data.AccountTransactions = append(data.AccountTransactions, account.AccountTransaction{
			ID:            uuid.New(),
			AccountID:     accountID,
			TransactionID: txID,
			Direction:     "in",
			Amount:        amount,
			CreatedAt:     time.Now(),
		})

		// TRANSFER DETAILS (New SNAP Metadata Table)
		td := transaction.TransferDetail{
			ID:                   uuid.New(),
			TransactionID:        txID,
			SourceAccountNo:      fmt.Sprintf("ACC%06d", i),
			BeneficiaryAccountNo: "ACC999999",
			FeeType:              "OUR",
			Remark:               "Seed Transfer",
			CreatedAt:            time.Now(),
		}
		data.TransferDetails = append(data.TransferDetails, td)

		// LEDGER (DOUBLE ENTRY)
		data.Ledgers = append(data.Ledgers,
			transaction.LedgerEntry{
				ID:            uuid.New(),
				TransactionID: txID,
				AccountID:     accountID,
				EntryType:     "debit",
				Amount:        amount,
				Currency:      "IDR",
				CreatedAt:     time.Now(),
			},
			transaction.LedgerEntry{
				ID:            uuid.New(),
				TransactionID: txID,
				AccountID:     accountID,
				EntryType:     "credit",
				Amount:        amount,
				Currency:      "IDR",
				CreatedAt:     time.Now(),
			},
		)

		// AUDIT
		ipStr := "127.0.0.1"
		data.AuditLogs = append(data.AuditLogs, transaction.AuditLog{
			ID:         uuid.New(),
			ActorID:    &customerID,
			EntityType: "transaction",
			EntityID:   &txID,
			Action:     "create",
			IPAddress:  &ipStr,
			CreatedAt:  time.Now(),
		})

		// IDEMPOTENCY
		data.IdempotencyKeys = append(data.IdempotencyKeys, transaction.IdempotencyKey{
			ID:              uuid.New(),
			Key:             fmt.Sprintf("KEY%06d", i),
			ResponseCode:    "2001700",
			ResponseMessage: "Successful",
			ResponseBody:    []byte(`{"status":"success"}`),
			CreatedAt:       time.Now(),
		})

		// FX
		data.FXRates = append(data.FXRates, transaction.FXRate{
			ID:            uuid.New(),
			BaseCurrency:  "USD",
			QuoteCurrency: "IDR",
			Rate:          15000,
			EffectiveAt:   time.Now(),
		})
	}

	return data
}
