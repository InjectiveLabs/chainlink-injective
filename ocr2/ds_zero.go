package ocr2

import (
	"context"
	"math/big"
)

type dsZero struct{}

func (d *dsZero) Observe(ctx context.Context) (*big.Int, error) {
	return new(big.Int), nil
}
