package account

import (
	"context"
	"core-banking/pkg/pagination"
	"core-banking/pkg/telemetry"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type AccountRepository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{DB: db}
}

func (r *AccountRepository) Create(ctx context.Context, acc *Account) error {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountRepository.Create")
	defer span.End()

	tx, err := r.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryAcc := `
	INSERT INTO accounts(
		id, customer_id, account_number, product_code, currency, status)
	VALUES(gen_random_uuid(), $1, $2, $3, $4, 'active')
	RETURNING id, created_at, updated_at, opened_at;`

	err = tx.QueryRowxContext(ctx, queryAcc,
		acc.CustomerID, acc.AccountNumber, acc.ProductCode, acc.Currency,
	).StructScan(acc)
	if err != nil {
		return err
	}

	queryBal := `
	INSERT INTO account_balances(account_id, available_balance, pending_balance, last_updated)
	VALUES($1, 0, 0, NOW());`

	_, err = tx.ExecContext(ctx, queryBal, acc.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AccountRepository) GetCustomerByID(ctx context.Context, id string) (*Customer, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountRepository.GetCustomerByID")
	defer span.End()
	span.SetAttributes(telemetry.RepoAttrs("AccountRepository", "GetCustomerByID", "Customer", id)...)

	var cust Customer
	err := r.DB.GetContext(ctx, &cust, "SELECT * FROM customers WHERE id = $1;", id)
	return &cust, err
}

func (r *AccountRepository) GetProduct(ctx context.Context, code string) (*Product, error) {
	var prod Product
	err := r.DB.GetContext(ctx, &prod, "SELECT * FROM products WHERE code = $1", code)
	return &prod, err
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*Account, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountRepository.GetByID")
	defer span.End()

	var acc Account
	err := r.DB.GetContext(ctx, &acc, `
		SELECT 	a.id, a.customer_id, a.account_number, a.product_code, a.currency, a.status,
				a.opened_at, a.closed_at, a.created_at, a.updated_at
		FROM accounts a WHERE a.id=$1;`, id)
	return &acc, err
}

func (r *AccountRepository) GetBalance(ctx context.Context, accountID string) (*AccountBalance, error) {
	var bal AccountBalance
	err := r.DB.GetContext(ctx, &bal, "SELECT * FROM account_balances WHERE account_id = $1", accountID)
	return &bal, err
}

func (r *AccountRepository) List(ctx context.Context, f ListFilter) ([]Account, int, *pagination.Cursor, *pagination.Cursor, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountRepository.List")
	defer span.End()

	var accounts []Account
	var total int

	base := `FROM accounts WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if f.CustomerID != nil {
		base += fmt.Sprintf(" AND customer_id = $%d", idx)
		args = append(args, *f.CustomerID)
		idx++
	}
	if f.Status != nil {
		base += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, *f.Status)
		idx++
	}
	if f.ProductCode != nil {
		base += fmt.Sprintf(" AND product_code = $%d", idx)
		args = append(args, *f.ProductCode)
		idx++
	}
	if f.Currency != nil {
		base += fmt.Sprintf(" AND currency = $%d", idx)
		args = append(args, *f.Currency)
		idx++
	}

	order := "ORDER BY created_at DESC, id DESC"
	if f.Cursor != nil {
		if f.Direction == "prev" {
			base += fmt.Sprintf(" AND (created_at, id) > ($%d, $%d)", idx, idx+1)
			order = "ORDER BY created_at ASC, id ASC"
		} else {
			base += fmt.Sprintf(" AND (created_at, id) < ($%d, $%d)", idx, idx+1)
		}
		args = append(args, f.Cursor.CreatedAt, f.Cursor.ID)
		idx += 2
	}

	err := r.DB.GetContext(ctx, &total, "SELECT COUNT(*) "+base, args...)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	query := `SELECT id, customer_id, account_number, product_code, currency, status,
				opened_at, closed_at, created_at, updated_at ` + base + ` ` + order + ` LIMIT $` + fmt.Sprint(idx)
	args = append(args, f.Limit)

	err = r.DB.SelectContext(ctx, &accounts, query, args...)
	if err != nil {
		return nil, 0, nil, nil, err
	}

	if f.Direction == "prev" {
		for i, j := 0, len(accounts)-1; i < j; i, j = i+1, j-1 {
			accounts[i], accounts[j] = accounts[j], accounts[i]
		}
	}

	var nextCursor, prevCursor *pagination.Cursor
	if len(accounts) > 0 {
		first, last := accounts[0], accounts[len(accounts)-1]
		prevCursor = &pagination.Cursor{CreatedAt: first.CreatedAt, ID: first.ID.String()}
		nextCursor = &pagination.Cursor{CreatedAt: last.CreatedAt, ID: last.ID.String()}
	}

	return accounts, total, nextCursor, prevCursor, nil
}

func (r *AccountRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE accounts
		SET status = $1,
			updated_at = NOW(),
			closed_at = CASE
				WHEN $1 = 'closed' AND closed_at IS NULL THEN NOW()
				ELSE closed_at
			END
		WHERE id = $2;`, status, id)
	return err
}

