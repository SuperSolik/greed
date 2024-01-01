package greed

import "math/big"

func ParseBigFloat(x string) (*big.Float, error) {
	parsedX, _, err := big.ParseFloat(x, 10, 53, big.ToNearestEven)
	return parsedX, err
}
