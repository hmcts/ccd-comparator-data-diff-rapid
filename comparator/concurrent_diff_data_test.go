package comparator

import (
	"reflect"
	"testing"
)

func TestConcurrentEventFieldDifferences_SetAndGet(t *testing.T) {
	ced := NewConcurrentEventFieldDifferences()

	key := "key1"
	value := []EventFieldChange{
		{OldRecord: "old1", NewRecord: "new1"},
		{OldRecord: "old2", NewRecord: "new2"},
	}
	ced.Set(key, value)

	retrievedValue, ok := ced.Get(key)
	if !ok {
		t.Errorf("TestConcurrentEventFieldDifferences_SetAndGet: Failed to retrieve value for key '%s'", key)
	}

	if !reflect.DeepEqual(retrievedValue, value) {
		t.Errorf("TestConcurrentEventFieldDifferences_SetAndGet: Expected value '%v', but got '%v'", value, retrievedValue)
	}

	nonExistentKey := "non_existent_key"
	_, ok = ced.Get(nonExistentKey)
	if ok {
		t.Errorf("TestConcurrentEventFieldDifferences_SetAndGet: Retrieved value for non-existent key '%s'", nonExistentKey)
	}
}

func TestConcurrentEventFieldDifferences_Delete(t *testing.T) {
	ced := NewConcurrentEventFieldDifferences()

	key := "key1"
	value := []EventFieldChange{
		{OldRecord: "old1", NewRecord: "new1"},
	}
	ced.Set(key, value)

	ced.Delete(key)

	_, ok := ced.Get(key)
	if ok {
		t.Errorf("TestConcurrentEventFieldDifferences_Delete: Retrieved value for deleted key '%s'", key)
	}
}

func TestConcurrentEventFieldDifferences_Size(t *testing.T) {
	ced := NewConcurrentEventFieldDifferences()

	size := ced.Size()
	if size != 0 {
		t.Errorf("TestConcurrentEventFieldDifferences_Size: Expected size 0 for empty map, but got %d", size)
	}

	key1 := "key1"
	value1 := []EventFieldChange{
		{OldRecord: "old1", NewRecord: "new1"},
	}
	ced.Set(key1, value1)

	key2 := "key2"
	value2 := []EventFieldChange{
		{OldRecord: "old2", NewRecord: "new2"},
		{OldRecord: "old3", NewRecord: "new3"},
	}
	ced.Set(key2, value2)

	size = ced.Size()
	if size != 2 {
		t.Errorf("TestConcurrentEventFieldDifferences_Size: Expected size 2, but got %d", size)
	}
}

func TestConcurrentEventFieldDifferences_Clear(t *testing.T) {
	ced := NewConcurrentEventFieldDifferences()

	key1 := "key1"
	value1 := []EventFieldChange{
		{OldRecord: "old1", NewRecord: "new1"},
	}
	ced.Set(key1, value1)

	key2 := "key2"
	value2 := []EventFieldChange{
		{OldRecord: "old2", NewRecord: "new2"},
		{OldRecord: "old3", NewRecord: "new3"},
	}
	ced.Set(key2, value2)

	ced.Clear()

	size := ced.Size()
	if size != 0 {
		t.Errorf("TestConcurrentEventFieldDifferences_Clear: Expected size 0 after clearing, but got %d", size)
	}
}

func TestConcurrentEventFieldDifferences_PutAll(t *testing.T) {
	ced := NewConcurrentEventFieldDifferences()

	otherMap := make(EventFieldChanges)
	otherMap["key1"] = []EventFieldChange{
		{OldRecord: "old1", NewRecord: "new1"},
	}
	otherMap["key2"] = []EventFieldChange{
		{OldRecord: "old2", NewRecord: "new2"},
		{OldRecord: "old3", NewRecord: "new3"},
	}

	ced.PutAll(otherMap)

	key1Value, ok := ced.Get("key1")
	if !ok {
		t.Errorf("TestConcurrentEventFieldDifferences_PutAll: Failed to retrieve value for key 'key1'")
	}
	if !reflect.DeepEqual(key1Value, otherMap["key1"]) {
		t.Errorf("TestConcurrentEventFieldDifferences_PutAll: Incorrect value for key 'key1'")
	}

	key2Value, ok := ced.Get("key2")
	if !ok {
		t.Errorf("TestConcurrentEventFieldDifferences_PutAll: Failed to retrieve value for key 'key2'")
	}
	if !reflect.DeepEqual(key2Value, otherMap["key2"]) {
		t.Errorf("TestConcurrentEventFieldDifferences_PutAll: Incorrect value for key 'key2'")
	}
}
