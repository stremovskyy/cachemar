package tests

import (
	"context"
	"github.com/stremovskyy/cachemar"
	"github.com/stremovskyy/cachemar/drivers/memory"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemoryCacheService(t *testing.T) {
	cache := memory.New()
	ctx := context.Background()
	tags := []string{"tag1", "tag2"}

	t.Run("Set and Get", func(t *testing.T) {
		err := cache.Set(ctx, "key1", "value1", 10*time.Minute, tags)
		require.NoError(t, err)

		var value string
		err = cache.Get(ctx, "key1", &value)
		require.NoError(t, err)
		require.Equal(t, "value1", value)
	})

	t.Run("Remove", func(t *testing.T) {
		err := cache.Remove(ctx, "key1")
		require.NoError(t, err)

		var value string
		err = cache.Get(ctx, "key1", &value)
		require.Equal(t, cachemar.ErrNotFound, err)
	})

	t.Run("Exists", func(t *testing.T) {
		exists, err := cache.Exists(ctx, "key1")
		require.NoError(t, err)
		require.False(t, exists)

		err = cache.Set(ctx, "key1", "value1", 10*time.Minute, tags)
		require.NoError(t, err)

		exists, err = cache.Exists(ctx, "key1")
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("Increment and Decrement", func(t *testing.T) {
		err := cache.Set(ctx, "key1", 0, 10*time.Minute, tags)
		require.NoError(t, err)

		err = cache.Increment(ctx, "key1")
		require.NoError(t, err)

		var value string
		err = cache.Get(ctx, "key1", &value)
		require.NoError(t, err)
		require.Equal(t, "1", value)

		err = cache.Decrement(ctx, "key1")
		require.NoError(t, err)

		err = cache.Get(ctx, "key1", &value)
		require.NoError(t, err)
		require.Equal(t, "0", value)
	})

	t.Run("GetKeysByTag", func(t *testing.T) {
		err := cache.Set(ctx, "key1", "value1", 10*time.Minute, tags)
		require.NoError(t, err)
		err = cache.Set(ctx, "key2", "value2", 10*time.Minute, tags)
		require.NoError(t, err)

		keys, err := cache.GetKeysByTag(ctx, "tag1")
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"key1", "key2"}, keys)
	})

	t.Run("RemoveByTag and RemoveByTags", func(t *testing.T) {
		err := cache.RemoveByTag(ctx, "tag1")
		require.NoError(t, err)

		var value string
		err = cache.Get(ctx, "key1", &value)
		require.Equal(t, cachemar.ErrNotFound, err)

		err = cache.Set(ctx, "key1", "value1", 10*time.Minute, []string{"tag1"})
		require.NoError(t, err)
		err = cache.Set(ctx, "key2", "value2", 10*time.Minute, []string{"tag2"})
		require.NoError(t, err)

		err = cache.RemoveByTags(ctx, []string{"tag1", "tag2"})
		require.NoError(t, err)

		err = cache.Get(ctx, "key1", &value)
		require.Equal(t, cachemar.ErrNotFound, err)
		err = cache.Get(ctx, "key2", &value)
		require.Equal(t, cachemar.ErrNotFound, err)
	})
}
