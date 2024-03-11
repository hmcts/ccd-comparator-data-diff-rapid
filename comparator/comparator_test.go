package comparator

import (
	"ccd-comparator-data-diff-rapid/helper"
	"ccd-comparator-data-diff-rapid/jsonx"
	"reflect"
	"testing"
	"time"
)

var layout = "2006-01-02"

func Test_compareNodes(t *testing.T) {
	type args struct {
		base        interface{}
		compareWith interface{}
		differences *differences
		parentPath  string
		eventId     int64
		createdDate time.Time
		eventName   string
	}

	createdDate := helper.MustParseTime("2006-01-02T15:04:05.000", "2023-01-01T00:00:00.000")

	expectedDifferences := &differences{
		differencesByPath: map[string][]EventFieldChange{
			".field1": {
				{
					OldRecord:       "123",
					NewRecord:       "456",
					CreatedDate:     createdDate,
					SourceEventId:   1,
					SourceEventName: "TestEvent",
					OperationType:   Modified,
					UserId:          "1",
				},
			},
			".field2": {
				{
					OldRecord:       "abc",
					NewRecord:       "xyz",
					CreatedDate:     createdDate,
					SourceEventId:   1,
					SourceEventName: "TestEvent",
					OperationType:   Modified,
					UserId:          "1",
				},
			},
			".field3": {
				{
					OldRecord:       "true",
					NewRecord:       "",
					CreatedDate:     createdDate,
					SourceEventId:   1,
					SourceEventName: "TestEvent",
					OperationType:   Deleted,
					UserId:          "1",
				},
			},
			".field4": {
				{
					OldRecord:       "",
					NewRecord:       "789",
					CreatedDate:     createdDate,
					SourceEventId:   1,
					SourceEventName: "TestEvent",
					OperationType:   Added,
					UserId:          "1",
				},
			},
		},
	}

	expectedDifferencesComplexObject := &differences{
		differencesByPath: map[string][]EventFieldChange{
			".field1": {
				{
					OldRecord:       "123",
					NewRecord:       "789",
					CreatedDate:     createdDate,
					SourceEventId:   2,
					SourceEventName: "ComplexEvent",
					OperationType:   Modified,
					UserId:          "1",
				},
			},
			".field2.subfield1": {
				{
					OldRecord:       "abc",
					NewRecord:       "xyz",
					CreatedDate:     createdDate,
					SourceEventId:   2,
					SourceEventName: "ComplexEvent",
					OperationType:   Modified,
					UserId:          "1",
				},
			},
			".field2.subfield2": {
				{
					OldRecord:       "456",
					NewRecord:       "789",
					CreatedDate:     createdDate,
					SourceEventId:   2,
					SourceEventName: "ComplexEvent",
					OperationType:   Modified,
					UserId:          "1",
				},
			},
		},
	}

	tests := []struct {
		name string
		args args
		want *differences
	}{
		{
			name: "PositiveScenario",
			args: args{
				base:        jsonx.NodeAny{"field1": 123, "field2": "abc", "field3": true},
				compareWith: jsonx.NodeAny{"field1": 456, "field2": "xyz", "field4": 789},
				differences: newDifferences(),
				parentPath:  "",
				eventId:     1,
				createdDate: createdDate,
				eventName:   "TestEvent",
			},
			want: expectedDifferences,
		},
		{
			name: "ComplexObject",
			args: args{
				base: jsonx.NodeAny{
					"field1": 123,
					"field2": jsonx.NodeAny{"subfield1": "abc", "subfield2": 456},
					"field3": []int{1, 2, 3},
					"field4": []string{"a", "b", "c"},
				},
				compareWith: jsonx.NodeAny{
					"field1": 789,
					"field2": jsonx.NodeAny{"subfield1": "xyz", "subfield2": 789},
					"field3": []int{4, 5, 6},
					"field4": []string{"a", "b"},
				},
				differences: newDifferences(),
				parentPath:  "",
				eventId:     2,
				createdDate: createdDate,
				eventName:   "ComplexEvent",
			},
			want: expectedDifferencesComplexObject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compareJsonNodes(tt.args.base, tt.args.parentPath, tt.args.eventId, tt.args.createdDate, tt.args.eventName, "1")

			if !reflect.DeepEqual(tt.args.differences, tt.want) {
				t.Errorf("Unexpected mergedDifferences:\nGot: %#v\nWant: %#v", tt.args.differences, tt.want)
			}
		})
	}
}

