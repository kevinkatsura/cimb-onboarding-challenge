package util

import (
	"fmt"
	"math"
	"strconv"

	"core-banking/internal/snap"
)

// ParseSNAPAmount converts "10000.00" -> 10000 (int64) assuming 2 decimal precision typically provided by SNAP
func ParseSNAPAmount(amount snap.SNAPAmount) (int64, error) {
	if amount.Value == "" {
		return 0, fmt.Errorf("amount is empty")
	}

	valFloat, err := strconv.ParseFloat(amount.Value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format: %w", err)
	}

	// Assuming the internal representation stores integer minor units or direct value.
	// For IDR, fractional cents are rarely used natively, but SNAP forces .00.
	// As this system uses int64 standard amounts, let's round float properly.
	// Usually 10000.00 translates to 10000.
	rounded := int64(math.Round(valFloat))

	if rounded < 0 {
		return 0, fmt.Errorf("negative amounts are not allowed")
	}

	return rounded, nil
}

// FormatSNAPAmount converts 10000 (int64) -> "10000.00" string representation
func FormatSNAPAmount(amount int64) string {
	return fmt.Sprintf("%.2f", float64(amount))
}
