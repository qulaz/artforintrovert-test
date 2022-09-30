package cache

import (
	"sync"
)

type MemoryEntityCache[V Hashable] struct {
	plainCache    []V
	keyToIndexMap map[string]int
	mutex         sync.RWMutex
}

func NewMemoryEntityCache[V Hashable]() *MemoryEntityCache[V] {
	return &MemoryEntityCache[V]{
		plainCache:    make([]V, 0, 100),
		keyToIndexMap: make(map[string]int),
		mutex:         sync.RWMutex{},
	}
}

func (c *MemoryEntityCache[V]) mutexLessSet(value V) error {
	idx, ok := c.keyToIndexMap[value.Hash()]
	if ok {
		c.plainCache[idx] = value

		return nil
	}

	c.plainCache = append(c.plainCache, value)
	c.keyToIndexMap[value.Hash()] = len(c.plainCache) - 1

	return nil
}

func (c *MemoryEntityCache[V]) Set(value V) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.mutexLessSet(value)
}

func (c *MemoryEntityCache[V]) Get(key string) (V, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var noop V

	idx, ok := c.keyToIndexMap[key]
	if !ok {
		return noop, ErrKeyNotFound
	}

	return c.plainCache[idx], nil
}

func (c *MemoryEntityCache[V]) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	idx, ok := c.keyToIndexMap[key]
	if !ok {
		return ErrKeyNotFound
	}

	var noop V
	c.plainCache[idx] = noop // avoid memory leak. noop is zero value of generic type T
	// cc: https://cs.opensource.google/go/x/exp/+/c76eaa36:slices/slices.go;l=162
	c.plainCache = append(c.plainCache[:idx], c.plainCache[idx+1:]...)

	for key, i := range c.keyToIndexMap {
		if i == idx {
			delete(c.keyToIndexMap, key)
		} else if i > idx {
			c.keyToIndexMap[key]--
		}
	}

	return nil
}

func (c *MemoryEntityCache[V]) GetList(limit uint, offset uint) ([]V, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	start := offset
	end := offset + limit

	if end > uint(len(c.plainCache)) {
		end = uint(len(c.plainCache))
	}

	if start > uint(len(c.plainCache)) {
		start = uint(len(c.plainCache))
	}

	return c.plainCache[start:end], nil
}

func (c *MemoryEntityCache[V]) Replace(values []V) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.plainCache = make([]V, len(values))
	c.keyToIndexMap = make(map[string]int, len(values))

	for i, value := range values {
		c.plainCache[i] = value
		c.keyToIndexMap[value.Hash()] = i
	}

	return nil
}
