/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package tools provides bscp common tools.
package tools

import (
	"sort"
	"strconv"
	"strings"
)

// GetIntList 获取Int列表, 解析BizID时使用
func GetIntList(input string) ([]int, error) {
	if input == "" {
		return []int{}, nil
	}

	items := strings.Split(input, ",")
	result := make([]int, 0, len(items))
	for _, v := range items {
		intValue, err := strconv.Atoi(v)
		if err != nil {
			return []int{}, err
		}
		result = append(result, intValue)
	}
	return result, nil
}

// GetUint32List convert string to uint32 list
func GetUint32List(input string) ([]uint32, error) {
	if input == "" {
		return []uint32{}, nil
	}

	items := strings.Split(input, ",")
	result := make([]uint32, 0, len(items))
	for _, v := range items {
		intValue, err := strconv.Atoi(v)
		if err != nil {
			return []uint32{}, err
		}
		result = append(result, uint32(intValue))
	}
	return result, nil
}

// SliceDiff get the difference between two slices, return slice1-slice2
func SliceDiff(slice1, slice2 []uint32) []uint32 {
	set := make(map[uint32]struct{})
	diff := make([]uint32, 0)

	for _, v := range slice2 {
		set[v] = struct{}{}
	}

	for _, v := range slice1 {
		if _, ok := set[v]; !ok {
			diff = append(diff, v)
		}
	}

	return diff
}

// SliceRepeatedElements get the repeated elements in a slice, and the keep the sequence of result
func SliceRepeatedElements(slice []uint32) []uint32 {
	frequencyMap := make(map[uint32]uint32)
	repeatedElements := make(map[uint32]bool) // To track if an element is already added
	var result []uint32

	// Iterate through the slice
	for _, num := range slice {
		// If the element has been counted before, and it's not already added to the result
		if _, exists := frequencyMap[num]; exists && !repeatedElements[num] {
			result = append(result, num)
			repeatedElements[num] = true // Mark as added
		}

		frequencyMap[num]++
	}

	return result
}

// RemoveDuplicates remove duplicate elements from a slice
func RemoveDuplicates(input []uint32) []uint32 {
	// Create a map to track unique elements
	uniqueMap := make(map[uint32]bool)
	uniqueSlice := make([]uint32, 0)

	// Iterate through the original slice
	for _, item := range input {
		// If the element is not in the map, add it to the map and new slice
		if !uniqueMap[item] {
			uniqueMap[item] = true
			uniqueSlice = append(uniqueSlice, item)
		}
	}

	return uniqueSlice
}

// RemoveDuplicateStrings remove duplicate elements from a slice of string
func RemoveDuplicateStrings(input []string) []string {
	// Create a map to track unique elements
	uniqueMap := make(map[string]bool)
	uniqueSlice := make([]string, 0)

	// Iterate through the original slice
	for _, item := range input {
		// If the element is not in the map, add it to the map and new slice
		if !uniqueMap[item] {
			uniqueMap[item] = true
			uniqueSlice = append(uniqueSlice, item)
		}
	}

	return uniqueSlice
}

// IsSameSlice judge whether two slices have same elements（including count）, no need care the order of the elements
func IsSameSlice(s1, s2 []uint32) bool {
	// Check if the lengths are equal
	if len(s1) != len(s2) {
		return false
	}

	// Create maps to count occurrences of elements in s1 and s2
	countMap1 := make(map[uint32]int)
	countMap2 := make(map[uint32]int)

	// Count occurrences in s1
	for _, num := range s1 {
		countMap1[num]++
	}

	// Count occurrences in s2
	for _, num := range s2 {
		countMap2[num]++
	}

	// Compare the occurrence counts
	for num, count1 := range countMap1 {
		count2, exists := countMap2[num]
		if !exists || count1 != count2 {
			return false
		}
	}

	return true
}

// MergeDoubleStringSlice merge [][]string elements into []string
// keep the elements in result is unique and sorted in ASCII character order
// eg: input [][]string{{"a", "b"}, {"a", "c"}} return []string{"a", "b", "c"}
func MergeDoubleStringSlice(input [][]string) []string {
	uniqueMap := make(map[string]bool)

	for _, subSlice := range input {
		for _, element := range subSlice {
			uniqueMap[element] = true
		}
	}

	uniqueElements := make([]string, 0, len(uniqueMap))
	for element := range uniqueMap {
		uniqueElements = append(uniqueElements, element)
	}

	// Sort the unique elements in ASCII character order
	sort.Strings(uniqueElements)

	return uniqueElements
}
