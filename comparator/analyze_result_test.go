package comparator

import (
	"testing"
)

func TestAnalyzeResult(t *testing.T) {
	analyzeResult := NewAnalyzeResult()

	// Test Put and Get methods
	analyzeResult.Put("ref1", 123, "Message1")
	analyzeResult.Put("ref2", 456, "Message2")

	expectedMessage1 := "Message1"
	expectedMessage2 := "Message2"
	actualMessage1 := analyzeResult.Get("ref1", 123)
	actualMessage2 := analyzeResult.Get("ref2", 456)

	if actualMessage1 != expectedMessage1 {
		t.Errorf("Expected message for ref1 and 123: %s, but got: %s", expectedMessage1, actualMessage1)
	}

	if actualMessage2 != expectedMessage2 {
		t.Errorf("Expected message for ref2 and 456: %s, but got: %s", expectedMessage2, actualMessage2)
	}

	// Test IsNotEmpty and IsEmpty methods
	if analyzeResult.IsEmpty() {
		t.Error("Expected IsEmpty to be false, but got true")
	}

	if !analyzeResult.IsNotEmpty() {
		t.Error("Expected IsNotEmpty to be true, but got false")
	}

	// Test Clear and Size methods
	analyzeResult.Clear()

	if analyzeResult.Size() != 0 {
		t.Errorf("Expected size after Clear: 0, but got: %d", analyzeResult.Size())
	}

	if !analyzeResult.IsEmpty() {
		t.Error("Expected IsEmpty to be true after Clear, but got false")
	}
}
