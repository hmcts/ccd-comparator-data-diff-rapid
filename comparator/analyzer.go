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
		pair := rule.CheckForViolation(fieldName, fieldChanges)
		if pair != nil {
			e.addAnalyzeDetail(combinedReference, pair)
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

func (e *EventChangesAnalyze) addAnalyzeDetail(combinedReference string, pair *Pair[int64, string]) {
	newMessage := pair.Right
	eventId := pair.Left
	existingMessage := e.analyzeResult.Get(combinedReference, eventId)
	if existingMessage != "" {
		e.analyzeResult.Put(combinedReference, eventId, appendMessages(existingMessage, newMessage))
	} else {
		e.analyzeResult.Put(combinedReference, eventId, newMessage)
	}
}

func appendMessages(existingMessage, newMessage string) string {
	if existingMessage == "" {
		return newMessage
	}
	return existingMessage + "\n" + newMessage
}
