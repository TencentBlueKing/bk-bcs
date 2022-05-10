/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
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
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// TimeSequence time sequence
func TimeSequence() uint64 {
	return uint64(time.Now().UnixNano() / 1e6)
}

// ParseCIDR parse cidr to ip and mask
func ParseCIDR(cidr string) (string, int, error) {
	if _, _, err := net.ParseCIDR(cidr); err != nil {
		blog.Errorf("parse cidr %s addr failed, err %s", cidr, err.Error())
		return "", 0, err
	}
	strs := strings.Split(cidr, "/")
	if len(strs) != 2 {
		blog.Errorf("cidr %s format error", cidr)
		return "", 0, fmt.Errorf("cidr %s format error", cidr)
	}
	mask, _ := strconv.Atoi(strs[1])
	return strs[0], mask, nil
}

// ContainsString to see if slice contains string
func ContainsString(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}

// RemoveString remove string from slice
func RemoveString(strs []string, str string) []string {
	var newSlice []string
	for _, s := range strs {
		if s != str {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}

// FormatTime format time to utc string
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

// ParseTimeString parse utc string to time object
func ParseTimeString(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

// MaxInt get max value between two int numbers
func MaxInt(x, y int) int {
	if x >= y {
		return x
	}
	return y
}

// MinInt get min value between two int numbers
func MinInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func ToJsonString(i interface{}) string {
	b, _ := json.Marshal(i)
	return string(b)
}
