package account

import (
	"context"
	"encoding/json"

	"core-banking/pkg/apperror"
	"core-banking/pkg/idgen"
	"core-banking/pkg/pagination"
	"core-banking/pkg/telemetry"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

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
	repo      Repository
	accNumGen idgen.AccountNumberGenerator
	cache     *RedisCache
}

func NewService(repo Repository, accNumGen idgen.AccountNumberGenerator) *Service {
	return &Service{
		repo:      repo,
		accNumGen: accNumGen,
		cache:     nil,
	}
}

func (s *Service) SetRedisClient(client *redis.Client) {
	s.cache = NewRedisCache(client)
}

func (s *Service) RegisterAccount(ctx context.Context, req RegistrationAccountCreationRequest) (*Account, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "accountService.RegisterAccount")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrs("RegisterAccount", "account", "onboarding")...)

	// Default values for registration-based account creation
	defaultProductCode := "savings"
	defaultCurrency := "IDR"

	// Map to internal CreateAccountRequest for reuse
	internalReq := CreateAccountRequest{
		PartnerReferenceNo: req.PartnerReferenceNo,
		CustomerID:         req.CustomerID,
		CountryCode:        req.CountryCode,
		DeviceInfo:         req.DeviceInfo,
		Name:               req.Name,
		Email:              req.Email,
		PhoneNo:            req.PhoneNo,
		OnboardingPartner:  req.OnboardingPartner,
		RedirectURL:        req.RedirectURL,
		Scopes:             req.Scopes,
		SeamlessData:       req.SeamlessData,
		SeamlessSign:       req.SeamlessSign,
		State:              req.State,
		Lang:               req.Lang,
		Locale:             req.Locale,
		MerchantID:         req.MerchantID,
		SubMerchantID:      req.SubMerchantID,
		TerminalType:       req.TerminalType,
		AdditionalInfo:     req.AdditionalInfo,
		ProductCode:        defaultProductCode,
		Currency:           defaultCurrency,
	}

	return s.CreateAccount(ctx, internalReq)
}

func (s *Service) CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "accountService.CreateAccount")
	defer span.End()
	span.SetAttributes(telemetry.ServiceAttrs("CreateAccount", "account", "onboarding")...)

	var customer *Customer
	var err error

	if req.CustomerID != "" {
		if _, err := uuid.Parse(req.CustomerID); err != nil {
			span.RecordError(err)
			return nil, apperror.Wrap(apperror.ErrInvalidFieldFormat, "invalid field format", err)
		}
		customer, err = s.repo.GetCustomerByID(ctx, req.CustomerID)
		if err != nil {
			span.RecordError(err)
			return nil, apperror.NewNotFound("customer not found")
		}
		// Update customer with new SNAP info if provided
		if req.Name != "" {
			customer.FullName = req.Name
		}
		if req.Email != "" {
			customer.Email = req.Email
		}
		if req.PhoneNo != "" {
			customer.PhoneNumber = req.PhoneNo
		}
	} else {
		// New Customer Onboarding
		customer = &Customer{
			FullName:    req.Name,
			Email:       req.Email,
			PhoneNumber: req.PhoneNo,
			KYCStatus:   "pending",
			RiskLevel:   "low",
		}
	}

	// Map SNAP fields
	customer.PartnerReferenceNo = req.PartnerReferenceNo
	customer.CountryCode = req.CountryCode
	customer.ExternalCustomerID = req.CustomerID // Original SNAP ID
	if req.DeviceInfo != nil {
		customer.DeviceOS = req.DeviceInfo.OS
		customer.DeviceOSVersion = req.DeviceInfo.OSVersion
		customer.DeviceModel = req.DeviceInfo.Model
		customer.DeviceManufacturer = req.DeviceInfo.Manufacturer
	}
	customer.Lang = req.Lang
	customer.Locale = req.Locale
	customer.OnboardingPartner = req.OnboardingPartner
	customer.RedirectURL = req.RedirectURL
	customer.Scopes = req.Scopes
	customer.SeamlessData = req.SeamlessData
	customer.SeamlessSign = req.SeamlessSign
	customer.State = req.State
	customer.MerchantID = req.MerchantID
	customer.SubMerchantID = req.SubMerchantID
	customer.TerminalType = req.TerminalType

	if req.AdditionalInfo != nil {
		customer.AdditionalInfo, _ = json.Marshal(req.AdditionalInfo)
	}

	if customer.ID == uuid.Nil {
		err = s.repo.CreateCustomer(ctx, customer)
	} else {
		err = s.repo.UpdateCustomer(ctx, customer)
	}

	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to process customer", err)
	}

	accNumber, err := s.accNumGen.Generate()
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to generate account number", err)
	}

	acc := Account{
		CustomerID:    customer.ID,
		AccountNumber: accNumber,
		ProductCode:   req.ProductCode,
		Currency:      req.Currency,
	}

	err = s.repo.Create(ctx, &acc)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewInternal("failed to create account", err)
	}

	if s.cache != nil {
		defer func() {
			s.cache.SetAccount(ctx, &acc)
			s.cache.SetAccountBalance(ctx, acc.ID.String(), 0)
		}()
	}

	accountCreateCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("product_code", req.ProductCode),
		attribute.String("currency", req.Currency),
	))

	span.SetStatus(codes.Ok, "account created")
	return &acc, nil
}

