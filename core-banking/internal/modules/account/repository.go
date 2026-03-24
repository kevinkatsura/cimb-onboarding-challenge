package account

import "github.com/jmoiron/sqlx"

type Repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) Create(tx *sqlx.Tx, acc *Account) error {
	query := `
	INSERT INTO accounts(
		id, 
		customer_id, 
		account_number,
		account_type,
		currency,
		status,
		overdraft_limit)
	VALUES(gen_random_uuid(), $1, $2, $3, $4, 'active', $5)
	RETURNING id, created_at, updated_at, opened_at`

	return tx.QueryRowx(
		query,
		acc.CustomerID,
		acc.AccountNumber,
		acc.AccountType,
		acc.Currency,
		acc.OverdraftLimit,
	).StructScan(acc)
}

func (r *Repository) GetByID(id string) (*Account, error) {
	var acc Account
	err := r.DB.Get(&acc, `
		SELECT 	id,
				customer_id,
				account_number,
				account_type,
				currency,
				status,
				available_balance,
				pending_balance,
				overdraft_limit,
				opened_at,
				closed_at,
				created_at,
				updated_at
		FROM accounts WHERE id=$1`, id)
	return &acc, err
}

func (r *Repository) ListAll(limit, offset int) ([]Account, error) {
	var accounts []Account

	query := `
		SELECT 	id,
				customer_id,
				account_number,
				account_type,
				currency,
				status,
				available_balance,
				pending_balance,
				overdraft_limit,
				opened_at,
				closed_at,
				created_at,
				updated_at
		FROM accounts WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		`
	err := r.DB.Select(&accounts, query, limit, offset)
	return accounts, err
}

func (r *Repository) ListByCustomer(customerID string) ([]Account, error) {
	var accounts []Account
	err := r.DB.Select(&accounts, `
		SELECT 	id,
				customer_id,
				account_number,
				account_type,
				currency,
				status,
				available_balance,
				pending_balance,
				overdraft_limit,
				opened_at,
				closed_at,
				created_at,
				updated_at
		FROM accounts WHERE customer_id=$1 ORDER BY created_at DESC`, customerID)
	return accounts, err
}

func (r *Repository) UpdateStatus(tx *sqlx.Tx, id string, status string) error {
	_, err := tx.Exec(`
		UPDATE accounts
		SET	status=$1,
			updated_at=NOW(),
			closed_at=CASE WHEN $1='closed' THEN NOW() ELSE NULL END
		WHERE id=$2`, status, id)
	return err
}

func (r *Repository) SoftDelete(tx *sqlx.Tx, id string) error {
	_, err := tx.Exec(`
		UPDATE accounts
		SET deleted_at=NOW(), status='closed', updated_at=NOW()
		WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}
