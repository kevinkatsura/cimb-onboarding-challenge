package account

import (
	"context"
	"core-banking/internal/domain/account"
	"core-banking/internal/dto"
	"core-banking/pkg/apperror"
	"core-banking/pkg/cache"
	"core-banking/pkg/idgen"
	"core-banking/pkg/pagination"
	"core-banking/pkg/telemetry"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"core-banking/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

var (
	accountCreateCounter metric.Int64Counter
)

func init() {
	meter := otel.Meter("core-banking.account")
	var err error
	accountCreateCounter, err = meter.Int64Counter("core_banking_accounts_created_total",
		metric.WithDescription("Total number of accounts successfully created"),
	)
	if err != nil {
		panic(err)
	}
}

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

func (s *Service) SetRedisClient(client *redis.Client) {
	s.cache = cache.NewRedisCache(client)
}

func (s *Service) CreateAccount(ctx context.Context, req dto.CreateAccountRequest) (*domain.Account, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "accountService.CreateAccount")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrs("CreateAccount", "account", "onboarding")...)
	span.SetAttributes(
		attribute.String("customer.id", req.CustomerID),
		attribute.String("account.type", req.AccountType),
		attribute.String("currency", req.Currency),
	)

	accNumber, err := s.accNumGen.Generate()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "account number generation failed")
		return nil, apperror.NewInternal("failed to generate account number", err)
	}

	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid customer ID format")
		return nil, apperror.NewBadRequest("invalid customer ID format")
	}

	acc := domain.Account{
		CustomerID:     customerID,
		AccountNumber:  accNumber,
		AccountType:    req.AccountType,
		Currency:       req.Currency,
		OverdraftLimit: req.OverdraftLimit,
	}

	err = s.repo.Create(ctx, &acc)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "database create failed")
		return nil, apperror.NewInternal("failed to create account", err)
	}

	if s.cache != nil {
		defer func() {
			s.cache.SetAccount(ctx, &acc)
			s.cache.SetAccountBalance(ctx, acc.ID.String(), 0)
		}()
	}

	accountCreateCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("account_type", req.AccountType),
		attribute.String("currency", req.Currency),
	))

	span.SetStatus(codes.Ok, "account created")
	return &acc, nil
}

func (s *Service) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "accountService.GetAccount")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrs("GetAccount", "account", "inquiry")...)
	span.SetAttributes(attribute.String("account.id", id))

	if s.cache != nil {
		_, cacheSpan := telemetry.Tracer.Start(ctx, "cache.GetAccount")
		cachedAccount, err := s.cache.GetAccount(ctx, id)
		cacheSpan.End()

		if err == nil && cachedAccount != nil {
			span.SetAttributes(attribute.Bool("cache.hit", true))
			span.SetStatus(codes.Ok, "cache hit")
			return cachedAccount, nil
		}
	}
	span.SetAttributes(attribute.Bool("cache.hit", false))

	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "account not found")
		return nil, apperror.NewNotFound("account not found")
	}

	if s.cache != nil {
		go func() {
			s.cache.SetAccount(ctx, account)
		}()
	}

	span.SetStatus(codes.Ok, "account retrieved")
	return account, nil
}

func (s *Service) ListAccounts(ctx context.Context, f domain.ListFilter) ([]domain.Account, int, string, string, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "accountService.ListAccounts")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrs("ListAccounts", "account", "inquiry")...)

	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}
	if f.Direction == "" {
		f.Direction = "next"
	}

	accounts, total, nextC, prevC, err := s.repo.List(ctx, f)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "listing failed")
		return nil, 0, "", "", apperror.NewInternal("failed to list accounts", err)
	}

	var nextCursor, prevCursor string
	if nextC != nil {
		nextCursor, _ = pagination.EncodeCursor(*nextC)
	}
	if prevC != nil {
		prevCursor, _ = pagination.EncodeCursor(*prevC)
	}

	span.SetStatus(codes.Ok, "accounts listed")
	return accounts, total, nextCursor, prevCursor, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status string) error {
	ctx, span := telemetry.Tracer.Start(ctx, "accountService.UpdateStatus")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrs("UpdateStatus", "account", "lifecycle")...)
	span.SetAttributes(
		attribute.String("account.id", id),
		attribute.String("account.status", status),
	)

	if status != "active" && status != "frozen" && status != "closed" {
		span.SetStatus(codes.Error, "invalid status")
		return apperror.NewBadRequest("invalid status: must be active, frozen, or closed")
	}

	err := s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "status update failed")
		return apperror.NewInternal("failed to update account status", err)
	}

	if s.cache != nil {
		go func() {
			ctx := context.Background()
			s.cache.DeleteAccount(ctx, id)
		}()
	}

	span.SetStatus(codes.Ok, "status updated")
	return nil
}

func (s *Service) DeleteAccount(ctx context.Context, id string) error {
	ctx, span := telemetry.Tracer.Start(ctx, "accountService.DeleteAccount")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrs("DeleteAccount", "account", "lifecycle")...)
	span.SetAttributes(attribute.String("account.id", id))

	acc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "account lookup failed")
		return apperror.NewNotFound("account not found")
	}

	if acc.AvailableBalance != 0 {
		err := fmt.Errorf("cannot delete account with non-zero balance (balance: %d)", acc.AvailableBalance)
		span.RecordError(err)
		span.SetStatus(codes.Error, "business validation failed")
		return apperror.NewBadRequest("cannot delete account with non-zero balance")
	}

	if acc.Status != "closed" {
		span.SetStatus(codes.Error, "account not closed")
		return apperror.NewBadRequest("account must be closed before deletion")
	}

	err = s.repo.SoftDelete(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "soft delete failed")
		return apperror.NewInternal("failed to delete account", err)
	}

	if s.cache != nil {
		go func() {
			ctx := context.Background()
			s.cache.DeleteAccount(ctx, id)
			s.cache.InvalidateAccountBalance(ctx, id)
		}()
	}

	span.SetStatus(codes.Ok, "account deleted")
	return nil
}

func (s *Service) UpdateAccountBalance(ctx context.Context, accountID string, amount int64) error {
	ctx, span := telemetry.Tracer.Start(ctx, "accountService.UpdateAccountBalance")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrs("UpdateAccountBalance", "account", "balance")...)
	span.SetAttributes(attribute.String("account.id", accountID))

	if s.cache == nil {
		span.SetStatus(codes.Error, "redis not configured")
		return apperror.NewUnavailable("redis client not configured")
	}

	lockAcquired, err := s.cache.AcquireLock(ctx, accountID, 30*time.Second)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "lock acquisition failed")
		return apperror.NewInternal("failed to acquire lock", err)
	}
	if !lockAcquired {
		span.SetStatus(codes.Error, "lock contention")
		return apperror.NewConflict("account is currently being modified by another process")
	}

	defer func() {
		go func() {
			ctx := context.Background()
			s.cache.ReleaseLock(ctx, accountID)
		}()
	}()

	account, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "account fetch failed")
		return apperror.NewNotFound("account not found")
	}

	newBalance := account.AvailableBalance + amount
	if newBalance < -account.OverdraftLimit {
		span.SetStatus(codes.Error, "insufficient funds")
		return apperror.NewBadRequest(fmt.Sprintf("insufficient funds: balance would be %d, overdraft limit %d", newBalance, account.OverdraftLimit))
	}

	if s.cache != nil {
		go func() {
			ctx := context.Background()
			s.cache.SetAccountBalance(ctx, accountID, newBalance)
		}()
	}

	span.SetStatus(codes.Ok, "balance updated")
	return nil
}
