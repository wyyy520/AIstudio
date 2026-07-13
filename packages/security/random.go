package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("random string: %w", err)
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

func RandomHex(length int) (string, error) {
	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("random hex: %w", err)
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func RandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("random bytes: %w", err)
	}
	return b, nil
}
