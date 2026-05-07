package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

const tokenFilePerm = 0o600

var _ TokenStore = (*FileTokenStore)(nil)

// FileTokenStore persists OAuth2 tokens to a JSON file.
type FileTokenStore struct {
	mu   sync.Mutex
	path string
}

// NewFileTokenStore returns a token store backed by path.
func NewFileTokenStore(path string) *FileTokenStore {
	return &FileTokenStore{path: path}
}

// Save writes tf to disk using a temporary file and atomic rename.
func (s *FileTokenStore) Save(ctx context.Context, tf TokenFile) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(tf)
	if err != nil {
		return fmt.Errorf("marshal token file: %w", err)
	}

	tmpPath := s.path + ".tmp"
	if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove stale temporary token file: %w", err)
	}

	tmpFile, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, tokenFilePerm)
	if err != nil {
		return fmt.Errorf("write temporary token file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temporary token file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temporary token file: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("replace token file: %w", err)
	}

	return nil
}

// Load reads and decodes a persisted token file from disk.
func (s *FileTokenStore) Load(ctx context.Context) (TokenFile, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return TokenFile{}, &AuthRequiredError{}
		}
		return TokenFile{}, fmt.Errorf("read token file: %w", err)
	}

	var tf TokenFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return TokenFile{}, fmt.Errorf("parse token file: %w", err)
	}

	return tf, nil
}
