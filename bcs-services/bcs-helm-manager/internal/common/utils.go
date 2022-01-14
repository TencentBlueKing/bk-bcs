/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"math/rand"
	"strings"
	"time"
)

// SplitAddrString split address string
func SplitAddrString(address string) []string {
	address = strings.Replace(address, ";", ",", -1)
	address = strings.Replace(address, " ", ",", -1)
	return strings.Split(address, ",")
}

// MapInt2MapIf convert map[string]int to map[string]interface{}
func MapInt2MapIf(m map[string]int) map[string]interface{} {
	newM := make(map[string]interface{})
	for k, v := range m {
		newM[k] = v
	}
	return newM
}

// GetStoredTimestamp return the timestamp of given time for storage
func GetStoredTimestamp(t time.Time) int64 {
	return t.UTC().Unix()
}

// GetStoredTime return the time.Time of given timestamp from storage
func GetStoredTime(t int64) time.Time {
	return time.Unix(t, 0).UTC()
}

// GetStringP return ptr of copied string
func GetStringP(s string) *string {
	p := s
	return &p
}

// GetInt64P return ptr of copied int64
func GetInt64P(s int64) *int64 {
	p := s
	return &p
}

// GetBoolP return ptr of copied bool
func GetBoolP(s bool) *bool {
	p := s
	return &p
}

// GetUint32P return ptr of copied uint32
func GetUint32P(s uint32) *uint32 {
	p := s
	return &p
}

// GetFloat64P return ptr of copied float64
func GetFloat64P(s float64) *float64 {
	p := s
	return &p
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomString get a random string made up of alphabet characters with given length.
func RandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// RandomInt get a random int less than limit
func RandomInt(limit int) int {
	return rand.Intn(limit)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
