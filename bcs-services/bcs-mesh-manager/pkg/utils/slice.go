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

package utils

// MergeSlices 合并多个切片，返回一个新的切片
func MergeSlices[T any](slices ...[]T) []T {
	if len(slices) == 0 {
		return nil
	}

	// 计算总长度
	totalLen := 0
	for _, slice := range slices {
		totalLen += len(slice)
	}

	// 预分配容量
	result := make([]T, 0, totalLen)

	// 合并所有切片
	for _, slice := range slices {
		result = append(result, slice...)
	}

	return result
}

// Contains 检查切片中是否包含指定元素
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Difference 返回在 sliceOne 中但不在 sliceTwo 中的元素
func Difference[T comparable](sliceOne, sliceTwo []T) []T {
	excludeMap := make(map[T]struct{}, len(sliceTwo))
	for _, item := range sliceTwo {
		excludeMap[item] = struct{}{}
	}

	// 预分配 result 切片容量，避免多次内存分配
	result := make([]T, 0, len(sliceOne))
	for _, item := range sliceOne {
		if _, exists := excludeMap[item]; !exists {
			result = append(result, item)
		}
	}
	return result
}
