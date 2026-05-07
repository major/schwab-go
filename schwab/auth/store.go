package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	authRequiredLoginMessage = "no token file found, login required"
	tokenDirPerm             = 0o700
	tokenFilePerm            = 0o600
)

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

// Save writes tf to disk using a temporary file and replace operation.
func (s *FileTokenStore) Save(ctx context.Context, tf TokenFile) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(tf)
	if err != nil {
		return fmt.Errorf("marshal token file: %w", err)
	}
	err = os.MkdirAll(filepath.Dir(s.path), tokenDirPerm)
	if err != nil {
		return fmt.Errorf("create token file directory: %w", err)
	}

	tmpPath := s.path + ".tmp"
	err = os.Remove(tmpPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove stale temporary token file: %w", err)
	}

	tmpFile, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, tokenFilePerm)
	if err != nil {
		return fmt.Errorf("create temporary token file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	n, err := tmpFile.Write(data)
	if err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temporary token file: %w", err)
	}
	if n != len(data) {
		_ = tmpFile.Close()
		return fmt.Errorf("write temporary token file: %w (%d/%d bytes)", io.ErrShortWrite, n, len(data))
	}

	err = tmpFile.Sync()
	if err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("sync temporary token file: %w", err)
	}

	err = tmpFile.Close()
	if err != nil {
		return fmt.Errorf("close temporary token file: %w", err)
	}

	err = replaceTokenFile(tmpPath, s.path, os.Rename)
	if err != nil {
		return fmt.Errorf("replace token file: %w", err)
	}

	err = syncParentDir(s.path)
	if err != nil {
		return fmt.Errorf("sync token file directory: %w", err)
	}

	return nil
}

func replaceTokenFile(tmpPath, targetPath string, rename func(string, string) error) error {
	renameErr := rename(tmpPath, targetPath)
	if renameErr == nil {
		return nil
	}

	// os.Rename replaces existing files on Unix, but not on all supported
	// platforms. Keep the same-directory rename as the durable fast path, and use
	// a remove-then-rename fallback only when replacing an existing token file.
	_, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return renameErr
		}
		return errors.Join(renameErr, fmt.Errorf("stat token file: %w", err))
	}

	err = os.Remove(targetPath)
	if err != nil {
		return errors.Join(renameErr, fmt.Errorf("remove token file before fallback rename: %w", err))
	}

	err = rename(tmpPath, targetPath)
	if err != nil {
		return errors.Join(renameErr, fmt.Errorf("fallback rename token file: %w", err))
	}

	return nil
}

func syncParentDir(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}

	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()

	return dir.Sync()
}

// Load reads and decodes a persisted token file from disk.
func (s *FileTokenStore) Load(ctx context.Context) (TokenFile, error) {
	if err := ctx.Err(); err != nil {
		return TokenFile{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return TokenFile{}, errors.Join(&AuthRequiredError{Msg: authRequiredLoginMessage}, err)
		}
		return TokenFile{}, fmt.Errorf("read token file: %w", err)
	}

	var tf TokenFile
	err = json.Unmarshal(data, &tf)
	if err != nil {
		return TokenFile{}, fmt.Errorf("parse token file: %w", err)
	}

	return tf, nil
}
