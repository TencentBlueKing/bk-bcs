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

package lib

import (
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	restful "github.com/emicklei/go-restful/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	defaultMaxMemory    = 32 << 20 // 32 MB
	labelSelectorTag    = "labelSelector"
	labelSelectorPrefix = "data.metadata.labels."
	extraTag            = "extra"
	fieldTag            = "field"
	offsetTag           = "offset"
	limitTag            = "limit"
	updateTimeQueryTag  = "updateTimeBefore"
	updateTimeTag       = "updateTime"
)

// GetQueryParamString get string from rest query parameter
func GetQueryParamString(req *restful.Request, key string) string {
	return req.QueryParameter(key)
}

// GetQueryParamStringArray get string array from restful query parameter
func GetQueryParamStringArray(req *restful.Request, key, sep string) []string {
	s := req.QueryParameter(key)
	if len(s) == 0 {
		return nil
	}
	fields := strings.Split(s, sep)
	return fields
}

// GetQueryParamInt get int from restful query parameter
func GetQueryParamInt(req *restful.Request, key string, defaultValue int) (int, error) {
	s := req.QueryParameter(key)
	if len(s) == 0 {
		return defaultValue, nil
	}
	return strconv.Atoi(s)
}

// GetQueryParamInt64 get int64 from restful query parameter
func GetQueryParamInt64(req *restful.Request, key string, defaultValue int64) (int64, error) {
	s := req.QueryParameter(key)
	if len(s) == 0 {
		return defaultValue, nil
	}
	return strconv.ParseInt(s, 10, 64)
}

// GetJsonParamStringArray get string array from json map parameter
func GetJsonParamStringArray(params map[string]string, key, sep string) []string {
	s := params[key]
	if len(s) == 0 {
		return nil
	}
	fields := strings.Split(s, sep)
	return fields
}

// GetJsonParamInt get int from json map parameter
func GetJsonParamInt(params map[string]string, key string, defaultValue int) (int, error) {
	s := params[key]
	if len(s) == 0 {
		return defaultValue, nil
	}
	return strconv.Atoi(s)
}

// GetJsonParamInt64 get int64 from json map parameter
func GetJsonParamInt64(params map[string]string, key string, defaultValue int64) (int64, error) {
	s := params[key]
	if len(s) == 0 {
		return defaultValue, nil
	}
	return strconv.ParseInt(s, 10, 64)
}

func buildLeafCondition(key, value, sep string, op operator.Operator) *operator.Condition {
	valueList := strings.Split(value, sep)
	if len(valueList) == 1 && strings.TrimSpace(valueList[0]) == "" {
		return operator.NewLeafCondition(operator.Ext, key)
	}
	return operator.NewLeafCondition(op, operator.M{key: valueList})
}

// buildSelectorCondition parse labelSelector
func buildSelectorCondition(prefix string, valueList []string) *operator.Condition {
	valueStr := strings.Join(valueList, ",")
	selector := &Selector{Prefix: prefix, SelectorStr: valueStr}
	conds := selector.GetAllConditions()
	// TODO error deal
	if conds == nil {
		return operator.NewLeafCondition(operator.Tr, operator.M{})
	}
	return operator.NewBranchCondition(operator.And, conds...)
}

// buildNormalCondition parse normal parameters
func buildNormalCondition(key string, valueList []string) *operator.Condition {
	if len(valueList) == 0 {
		return operator.NewLeafCondition(operator.Ext, key)
	}
	if len(valueList) == 1 && strings.TrimSpace(valueList[0]) == "" {
		return operator.NewLeafCondition(operator.Ext, key)
	}
	conds := make([]*operator.Condition, 0)
	for _, value := range valueList {
		conds = append(conds, buildLeafCondition(key, value, ",", operator.In))
	}
	return operator.NewBranchCondition(operator.And, conds...)
}

// GetCustomCondition get custom condition from req url and parameter
func GetCustomCondition(req *restful.Request) *operator.Condition {
	if req.Request.Form == nil {
		_ = req.Request.ParseMultipartForm(defaultMaxMemory)
	}
	if len(req.Request.Form) == 0 {
		return nil
	}
	return GetCustomConditionFromBody(req.Request.Form)
}

// GetCustomConditionFromBody get custom condition from req body
func GetCustomConditionFromBody(body map[string][]string) *operator.Condition {
	conds := make([]*operator.Condition, 0)
	// empty Condition
	rootCondition := operator.NewLeafCondition(operator.Tr, operator.M{})
	for key, valueList := range body {
		switch key {
		// labelSelector=tag1=val1,tag2+in+(v1,v2),tag3+notin+(v1,v2),tag4+!=+(v1,v2),tag5
		case labelSelectorTag:
			conds = append(conds, buildSelectorCondition(labelSelectorPrefix, valueList))
		case extraTag, fieldTag, limitTag, offsetTag:
			// nolint
			break
		// updateTimeBefore=
		case updateTimeQueryTag:
			var t time.Time
			ts, err := strconv.ParseInt(valueList[0], 10, 64)
			if err != nil {
				var innerErr error
				t, innerErr = time.Parse("2006-01-02T15:04:05.000Z", valueList[0])
				if innerErr != nil {
					// nolint
					blog.Errorf("Unrecognized update time (%s) format, neither timestamp in seconds format nor time expression like 2006-01-02T15:04:05.000Z", valueList[0])
					break
				}
			} else {
				t = time.Unix(ts, 0)
			}
			conds = append(conds, operator.NewLeafCondition(operator.Lt, operator.M{updateTimeTag: t}))
		// col1=val1,val2
		default:
			conds = append(conds, buildNormalCondition(key, valueList))
		}
	}
	if len(conds) != 0 {
		rootCondition = operator.NewBranchCondition(operator.And, conds...)
	}
	return rootCondition
}

// FormatTime format time
func FormatTime(data []operator.M, needTimeFormatList []string) {
	// Some time-field need to be format before return
	for i := range data {
		for _, t := range needTimeFormatList {
			tmp, ok := data[i][t].(primitive.DateTime)
			if !ok {
				continue
			}
			data[i][t] = tmp.Time()
		}
	}
}
