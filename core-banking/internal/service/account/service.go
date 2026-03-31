package account

import (
	"context"
	"core-banking/internal/domain/account"
	"core-banking/internal/dto"
	"core-banking/pkg/cache"
	"core-banking/pkg/idgen"
	"core-banking/pkg/logging"
	"core-banking/pkg/pagination"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"core-banking/internal/domain"
)

type Service struct {
	repo      account.Repository
	accNumGen idgen.AccountNumberGenerator
	cache     *cache.RedisCache
}

func NewService(repo account.Repository, accNumGen idgen.AccountNumberGenerator) *Service {
	return &Service{
		repo:      repo,
		accNumGen: accNumGen,
		cache:     nil,
	}
}

// SetRedisClient sets the Redis client for caching
func (s *Service) SetRedisClient(client *redis.Client) {
	s.cache = cache.NewRedisCache(client)
}

func (s *Service) CreateAccount(ctx context.Context, req dto.CreateAccountRequest) (*domain.Account, error) {
	var acc domain.Account

	// 1. Get account number
	accNumber, err := s.accNumGen.Generate()
	if err != nil {
		return nil, err
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID format: %w", err)
	}

	acc = domain.Account{
		CustomerID:     customerID,
		AccountNumber:  accNumber,
		AccountType:    req.AccountType,
		Currency:       req.Currency,
		OverdraftLimit: req.OverdraftLimit,
	}

	// 2. Create account
	err = s.repo.Create(&acc)
	if err != nil {
		logging.Logger().Errorw("failed_to_create_account",
			"customer_id", req.CustomerID,
			"account_type", req.AccountType,
			"error", err,
		)
		return nil, err
	}

	// 3. Cache the new account (write-through)
	if s.cache != nil {
		go func() {
			ctx := context.Background()
			s.cache.SetAccount(ctx, &acc)
			s.cache.SetAccountBalance(ctx, acc.ID.String(), 0)
		}()
	}

	logging.Logger().Infow("account_created_successfully",
		"account_id", acc.ID,
		"account_number", acc.AccountNumber,
		"customer_id", req.CustomerID,
		"account_type", req.AccountType,
	)
	return &acc, err
}

func (s *Service) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	// Try cache first (cache-aside pattern)
	if s.cache != nil {
		if cachedAccount, err := s.cache.GetAccount(ctx, id); err == nil && cachedAccount != nil {
			return cachedAccount, nil
		}
	}

	// Get from database
	account, err := s.repo.GetByID(id)
	if err != nil {
		logging.Logger().Warnw("account_not_found",
			"account_id", id,
			"error", err,
		)
		return nil, err
	}

	// Cache the result asynchronously (don't block response)
	if s.cache != nil {
		go func() {
			ctx := context.Background()
			s.cache.SetAccount(ctx, account)
		}()
	}

	logging.Logger().Debugw("account_retrieved_from_db",
		"account_id", id,
		"account_number", account.AccountNumber,
	)
	return account, nil
}

func (s *Service) ListAccounts(ctx context.Context, f domain.ListFilter) ([]domain.Account, int, string, string, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Direction == "" {
		f.Direction = "next"
	}

	// // Try cache first (cache-aside pattern)
	// if s.cache != nil {
	// 	// cache hit
	// 	if cachedAccount, total, err := s.cache.GetAccountList(ctx, "id"); err == nil && cachedAccount != nil {
	// 		return cachedAccount, total, "", "", err
	// 	}
	// }

	accounts, total, nextC, prevC, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, "", "", err
	}

	var nextCursor, prevCursor string
	if nextC != nil {
		nextCursor, _ = pagination.EncodeCursor(*nextC)
	}
	if prevC != nil {
		prevCursor, _ = pagination.EncodeCursor(*prevC)
	}

	// if s.cache != nil {
	// 	go func() {
	// 		ctx := context.Background()
	// 		s.cache.SetAccountList(ctx, "", accounts, total)
	// 	}()
	// }

	return accounts, total, nextCursor, prevCursor, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status string) error {
	// business rule
	if status != "active" && status != "frozen" && status != "closed" {
		logging.Logger().Errorw("invalid_account_status_requested",
			"account_id", id,
			"status", status,
			"valid_statuses", []string{"active", "frozen", "closed"},
		)
		return fmt.Errorf("invalid status")
	}

	// Update database
	err := s.repo.UpdateStatus(id, status)
	if err != nil {
		logging.Logger().Errorw("failed_to_update_account_status",
			"account_id", id,
			"new_status", status,
			"error", err,
		)
		return err
	}

	// Invalidate cache (write-through)
	if s.cache != nil {
		go func() {
			ctx := context.Background()
			s.cache.DeleteAccount(ctx, id)
		}()
	}

	logging.Logger().Infow("account_status_updated",
		"account_id", id,
		"new_status", status,
	)
	return nil
}

