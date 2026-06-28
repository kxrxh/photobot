package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTTLCache(t *testing.T) {
	ttl := 5 * time.Minute
	c := NewTTLCache[string, int](ttl)

	require.NotNil(t, c)
	assert.Equal(t, ttl, c.ttl)
	assert.NotNil(t, c.entries)
}

func TestTTLCache_Get_EmptyCache(t *testing.T) {
	c := NewTTLCache[string, string](time.Minute)

	val, ok := c.Get("missing")
	assert.False(t, ok, "Get on empty cache should return false")
	assert.Empty(t, val)
}

func TestTTLCache_Get_NonExistentKey(t *testing.T) {
	c := NewTTLCache[string, int](time.Minute)
	c.Set("existing", 42)

	val, ok := c.Get("missing")
	assert.False(t, ok, "Get on non-existent key should return false")
	assert.Zero(t, val)
}

func TestTTLCache_SetGet_HappyPath(t *testing.T) {
	c := NewTTLCache[string, string](time.Minute)

	c.Set("key1", "value1")
	val, ok := c.Get("key1")
	assert.True(t, ok, "Get should find freshly set value")
	assert.Equal(t, "value1", val)

	c.Set("key2", "value2")
	val2, ok2 := c.Get("key2")
	assert.True(t, ok2)
	assert.Equal(t, "value2", val2)
}

func TestTTLCache_SetGet_Overwrite(t *testing.T) {
	c := NewTTLCache[string, int](time.Minute)

	c.Set("key", 1)
	val, ok := c.Get("key")
	assert.True(t, ok)
	assert.Equal(t, 1, val)

	c.Set("key", 99)
	val, ok = c.Get("key")
	assert.True(t, ok)
	assert.Equal(t, 99, val)
}

func TestTTLCache_Get_ExpiredEntry(t *testing.T) {
	ttl := 10 * time.Millisecond
	c := NewTTLCache[string, string](ttl)

	c.Set("key", "value")
	val, ok := c.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "value", val)

	time.Sleep(ttl + 5*time.Millisecond)

	val, ok = c.Get("key")
	assert.False(t, ok, "Get should return false for expired entry")
	assert.Empty(t, val)
}

func TestTTLCache_Get_StructValue(t *testing.T) {
	type item struct {
		ID   int
		Name string
	}
	c := NewTTLCache[string, item](time.Minute)

	c.Set("a", item{ID: 1, Name: "foo"})
	val, ok := c.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, val.ID)
	assert.Equal(t, "foo", val.Name)

	val, ok = c.Get("b")
	assert.False(t, ok)
	assert.Zero(t, val.ID)
	assert.Empty(t, val.Name)
}

func TestTTLCache_IntKey(t *testing.T) {
	c := NewTTLCache[int, string](time.Minute)

	c.Set(1, "one")
	c.Set(2, "two")

	val, ok := c.Get(1)
	assert.True(t, ok)
	assert.Equal(t, "one", val)

	val, ok = c.Get(2)
	assert.True(t, ok)
	assert.Equal(t, "two", val)

	_, ok = c.Get(99)
	assert.False(t, ok)
}

func TestTTLCache_ConcurrentAccess(t *testing.T) {
	c := NewTTLCache[string, int](time.Minute)
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Set("key", n)
		}(i)
	}

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Get("key")
		}()
	}

	wg.Wait()
	val, ok := c.Get("key")
	assert.True(t, ok)
	assert.GreaterOrEqual(t, val, 0)
	assert.Less(t, val, 100)
}

func TestTTLCache_ConcurrentDifferentKeys(t *testing.T) {
	c := NewTTLCache[string, int](time.Minute)
	var wg sync.WaitGroup

	for i := range 50 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := string(rune('a' + (n % 26)))
			c.Set(key, n)
			c.Get(key)
		}(i)
	}

	wg.Wait()
}
