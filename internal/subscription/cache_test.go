package subscription

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitNewCache(t *testing.T) {
	c := NewCache()
	require.NotNil(t, c)
	require.NotNil(t, c.data)
}

func TestUnitAddAndGetItems(t *testing.T) {
	c := NewCache()
	c.AddItems("key-1", "val-1", "val-2")
	c.AddItems("key-2", "val-0")
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
		require.Contains(t, items, "val-1")
		require.Contains(t, items, "val-2")
	})
}

func TestUnitRemoveKey(t *testing.T) {
	c := NewCache()

	c.AddItems("key-1", "val-1", "val-2")
	c.RemoveKey("key-1")

	items, ok := c.GetItems("key-1")
	require.False(t, ok)
	require.Empty(t, items)
}

func TestUnitRemoveItem(t *testing.T) {
	c := NewCache()

	c.AddItems("key-1", "val-1", "val-2")
	c.RemoveItem("key-1", "val-2")

	items, ok := c.GetItems("key-1")
	require.True(t, ok)
	require.Len(t, items, 1)
	require.Equal(t, items, []string{"val-1"})
}
