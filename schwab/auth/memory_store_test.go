package auth

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryTokenStore(t *testing.T) {
	t.Parallel()

	t.Run("load before save returns auth required", func(t *testing.T) {
		t.Parallel()

		store := NewMemoryTokenStore()
		_, err := store.Load(context.Background())
		require.Error(t, err)

		var authRequired *AuthRequiredError
		require.ErrorAs(t, err, &authRequired)
		assert.Equal(t, authRequiredLoginMessage, authRequired.Msg)
	})

	t.Run("save then load preserves all fields", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := NewMemoryTokenStore()
		want := testTokenFile()

		require.NoError(t, store.Save(ctx, want))

		got, err := store.Load(ctx)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("zero value can save and load", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		var store MemoryTokenStore
		want := testTokenFile()

		require.NoError(t, store.Save(ctx, want))

		got, err := store.Load(ctx)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("save replaces previous token", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := NewMemoryTokenStore()
		first := testTokenFile()
		second := testTokenFile()
		second.Token.AccessToken = "replacement-access-token"

		require.NoError(t, store.Save(ctx, first))
		require.NoError(t, store.Save(ctx, second))

		got, err := store.Load(ctx)
		require.NoError(t, err)
		assert.Equal(t, second, got)
	})

	t.Run("save respects canceled context", func(t *testing.T) {
		t.Parallel()

		store := NewMemoryTokenStore()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := store.Save(ctx, testTokenFile())
		require.ErrorIs(t, err, context.Canceled)

		_, err = store.Load(context.Background())
		require.Error(t, err)
		var authRequired *AuthRequiredError
		assert.ErrorAs(t, err, &authRequired)
	})

	t.Run("load respects canceled context", func(t *testing.T) {
		t.Parallel()

		store := NewMemoryTokenStore()
		require.NoError(t, store.Save(context.Background(), testTokenFile()))
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := store.Load(ctx)
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("safe for concurrent use", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		store := NewMemoryTokenStore()
		require.NoError(t, store.Save(ctx, testTokenFile()))

		const workers = 16
		errCh := make(chan error, workers)
		var wg sync.WaitGroup
		for i := range workers {
			wg.Go(func() {
				tf := testTokenFile()
				tf.Token.AccessToken = fmt.Sprintf("access-token-%d", i)
				if err := store.Save(ctx, tf); err != nil {
					errCh <- fmt.Errorf("MemoryTokenStore.Save(worker=%d): %w", i, err)
					return
				}

				if _, err := store.Load(ctx); err != nil {
					errCh <- fmt.Errorf("MemoryTokenStore.Load(worker=%d): %w", i, err)
				}
			})
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			assert.NoError(t, err)
		}
	})
}

func TestMemoryTokenStoreDoesNotExposeFileErrors(t *testing.T) {
	t.Parallel()

	_, err := NewMemoryTokenStore().Load(context.Background())
	require.Error(t, err)
	assert.NotErrorIs(t, err, os.ErrNotExist)
}
