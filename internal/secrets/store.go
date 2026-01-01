package secrets

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/99designs/keyring"
)

const (
	serviceName = "line-cli"
)

// storedCredentials is the internal format stored in keyring (includes metadata)
type storedCredentials struct {
	ChannelAccessToken string    `json:"channel_access_token"`
	ChannelID          string    `json:"channel_id,omitempty"`
	ChannelSecret      string    `json:"channel_secret,omitempty"`
	CreatedAt          time.Time `json:"created_at,omitempty"`
	IsPrimary          bool      `json:"is_primary,omitempty"`
	BotName            string    `json:"bot_name,omitempty"`
}

// Credentials holds the authentication information for a LINE Official Account
type Credentials struct {
	ChannelAccessToken string `json:"-"` // Never serialize to JSON responses
	ChannelID          string `json:"channel_id,omitempty"`
	ChannelSecret      string `json:"channel_secret,omitempty"`
}

// AccountInfo represents a stored account
type AccountInfo struct {
	Name      string
	CreatedAt time.Time
	IsPrimary bool
	BotName   string
}

// Store provides secure credential storage
type Store interface {
	Set(name string, creds Credentials, botName string) error
	Get(name string) (*Credentials, error)
	Delete(name string) error
	List() ([]AccountInfo, error)
	SetPrimary(name string) error
	GetPrimary() (string, error)
}

// KeychainStore implements Store using the system keychain
type KeychainStore struct {
	ring keyring.Keyring
}

// NewKeychainStore creates a new keychain-backed store
func NewKeychainStore() (*KeychainStore, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName:              serviceName,
		KeychainTrustApplication: true,
		// Fallback to file-based if no keychain available
		FileDir:          "~/.line-cli/credentials",
		FilePasswordFunc: keyring.FixedStringPrompt(""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	return &KeychainStore{ring: ring}, nil
}

// Set stores credentials for an account
func (s *KeychainStore) Set(name string, creds Credentials, botName string) error {
	name = normalize(name)

	// Check if this is the first account (auto-set as primary)
	isPrimary := false
	accounts, err := s.List()
	if err == nil && len(accounts) == 0 {
		isPrimary = true
	}

	stored := storedCredentials{
		ChannelAccessToken: creds.ChannelAccessToken,
		ChannelID:          creds.ChannelID,
		ChannelSecret:      creds.ChannelSecret,
		CreatedAt:          time.Now().UTC(),
		IsPrimary:          isPrimary,
		BotName:            botName,
	}

	data, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	err = s.ring.Set(keyring.Item{
		Key:  tokenKey(name),
		Data: data,
	})
	if err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	return nil
}

// Get retrieves credentials for an account
func (s *KeychainStore) Get(name string) (*Credentials, error) {
	name = normalize(name)
	if name == "" {
		return nil, fmt.Errorf("account name cannot be empty")
	}

	item, err := s.ring.Get(tokenKey(name))
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil, fmt.Errorf("account not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	var stored storedCredentials
	if err := json.Unmarshal(item.Data, &stored); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	creds := &Credentials{
		ChannelAccessToken: stored.ChannelAccessToken,
		ChannelID:          stored.ChannelID,
		ChannelSecret:      stored.ChannelSecret,
	}

	return creds, nil
}

// Delete removes credentials for an account
func (s *KeychainStore) Delete(name string) error {
	name = normalize(name)

	err := s.ring.Remove(tokenKey(name))
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete credentials: %w", err)
	}
	return nil
}

// List returns all stored accounts
func (s *KeychainStore) List() ([]AccountInfo, error) {
	keys, err := s.ring.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}

	var accounts []AccountInfo
	for _, key := range keys {
		name, ok := parseTokenKey(key)
		if !ok {
			continue // Skip non-token keys
		}

		item, err := s.ring.Get(key)
		if err != nil {
			continue // Skip if we can't read the item
		}

		var stored storedCredentials
		if err := json.Unmarshal(item.Data, &stored); err != nil {
			// Fallback for legacy entries without metadata
			accounts = append(accounts, AccountInfo{Name: name})
			continue
		}

		accounts = append(accounts, AccountInfo{
			Name:      name,
			CreatedAt: stored.CreatedAt,
			IsPrimary: stored.IsPrimary,
			BotName:   stored.BotName,
		})
	}

	return accounts, nil
}

// SetPrimary sets the specified account as the primary account
func (s *KeychainStore) SetPrimary(name string) error {
	name = normalize(name)

	// Get all keys
	keys, err := s.ring.Keys()
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	found := false
	for _, key := range keys {
		accountName, ok := parseTokenKey(key)
		if !ok {
			continue
		}

		item, err := s.ring.Get(key)
		if err != nil {
			continue
		}

		var stored storedCredentials
		if err := json.Unmarshal(item.Data, &stored); err != nil {
			continue
		}

		// Set IsPrimary based on whether this is the target account
		isPrimary := accountName == name
		if isPrimary {
			found = true
		}

		if stored.IsPrimary != isPrimary {
			stored.IsPrimary = isPrimary
			data, err := json.Marshal(stored)
			if err != nil {
				return fmt.Errorf("failed to marshal credentials for %s: %w", accountName, err)
			}
			if err := s.ring.Set(keyring.Item{
				Key:  key,
				Data: data,
			}); err != nil {
				return fmt.Errorf("failed to update primary status for %s: %w", accountName, err)
			}
		}
	}

	if !found {
		return fmt.Errorf("account not found: %s", name)
	}

	return nil
}

// GetPrimary returns the name of the primary account.
// If no account is explicitly marked as primary, it falls back to returning
// the first account in the list. This ensures single-account setups work
// without requiring explicit primary designation.
func (s *KeychainStore) GetPrimary() (string, error) {
	accounts, err := s.List()
	if err != nil {
		return "", err
	}

	if len(accounts) == 0 {
		return "", nil
	}

	// Find the primary account
	for _, account := range accounts {
		if account.IsPrimary {
			return account.Name, nil
		}
	}

	// Fallback: return the first account if none is marked as primary
	return accounts[0].Name, nil
}

// tokenKey returns the keyring key for a token
func tokenKey(name string) string {
	return fmt.Sprintf("token:%s", name)
}

// parseTokenKey extracts the account name from a token key
func parseTokenKey(key string) (string, bool) {
	const prefix = "token:"
	if !strings.HasPrefix(key, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(key, prefix)
	if strings.TrimSpace(rest) == "" {
		return "", false
	}
	return rest, true
}

// normalize normalizes an account name by trimming whitespace and converting
// to lowercase. Case-insensitivity is intentional to prevent duplicate accounts
// with different casing (e.g., "MyBot" vs "mybot").
func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
