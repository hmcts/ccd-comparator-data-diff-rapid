package comparator

import (
	"sync"
)

type ConcurrentEventFieldDifferences struct {
	dataMap EventFieldChanges
	mutex   sync.RWMutex
}

func NewConcurrentEventFieldDifferences() *ConcurrentEventFieldDifferences {
	return &ConcurrentEventFieldDifferences{
		dataMap: make(EventFieldChanges),
	}
}

func (c *ConcurrentEventFieldDifferences) Set(key string, value []EventFieldChange) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.dataMap[key] = value
}

func (c *ConcurrentEventFieldDifferences) Get(key string) ([]EventFieldChange, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, ok := c.dataMap[key]
	return value, ok
}

func (c *ConcurrentEventFieldDifferences) GetAll() EventFieldChanges {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.dataMap
}

func (c *ConcurrentEventFieldDifferences) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.dataMap, key)
}

func (c *ConcurrentEventFieldDifferences) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.dataMap)
}

func (c *ConcurrentEventFieldDifferences) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k := range c.dataMap {
		delete(c.dataMap, k)
	}
}

func (c *ConcurrentEventFieldDifferences) PutAll(otherMap EventFieldChanges) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, value := range otherMap {
		c.dataMap[key] = value
	}
}
