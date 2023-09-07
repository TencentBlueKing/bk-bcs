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
 *
 */

package audit

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"
)

// Logger is an interface for logging.
type Logger interface {
	Info(args ...interface{})
}

// Log is a logger.
type Log struct {
}

// Info log info
func (l *Log) Info(args ...interface{}) {
	fmt.Println(args...)
}

// SplitSlice splits a slice of any type into smaller slices of given length.
// It returns a 2D slice of the same type as the input slice.
// If the length of the input slice is not divisible by the given length, the last slice may have fewer elements.
func SplitSlice[T any](slice []T, length int) [][]T {
	var result [][]T
	for i := 0; i < len(slice); i += length {
		end := i + length
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}
	return result
}

// GetEnvWithDefault takes two string parameters, key and defaultValue.
// It uses the os.Getenv function to retrieve the value of the environment variable specified by key.
// If the value is an empty string, it returns the defaultValue parameter.
// Otherwise, it returns the value of the environment variable.
func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GenerateEventID generate event id, format: app_code-YYYYMMDDHHMMSS-substring(MD5(随机因子)),8,24)
func GenerateEventID(appCode, factor string) string {
	currentTime := time.Now().Format("20060102150405")
	// NOCC:gas/crypto(设计如此)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(factor)))
	result := fmt.Sprintf("%s-%s-%s", appCode, currentTime, hash[8:24])
	return result
}