func (s *Service) GetAccount(ctx context.Context, id string) (*Account, error) {
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

func (s *Service) ListAccounts(ctx context.Context, f ListFilter) ([]Account, int, string, string, error) {
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
		return apperror.NewNotFound("account not found")
	}

	bal, err := s.repo.GetBalance(ctx, id)
	if err != nil {
		span.RecordError(err)
		return apperror.NewInternal("failed to fetch balance", err)
	}

	if bal.AvailableBalance != 0 {
		err := fmt.Errorf("cannot delete account with non-zero balance (balance: %d)", bal.AvailableBalance)
		span.RecordError(err)
		return apperror.NewBadRequest("cannot delete account with non-zero balance")
	}

	if acc.Status != "closed" {
		return apperror.NewBadRequest("account must be closed before deletion")
	}

	err = s.repo.SoftDelete(ctx, id)
	if err != nil {
		span.RecordError(err)
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
		return apperror.NewUnavailable("redis client not configured")
	}

	lockAcquired, err := s.cache.AcquireLock(ctx, accountID, 30*time.Second)
	if err != nil {
		span.RecordError(err)
		return apperror.NewInternal("failed to acquire lock", err)
	}
	if !lockAcquired {
		return apperror.NewConflict("account is currently being modified by another process")
	}

	defer func() {
		go func() {
			ctx := context.Background()
			s.cache.ReleaseLock(ctx, accountID)
		}()
	}()

	acc, err := s.repo.GetByID(ctx, accountID)
	if err != nil {
		span.RecordError(err)
		return apperror.NewNotFound("account not found")
	}

	prod, err := s.repo.GetProduct(ctx, acc.ProductCode)
	if err != nil {
		span.RecordError(err)
		return apperror.NewInternal("failed to fetch product info", err)
	}

	bal, err := s.repo.GetBalance(ctx, accountID)
	if err != nil {
		span.RecordError(err)
		return apperror.NewInternal("failed to fetch balance", err)
	}

	newBalance := bal.AvailableBalance + amount
	if newBalance < -prod.OverdraftLimit {
		return apperror.NewBadRequest(fmt.Sprintf("insufficient funds: balance would be %d, overdraft limit %d", newBalance, prod.OverdraftLimit))
	}

	err = s.repo.UpdateBalance(ctx, accountID, amount)
	if err != nil {
		span.RecordError(err)
		return apperror.NewInternal("failed to update balance", err)
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
