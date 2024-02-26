package comparator

import (
	"strings"
)

type EventChangesAnalyze struct {
	activeRules       *[]Rule
	eventFieldChanges map[string][]EventFieldChange
	analyzeResult     *AnalyzeResult
}

func NewEventChangesAnalyze(activeRules *[]Rule, eventFieldChanges map[string][]EventFieldChange) *EventChangesAnalyze {
	return &EventChangesAnalyze{
		activeRules:       activeRules,
		eventFieldChanges: eventFieldChanges,
		analyzeResult:     NewAnalyzeResult(),
	}
}

func (e *EventChangesAnalyze) AnalyzeEventFieldChanges() *AnalyzeResult {
	if e.eventFieldChanges != nil {
		for combinedReference, fieldChanges := range e.eventFieldChanges {
			e.analyzeFieldDifferencesForCase(combinedReference, fieldChanges)
		}
	}

	return e.analyzeResult
}

func (e *EventChangesAnalyze) analyzeFieldDifferencesForCase(combinedReference string, fieldChanges []EventFieldChange) {
	fieldName := getFieldFromCombinedReference(combinedReference)
	for _, rule := range *e.activeRules {
		violations := rule.CheckForViolation(fieldName, fieldChanges)
		if len(violations) > 0 {
			for _, violation := range violations {
				e.addAnalyzeDetail(combinedReference, violation)
			}
		}
	}
}

func getFieldFromCombinedReference(combinedReference string) string {
	referenceParts := strings.Split(combinedReference, "->")
	if len(referenceParts) > 1 {
		return referenceParts[1]
	}
	return ""
}

func (e *EventChangesAnalyze) addAnalyzeDetail(combinedReference string, violation Violation) {
	newMessage := violation.message
	existingViolation := e.analyzeResult.Get(combinedReference, violation.sourceEventId)

	if existingViolation.message != "" {
		newMessage = appendMessages(existingViolation.message, newMessage)
	}
	violation.message = newMessage
	e.analyzeResult.Put(combinedReference, violation)
}

func appendMessages(existingMessage, newMessage string) string {
	if existingMessage == "" {
		return newMessage
	}
	return existingMessage + "\n" + newMessage
}
