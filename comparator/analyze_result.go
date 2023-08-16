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

func (ar *AnalyzeResult) Get(combinedReference string, sourceEventId int64) string {
	ar.mutex.RLock()
	defer ar.mutex.RUnlock()
	key := ar.generateKey(combinedReference, sourceEventId)
	return ar.result[key]
}

func (ar *AnalyzeResult) Put(combinedReference string, sourceEventId int64, analyzeMessage string) {
	ar.mutex.Lock()
	defer ar.mutex.Unlock()
	key := ar.generateKey(combinedReference, sourceEventId)
	ar.result[key] = analyzeMessage
}

func (ar *AnalyzeResult) IsNotEmpty() bool {
	ar.mutex.RLock()
	defer ar.mutex.RUnlock()
	return len(ar.result) > 0
}

func (ar *AnalyzeResult) IsEmpty() bool {
	ar.mutex.RLock()
	defer ar.mutex.RUnlock()
	return len(ar.result) == 0
}

func (ar *AnalyzeResult) Clear() {
	ar.mutex.Lock()
	defer ar.mutex.Unlock()
	for k := range ar.result {
		delete(ar.result, k)
	}
}

func (ar *AnalyzeResult) Size() int {
	ar.mutex.RLock()
	defer ar.mutex.RUnlock()
	return len(ar.result)
}

func (ar *AnalyzeResult) generateKey(combinedReference string, sourceEventId int64) string {
	return fmt.Sprintf("%s_%d", combinedReference, sourceEventId)
}
