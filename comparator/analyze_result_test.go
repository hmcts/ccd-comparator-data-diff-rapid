package comparator

import (
	"testing"
)

func TestAnalyzeResult(t *testing.T) {
	analyzeResult := NewAnalyzeResult()

	v1 := Violation{
		sourceEventId:            123,
		previousEventCreatedDate: "1",
		previousEventUserId:      "1",
		message:                  "Message1",
	}
	v2 := Violation{
		sourceEventId:            456,
		previousEventCreatedDate: "2",
		previousEventUserId:      "2",
		message:                  "Message2",
	}
	// Test Put and Get methods
	analyzeResult.Put("ref1", v1)
	analyzeResult.Put("ref2", v2)

	expectedMessage1 := "Message1"
	expectedMessage2 := "Message2"
	v1Return := analyzeResult.Get("ref1", 123)
	v2Return := analyzeResult.Get("ref2", 456)

	if v1Return.message != expectedMessage1 {
		t.Errorf("Expected message for ref1 and 123: %s, but got: %s", expectedMessage1, v1Return.message)
	}

	if v2Return.message != expectedMessage2 {
		t.Errorf("Expected message for ref2 and 456: %s, but got: %s", expectedMessage2, v2Return.message)
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
