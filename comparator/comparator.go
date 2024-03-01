package comparator

import (
	"ccd-comparator-data-diff-rapid/helper"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"
)

type OperationType string

const (
	Added         OperationType = "ADDED"
	Deleted       OperationType = "DELETED"
	Modified      OperationType = "MODIFIED"
	ArrayModified OperationType = "ARRAY_MODIFIED"
	ArrayExtended OperationType = "ARRAY_EXTENDED"
	ArrayShrunk   OperationType = "ARRAY_SHRUNK"
	NoChange      OperationType = "NO_CHANGE"
)

type EventDetails struct {
	Id          int64
	Name        string
	CreatedDate time.Time
	Data        string
	CaseDataId  int64
	UserId      string
}

type EventFieldChange struct {
	OldRecord       string
	NewRecord       string
	CreatedDate     time.Time
	SourceEventId   int64
	SourceEventName string
	OperationType   OperationType
	UserId          string
}

type CasesWithEventDetails map[int64]map[int64]EventDetails

type EventFieldChanges map[string][]EventFieldChange

type differences struct {
	differencesByPath EventFieldChanges
}

func newDifferences() *differences {
	return &differences{
		differencesByPath: make(EventFieldChanges),
	}
}

func CompareEventsByCaseReference(transactionId string, caseEvents CasesWithEventDetails) EventFieldChanges {
	mergedDifferences := NewConcurrentEventFieldDifferences()

	var wg sync.WaitGroup

	for caseReference, events := range caseEvents {
		wg.Add(1)
		go func(caseReference int64, events map[int64]EventDetails) {
			defer func() {
				if r := recover(); r != nil {
					log.Error().Msgf("tid:%s - Recovered from panic: %s. %d hasn't been processed.", transactionId,
						r, caseReference)
				}
				wg.Done()
			}()

			differences := detectEventModifications(caseReference, events)
			mergedDifferences.PutAll(differences)
		}(caseReference, events)
	}
	wg.Wait()

	log.Info().Msgf("tid:%s - All cases have been run successfully!", transactionId)

	return mergedDifferences.GetAll()
}

type jsonNode map[string]interface{}

func detectEventModifications(caseReference int64, eventDetails map[int64]EventDetails) EventFieldChanges {
	differences := newDifferences()
	var base jsonNode

	keys := make([]int64, 0, len(eventDetails))
	for eventId := range eventDetails {
		keys = append(keys, eventId)
	}

	// Sort the keys to ensure they are in ascending order
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, eventId := range keys {
		eventDetail := eventDetails[eventId]
		if base == nil {
			// Unmarshal the base data from the first event
			helper.MustUnmarshal([]byte(eventDetail.Data), &base)
			continue
		}

		var compareWith jsonNode
		helper.MustUnmarshal([]byte(eventDetail.Data), &compareWith)
		compareJsonNodes(base, compareWith, differences, strconv.FormatInt(caseReference, 10)+"->", eventId,
			eventDetail.CreatedDate, eventDetail.Name, eventDetail.UserId)
		base = compareWith
	}

	return differences.differencesByPath
}

func compareJsonNodes(base, compareWith interface{}, differences *differences, parentPath string, eventId int64,
	createdDate time.Time, eventName string, userId string) {
	baseNode, isBaseObject := convertToMap(base)
	compareNode, isCompareObject := convertToMap(compareWith)

	if isBaseObject && isCompareObject {
		for key, value := range baseNode {
			currentPath := fmt.Sprintf("%s.%s", parentPath, key)
			if compareValue, ok := compareNode[key]; ok {
				compareJsonNodes(value, compareValue, differences, currentPath, eventId, createdDate, eventName, userId)
			} else {
				differences.recordDifferenceAtPath(currentPath, createDifference(value, "", eventId,
					createdDate, Deleted, eventName, userId))
			}
		}
		for key, value := range compareNode {
			currentPath := fmt.Sprintf("%s.%s", parentPath, key)
			if _, ok := baseNode[key]; !ok {
				differences.recordDifferenceAtPath(currentPath, createDifference("", value, eventId,
					createdDate, Added, eventName, userId))
			}
		}
	} else {
		baseArray, isBaseArray := convertToSlice(base)
		compareArray, isCompareArray := convertToSlice(compareWith)
		if isBaseArray && isCompareArray {
			if len(baseArray) > len(compareArray) {
				differences.recordDifferenceAtPath(parentPath, createDifference(base, compareWith, eventId, createdDate,
					ArrayShrunk, eventName, userId))
			} else if len(baseArray) < len(compareArray) {
				differences.recordDifferenceAtPath(parentPath, createDifference(base, compareWith, eventId, createdDate,
					ArrayExtended, eventName, userId))
			} else if !compareArrays(baseArray, compareArray) {
				differences.recordDifferenceAtPath(parentPath, createDifference(base, compareWith, eventId, createdDate,
					ArrayModified, eventName, userId))
			}
		} else if !compareWithEqual(base, compareWith) {
			differences.recordDifferenceAtPath(parentPath, createDifference(base, compareWith, eventId, createdDate,
				Modified, eventName, userId))
		} else {
			differences.recordDifferenceAtPath(parentPath, createDifference(base, compareWith, eventId, createdDate,
				NoChange, eventName, userId))
		}
	}
}

