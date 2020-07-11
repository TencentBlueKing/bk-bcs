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

package mongodb

import (
	"bytes"
	"math"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
)

const (
	// "\uff0e" is the unicode of ".", which is the official recommend of Mongodb
	dotReplacement = "\uff0e"
)

// dotHandler convert the "." in keys of raw to the dotReplacement for
// Mongodb does not support key with "."
func dotHandler(raw interface{}) interface{} {
	s, ok := raw.(map[string]interface{})
	if !ok {
		l, ok2 := raw.([]interface{})
		if !ok2 {
			return raw
		}

		for i, field := range l {
			l[i] = dotHandler(field)
		}
		return l
	}

	for k, v := range s {
		if vv, ok3 := v.(float64); ok3 {
			if vv == math.Trunc(vv) {
				s[k] = int64(vv)
			}
			continue
		}
		delete(s, k)
		key := strings.Replace(k, ".", dotReplacement, -1)
		s[key] = dotHandler(v)
	}
	return s
}

// dotRecover recover from the dotHandler
func dotRecover(s []interface{}) (r []interface{}) {
	var tmp []byte
	if err := codec.EncJson(s, &tmp); err != nil {
		return s
	}

	tmp = bytes.Replace(tmp, []byte(dotReplacement), []byte("."), -1)

	if err := codec.DecJson(tmp, &r); err != nil {
		return s
	}
	return
}
