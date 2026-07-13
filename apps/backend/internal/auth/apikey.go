package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/aistudio/backend/internal/database/models"
	"gorm.io/gorm"
)

type APIKeyManager struct {
	db       *gorm.DB
	encKey   []byte
}

func NewAPIKeyManager(db *gorm.DB, encryptionKey string) *APIKeyManager {
	hash := sha256.Sum256([]byte(encryptionKey))
	return &APIKeyManager{
		db:     db,
		encKey: hash[:],
	}
}

type CreateAPIKeyParams struct {
	UserID   uint
	Provider string
	Name     string
	Key      string
}

func (m *APIKeyManager) Create(params CreateAPIKeyParams) (*models.APIKey, error) {
	encrypted, err := m.encrypt(params.Key)
	if err != nil {
		return nil, fmt.Errorf("encrypt API key: %w", err)
	}

	prefix := ""
	if len(params.Key) >= 8 {
		prefix = params.Key[:8]
	}

	apiKey := &models.APIKey{
		UserID:    params.UserID,
		Provider:  params.Provider,
		Name:      params.Name,
		KeyPrefix: prefix,
		KeyHash:   encrypted,
		Status:    "active",
	}
	if err := m.db.Create(apiKey).Error; err != nil {
		return nil, fmt.Errorf("create API key: %w", err)
	}
	return apiKey, nil
}

func (m *APIKeyManager) GetByID(id uint) (*models.APIKey, error) {
	var apiKey models.APIKey
	if err := m.db.First(&apiKey, id).Error; err != nil {
		return nil, ErrAPIKeyNotFound
	}
	return &apiKey, nil
}

func (m *APIKeyManager) GetByUser(userID uint) ([]models.APIKey, error) {
	var keys []models.APIKey
	if err := m.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("get user API keys: %w", err)
	}
	return keys, nil
}

func (m *APIKeyManager) GetByUserAndProvider(userID uint, provider string) (*models.APIKey, error) {
	var apiKey models.APIKey
	if err := m.db.Where("user_id = ? AND provider = ? AND status = ?",
		userID, provider, "active").First(&apiKey).Error; err != nil {
		return nil, ErrAPIKeyNotFound
	}
	return &apiKey, nil
}

func (m *APIKeyManager) DecryptKey(apiKey *models.APIKey) (string, error) {
	return m.decrypt(apiKey.KeyHash)
}

func (m *APIKeyManager) Delete(id uint, userID uint) error {
	result := m.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.APIKey{})
	if result.Error != nil {
		return fmt.Errorf("delete API key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

func (m *APIKeyManager) UpdateStatus(id uint, userID uint, status string) error {
	validStatus := map[string]bool{"active": true, "disabled": true}
	if !validStatus[status] {
		return fmt.Errorf("invalid status: %s", status)
	}
	result := m.db.Model(&models.APIKey{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("update API key status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

func (m *APIKeyManager) ProviderList() []string {
	return []string{"openai", "claude", "gemini", "deepseek", "moonshot", "qwen", "glm", "baidu", "tencent"}
}

func (m *APIKeyManager) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(m.encKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (m *APIKeyManager) decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(m.encKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, cipherBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func MaskKey(key string) string {
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}
