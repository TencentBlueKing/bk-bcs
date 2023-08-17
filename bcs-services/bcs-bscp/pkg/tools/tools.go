/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package tools

import (
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