func TestGetDifferences(t *testing.T) {
	// Test case data
	eventData1 := `{"field1": 123, "field2": "abc"}`
	eventData2 := `{"field1": 456, "field2": "def"}`
	eventData3 := `{"field1": 789, "field2": "ghi"}`
	eventData4 := `{"field1": 123, "field2": "xyz"}`
	eventData5 := `{"field1": 456, "field2": "def"}`

	eventDetails := map[int64]EventDetails{
		1: {Id: 1, Name: "Event1", CreatedDate: helper.MustParseTime(layout, "2023-07-25"), Data: eventData1},
		5: {Id: 5, Name: "Event5", CreatedDate: helper.MustParseTime(layout, "2023-07-25"), Data: eventData5},
		3: {Id: 3, Name: "Event3", CreatedDate: helper.MustParseTime(layout, "2023-07-25"), Data: eventData3},
		2: {Id: 2, Name: "Event2", CreatedDate: helper.MustParseTime(layout, "2023-07-25"), Data: eventData2},
		4: {Id: 4, Name: "Event4", CreatedDate: helper.MustParseTime(layout, "2023-07-25"), Data: eventData4},
	}

	differences := detectEventModifications(123, eventDetails)

	expectedDifferences := EventFieldChanges{
		"123->.field1": {
			{OldRecord: "123", NewRecord: "456", CreatedDate: eventDetails[2].CreatedDate, SourceEventId: 2, SourceEventName: "Event2", OperationType: Modified},
			{OldRecord: "456", NewRecord: "789", CreatedDate: eventDetails[3].CreatedDate, SourceEventId: 3, SourceEventName: "Event3", OperationType: Modified},
			{OldRecord: "789", NewRecord: "123", CreatedDate: eventDetails[4].CreatedDate, SourceEventId: 4, SourceEventName: "Event4", OperationType: Modified},
			{OldRecord: "123", NewRecord: "456", CreatedDate: eventDetails[5].CreatedDate, SourceEventId: 5, SourceEventName: "Event5", OperationType: Modified},
		},
		"123->.field2": {
			{OldRecord: "abc", NewRecord: "def", CreatedDate: eventDetails[2].CreatedDate, SourceEventId: 2, SourceEventName: "Event2", OperationType: Modified},
			{OldRecord: "def", NewRecord: "ghi", CreatedDate: eventDetails[3].CreatedDate, SourceEventId: 3, SourceEventName: "Event3", OperationType: Modified},
			{OldRecord: "ghi", NewRecord: "xyz", CreatedDate: eventDetails[4].CreatedDate, SourceEventId: 4, SourceEventName: "Event4", OperationType: Modified},
			{OldRecord: "xyz", NewRecord: "def", CreatedDate: eventDetails[5].CreatedDate, SourceEventId: 5, SourceEventName: "Event5", OperationType: Modified},
		},
	}

	if !reflect.DeepEqual(differences, expectedDifferences) {
		t.Errorf("Unexpected differences. \nGot: %v\nWan: %v", differences, expectedDifferences)
	}
}

func TestCompareCaseEvents(t *testing.T) {
	caseEvents := CasesWithEventDetails{
		1: {
			1: {
				Id:          1,
				Name:        "Event 1",
				CreatedDate: helper.MustParseTime(layout, "2023-07-25"),
				Data:        `{"field1": "value1"}`,
				CaseDataId:  111,
			},
			2: {
				Id:          2,
				Name:        "Event 2",
				CreatedDate: helper.MustParseTime(layout, "2023-07-25"),
				Data:        `{"field1": "value2"}`,
				CaseDataId:  111,
			},
		},
		2: {
			1: {
				Id:          1,
				Name:        "Event 1",
				CreatedDate: helper.MustParseTime(layout, "2023-07-25"),
				Data:        `{"field1": "value3"}`,
				CaseDataId:  222,
			},
			2: {
				Id:          2,
				Name:        "Event 1",
				CreatedDate: helper.MustParseTime(layout, "2023-07-25"),
				Data:        `{"field1": "value3"}`,
				CaseDataId:  333,
			},
		},
		3: {
			1: {
				Id:          1,
				Name:        "Event 1",
				CreatedDate: helper.MustParseTime(layout, "2023-07-25"),
				Data:        `{"field1": "value3"}`,
				CaseDataId:  222,
			},
			3: {
				Id:          3,
				Name:        "Event 3",
				CreatedDate: helper.MustParseTime(layout, "2023-07-25"),
				Data:        `{"field2": "value3"}`,
				CaseDataId:  333,
			},
		},
	}

	expectedResult := EventFieldChanges{
		"1->.field1": {
			{
				OldRecord:       "value1",
				NewRecord:       "value2",
				CreatedDate:     helper.MustParseTime(layout, "2023-07-25"),
				SourceEventId:   2,
				SourceEventName: "Event 2",
				OperationType:   Modified,
			},
		},
		"3->.field1": {
			{
				OldRecord:       "value3",
				NewRecord:       "",
				CreatedDate:     helper.MustParseTime(layout, "2023-07-25"),
				SourceEventId:   3,
				SourceEventName: "Event 3",
				OperationType:   Deleted,
			},
		},
		"3->.field2": {
			{
				OldRecord:       "",
				NewRecord:       "value3",
				CreatedDate:     helper.MustParseTime(layout, "2023-07-25"),
				SourceEventId:   3,
				SourceEventName: "Event 3",
				OperationType:   Added,
			},
		},
	}

	result := CompareEventsByCaseReference("", caseEvents)

	// Compare the result with the expected result
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Unexpected result.\nwan: %+v\nGot: %+v", expectedResult, result)
	}
}
