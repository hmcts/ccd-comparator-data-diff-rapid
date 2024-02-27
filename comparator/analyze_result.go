package comparator

import (
	"fmt"
	"sync"
)

type AnalyzeResult struct {
	result map[string]Violation
	mutex  sync.RWMutex
}

func NewAnalyzeResult() *AnalyzeResult {
	return &AnalyzeResult{
		result: make(map[string]Violation),
	}
}

func (a *AnalyzeResult) Get(combinedReference string, sourceEventId int64) Violation {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	key := a.generateKey(combinedReference, sourceEventId)
	return a.result[key]
}

func (a *AnalyzeResult) Put(combinedReference string, violation Violation) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	key := a.generateKey(combinedReference, violation.sourceEventId)
	a.result[key] = violation
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
