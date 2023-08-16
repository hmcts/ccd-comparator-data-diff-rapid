package comparator

import (
	"fmt"
	"sync"
)

type AnalyzeResult struct {
	result map[string]string
	mutex  sync.RWMutex
}

func NewAnalyzeResult() *AnalyzeResult {
	return &AnalyzeResult{
		result: make(map[string]string),
	}
}

func (a *AnalyzeResult) Get(combinedReference string, sourceEventId int64) string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	key := a.generateKey(combinedReference, sourceEventId)
	return a.result[key]
}

func (a *AnalyzeResult) Put(combinedReference string, sourceEventId int64, analyzeMessage string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	key := a.generateKey(combinedReference, sourceEventId)
	a.result[key] = analyzeMessage
}

func (a *AnalyzeResult) IsNotEmpty() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.result) > 0
}

func (a *AnalyzeResult) IsEmpty() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.result) == 0
}

func (a *AnalyzeResult) Clear() {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	for key := range a.result {
		delete(a.result, key)
	}
}

func (a *AnalyzeResult) Size() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return len(a.result)
}

func (a *AnalyzeResult) generateKey(combinedReference string, sourceEventId int64) string {
	return fmt.Sprintf("%s_%d", combinedReference, sourceEventId)
}
