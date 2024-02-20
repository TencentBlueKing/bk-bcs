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

package dynamicquery

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

type qFilter interface {
	// GetCondition get the condition for querying
	GetCondition() *operator.Condition
}

// NOCC:golint/unparam(设计如此:)
// nolint
func qGenerate(q qFilter, timeLayout string) *operator.Condition {
	typeOf := reflect.TypeOf(q)
	n := typeOf.NumField()

	condList := make([]*operator.Condition, 0)
	for i := 0; i < n; i++ {
		field := typeOf.Field(i)

		var tag, op, value string
		var allowNoExists bool
		// NOCC:revive/early-return(设计如此:)
		// nolint
		if tagRaw := field.Tag.Get("filter"); tagRaw != "" {
			tagList := strings.Split(tagRaw, ",")
			tag = tagList[0]
			if len(tagList) > 1 {
				op = tagList[1]
			}
			if len(tagList) > 2 && tagList[2] == "allowNoExists" {
				allowNoExists = true
			}
		} else {
			continue
		}

		// NOCC:revive/early-return(设计如此:)
		// nolint
		if valueRaw := reflect.ValueOf(q).FieldByName(field.Name); valueRaw.Type().Kind() == reflect.String {
			if value = valueRaw.String(); value == "" {
				continue
			}
		} else {
			continue
		}

		var sub *operator.Condition
		// extra filter
		switch op {
		case "timeL":
			// NOCC:revive/early-return(设计如此:)
			// nolint
			if t, err := getTime(value, timeLayout); err == nil {
				sub = operator.NewLeafCondition(operator.Gt, operator.M{tag: t})
			} else {
				continue
			}
		case "timeR":
			// NOCC:revive/early-return(设计如此:)
			// nolint
			if t, err := getTime(value, timeLayout); err == nil {
				sub = operator.NewLeafCondition(operator.Lt, operator.M{tag: t})
			} else {
				continue
			}
		case "int64":
			if v, err := strconv.ParseInt(value, 10, 0); err == nil {
				sub = operator.NewLeafCondition(operator.Eq, operator.M{tag: v})
			}
		case "int":
			if v, err := strconv.Atoi(value); err == nil {
				sub = operator.NewLeafCondition(operator.Eq, operator.M{tag: v})
			}
		case "bool":
			v := bool(strings.ToUpper(value) == "TRUE" || value == "1")
			sub = operator.NewLeafCondition(operator.Eq, operator.M{tag: v})
		default:
			sub = operator.NewLeafCondition(operator.In, operator.M{tag: strings.Split(value, ",")})
		}

		if allowNoExists {
			sub = operator.NewBranchCondition(operator.Or,
				sub, operator.NewLeafCondition(operator.Ext, operator.M{tag: false}))
		}
		condList = append(condList, sub)
	}
	r := operator.NewBranchCondition(operator.And, condList...)
	return r
}

func getTime(timeStr, layout string) (r interface{}, err error) {
	var tmp int64
	tmp, err = strconv.ParseInt(timeStr, 10, 64)
	t := time.Unix(tmp, 0)

	if err != nil {
		return
	}

	if layout == timestampsLayout {
		r = t.Unix()
	} else {
		r = t.Format(layout)
	}
	return r, nil
}
