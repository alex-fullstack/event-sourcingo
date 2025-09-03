package helpers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const (
	numberTwo = 2
	maxIntVal = 9
)

func RandomInt(n int) (string, error) {
	digits := make([]*big.Int, n+numberTwo)
	var err error
	for i := range digits {
		digits[i], err = rand.Int(rand.Reader, big.NewInt(maxIntVal))
		if err != nil {
			return "", err
		}
	}
	fields := strings.Fields(fmt.Sprint(digits))
	return strings.Join(fields[1:len(fields)-1], ""), nil
}
