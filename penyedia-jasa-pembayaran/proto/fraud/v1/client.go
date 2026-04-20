package fraudpb

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

// FraudDetectionClient is a wrapper around the generated gRPC client.
// It maintains the same interface for the caller but uses gRPC internally.
type FraudDetectionClient struct {
	client FraudDetectionServiceClient
}

// NewFraudDetectionClient creates a new fraud detection client using an existing gRPC connection.
func NewFraudDetectionClient(cc grpc.ClientConnInterface) *FraudDetectionClient {
	return &FraudDetectionClient{
		client: NewFraudDetectionServiceClient(cc),
	}
}

// EvaluateTransaction calls the fraud detection service via gRPC.
func (c *FraudDetectionClient) EvaluateTransaction(ctx context.Context, req *FraudEvaluationRequest) (*FraudEvaluationResponse, error) {
	resp, err := c.client.EvaluateTransaction(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fraud gRPC call failed: %w", err)
	}
	return resp, nil
}
