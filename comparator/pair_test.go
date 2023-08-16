package comparator

import (
	"reflect"
	"testing"
)

func TestNewPair(t *testing.T) {
	pair1 := NewPair(42, "hello")
	expectedPair1 := &Pair[int, string]{Left: 42, Right: "hello"}

	if !reflect.DeepEqual(pair1, expectedPair1) {
		t.Errorf("TestNewPair: Expected %v, but got %v", expectedPair1, pair1)
	}

	pair2 := NewPair(true, 3.14)
	expectedPair2 := &Pair[bool, float64]{Left: true, Right: 3.14}

	if !reflect.DeepEqual(pair2, expectedPair2) {
		t.Errorf("TestNewPair: Expected %v, but got %v", expectedPair2, pair2)
	}

	type MyStruct struct {
		Field1 string
		Field2 int
	}

	myStruct1 := MyStruct{"abc", 123}
	myStruct2 := MyStruct{"xyz", 456}

	pair3 := NewPair(myStruct1, myStruct2)
	expectedPair3 := &Pair[MyStruct, MyStruct]{Left: myStruct1, Right: myStruct2}

	if !reflect.DeepEqual(pair3, expectedPair3) {
		t.Errorf("TestNewPair: Expected %v, but got %v", expectedPair3, pair3)
	}
}
