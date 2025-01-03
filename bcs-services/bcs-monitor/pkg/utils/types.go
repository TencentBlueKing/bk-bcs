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

import (
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

const (
	// FSType file system type
	FSType = "ext[234]|btrfs|xfs|zfs"
)

// JSONTime format time in json marshal
type JSONTime struct {
	time.Time
}

const timeLayout = "2006-01-02 15:04:05"

// MarshalJSON marshal json
func (t *JSONTime) MarshalJSON() ([]byte, error) {
	loc, err := time.LoadLocation(config.G.Base.TimeZone)
	if err != nil {
		return nil, err
	}
	tt := t.Time.In(loc)
	return []byte(fmt.Sprintf("\"%s\"", tt.Format(timeLayout))), nil
}

// UnmarshalJSON unmarshal json
func (t *JSONTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		t.Time = time.Time{}
		return nil
	}
	var err error
	t.Time, err = time.Parse(timeLayout, s)
	return err
}
