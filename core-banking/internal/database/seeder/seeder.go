package seeder

import (
	"context"

	"core-banking/pkg/logging"

	"github.com/jmoiron/sqlx"
)

type Seeder struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) Seed(ctx context.Context) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	data := GenerateAll(50)

	// Insert in dependency order
	if err := InsertCustomers(tx, data.Customers); err != nil {
		return err
	}
	// if err := InsertCustomerDocuments(tx, data.Documents); err != nil {
	// 	return err
	// }
	if err := InsertAccounts(tx, data.Accounts); err != nil {
		return err
	}
	if err := InsertTransactions(tx, data.Transactions); err != nil {
		return err
	}
	if err := InsertTransferDetails(tx, data.TransferDetails); err != nil {
		return err
	}
	if err := InsertJournals(tx, data.Journals); err != nil {
		return err
	}
	if err := InsertLedgerEntries(tx, data.Ledgers); err != nil {
		return err
	}
	if err := InsertPayments(tx, data.Payments); err != nil {
		return err
	}
	if err := InsertAuditLogs(tx, data.AuditLogs); err != nil {
		return err
	}
	if err := InsertIdempotencyKeys(tx, data.IdempotencyKeys); err != nil {
		return err
	}
	if err := InsertFXRates(tx, data.FXRates); err != nil {
		return err
	}

	logging.Logger().Info("Seeding completed")

	return tx.Commit()
}
