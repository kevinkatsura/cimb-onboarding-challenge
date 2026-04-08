package cache

import (
	"context"
	"core-banking/pkg/telemetry"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"core-banking/internal/domain"

	"go.opentelemetry.io/otel/codes"
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
	ctx, span := telemetry.Tracer.Start(ctx, "cache.SetAccount")
	defer span.End()

	key := AccountKeyPrefix + account.ID.String()
	span.SetAttributes(telemetry.CacheAttrsNoHit("redis", "SET", key, int(c.ttl.Seconds()))...)

	data, err := json.Marshal(account)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "marshal failed")
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis SET failed")
		return err
	}

	span.SetStatus(codes.Ok, "cached")
	return nil
}

// GetAccount retrieves cached account
func (c *RedisCache) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "cache.GetAccount")
	defer span.End()

	key := AccountKeyPrefix + id
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		span.SetAttributes(telemetry.CacheAttrs("redis", "GET", key, false, int(c.ttl.Seconds()))...)
		span.SetStatus(codes.Ok, "cache miss")
		return nil, nil // Cache miss
	}
	if err != nil {
		span.SetAttributes(telemetry.CacheAttrs("redis", "GET", key, false, int(c.ttl.Seconds()))...)
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis GET failed")
		return nil, fmt.Errorf("failed to get account from cache: %w", err)
	}

	span.SetAttributes(telemetry.CacheAttrs("redis", "GET", key, true, int(c.ttl.Seconds()))...)

	var account domain.Account
	if err := json.Unmarshal([]byte(data), &account); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unmarshal failed")
		return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	}

	span.SetStatus(codes.Ok, "cache hit")
	return &account, nil
}

// DeleteAccount removes account from cache
func (c *RedisCache) DeleteAccount(ctx context.Context, id string) error {
	ctx, span := telemetry.Tracer.Start(ctx, "cache.DeleteAccount")
	defer span.End()

	key := AccountKeyPrefix + id
	span.SetAttributes(telemetry.CacheAttrsNoHit("redis", "DEL", key, 0)...)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis DEL failed")
		return err
	}

	span.SetStatus(codes.Ok, "deleted")
	return nil
}

// SetAccountBalance caches account balance separately (shorter TTL)
func (c *RedisCache) SetAccountBalance(ctx context.Context, accountID string, balance int64) error {
	ctx, span := telemetry.Tracer.Start(ctx, "cache.SetAccountBalance")
	defer span.End()

	key := BalanceKeyPrefix + accountID
	span.SetAttributes(telemetry.CacheAttrsNoHit("redis", "SET", key, int(c.ttl.Seconds()))...)

	if err := c.client.Set(ctx, key, balance, c.ttl).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis SET failed")
		return err
	}

	span.SetStatus(codes.Ok, "balance cached")
	return nil
}

// GetAccountBalance retrieves cached balance
func (c *RedisCache) GetAccountBalance(ctx context.Context, accountID string) (int64, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "cache.GetAccountBalance")
	defer span.End()

	key := BalanceKeyPrefix + accountID
	balance, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		span.SetAttributes(telemetry.CacheAttrs("redis", "GET", key, false, int(c.ttl.Seconds()))...)
		span.SetStatus(codes.Ok, "cache miss")
		return 0, nil // Cache miss
	}
	if err != nil {
		span.SetAttributes(telemetry.CacheAttrs("redis", "GET", key, false, int(c.ttl.Seconds()))...)
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis GET failed")
		return 0, err
	}

	span.SetAttributes(telemetry.CacheAttrs("redis", "GET", key, true, int(c.ttl.Seconds()))...)
	span.SetStatus(codes.Ok, "cache hit")
	return balance, nil
}

// InvalidateAccountBalance removes balance from cache
func (c *RedisCache) InvalidateAccountBalance(ctx context.Context, accountID string) error {
	ctx, span := telemetry.Tracer.Start(ctx, "cache.InvalidateAccountBalance")
	defer span.End()

	key := BalanceKeyPrefix + accountID
	span.SetAttributes(telemetry.CacheAttrsNoHit("redis", "DEL", key, 0)...)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis DEL failed")
		return err
	}

	span.SetStatus(codes.Ok, "invalidated")
	return nil
}

// AcquireLock acquires a distributed lock for account operations
func (c *RedisCache) AcquireLock(ctx context.Context, accountID string, ttl time.Duration) (bool, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "cache.AcquireLock")
	defer span.End()

	key := AccountLockPrefix + accountID
	span.SetAttributes(telemetry.CacheAttrsNoHit("redis", "SETNX", key, int(ttl.Seconds()))...)

	acquired, err := c.client.SetNX(ctx, key, "locked", ttl).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis SETNX failed")
		return false, err
	}

	span.SetAttributes(telemetry.KeyCacheHit.Bool(acquired))
	span.SetStatus(codes.Ok, "lock attempted")
	return acquired, nil
}

// ReleaseLock releases the distributed lock
func (c *RedisCache) ReleaseLock(ctx context.Context, accountID string) error {
	ctx, span := telemetry.Tracer.Start(ctx, "cache.ReleaseLock")
	defer span.End()

	key := AccountLockPrefix + accountID
	span.SetAttributes(telemetry.CacheAttrsNoHit("redis", "DEL", key, 0)...)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis DEL failed")
		return err
	}

	span.SetStatus(codes.Ok, "lock released")
	return nil
}

// SetAccountList caches paginated account lists
func (c *RedisCache) SetAccountList(ctx context.Context, cacheKey string, accounts []domain.Account, total int) error {
	ctx, span := telemetry.Tracer.Start(ctx, "cache.SetAccountList")
	defer span.End()

	key := AccountListPrefix + cacheKey
	span.SetAttributes(telemetry.CacheAttrsNoHit("redis", "SET", key, int(c.ttl.Seconds()))...)

	data := struct {
		Accounts []domain.Account `json:"accounts"`
		Total    int              `json:"total"`
	}{
		Accounts: accounts,
		Total:    total,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "marshal failed")
		return fmt.Errorf("failed to marshal account list: %w", err)
	}

	if err := c.client.Set(ctx, key, jsonData, c.ttl).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis SET failed")
		return err
	}

	span.SetStatus(codes.Ok, "list cached")
	return nil
}

// GetAccountList retrieves cached account list
func (c *RedisCache) GetAccountList(ctx context.Context, cacheKey string) ([]domain.Account, int, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "cache.GetAccountList")
	defer span.End()

	key := AccountListPrefix + cacheKey
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		span.SetAttributes(telemetry.CacheAttrs("redis", "GET", key, false, int(c.ttl.Seconds()))...)
		span.SetStatus(codes.Ok, "cache miss")
		return nil, 0, nil // Cache miss
	}
	if err != nil {
		span.SetAttributes(telemetry.CacheAttrs("redis", "GET", key, false, int(c.ttl.Seconds()))...)
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis GET failed")
		return nil, 0, fmt.Errorf("failed to get account list from cache: %w", err)
	}

	span.SetAttributes(telemetry.CacheAttrs("redis", "GET", key, true, int(c.ttl.Seconds()))...)

	var result struct {
		Accounts []domain.Account `json:"accounts"`
		Total    int              `json:"total"`
	}

	if err := json.Unmarshal([]byte(data), &result); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unmarshal failed")
		return nil, 0, fmt.Errorf("failed to unmarshal account list: %w", err)
	}

	span.SetStatus(codes.Ok, "cache hit")
	return result.Accounts, result.Total, nil
}
