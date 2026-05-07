package auth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileTokenStore(t *testing.T) {
	t.Parallel()

	t.Run("save then load preserves all fields", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		path := filepath.Join(t.TempDir(), "token.json")
		store := NewFileTokenStore(path)
		want := testTokenFile()

		require.NoError(t, store.Save(ctx, want))

		got, err := store.Load(ctx)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("save writes file with owner only permissions", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		path := filepath.Join(t.TempDir(), "token.json")
		store := NewFileTokenStore(path)

		require.NoError(t, store.Save(ctx, testTokenFile()))

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("save creates parent directory with owner only permissions", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		path := filepath.Join(t.TempDir(), "nested", "token.json")
		store := NewFileTokenStore(path)

		require.NoError(t, store.Save(ctx, testTokenFile()))

		info, err := os.Stat(filepath.Dir(path))
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
	})

	t.Run("save returns directory creation errors", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		basePath := filepath.Join(t.TempDir(), "token-parent")
		require.NoError(t, os.WriteFile(basePath, []byte("not a directory"), 0o600))
		store := NewFileTokenStore(filepath.Join(basePath, "token.json"))

		err := store.Save(ctx, testTokenFile())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create token file directory")
	})

	t.Run("load missing file returns auth required", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		path := filepath.Join(t.TempDir(), "missing-token.json")
		store := NewFileTokenStore(path)

		_, err := store.Load(ctx)
		require.Error(t, err)

		var authRequired *AuthRequiredError
		require.ErrorAs(t, err, &authRequired)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("load invalid json returns parse error", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		path := filepath.Join(t.TempDir(), "token.json")
		require.NoError(t, os.WriteFile(path, []byte("not json"), 0o600))

		store := NewFileTokenStore(path)
		_, err := store.Load(ctx)
		require.Error(t, err)

		var authRequired *AuthRequiredError
		assert.NotErrorAs(t, err, &authRequired)
	})

	t.Run("save removes temporary file after rename", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		path := filepath.Join(t.TempDir(), "token.json")
		store := NewFileTokenStore(path)

		require.NoError(t, store.Save(ctx, testTokenFile()))

		_, err := os.Stat(path + ".tmp")
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("save replaces stale temporary file with owner only permissions", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		path := filepath.Join(t.TempDir(), "token.json")
		staleTempPath := path + ".tmp"
		require.NoError(t, os.WriteFile(staleTempPath, []byte("stale token"), 0o644))

		store := NewFileTokenStore(path)
		require.NoError(t, store.Save(ctx, testTokenFile()))

		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

		_, err = os.Stat(staleTempPath)
		require.ErrorIs(t, err, os.ErrNotExist)
	})
}

func TestReplaceTokenFileFallsBackToRemoveThenRename(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	tmpPath := filepath.Join(dir, "token.json.tmp")
	targetPath := filepath.Join(dir, "token.json")
	require.NoError(t, os.WriteFile(tmpPath, []byte("new token"), 0o600))
	require.NoError(t, os.WriteFile(targetPath, []byte("old token"), 0o644))

	renameErr := errors.New("rename replacement unsupported")
	renames := 0
	err := replaceTokenFile(tmpPath, targetPath, func(oldPath, newPath string) error {
		renames++
		if renames == 1 {
			return renameErr
		}
		return os.Rename(oldPath, newPath)
	})
	require.NoError(t, err)

	contents, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, "new token", string(contents))

	info, err := os.Stat(targetPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	_, err = os.Stat(tmpPath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestReplaceTokenFileReturnsRenameErrorWhenTargetIsMissing(t *testing.T) {
	t.Parallel()

	renameErr := errors.New("rename replacement unsupported")
	targetPath := filepath.Join(t.TempDir(), "token.json")
	err := replaceTokenFile(filepath.Join(t.TempDir(), "missing.tmp"), targetPath, func(string, string) error {
		return renameErr
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, renameErr)
}

func TestReplaceTokenFileReportsFallbackRenameErrors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	tmpPath := filepath.Join(dir, "token.json.tmp")
	targetPath := filepath.Join(dir, "token.json")
	require.NoError(t, os.WriteFile(tmpPath, []byte("new token"), 0o600))
	require.NoError(t, os.WriteFile(targetPath, []byte("old token"), 0o600))

	renameErr := errors.New("rename replacement unsupported")
	fallbackErr := errors.New("fallback rename failed")
	renames := 0
	err := replaceTokenFile(tmpPath, targetPath, func(string, string) error {
		renames++
		if renames == 1 {
			return renameErr
		}
		return fallbackErr
	})
	require.Error(t, err)
	require.ErrorIs(t, err, renameErr)
	require.ErrorIs(t, err, fallbackErr)
	assert.Contains(t, err.Error(), "fallback rename token file")
}

func TestSyncParentDir(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "token.json")
	require.NoError(t, os.WriteFile(path, []byte("token"), tokenFilePerm))

	require.NoError(t, syncParentDir(path))
}

func testTokenFile() TokenFile {
	return TokenFile{
		CreationTimestamp: 1715000000,
		Token: TokenData{
			AccessToken:  "access-token",
			TokenType:    "Bearer",
			ExpiresIn:    1800,
			RefreshToken: "refresh-token",
			Scope:        "readonly",
			ExpiresAt:    1715001800,
		},
	}
}
