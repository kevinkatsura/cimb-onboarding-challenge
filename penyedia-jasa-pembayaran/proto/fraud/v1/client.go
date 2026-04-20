package fraudpb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// FraudDetectionClient is a lightweight HTTP-based client for the fraud
// detection service. We use HTTP/JSON instead of gRPC on the Go client side
// because protoc-generated stubs are unavailable. The Python server exposes
// both gRPC and a /evaluate REST endpoint for this purpose.
type FraudDetectionClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewFraudDetectionClient creates a new fraud detection client.
// addr should be like "fraud-detection:8085" (the HTTP port).
func NewFraudDetectionClient(addr string) *FraudDetectionClient {
	return &FraudDetectionClient{
		baseURL: fmt.Sprintf("http://%s", addr),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// EvaluateTransaction calls the fraud detection service to evaluate a transaction.
func (c *FraudDetectionClient) EvaluateTransaction(ctx context.Context, req *FraudEvaluationRequest) (*FraudEvaluationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal fraud request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/evaluate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create fraud request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Inject tracing headers
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpReq.Header))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("fraud service call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fraud service returned status %d", resp.StatusCode)
	}

	var fraudResp FraudEvaluationResponse
	if err := json.NewDecoder(resp.Body).Decode(&fraudResp); err != nil {
		return nil, fmt.Errorf("decode fraud response: %w", err)
	}

	return &fraudResp, nil
}
