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
	Customers       []account.Customer
	Documents       []account.CustomerDocument
	Accounts        []account.Account
	Transactions    []transaction.Transaction
	TransferDetails []transaction.TransferDetail
	Journals        []transaction.Journal
	Ledgers         []transaction.LedgerEntry
	Payments        []transaction.Payment
	AuditLogs       []transaction.AuditLog
	IdempotencyKeys []transaction.IdempotencyKey
	FXRates         []transaction.FXRate
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
		})

		// ACCOUNT
		account := account.Account{
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
		tx := transaction.Transaction{
			ID:                 txID,
			PartnerReferenceNo: fmt.Sprintf("REF%06d", i),
			TransactionType:    "deposit",
			Status:             "completed",
			Amount:             amount,
			Currency:           "IDR",
		}
		data.Transactions = append(data.Transactions, tx)

		// TRANSFER DETAILS (New SNAP Metadata Table)
		td := transaction.TransferDetail{
			ID:                   uuid.New(),
			TransactionID:        txID,
			SourceAccountNo:      fmt.Sprintf("ACC%06d", i),
			BeneficiaryAccountNo: "ACC999999",
			FeeType:              "OUR",
			Remark:               "Seed Transfer",
		}
		data.TransferDetails = append(data.TransferDetails, td)

		// JOURNAL
		data.Journals = append(data.Journals, transaction.Journal{
			ID:            journalID,
			TransactionID: txID,
			JournalType:   "deposit",
			Status:        "posted",
		})

		// LEDGER (DOUBLE ENTRY)
		data.Ledgers = append(data.Ledgers,
			transaction.LedgerEntry{
				ID:        uuid.New(),
				JournalID: journalID,
				AccountID: accountID,
				EntryType: "debit",
				Amount:    amount,
				Currency:  "IDR",
			},
			transaction.LedgerEntry{
				ID:        uuid.New(),
				JournalID: journalID,
				AccountID: accountID,
				EntryType: "credit",
				Amount:    amount,
				Currency:  "IDR",
			},
		)

		// PAYMENT
		data.Payments = append(data.Payments, transaction.Payment{
			ID:            uuid.New(),
			TransactionID: txID,
			PaymentMethod: "bank_transfer",
			Provider:      "bca",
			Status:        "settled",
			FeeAmount:     0,
		})

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
			RequestHash:     "hash",
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
