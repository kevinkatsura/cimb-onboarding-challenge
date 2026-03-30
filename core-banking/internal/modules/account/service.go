package account

import (
	"context"
	"core-banking/internal/pkg/pagination"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo      AccountRepositoryInterface
	accNumGen AccountNumberGenerator
	cache     *RedisCache
}

func NewService(repo AccountRepositoryInterface, accNumGen AccountNumberGenerator) *Service {
	return &Service{
		repo:      repo,
		accNumGen: accNumGen,
		cache:     nil, // Will be set via dependency injection
	}
}

// SetRedisClient sets the Redis client for caching
func (s *Service) SetRedisClient(client *redis.Client) {
	s.cache = NewRedisCache(client)
}

func (s *Service) CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error) {
	var acc Account

	// 1. Get account number
	accNumber, err := s.accNumGen.Generate()
	if err != nil {
		return nil, err
	}

	acc = Account{
		CustomerID:     req.CustomerID,
		AccountNumber:  accNumber,
		AccountType:    req.AccountType,
		Currency:       req.Currency,
		OverdraftLimit: req.OverdraftLimit,
	}

	// 2. Create account
	err = s.repo.Create(&acc)
	if err != nil {
		return nil, err
	}

	// 3. Cache the new account (write-through)
	if s.cache != nil {
		go func() {
			ctx := context.Background()
			s.cache.SetAccount(ctx, &acc)
			// Initialize balance cache
			s.cache.SetAccountBalance(ctx, acc.ID, 0)
		}()
	}

	return &acc, err
}

func (s *Service) GetAccount(ctx context.Context, id string) (*Account, error) {
	// Try cache first (cache-aside pattern)
	if s.cache != nil {
		if cachedAccount, err := s.cache.GetAccount(ctx, id); err == nil && cachedAccount != nil {
			return cachedAccount, nil
		}
		// Cache miss or error - continue to database
	}

	// Get from database
	account, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Cache the result asynchronously (don't block response)
	if s.cache != nil {
		go func() {
			ctx := context.Background() // Use background context for async operation
			s.cache.SetAccount(ctx, account)
		}()
	}

	return account, nil
}

func (s *Service) ListAccounts(ctx context.Context, f ListFilter) ([]Account, int, string, string, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Direction == "" {
		f.Direction = "next"
	}

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

	return accounts, total, nextCursor, prevCursor, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status string) error {
	// business rule
	if status != "active" && status != "frozen" && status != "closed" {
		return fmt.Errorf("invalid status")
	}

	// Update database
	err := s.repo.UpdateStatus(id, status)
	if err != nil {
		return err
	}

	// Invalidate cache (write-through)
	if s.cache != nil {
		go func() {
			ctx := context.Background()
			s.cache.DeleteAccount(ctx, id)
		}()
	}

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
		return fmt.Errorf("cannot delete account with non-zero balance")
	}

	if acc.Status != "closed" {
		return fmt.Errorf("account must be closed before deletion")
	}

	// 3. Soft delete
	err = s.repo.SoftDelete(id)
	if err != nil {
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

	return nil
}

// UpdateAccountBalance demonstrates distributed locking for balance updates
func (s *Service) UpdateAccountBalance(ctx context.Context, accountID string, amount int64) error {
	if s.cache == nil {
		return fmt.Errorf("redis client not configured")
	}

	// Acquire distributed lock (30 second TTL)
	lockAcquired, err := s.cache.AcquireLock(ctx, accountID, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	if !lockAcquired {
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
		return err
	}

	// Update balance
	newBalance := account.AvailableBalance + amount
	if newBalance < -account.OverdraftLimit {
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

	return nil
}
