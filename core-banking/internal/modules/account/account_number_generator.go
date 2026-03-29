package account

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type AccountNumberGenerator interface {
	Generate() (string, error)
}

type RandomAccountNumberGenerator struct{}

func (g *RandomAccountNumberGenerator) Generate() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1e12))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("10%010d", n.Int64()), nil
}
