package database

// import (
// 	"log"

// 	"github.com/google/uuid"
// 	"github.com/jmoiron/sqlx"
// )

// func Seed(db *sqlx.DB) {
// 	tx, err := db.Beginx()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer tx.Rollback()

// 	// ---- Customers ----
// 	customerID := uuid.New()
// 	_, err = tx.Exec(`
// 		INSERT INTO customers (
// 			id, full_name, data_of_birth, nationality,
// 			email, phone_number, kyc_status, risk_level
// 		)
// 		VALUES ($1, 'John Doe', '1990-01-01', 'ID',
// 		        'john@example.com', '+6281234567890', 'verified', 'low')
// 	`, customerID)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// ---- Account ----
// 	accountID := uuid.New()
// 	_, err = tx.Exec(`
// 		INSERT INTO accounts (
// 			id, customer_id, account_number, account_type,
// 			currency, status, available_balance
// 		)
// 		VALUES ($1, $2, '1234567890', 'savings', 'IDR', 'active', 1000000)
// 	`, accountID, customerID)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// ---- Transaction ----
// 	txID := uuid.New()
// 	_, err = tx.Exec(`
// 		INSERT INTO transactions (
// 			id, reference_id, transaction_type,
// 			status, amount, currency
// 		)
// 		VALUES ($1, 'ref-001', 'deposit', 'completed', 1000000, 'IDR')
// 	`, txID)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// ---- Journal Entry ----
// 	journalID := uuid.New()
// 	_, err = tx.Exec(`
// 		INSERT INTO journal_entries (
// 			id, transaction_id, journal_type
// 		)
// 		VALUES ($1, $2, 'deposit')
// 	`, journalID, txID)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// ---- Ledger Entries (double entry) ----
// 	ledger1 := uuid.New()
// 	ledger2 := uuid.New()

// 	_, err = tx.Exec(`
// 		INSERT INTO ledger_entries (
// 			id, journal_id, account_id, entry_type,
// 			amount, currency
// 		)
// 		VALUES
// 		($1, $2, $3, 'debit', 1000000, 'IDR'),
// 		($4, $2, $3, 'credit', 1000000, 'IDR')
// 	`, ledger1, journalID, accountID, ledger2)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// ---- Payment ----
// 	paymentID := uuid.New()
// 	_, err = tx.Exec(`
// 		INSERT INTO payments (
// 			id, transaction_id, payment_method, status
// 		)
// 		VALUES ($1, $2, 'bank_transfer', 'settled')
// 	`, paymentID, txID)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	err = tx.Commit()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Println("Seed data inserted successfully")
// }
