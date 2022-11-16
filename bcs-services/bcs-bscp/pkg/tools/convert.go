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

// Itoa convert uint32 to string
func Itoa(v uint32) string {
	return strconv.FormatUint(uint64(v), 10)
}

// SinceMS calculate the lag time since the start time, and convert the duration to float64
func SinceMS(start time.Time) float64 {
	return float64(time.Since(start).Milliseconds())
}
