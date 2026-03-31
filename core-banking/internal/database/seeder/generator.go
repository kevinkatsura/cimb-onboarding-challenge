package seeder

import (
	"core-banking/internal/domain"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type SeedData struct {
	Customers       []domain.Customer
	Documents       []domain.CustomerDocument
	Accounts        []domain.Account
	Transactions    []domain.Transaction
	Journals        []domain.JournalEntry
	Ledgers         []domain.LedgerEntry
	Payments        []domain.Payment
	AuditLogs       []domain.AuditLog
	IdempotencyKeys []domain.IdempotencyKey
	FXRates         []domain.FXRate
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func ptrDate(t time.Time) *time.Time {
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return &d
}

func GenerateAll(n int) SeedData {
	rand.Seed(42)

	var data SeedData

	for i := 0; i < n; i++ {
		customerID := uuid.New()
		accountID := uuid.New()
		txID := uuid.New()
		journalID := uuid.New()

		amount := int64(rand.Intn(1_000_000) + 1000)

		// CUSTOMER
		customer := domain.Customer{
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
		}
		data.Customers = append(data.Customers, customer)

		// DOCUMENT
		data.Documents = append(data.Documents, domain.CustomerDocument{
			ID:             uuid.New(),
			CustomerID:     customerID,
			DocumentType:   "KTP",
			DocumentNumber: fmt.Sprintf("DOC%06d", i),
			IssuingCountry: "ID",
			ExpiresAt:      ptrDate(time.Now().AddDate(5, 0, 0)),
		})

		// ACCOUNT
		account := domain.Account{
			ID:               accountID,
			CustomerID:       customerID,
			AccountNumber:    fmt.Sprintf("ACC%06d", i),
			AccountType:      "savings",
			Currency:         "IDR",
			Status:           "active",
			AvailableBalance: amount,
			PendingBalance:   0,
			OverdraftLimit:   0,
		}
		data.Accounts = append(data.Accounts, account)

		// TRANSACTION
		tx := domain.Transaction{
			ID:              txID,
			ReferenceID:     fmt.Sprintf("REF%06d", i),
			TransactionType: "deposit",
			Status:          "completed",
			Amount:          amount,
			Currency:        "IDR",
			InitiatedBy:     &customerID,
		}
		data.Transactions = append(data.Transactions, tx)

		// JOURNAL
		data.Journals = append(data.Journals, domain.JournalEntry{
			ID:            journalID,
			TransactionID: txID,
			JournalType:   "deposit",
		})

		// LEDGER (DOUBLE ENTRY)
		data.Ledgers = append(data.Ledgers,
			domain.LedgerEntry{
				ID:        uuid.New(),
				JournalID: journalID,
				AccountID: accountID,
				EntryType: "debit",
				Amount:    amount,
				Currency:  "IDR",
			},
			domain.LedgerEntry{
				ID:        uuid.New(),
				JournalID: journalID,
				AccountID: accountID,
				EntryType: "credit",
				Amount:    amount,
				Currency:  "IDR",
			},
		)

		// PAYMENT
		data.Payments = append(data.Payments, domain.Payment{
			ID:            uuid.New(),
			TransactionID: txID,
			PaymentMethod: "bank_transfer",
			Provider:      "bca",
			Status:        "settled",
			FeeAmount:     0,
		})

		// AUDIT
		data.AuditLogs = append(data.AuditLogs, domain.AuditLog{
			ID:         uuid.New(),
			ActorID:    &customerID,
			EntityType: "transaction",
			EntityID:   &txID,
			Action:     "create",
		})

		// IDEMPOTENCY
		data.IdempotencyKeys = append(data.IdempotencyKeys, domain.IdempotencyKey{
			ID:          uuid.New(),
			Key:         fmt.Sprintf("KEY%06d", i),
			RequestHash: "hash",
		})

		// FX
		data.FXRates = append(data.FXRates, domain.FXRate{
			ID:            uuid.New(),
			BaseCurrency:  "USD",
			QuoteCurrency: "IDR",
			Rate:          15000,
			EffectiveAt:   time.Now(),
		})
	}

	return data
}
