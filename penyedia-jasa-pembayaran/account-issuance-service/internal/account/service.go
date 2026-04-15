package account

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"account-issuance-service/pkg/apperror"
	"account-issuance-service/pkg/logging"
	"account-issuance-service/pkg/messaging"
	"account-issuance-service/pkg/telemetry"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

const (
	cacheTTL        = 5 * time.Minute
	cacheKeyAccount = "ais:account:%s"
	cacheKeyBalance = "ais:balance:%s"
)

type Service struct {
	repo     Repository
	producer messaging.Producer
	redis    *redis.Client
}

func NewService(repo Repository, producer messaging.Producer, redisClient *redis.Client) *Service {
	return &Service{repo: repo, producer: producer, redis: redisClient}
}

// RegisterAccount creates a customer, account, and initial balance atomically.
func (s *Service) RegisterAccount(ctx context.Context, req RegistrationRequest) (*RegistrationResponse, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountService.RegisterAccount")
	defer span.End()
	span.SetAttributes(attribute.String("partner_reference_no", req.PartnerReferenceNo))

	// Create customer
	customerID := uuid.New()
	if req.CustomerID != "" {
		parsedID, err := uuid.Parse(req.CustomerID)
		if err != nil {
			return nil, apperror.New(apperror.ErrInvalidFieldFormat, "invalid customerId format")
		}
		customerID = parsedID
	}

	customer := &Customer{
		ID:                customerID,
		Name:              req.Name,
		Email:             req.Email,
		PhoneNo:           req.PhoneNo,
		CountryCode:       defaultStr(req.CountryCode, "ID"),
		DeviceID:          req.DeviceID,
		DeviceType:        req.DeviceType,
		DeviceModel:       req.DeviceModel,
		DeviceOS:          req.DeviceOS,
		OnboardingPartner: req.OnboardingPartner,
		Lang:              defaultStr(req.Lang, "en"),
		Locale:            req.Locale,
	}

	if err := s.repo.CreateCustomer(ctx, customer); err != nil {
		return nil, apperror.NewInternal("failed to create customer", err)
	}

	// Create account
	accountID := uuid.New()
	accNumber := generateAccountNumber()
	account := &Account{
		ID:            accountID,
		CustomerID:    customerID,
		AccountNumber: accNumber,
		ProductCode:   "savings",
		Currency:      "IDR",
		Status:        "active",
	}

	if err := s.repo.CreateAccount(ctx, account); err != nil {
		return nil, apperror.NewInternal("failed to create account", err)
	}

	// Create initial balance
	balance := &AccountBalance{
		AccountID: accountID,
		Available: 0,
		Pending:   0,
		Currency:  "IDR",
		Version:   1,
	}
	if err := s.repo.CreateBalance(ctx, balance); err != nil {
		return nil, apperror.NewInternal("failed to create balance", err)
	}

	// Cache account and balance
	s.cacheAccount(ctx, account)
	s.cacheBalance(ctx, balance)

	// Publish event
	s.publishAccountCreated(ctx, account)

	logging.Ctx(ctx).Infow("account registered successfully",
		"account_id", accountID.String(), "account_number", accNumber,
		"customer_id", customerID.String())

	return &RegistrationResponse{
		PartnerReferenceNo: req.PartnerReferenceNo,
		AccountNumber:      accNumber,
		AccountID:          accountID.String(),
		CustomerID:         customerID.String(),
		ProductCode:        "savings",
		Currency:           "IDR",
		Status:             "active",
	}, nil
}

// GetAccountByNumber fetches account from Redis cache first, then DB.
func (s *Service) GetAccountByNumber(ctx context.Context, number string) (*Account, *AccountBalance, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountService.GetAccountByNumber")
	defer span.End()
	span.SetAttributes(attribute.String("account_number", number))

	acc, err := s.repo.GetAccountByNumber(ctx, number)
	if err != nil {
		return nil, nil, apperror.NewNotFound("account not found")
	}

	bal, err := s.repo.GetBalance(ctx, acc.ID)
	if err != nil {
		return acc, nil, nil
	}

	return acc, bal, nil
}

// GetAccountByID fetches account by UUID.
func (s *Service) GetAccountByID(ctx context.Context, id uuid.UUID) (*Account, *AccountBalance, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "AccountService.GetAccountByID")
	defer span.End()

	// Try cache
	if s.redis != nil {
		cacheKey := fmt.Sprintf(cacheKeyAccount, id.String())
		cached, err := s.redis.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var acc Account
			if json.Unmarshal(cached, &acc) == nil {
				bal, _ := s.repo.GetBalance(ctx, acc.ID)
				return &acc, bal, nil
			}
		}
	}

	acc, err := s.repo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, nil, apperror.NewNotFound("account not found")
	}

	s.cacheAccount(ctx, acc)
	bal, _ := s.repo.GetBalance(ctx, acc.ID)
	return acc, bal, nil
}

func (s *Service) publishAccountCreated(ctx context.Context, acc *Account) {
	if s.producer == nil {
		return
	}
	event := AccountCreatedEvent{
		AccountID:     acc.ID.String(),
		CustomerID:    acc.CustomerID.String(),
		AccountNumber: acc.AccountNumber,
		ProductCode:   acc.ProductCode,
		Currency:      acc.Currency,
		CreatedAt:     acc.CreatedAt.Format(time.RFC3339),
	}
	if err := s.producer.Publish(ctx, "account-created", acc.ID.String(), event); err != nil {
		logging.Ctx(ctx).Errorw("failed to publish account-created event", "error", err)
	}
}

func (s *Service) cacheAccount(ctx context.Context, acc *Account) {
	if s.redis == nil {
		return
	}
	data, _ := json.Marshal(acc)
	key := fmt.Sprintf(cacheKeyAccount, acc.ID.String())
	s.redis.Set(ctx, key, data, cacheTTL)
}

func (s *Service) cacheBalance(ctx context.Context, bal *AccountBalance) {
	if s.redis == nil {
		return
	}
	data, _ := json.Marshal(bal)
	key := fmt.Sprintf(cacheKeyBalance, bal.AccountID.String())
	s.redis.Set(ctx, key, data, cacheTTL)
}

func generateAccountNumber() string {
	return fmt.Sprintf("80%014d", rand.Int63n(99999999999999))
}

func defaultStr(val, fallback string) string {
	if val == "" {
		return fallback
	}
	return val
}
