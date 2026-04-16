package grpcclient

import (
	"context"
	"fmt"

	"payment-initiation-acquiring-service/pkg/logging"
	"payment-initiation-acquiring-service/pkg/telemetry"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AccountClient communicates with the Account Issuance Service via gRPC.
type AccountClient struct {
	conn *grpc.ClientConn
}

type AccountInfo struct {
	AccountID        string
	CustomerID       string
	AccountNumber    string
	ProductCode      string
	Currency         string
	Status           string
	AvailableBalance int64
}

func NewAccountClient(addr string) (*AccountClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to account service: %w", err)
	}
	logging.Logger().Infow("connected to Account Issuance gRPC", "addr", addr)
	return &AccountClient{conn: conn}, nil
}

func (c *AccountClient) GetAccount(ctx context.Context, accountNumber string) (*AccountInfo, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPCClient.GetAccount")
	defer span.End()
	span.SetAttributes(attribute.String("account_number", accountNumber))

	// Manual gRPC invoke (proto stubs will replace this)
	var resp AccountInfo
	err := c.conn.Invoke(ctx, "/account.v1.AccountService/GetAccount",
		&struct{ AccountNumber string }{AccountNumber: accountNumber}, &resp, grpc.CallContentSubtype("json"))
	if err != nil {
		return nil, fmt.Errorf("account lookup failed: %w", err)
	}
	return &resp, nil
}

func (c *AccountClient) Close() error {
	return c.conn.Close()
}

// LedgerClient communicates with the Core Banking System via gRPC.
type LedgerClient struct {
	conn *grpc.ClientConn
}

type JournalLine struct {
	AccountID string
	Debit     int64
	Credit    int64
	Currency  string
}

type CreateJournalEntryResponse struct {
	JournalEntryID string
}

type GetBalanceResponse struct {
	AccountID      string
	CurrentBalance int64
	Currency       string
}

func NewLedgerClient(addr string) (*LedgerClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ledger service: %w", err)
	}
	logging.Logger().Infow("connected to Core Banking System gRPC", "addr", addr)
	return &LedgerClient{conn: conn}, nil
}

func (c *LedgerClient) CreateJournalEntry(ctx context.Context, transactionRef, description string, lines []JournalLine) (string, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPCClient.CreateJournalEntry")
	defer span.End()

	req := struct {
		TransactionRef string
		Description    string
		Lines          []JournalLine
	}{TransactionRef: transactionRef, Description: description, Lines: lines}

	var resp CreateJournalEntryResponse
	err := c.conn.Invoke(ctx, "/ledger.v1.LedgerService/CreateJournalEntry", &req, &resp, grpc.CallContentSubtype("json"))
	if err != nil {
		return "", fmt.Errorf("journal entry creation failed: %w", err)
	}
	return resp.JournalEntryID, nil
}

func (c *LedgerClient) GetBalance(ctx context.Context, accountID string) (int64, error) {
	ctx, span := telemetry.Tracer.Start(ctx, "gRPCClient.GetBalance")
	defer span.End()

	req := struct{ AccountID string }{AccountID: accountID}
	var resp GetBalanceResponse
	err := c.conn.Invoke(ctx, "/ledger.v1.LedgerService/GetBalance", &req, &resp, grpc.CallContentSubtype("json"))
	if err != nil {
		return 0, err
	}
	return resp.CurrentBalance, nil
}

func (c *LedgerClient) Close() error {
	return c.conn.Close()
}