func (r *AccountRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE accounts 
		SET status = 'closed',
			updated_at = NOW(),
			closed_at = COALESCE(closed_at, NOW()) WHERE id = $1;`, id)
	return err
}

func (r *AccountRepository) UpdateBalance(ctx context.Context, accountID string, amount int64) error {
	_, err := r.DB.ExecContext(ctx, `
		UPDATE account_balances
		SET available_balance = available_balance + $1,
			last_updated = NOW()
		WHERE account_id = $2;`, amount, accountID)
	return err
}

func (r *AccountRepository) CreateCustomer(ctx context.Context, cust *Customer) error {
	query := `
	INSERT INTO customers (
		id, full_name, date_of_birth, nationality, email, phone_number,
		kyc_status, kyc_verified_at, risk_level, pep_flag,
		partner_reference_no, country_code, external_customer_id,
		device_os, device_os_version, device_model, device_manufacturer,
		lang, locale, onboarding_partner, redirect_url,
		scopes, seamless_data, seamless_sign, state,
		merchant_id, sub_merchant_id, terminal_type,
		additional_info,
		created_at, updated_at
	) VALUES (
		gen_random_uuid(), :full_name, :date_of_birth, :nationality, :email, :phone_number,
		:kyc_status, :kyc_verified_at, :risk_level, :pep_flag,
		:partner_reference_no, :country_code, :external_customer_id,
		:device_os, :device_os_version, :device_model, :device_manufacturer,
		:lang, :locale, :onboarding_partner, :redirect_url,
		:scopes, :seamless_data, :seamless_sign, :state,
		:merchant_id, :sub_merchant_id, :terminal_type,
		:additional_info,
		NOW(), NOW()
	) RETURNING id, created_at, updated_at;`

	rows, err := r.DB.NamedQueryContext(ctx, query, cust)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.StructScan(cust)
	}
	return nil
}

func (r *AccountRepository) UpdateCustomer(ctx context.Context, cust *Customer) error {
	query := `
	UPDATE customers SET
		full_name = :full_name, date_of_birth = :date_of_birth, nationality = :nationality,
		email = :email, phone_number = :phone_number, kyc_status = :kyc_status,
		kyc_verified_at = :kyc_verified_at, risk_level = :risk_level, pep_flag = :pep_flag,
		partner_reference_no = :partner_reference_no, country_code = :country_code,
		external_customer_id = :external_customer_id, device_os = :device_os,
		device_os_version = :device_os_version, device_model = :device_model,
		device_manufacturer = :device_manufacturer, lang = :lang, locale = :locale,
		onboarding_partner = :onboarding_partner, redirect_url = :redirect_url,
		scopes = :scopes, seamless_data = :seamless_data, seamless_sign = :seamless_sign,
		state = :state, merchant_id = :merchant_id, sub_merchant_id = :sub_merchant_id,
		terminal_type = :terminal_type, additional_info = :additional_info,
		updated_at = NOW()
	WHERE id = :id`

	_, err := r.DB.NamedExecContext(ctx, query, cust)
	return err
}
