package tools

import (
	"reflect"
	"testing"
)

func TestGetIntList(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
		err      bool
	}{
		{"1,2,3,4,5", []int{1, 2, 3, 4, 5}, false},
		{"", []int{}, false},
		{"1,invalid,3", []int{}, true},
	}

	for _, test := range tests {
		result, err := GetIntList(test.input)

		if test.err {
			if err == nil {
				t.Errorf("Expected error for input: %s", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input: %s, err: %s", test.input, err)
			}

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("For input %s, expected %v, but got %v", test.input, test.expected, result)
			}
		}
	}
}

func TestGetUint32List(t *testing.T) {
	tests := []struct {
		input    string
		expected []uint32
		err      bool
	}{
		{"1,2,3,4,5", []uint32{1, 2, 3, 4, 5}, false},
		{"", []uint32{}, false},
		{"1,invalid,3", []uint32{}, true},
	}

	for _, test := range tests {
		result, err := GetUint32List(test.input)

		if test.err {
			if err == nil {
				t.Errorf("Expected error for input: %s", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input: %s, err: %s", test.input, err)
			}

			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("For input %s, expected %v, but got %v", test.input, test.expected, result)
			}
		}
	}
}

func TestSliceDiff(t *testing.T) {
	slice1 := []uint32{1, 2, 3, 4, 5}
	slice2 := []uint32{3, 4, 5, 6, 7}
	expected := []uint32{1, 2}

	result := SliceDiff(slice1, slice2)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected difference: %v, but got %v", expected, result)
	}
}

func TestSliceRepeatedElements(t *testing.T) {
	slice := []uint32{1, 2, 2, 3, 4, 4, 5, 5, 5}
	expected := []uint32{2, 4, 5}

	result := SliceRepeatedElements(slice)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected repeated elements: %v, but got %v", expected, result)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		input    []uint32
		expected []uint32
	}{
		{[]uint32{1, 2, 3, 2, 4, 1, 5, 6, 3}, []uint32{1, 2, 3, 4, 5, 6}},
		{[]uint32{2, 2, 2, 2, 2}, []uint32{2}},
		{[]uint32{}, []uint32{}},
		{[]uint32{1, 2, 3, 4}, []uint32{1, 2, 3, 4}},
	}

	for _, test := range tests {
		result := RemoveDuplicates(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("For input %v, expected %v, but got %v", test.input, test.expected, result)
		}
	}
}
