package subscription

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnitNewCache(t *testing.T) {
	c := NewCache()
	require.NotNil(t, c)
	require.NotNil(t, c.data)
}

func TestUnitAddAndGetItems(t *testing.T) {
	c := NewCache()

	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()

	c.AddItems("key-1", id1, id2)
	c.AddItems("key-2", id3)
	c.AddItems("key-3")

	t.Run("length is valid", func(t *testing.T) {
		items, ok := c.GetItems("key-1")
		require.True(t, ok)
		require.Len(t, items, 2)

		itemsByKey3, ok := c.GetItems("key-3")
		require.True(t, ok)
		require.Len(t, itemsByKey3, 0)
	})

	t.Run("contains correct values", func(t *testing.T) {
		items, ok := c.GetItems("key-1")
		require.True(t, ok)
		require.Contains(t, items, id1)
		require.Contains(t, items, id2)
	})
}

func TestUnitRemoveKey(t *testing.T) {
	c := NewCache()

	c.AddItems("key-1", uuid.New(), uuid.New())
	c.RemoveKey("key-1")

	items, ok := c.GetItems("key-1")
	require.False(t, ok)
	require.Empty(t, items)
}

func TestUnitRemoveItem(t *testing.T) {
	c := NewCache()

	id1 := uuid.New()
	id2 := uuid.New()

	c.AddItems("key-1", id1, id2)
	c.RemoveItem("key-1", id1)

	items, ok := c.GetItems("key-1")
	require.True(t, ok)
	require.Len(t, items, 1)
	require.Equal(t, items, []uuid.UUID{id2})
}
