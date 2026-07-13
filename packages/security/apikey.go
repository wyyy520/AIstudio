package security

import (
	"fmt"
	"strings"
)

const APIKeyPrefix = "ask-"
const APIKeyLength = 48

func GenerateAPIKey() (string, error) {
	randomPart, err := RandomHex(APIKeyLength - len(APIKeyPrefix))
	if err != nil {
		return "", fmt.Errorf("generate API key: %w", err)
	}
	return APIKeyPrefix + randomPart, nil
}

func ValidateAPIKeyFormat(key string) bool {
	if !strings.HasPrefix(key, APIKeyPrefix) {
		return false
	}
	randomPart := strings.TrimPrefix(key, APIKeyPrefix)
	if len(randomPart) == 0 {
		return false
	}
	for _, c := range randomPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func MaskKey(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func KeyPrefix(key string) string {
	if len(key) >= 8 {
		return key[:8]
	}
	return key
}