func (s *Service) DeleteAccount(ctx context.Context, id string) error {
	// 1. Lock account
	acc, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 2. Business rules (CRITICAL)
	if acc.AvailableBalance != 0 {
		logging.Logger().Warnw("account_deletion_blocked_by_balance",
			"account_id", id,
			"available_balance", acc.AvailableBalance,
		)
		return fmt.Errorf("cannot delete account with non-zero balance")
	}

	if acc.Status != "closed" {
		logging.Logger().Warnw("account_deletion_blocked_by_status",
			"account_id", id,
			"current_status", acc.Status,
			"required_status", "closed",
		)
		return fmt.Errorf("account must be closed before deletion")
	}

	// 3. Soft delete
	err = s.repo.SoftDelete(id)
	if err != nil {
		logging.Logger().Errorw("failed_to_delete_account",
			"account_id", id,
			"error", err,
		)
		return err
	}

	// 4. Invalidate cache (write-through)
	if s.cache != nil {
		go func() {
			ctx := context.Background()
			s.cache.DeleteAccount(ctx, id)
			s.cache.InvalidateAccountBalance(ctx, id)
		}()
	}

	logging.Logger().Infow("account_deleted_successfully",
		"account_id", id,
	)
	return nil
}

// UpdateAccountBalance demonstrates distributed locking for balance updates
func (s *Service) UpdateAccountBalance(ctx context.Context, accountID string, amount int64) error {
	if s.cache == nil {
		logging.Logger().Errorw("redis_client_not_configured",
			"account_id", accountID,
		)
		return fmt.Errorf("redis client not configured")
	}

	// Acquire distributed lock (30 second TTL)
	lockAcquired, err := s.cache.AcquireLock(ctx, accountID, 30*time.Second)
	if err != nil {
		logging.Logger().Errorw("failed_to_acquire_lock",
			"account_id", accountID,
			"error", err,
		)
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !lockAcquired {
		logging.Logger().Warnw("account_locked_by_another_process",
			"account_id", accountID,
		)
		return fmt.Errorf("account is currently being modified by another process")
	}

	// Ensure lock is released
	defer func() {
		go func() {
			ctx := context.Background()
			s.cache.ReleaseLock(ctx, accountID)
		}()
	}()

	// Get current account
	account, err := s.repo.GetByID(accountID)
	if err != nil {
		logging.Logger().Errorw("failed_to_fetch_account_for_balance_update",
			"account_id", accountID,
			"error", err,
		)
		return err
	}

	// Update balance
	newBalance := account.AvailableBalance + amount
	if newBalance < -account.OverdraftLimit {
		logging.Logger().Warnw("insufficient_balance_for_operation",
			"account_id", accountID,
			"current_balance", account.AvailableBalance,
			"requested_amount", amount,
			"resulting_balance", newBalance,
			"overdraft_limit", account.OverdraftLimit,
		)
		return fmt.Errorf("insufficient funds: balance would be %d, overdraft limit %d", newBalance, account.OverdraftLimit)
	}

	// In a real implementation, this would be done in a transaction
	// For demo purposes, we'll just update the cache
	if s.cache != nil {
		go func() {
			ctx := context.Background()
			s.cache.SetAccountBalance(ctx, accountID, newBalance)
		}()
	}

	logging.Logger().Infow("account_balance_updated",
		"account_id", accountID,
		"previous_balance", account.AvailableBalance,
		"amount_adjusted", amount,
		"new_balance", newBalance,
	)
	return nil
}
