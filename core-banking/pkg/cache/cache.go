package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"core-banking/internal/domain"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client: client,
		ttl:    1 * time.Minute,
	}
}

// Cache key patterns
const (
	AccountKeyPrefix  = "account:"
	AccountListPrefix = "account:list:"
	BalanceKeyPrefix  = "balance:"
	AccountLockPrefix = "lock:account:"
)

// SetAccount caches an account with TTL
func (c *RedisCache) SetAccount(ctx context.Context, account *domain.Account) error {
	key := AccountKeyPrefix + account.ID.String()
	data, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

// GetAccount retrieves cached account
func (c *RedisCache) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	key := AccountKeyPrefix + id
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account from cache: %w", err)
	}

	var account domain.Account
	if err := json.Unmarshal([]byte(data), &account); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	return &account, nil
}

// DeleteAccount removes account from cache
func (c *RedisCache) DeleteAccount(ctx context.Context, id string) error {
	key := AccountKeyPrefix + id
	return c.client.Del(ctx, key).Err()
}

// SetAccountBalance caches account balance separately (shorter TTL)
func (c *RedisCache) SetAccountBalance(ctx context.Context, accountID string, balance int64) error {
	key := BalanceKeyPrefix + accountID
	return c.client.Set(ctx, key, balance, c.ttl).Err()
}

// GetAccountBalance retrieves cached balance
func (c *RedisCache) GetAccountBalance(ctx context.Context, accountID string) (int64, error) {
	key := BalanceKeyPrefix + accountID
	balance, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil // Cache miss
	}
	return balance, err
}

// InvalidateAccountBalance removes balance from cache
func (c *RedisCache) InvalidateAccountBalance(ctx context.Context, accountID string) error {
	key := BalanceKeyPrefix + accountID
	return c.client.Del(ctx, key).Err()
}

// AcquireLock acquires a distributed lock for account operations
func (c *RedisCache) AcquireLock(ctx context.Context, accountID string, ttl time.Duration) (bool, error) {
	key := AccountLockPrefix + accountID
	return c.client.SetNX(ctx, key, "locked", ttl).Result()
}

// ReleaseLock releases the distributed lock
func (c *RedisCache) ReleaseLock(ctx context.Context, accountID string) error {
	key := AccountLockPrefix + accountID
	return c.client.Del(ctx, key).Err()
}

// CacheAccountList caches paginated account lists
func (c *RedisCache) SetAccountList(ctx context.Context, cacheKey string, accounts []domain.Account, total int) error {
	data := struct {
		Accounts []domain.Account `json:"accounts"`
		Total    int              `json:"total"`
	}{
		Accounts: accounts,
		Total:    total,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal account list: %w", err)
	}

	key := AccountListPrefix + cacheKey
	return c.client.Set(ctx, key, jsonData, c.ttl).Err()
}

// GetAccountList retrieves cached account list
func (c *RedisCache) GetAccountList(ctx context.Context, cacheKey string) ([]domain.Account, int, error) {
	key := AccountListPrefix + cacheKey
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, 0, nil // Cache miss
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get account list from cache: %w", err)
	}

	var result struct {
		Accounts []domain.Account `json:"accounts"`
		Total    int              `json:"total"`
	}

	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, 0, fmt.Errorf("failed to unmarshal account list: %w", err)
	}

	return result.Accounts, result.Total, nil
}
