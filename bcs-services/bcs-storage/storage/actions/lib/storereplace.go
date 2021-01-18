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

package lib

import (
	"bytes"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

const (
	// "\uff04" is the unicode of "＄", which is the official recommend of Mongodb
	dollarReplacement = "\uff04"
)

// dollarHandler convert the "." in keys of raw to the dollarReplacement for
// Mongodb does not support key with "."
func dollarHandler(raw operator.M) operator.M {
	for k, v := range raw {
		delete(raw, k)
		key := strings.Replace(k, "$", dollarReplacement, -1)
		raw[key] = dollarHandlerIf(v)
	}
	return raw
}

func dollarHandlerIf(raw interface{}) interface{} {
	s, ok := raw.(map[string]interface{})
	if !ok {
		l, ok2 := raw.([]interface{})
		if !ok2 {
			return raw
		}

		for i, field := range l {
			l[i] = dollarHandlerIf(field)
		}
		return l
	}

	for k, v := range s {
		delete(s, k)
		key := strings.Replace(k, "$", dollarReplacement, -1)
		s[key] = dollarHandlerIf(v)
	}
	return s
}

// dollarRecover recover from the dotHandler
func dollarRecover(s operator.M) operator.M {
	var tmp []byte
	if err := codec.EncJson(s, &tmp); err != nil {
		return s
	}

	tmp = bytes.Replace(tmp, []byte(dollarReplacement), []byte("$"), -1)

	out := make(operator.M)
	if err := codec.DecJson(tmp, &out); err != nil {
		return s
	}
	return out
}
