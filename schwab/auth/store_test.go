package auth

import (
	"context"
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

	t.Run("load missing file returns auth required", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		path := filepath.Join(t.TempDir(), "missing-token.json")
		store := NewFileTokenStore(path)

		_, err := store.Load(ctx)
		require.Error(t, err)

		var authRequired *AuthRequiredError
		assert.ErrorAs(t, err, &authRequired)
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
