package comparator

import (
	"sync"
)

type ConcurrentEventFieldDifferences struct {
	dataMap EventFieldDifferences
	mutex   sync.RWMutex
}

func NewConcurrentEventFieldDifferences() *ConcurrentEventFieldDifferences {
	return &ConcurrentEventFieldDifferences{
		dataMap: make(EventFieldDifferences),
	}
}

func (ced *ConcurrentEventFieldDifferences) Set(key string, value []EventFieldDifference) {
	ced.mutex.Lock()
	defer ced.mutex.Unlock()
	ced.dataMap[key] = value
}

func (ced *ConcurrentEventFieldDifferences) Get(key string) ([]EventFieldDifference, bool) {
	ced.mutex.RLock()
	defer ced.mutex.RUnlock()
	value, ok := ced.dataMap[key]
	return value, ok
}

func (ced *ConcurrentEventFieldDifferences) GetAll() EventFieldDifferences {
	ced.mutex.RLock()
	defer ced.mutex.RUnlock()
	return ced.dataMap
}

func (ced *ConcurrentEventFieldDifferences) Delete(key string) {
	ced.mutex.Lock()
	defer ced.mutex.Unlock()
	delete(ced.dataMap, key)
}

func (ced *ConcurrentEventFieldDifferences) Size() int {
	ced.mutex.RLock()
	defer ced.mutex.RUnlock()
	return len(ced.dataMap)
}

func (ced *ConcurrentEventFieldDifferences) Clear() {
	ced.mutex.Lock()
	defer ced.mutex.Unlock()
	for k := range ced.dataMap {
		delete(ced.dataMap, k)
	}
}

func (ced *ConcurrentEventFieldDifferences) PutAll(otherMap EventFieldDifferences) {
	ced.mutex.Lock()
	defer ced.mutex.Unlock()
	for key, value := range otherMap {
		ced.dataMap[key] = value
	}
}
