package comparator

import (
	"ccd-comparator-data-diff-rapid/jsonx"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type OperationType string

// IsArrayOperation checks if the OperationType represents an array operation.
func (o OperationType) IsArrayOperation() bool {
	return strings.HasPrefix(string(o), "ARRAY_")
}

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

func detectEventModifications(caseReference int64, eventDetails map[int64]EventDetails) EventFieldChanges {
	differences := newDifferences()
	var base jsonx.NodeAny

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
			jsonx.MustUnmarshal([]byte(eventDetail.Data), &base)
			continue
		}

		var compareWith jsonx.NodeAny
		jsonx.MustUnmarshal([]byte(eventDetail.Data), &compareWith)
		compareJsonNodes(base, compareWith, differences, strconv.FormatInt(caseReference, 10)+"->", eventId,
			eventDetail.CreatedDate, eventDetail.Name, eventDetail.UserId)
		base = compareWith
	}

	return differences.differencesByPath
}

func compareJsonNodes(base, compareWith any, differences *differences, parentPath string, eventId int64,
	createdDate time.Time, eventName string, userId string) {
	baseNode, isBaseObject := convertToMap(base)
	compareNode, isCompareObject := convertToMap(compareWith)
	if isBaseObject && isCompareObject && !baseNode.IsEmpty() && !compareNode.IsEmpty() {
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
		var changeType OperationType
		if isBaseArray && isCompareArray && !baseArray.IsEmpty() && !compareArray.IsEmpty() {
			if len(baseArray) > len(compareArray) {
				changeType = ArrayShrunk
			} else if len(baseArray) < len(compareArray) {
				changeType = ArrayExtended
			} else {
				changeType = ArrayModified
			}

			changes := compareArrays(baseArray, compareArray)
			if !changes.IsEmpty() {
				for key, items := range changes {
					itemsJson := jsonx.MustMarshal(items)
					differences.recordDifferenceAtPath(parentPath+key, createDifference("", string(itemsJson),
						eventId, createdDate, changeType, eventName, userId))
				}
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

func convertToMap(t any) (jsonx.NodeAny, bool) {
	if t == nil {
		return nil, false
	}

	if m, ok := t.(map[string]any); ok {
		return m, true
	}

	if n, ok := t.(jsonx.NodeAny); ok {
		return n, true
	}

	return nil, false
}

func convertToSlice(v any) (jsonx.ArrayAny, bool) {
	if v == nil {
		return nil, false
	}

	rv := reflect.ValueOf(v)
	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		out := make(jsonx.ArrayAny, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			out = append(out, rv.Index(i).Interface())
		}
		return out, true
	default:
		return nil, false
	}
}

func compareArrays(base, compare []any) jsonx.NodeChange {
	// https://pkg.go.dev/encoding/json#Marshal keys are sorted
	baseMap := printFieldValuesMap(base)
	compareMap := printFieldValuesMap(compare)
	result := make(jsonx.NodeChange)

	for key, compareItem := range compareMap {
		var changes jsonx.ArrayChange
		baseItem, ok := baseMap[key]
		if ok {
			changes = compareArrayObjects(baseItem, compareItem)
		} else {
			for _, value := range compareItem {
				changes = append(changes, jsonx.Change{Value: value, Added: true})
			}
		}
		if !changes.IsEmpty() {
			result[key] = changes
		}
	}

	for key, baseItem := range baseMap {
		var changes jsonx.ArrayChange
		_, ok := compareMap[key]
		if !ok {
			for _, value := range baseItem {
				changes = append(changes, jsonx.Change{Value: value, Deleted: true})
			}
		}
		if !changes.IsEmpty() {
			result[key] = changes
		}
	}

	return result
}

func printFieldValuesMap(data []any) map[string][]string {
	fieldValuesMap := make(map[string][]string)
	for _, obj := range data {
		if objMap, ok := obj.(map[string]any); ok {
			traverseArrayFields(objMap, "", fieldValuesMap)
		}
	}
	return fieldValuesMap
}

func traverseArrayFields(data map[string]any, prefix string, fieldValuesMap map[string][]string) {
	for key, value := range data {
		path := fmt.Sprintf("%s.%s", prefix, key)
		if reflect.TypeOf(value).Kind() == reflect.Map {
			if subMap, ok := value.(map[string]any); ok {
				traverseArrayFields(subMap, path, fieldValuesMap)
			}
		} else if reflect.TypeOf(value).Kind() == reflect.Slice {
			if subSlice, ok := value.([]any); ok {
				var subValues []string
				for _, subValue := range subSlice {
					if subMap, ok := subValue.(map[string]any); ok {
						traverseArrayFields(subMap, path, fieldValuesMap)
					} else {
						subValues = append(subValues, fmt.Sprintf("%v", subValue))
					}
				}
				if len(subValues) > 0 {
					valueStr := strings.Join(subValues, ", ")
					fieldValuesMap[path] = append(fieldValuesMap[path], valueStr)
				}
			}
		} else {
			valueStr := fmt.Sprintf("%v", value)
			fieldValuesMap[path] = append(fieldValuesMap[path], valueStr)
		}
	}
}

func compareArrayObjects(baseArray, compareArray []string) jsonx.ArrayChange {
	baseCounts := sliceToMap(baseArray)
	compareCounts := sliceToMap(compareArray)

	// Find deleted elements
	deleted := make(jsonx.ArrayChange, 0)
	for k, baseCount := range baseCounts {
		compareCount, found := compareCounts[k]
		if found {
			diff := baseCount - compareCount
			if diff > 0 {
				deleted = append(deleted, jsonx.Change{Value: k, Deleted: true})
			}
		} else {
			deleted = append(deleted, jsonx.Change{Value: k, Deleted: true})
		}
	}

	// Find added elements
	added := make(jsonx.ArrayChange, 0)
	for k, compareCount := range compareCounts {
		baseCount, found := baseCounts[k]
		if found {
			diff := compareCount - baseCount
			if diff > 0 {
				added = append(added, jsonx.Change{Value: k, Added: true})
			}
		} else {
			added = append(added, jsonx.Change{Value: k, Added: true})
		}
	}

	// Combine deleted and added elements
	changes := append(deleted, added...)

	return changes
}

func sliceToMap(slice []string) map[string]int {
	counts := make(map[string]int)
	for _, v := range slice {
		counts[v]++
	}
	return counts
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

func compareWithEqual(base, compareWith any) bool {
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

func createDifference(oldRecord, newRecord any, eventId int64, createdDate time.Time,
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