func convertToMap(t interface{}) (jsonNode, bool) {
	if t == nil {
		return nil, false
	}

	switch t.(type) {
	case map[string]interface{}:
		return t.(map[string]interface{}), true
	case jsonNode:
		return t.(jsonNode), true
	default:
		return nil, false
	}
}

func convertToSlice(v interface{}) ([]interface{}, bool) {
	if v == nil {
		return nil, false
	}
	var out []interface{}
	rv := reflect.ValueOf(v)
	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			out = append(out, rv.Index(i).Interface())
		}
		return out, true
	default:
		return nil, false
	}
}

func compareArrays(base, compare []interface{}) bool {
	// https://pkg.go.dev/encoding/json#Marshal keys are sorted
	baseMap := make(map[string]interface{})
	compareMap := make(map[string]interface{})

	for _, item := range base {
		switch item := item.(type) {
		case map[string]interface{}:
			delete(item, "id")
			itemJSON, _ := json.Marshal(item)
			baseMap[string(itemJSON)] = item
		default:
			if strItem, ok := item.(string); ok {
				baseMap[strItem] = item
			}
		}
	}

	for _, item := range compare {
		switch item := item.(type) {
		case map[string]interface{}:
			delete(item, "id")
			itemJSON, _ := json.Marshal(item)
			compareMap[string(itemJSON)] = item
		default:
			if strItem, ok := item.(string); ok {
				compareMap[strItem] = item
			}
		}
	}

	for key, baseItem := range baseMap {
		compareItem, ok := compareMap[key]
		if !ok {
			return false
		} else if !reflect.DeepEqual(baseItem, compareItem) {
			return false
		}
	}

	for key := range compareMap {
		_, ok := baseMap[key]
		if !ok {
			return false
		}
	}

	return true
}

func (d *differences) recordDifferenceAtPath(path string, difference EventFieldChange) {
	if !isNotEmpty(difference.OldRecord, difference.NewRecord) {
		difference.OperationType = NoChange
	}

	if _, ok := d.differencesByPath[path]; !ok {
		if difference.OperationType != NoChange {
			d.differencesByPath[path] = make([]EventFieldChange, 0)
		} else {
			return
		}
	}
	d.differencesByPath[path] = append(d.differencesByPath[path], difference)
}

func isNotEmpty(oldValue, newValue string) bool {
	return (oldValue != "" && oldValue != "null" && oldValue != "{}" && oldValue != "[]") ||
		(newValue != "" && newValue != "null" && newValue != "{}" && newValue != "[]")
}

func compareWithEqual(base, compareWith interface{}) bool {
	baseBytes, err := json.Marshal(base)
	if err != nil {
		return false
	}

	compareBytes, err := json.Marshal(compareWith)
	if err != nil {
		return false
	}
	return string(baseBytes) == string(compareBytes)
}

func createDifference(oldRecord, newRecord interface{}, eventId int64, createdDate time.Time,
	operationType OperationType, eventName string, userId string) EventFieldChange {

	oldRecordValue, oBase := oldRecord.(string)
	if !oBase {
		oldRecordBytes, _ := json.Marshal(oldRecord)
		oldRecordValue = string(oldRecordBytes)
	}

	newRecordValue, nBase := newRecord.(string)
	if !nBase {
		newRecordBytes, _ := json.Marshal(newRecord)
		newRecordValue = string(newRecordBytes)
	}

	return EventFieldChange{
		OldRecord:       oldRecordValue,
		NewRecord:       newRecordValue,
		SourceEventId:   eventId,
		SourceEventName: eventName,
		CreatedDate:     createdDate,
		OperationType:   operationType,
		UserId:          userId,
	}
}
