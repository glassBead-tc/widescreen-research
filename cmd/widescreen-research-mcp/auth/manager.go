package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// SecretStore interface for secret storage backends
type SecretStore interface {
	Retrieve(key string) (string, error)
	Store(key, value string) error
	ListExpiring(duration time.Duration) []Secret
	Delete(key string) error
}

// Secret represents a stored secret
type Secret struct {
	Name      string
	Type      string
	Value     string
	ExpiresAt time.Time
}

// TokenCache provides in-memory caching for tokens
type TokenCache struct {
	tokens map[string]*CachedToken
	mu     sync.RWMutex
}

// CachedToken represents a cached token with expiry
type CachedToken struct {
	Value     string
	ExpiresAt time.Time
}

// AuthManager manages authentication and secrets
type AuthManager struct {
	secretStore SecretStore
	tokenCache  *TokenCache
	encryptKey  []byte
	mu          sync.RWMutex
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() (*AuthManager, error) {
	// Generate or retrieve encryption key
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	return &AuthManager{
		secretStore: NewEnvSecretStore(), // Default to environment variables
		tokenCache: &TokenCache{
			tokens: make(map[string]*CachedToken),
		},
		encryptKey: key,
	}, nil
}

// GetCredential retrieves a credential with caching
func (a *AuthManager) GetCredential(ctx context.Context, key string) (string, error) {
	// Check cache first
	if token := a.tokenCache.Get(key); token != nil && !token.IsExpired() {
		return token.Value, nil
	}

	// Retrieve from secure store
	secret, err := a.secretStore.Retrieve(key)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve credential %s: %w", key, err)
	}

	// Validate credential
	if err := a.validateCredential(key, secret); err != nil {
		return "", fmt.Errorf("invalid credential %s: %w", key, err)
	}

	// Cache with appropriate TTL
	ttl := a.getCredentialTTL(key)
	a.tokenCache.Set(key, secret, ttl)

	return secret, nil
}

// RotateSecrets performs zero-downtime secret rotation
func (a *AuthManager) RotateSecrets(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Get secrets expiring within 7 days
	expiringSecrets := a.secretStore.ListExpiring(7 * 24 * time.Hour)

	for _, secret := range expiringSecrets {
		// Generate new secret
		newSecret, err := a.generateNewSecret(secret.Type)
		if err != nil {
			return fmt.Errorf("failed to generate new secret for %s: %w", secret.Name, err)
		}

		// Perform zero-downtime rotation
		if err := a.performRotation(ctx, secret, newSecret); err != nil {
			return fmt.Errorf("rotation failed for %s: %w", secret.Name, err)
		}
	}

	return nil
}

// validateCredential validates a credential
func (a *AuthManager) validateCredential(key, value string) error {
	if value == "" {
		return fmt.Errorf("credential is empty")
	}

	// Additional validation based on credential type
	switch key {
	case "CLAUDE_API_KEY":
		if len(value) < 20 {
			return fmt.Errorf("API key too short")
		}
	case "GCP_SERVICE_ACCOUNT_JSON":
		// Validate JSON structure
		if value[0] != '{' {
			return fmt.Errorf("invalid JSON format")
		}
	}

	return nil
}

// getCredentialTTL returns the TTL for a credential type
func (a *AuthManager) getCredentialTTL(key string) time.Duration {
	switch key {
	case "GITHUB_TOKEN", "GOOGLE_OAUTH_TOKEN":
		return 1 * time.Hour // Short-lived tokens
	default:
		return 24 * time.Hour // Default TTL
	}
}

// generateNewSecret generates a new secret of the given type
func (a *AuthManager) generateNewSecret(secretType string) (string, error) {
	// In production, this would integrate with secret generation services
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// performRotation performs zero-downtime rotation
func (a *AuthManager) performRotation(ctx context.Context, old Secret, new string) error {
	// Phase 1: Store new secret alongside old
	tempKey := old.Name + "_new"
	if err := a.secretStore.Store(tempKey, new); err != nil {
		return fmt.Errorf("failed to store new secret: %w", err)
	}

	// Phase 2: Test new secret
	if err := a.validateCredential(old.Name, new); err != nil {
		a.secretStore.Delete(tempKey)
		return fmt.Errorf("new secret validation failed: %w", err)
	}

	// Phase 3: Atomic swap
	if err := a.secretStore.Store(old.Name, new); err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}

	// Phase 4: Cleanup
	a.secretStore.Delete(tempKey)
	a.tokenCache.Delete(old.Name)

	return nil
}

// Encrypt encrypts data using AES-256-GCM
func (a *AuthManager) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(a.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// TokenCache methods

// Get retrieves a token from cache
func (tc *TokenCache) Get(key string) *CachedToken {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.tokens[key]
}

// Set stores a token in cache
func (tc *TokenCache) Set(key, value string, ttl time.Duration) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tokens[key] = &CachedToken{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a token from cache
func (tc *TokenCache) Delete(key string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	delete(tc.tokens, key)
}

// IsExpired checks if token is expired
func (ct *CachedToken) IsExpired() bool {
	return time.Now().After(ct.ExpiresAt)
}

// EnvSecretStore implements SecretStore using environment variables
type EnvSecretStore struct{}

// NewEnvSecretStore creates a new environment variable secret store
func NewEnvSecretStore() *EnvSecretStore {
	return &EnvSecretStore{}
}

// Retrieve gets a secret from environment
func (e *EnvSecretStore) Retrieve(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not set", key)
	}
	return value, nil
}

// Store sets an environment variable (for testing)
func (e *EnvSecretStore) Store(key, value string) error {
	return os.Setenv(key, value)
}

// ListExpiring returns secrets expiring soon (not applicable for env vars)
func (e *EnvSecretStore) ListExpiring(duration time.Duration) []Secret {
	return []Secret{} // Env vars don't expire
}

// Delete removes an environment variable
func (e *EnvSecretStore) Delete(key string) error {
	return os.Unsetenv(key)
}