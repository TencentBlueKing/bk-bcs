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

package tools

import (
	"strconv"
	"strings"
	"time"
)

// JoinUint32 joint the uint slice to a string
func JoinUint32(list []uint32, sep string) string {

	b := make([]string, len(list))
	for i, v := range list {
		b[i] = strconv.FormatUint(uint64(v), 10)
	}

	return strings.Join(b, sep)
}

// StringSliceToUint32Slice converts a string slice to a uint32 slice
func StringSliceToUint32Slice(strSlice []string) ([]uint32, error) {
	uintSlice := make([]uint32, len(strSlice))
	for _, str := range strSlice {
		i, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}
		uintSlice = append(uintSlice, uint32(i))
	}
	return uintSlice, nil
}

// Itoa convert uint32 to string
func Itoa(v uint32) string {
	return strconv.FormatUint(uint64(v), 10)
}

// SinceMS calculate the lag time since the start time, and convert the duration to float64
func SinceMS(start time.Time) float64 {
	return float64(time.Since(start).Milliseconds())
}

// StrToUint32Slice the comma separated string goes to uint32 slice
func StrToUint32Slice(str string) ([]uint32, error) {
	strValues := strings.Split(str, ",")
	// 创建一个切片用于存储uint32值
	var uint32Slice []uint32
	// 遍历拆分后的字符串切片
	for _, strValue := range strValues {
		// 使用strconv.Atoi将字符串转换为int
		intValue, err := strconv.Atoi(strValue)
		if err != nil {
			return nil, err
		}
		// 将int值转换为uint32并添加到切片中
		uint32Slice = append(uint32Slice, uint32(intValue))
	}
	return uint32Slice, nil
}

// Contains checks if a given value exists in a slice
func Contains(slice []uint32, value uint32) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// IsNumber check if the string is a number
func IsNumber(s string) bool {
	// Try to convert to an integer
	_, err := strconv.Atoi(s)
	if err == nil {
		return true
	}

	// Try to convert to a float64
	_, err = strconv.ParseFloat(s, 64)
	return err == nil
}
