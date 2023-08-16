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

type EventDetails struct {
	Id          int64
	Name        string
	CreatedDate time.Time
	Data        string
	CaseDataId  int64
}

type EventFieldDifference struct {
	OldRecord       string
	NewRecord       string
	CreatedDate     time.Time
	SourceEventId   int64
	SourceEventName string
	OperationType   helper.OperationType
}

type CasesWithEventDetails map[int64]map[int64]EventDetails

type EventFieldDifferences map[string][]EventFieldDifference

type differences struct {
	differencesByPath EventFieldDifferences
}

func newDifferences() *differences {
	return &differences{
		differencesByPath: make(EventFieldDifferences),
	}
}

func CompareCaseEvents(transactionId string, caseEvents CasesWithEventDetails) EventFieldDifferences {
	mergedDifferences := NewConcurrentEventFieldDifferences()

	var wg sync.WaitGroup

	for caseID, events := range caseEvents {
		wg.Add(1)
		go func(caseID int64, events map[int64]EventDetails) {
			defer wg.Done()

			differences := getDifferences(caseID, events)
			mergedDifferences.PutAll(differences)
		}(caseID, events)
	}
	wg.Wait()

	log.Info().Msgf("tid:%s - All cases have been run successfully!", transactionId)

	return mergedDifferences.GetAll()
}

type jsonNode map[string]interface{}

func getDifferences(caseReference int64, eventDetails map[int64]EventDetails) EventFieldDifferences {
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
		compareNodes(base, compareWith, differences, strconv.FormatInt(caseReference, 10)+"->", eventId,
			eventDetail.CreatedDate, eventDetail.Name)
		base = compareWith
	}

	return differences.differencesByPath
}

func (d *differences) recordDifferenceAtPath(path string, difference EventFieldDifference) {
	if _, ok := d.differencesByPath[path]; !ok {
		d.differencesByPath[path] = make([]EventFieldDifference, 0)
	}
	d.differencesByPath[path] = append(d.differencesByPath[path], difference)
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

func compareNodes(base, compareWith interface{}, differences *differences, parentPath string, eventId int64,
	createdDate time.Time, eventName string) {
	baseNode, isBaseObject := convertToMap(base)
	compareNode, isCompareObject := convertToMap(compareWith)

	if isBaseObject && isCompareObject {
		for key, value := range baseNode {
			currentPath := fmt.Sprintf("%s.%s", parentPath, key)
			if compareValue, ok := compareNode[key]; ok {
				compareNodes(value, compareValue, differences, currentPath, eventId, createdDate, eventName)
			} else {
				difference := createDifference(value, "", eventId, createdDate, helper.Deleted, eventName)
				differences.recordDifferenceAtPath(currentPath, difference)
			}
		}
		for key, value := range compareNode {
			currentPath := fmt.Sprintf("%s.%s", parentPath, key)
			if _, ok := baseNode[key]; !ok {
				difference := createDifference("", value, eventId, createdDate, helper.Added, eventName)
				differences.recordDifferenceAtPath(currentPath, difference)
			}
		}
	} else {
		baseArray, isBaseArray := convertToSlice(base)
		compareArray, isCompareArray := convertToSlice(compareWith)
		if isBaseArray && isCompareArray {
			if len(baseArray) != len(compareArray) {
				difference := createDifference(base, compareWith, eventId, createdDate, helper.ArrayModified, eventName)
				differences.recordDifferenceAtPath(parentPath, difference)
			} else {
				for i := 0; i < len(baseArray); i++ {
					compareNodes(baseArray[i], compareArray[i], differences, fmt.Sprintf("%s[%d]", parentPath, i),
						eventId, createdDate, eventName)
				}
			}
		} else {
			if !compareWithEqual(base, compareWith) {
				difference := createDifference(base, compareWith, eventId, createdDate, helper.Modified, eventName)
				differences.recordDifferenceAtPath(parentPath, difference)
			}
		}
	}
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
	operationType helper.OperationType, eventName string) EventFieldDifference {

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

	return EventFieldDifference{
		OldRecord:       oldRecordValue,
		NewRecord:       newRecordValue,
		SourceEventId:   eventId,
		SourceEventName: eventName,
		CreatedDate:     createdDate,
		OperationType:   operationType,
	}
}
