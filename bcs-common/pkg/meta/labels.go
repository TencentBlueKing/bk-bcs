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

package meta

import (
	"sort"
	"strings"
)

//Labels label definition
type Labels map[string]string

//String implement string interface
func (lb Labels) String() string {
	s := make([]string, 0, len(lb))
	for k, v := range lb {
		s = append(s, k+"="+v)
	}
	sort.StringSlice(s).Sort()
	return strings.Join(s, ",")
}

//Has check key existence
func (lb Labels) Has(key string) bool {
	_, ok := lb[key]
	return ok
}

//Get get key value
func (lb Labels) Get(key string) string {
	return lb[key]
}

//LabelsMerge merge two Labels into one
//if keys are conflict, keys in y will be reserved
func LabelsMerge(x, y Labels) Labels {
	z := Labels{}
	for k, v := range x {
		z[k] = v
	}
	for k, v := range y {
		z[k] = v
	}
	return z
}

//LabelsConflict means two labels have some key but different value
func LabelsConflict(x, y Labels) bool {
	small := x
	big := y
	if len(x) > len(y) {
		small = y
		big = x
	}
	for k, v := range small {
		if other, ok := big[k]; ok {
			if other != v {
				return true
			}
		}
	}
	return false
}

//LabelsAllMatch check key/value in Labels X are totally matched
// in Y, **please pay more attention**: function return
//true even if x is nil or empty
func LabelsAllMatch(x, y Labels) bool {
	if len(x) == 0 {
		return true
	}
	for k, v := range x {
		other, ok := y[k]
		if !ok {
			return false
		}
		if v != other {
			return false
		}
	}
	return true
}

//StringToLabels converts string to Labels
//string formate likes a=vlaue,b=value2, no space inside
func StringToLabels(s string) Labels {
	kvs := strings.Split(s, ",")
	if len(kvs) == 0 {
		return nil
	}
	m := make(map[string]string)
	for _, raw := range kvs {
		kv := strings.Split(raw, "=")
		if len(kv) != 2 {
			continue
		}
		m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	if len(m) == 0 {
		return nil
	}
	return Labels(m)
}
