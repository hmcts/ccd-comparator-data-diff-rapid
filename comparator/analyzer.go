package comparator

import (
	"strings"
)

type EventDifferencesData struct {
	activeRules      *[]Rule
	eventDifferences map[string][]EventFieldDifference
	analyzeResult    *AnalyzeResult
}

func NewEventDifferencesData(activeRules *[]Rule, eventDifferences map[string][]EventFieldDifference) *EventDifferencesData {
	return &EventDifferencesData{
		activeRules:      activeRules,
		eventDifferences: eventDifferences,
		analyzeResult:    NewAnalyzeResult(),
	}
}

func (e *EventDifferencesData) ProcessEventDiff() *AnalyzeResult {
	if e.eventDifferences != nil {
		for combinedReference, fieldDifferences := range e.eventDifferences {
			e.analyzeFieldDifferencesForCase(combinedReference, fieldDifferences)
		}
	}

	return e.analyzeResult
}

func (e *EventDifferencesData) analyzeFieldDifferencesForCase(combinedReference string, fieldDifferences []EventFieldDifference) {
	fieldName := strings.Split(combinedReference, "->")[1]
	for _, rule := range *e.activeRules {
		pair := rule.CheckForViolation(fieldName, fieldDifferences)
		if pair != nil {
			e.addAnalyzeDetail(combinedReference, pair)
		}
	}
}

func (e *EventDifferencesData) addAnalyzeDetail(combinedReference string, pair *Pair[int64, string]) {
	newMessage := pair.Right
	eventId := pair.Left
	existingMessage := e.analyzeResult.Get(combinedReference, eventId)
	if existingMessage != "" {
		e.analyzeResult.Put(combinedReference, eventId, checkAndUpdateMessage(existingMessage, newMessage))
	} else {
		e.analyzeResult.Put(combinedReference, eventId, newMessage)
	}
}

func checkAndUpdateMessage(existingMessage, newMessage string) string {
	if existingMessage == "" {
		return newMessage
	}
	return existingMessage + "\n" + newMessage
}
