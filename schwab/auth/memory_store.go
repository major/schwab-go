package auth

import (
	"context"
	"sync"
)

var _ TokenStore = (*MemoryTokenStore)(nil)

// MemoryTokenStore persists OAuth2 tokens in process memory.
//
// MemoryTokenStore is safe for concurrent use. It is intended for tests,
// examples, and short-lived applications that do not need token durability
// across process restarts. Its zero value is ready to use.
type MemoryTokenStore struct {
	mu       sync.Mutex
	token    TokenFile
	hasToken bool
}

// NewMemoryTokenStore returns an empty in-memory token store.
func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{}
}

// Save stores tf in memory for later Load calls.
func (s *MemoryTokenStore) Save(ctx context.Context, tf TokenFile) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.token = tf
	s.hasToken = true
	return nil
}

// Load returns the most recently saved token file.
func (s *MemoryTokenStore) Load(ctx context.Context) (TokenFile, error) {
	if err := ctx.Err(); err != nil {
		return TokenFile{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.hasToken {
		return TokenFile{}, &AuthRequiredError{Msg: authRequiredLoginMessage}
	}

	return s.token, nil
}
