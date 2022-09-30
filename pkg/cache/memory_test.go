package cache

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Entity int

func (e Entity) Hash() string {
	return strconv.Itoa(int(e))
}

func TestMemoryEntityCache_Set(t *testing.T) {
	c := NewMemoryEntityCache[Entity]()

	t.Run("set first item", func(t *testing.T) {
		err := c.Set(1)
		require.NoError(t, err)
		assert.Len(t, c.plainCache, 1)
		assert.Len(t, c.keyToIndexMap, 1)
		assert.Equal(t, 0, c.keyToIndexMap[Entity(1).Hash()])
		assert.Equal(t, Entity(1), c.plainCache[0])
	})
	t.Run("set second item", func(t *testing.T) {
		err := c.Set(2)
		require.NoError(t, err)
		assert.Len(t, c.plainCache, 2)
		assert.Len(t, c.keyToIndexMap, 2)
		assert.Equal(t, 1, c.keyToIndexMap[Entity(2).Hash()])
		assert.Equal(t, Entity(2), c.plainCache[1])
	})
	t.Run("set first item again", func(t *testing.T) {
		err := c.Set(1)
		require.NoError(t, err)
		assert.Len(t, c.plainCache, 2)
		assert.Len(t, c.keyToIndexMap, 2)
		assert.Equal(t, 0, c.keyToIndexMap[Entity(1).Hash()])
		assert.Equal(t, Entity(1), c.plainCache[0])
	})
}

func TestMemoryEntityCache_Get(t *testing.T) {
	c := NewMemoryEntityCache[Entity]()

	values := []Entity{1, 2, 3, 4, 5}
	err := c.Replace(values)
	require.NoError(t, err)

	for _, value := range values {
		value := value
		t.Run(fmt.Sprintf("%d", value), func(t *testing.T) {
			t.Parallel()

			v, err := c.Get(value.Hash())
			require.NoError(t, err)
			assert.Equal(t, value, v)
		})
	}
}

func TestMemoryEntityCache_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := NewMemoryEntityCache[Entity]()
		values := []Entity{1, 2, 3, 4, 5, 6}
		err := c.Replace(values)
		require.NoError(t, err)

		err = c.Delete(Entity(3).Hash())
		require.NoError(t, err)
		assert.Len(t, c.plainCache, len(values)-1)
		assert.Len(t, c.keyToIndexMap, len(values)-1)
		assert.NotContains(t, c.plainCache, Entity(3))
	})
	t.Run("not found", func(t *testing.T) {
		c := NewMemoryEntityCache[Entity]()
		values := []Entity{1, 2, 3, 4, 5, 6}
		err := c.Replace(values)
		require.NoError(t, err)

		err = c.Delete(Entity(10).Hash())
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrKeyNotFound))
		assert.Len(t, c.plainCache, len(values))
		assert.Len(t, c.keyToIndexMap, len(values))
		assert.Equal(t, values, c.plainCache)
	})
}

func TestMemoryEntityCache_GetList(t *testing.T) {
	c := NewMemoryEntityCache[Entity]()

	values := []Entity{1, 2, 3, 4, 5, 6}
	err := c.Replace(values)
	require.NoError(t, err)

	list, err := c.GetList(uint(len(values)), 0)
	require.NoError(t, err)
	assert.Equal(t, values, list)
}

func TestMemoryEntityCache_Replace(t *testing.T) {
	c := NewMemoryEntityCache[Entity]()

	values := []Entity{1, 2, 3, 4, 5, 6}
	err := c.Replace(values)
	require.NoError(t, err)
	assert.Equal(t, values, c.plainCache)
	assert.Len(t, c.keyToIndexMap, len(values))

	for i, value := range values {
		assert.Equal(t, i, c.keyToIndexMap[value.Hash()])
	}

	newValues := []Entity{1, 4, 6, 7, 8}
	err = c.Replace(newValues)
	require.NoError(t, err)
	assert.Equal(t, newValues, c.plainCache)
	assert.Len(t, c.keyToIndexMap, len(newValues))

	for i, value := range c.plainCache {
		assert.Equal(t, i, c.keyToIndexMap[value.Hash()])
	}
}

func TestMemoryEntityCache_DeleteParallel(t *testing.T) {
	c := NewMemoryEntityCache[Entity]()
	values := []Entity{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	deleteValues := []Entity{1, 20, 8, 13, 5, 18, 44, 58}
	expectedErrorsCount := 2 // deleteValues 44, 58 not presented in cache
	errorsCount := 0

	err := c.Replace(values)
	require.NoError(t, err)

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, deleteValue := range deleteValues {
		deleteValue := deleteValue
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := c.Delete(deleteValue.Hash())
			if err != nil {
				mu.Lock()
				defer mu.Unlock()
				errorsCount++
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, expectedErrorsCount, errorsCount)
	assert.Len(t, c.plainCache, len(values)-len(deleteValues)+expectedErrorsCount)

	// проверяем, что индексы в keyToIndexMap пересчитались правильно с учетом удаленных элементов
	for i, value := range c.plainCache {
		assert.Equal(t, c.keyToIndexMap[value.Hash()], i, "value: %+v; i = %d", value, i)
	}
}
