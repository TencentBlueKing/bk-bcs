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

package check

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-public-cluster-webhook/pkg/util"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v1"
)

const (
	// OperationIn 匹配符号
	OperationIn = "in"
	// OperationNotIn 不配配符号
	OperationNotIn = "notIn"
)

// BlackList 检查资源是否在过滤列表中
type BlackList struct {
	cfg *BlackListConfig
}

// NewBlackList new
func NewBlackList(configFile string) (*BlackList, error) {
	var cfg BlackListConfig
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read config file %s error, err %v", configFile, err)
	}
	if err := yaml.Unmarshal(configData, &cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal error, err %v", err)
	}
	return &BlackList{&cfg}, nil
}

// Check 检查k8s对象是否在黑名单中
func (blackList *BlackList) Check(req *RequestCheck) (*ResponseCheck, error) {
	allowedRes := &ResponseCheck{
		Allowed: true,
	}
	notAllowedRes := &ResponseCheck{
		Allowed: false,
		Message: "not allowed",
	}
	//过滤命名空间
	if util.CheckInStringArray(blackList.cfg.ExcludeNamespace, req.Namespace) {
		return allowedRes, nil
	}

	//pretty.Println(blackList.cfg)
	//匹配单个规则
	for _, rule := range blackList.cfg.List {
		//当前规则是否不校验命名空间
		if util.CheckInStringArray(rule.ExcludeNamespace, req.Namespace) {
			continue
		}
		//是否匹配资源类型
		if !util.CheckInStringArray(rule.ResourceType, req.Kind) {
			continue
		}

		//matchJson 是否匹配
		match, err := blackList.matchJSON(req.Object, rule.MatchQuery)
		if err != nil {
			return notAllowedRes, fmt.Errorf("json match error, err=%v", err)
		}
		if !match {
			continue
		}
		//匹配返回结果
		notAllowedRes.Message = rule.Message
		return notAllowedRes, nil
	}
	return allowedRes, nil
}

func (blackList *BlackList) matchJSON(data []byte, matchQuerys []*MatchQuery) (bool, error) {
	//pretty.Println(matchQuerys)
	match := true
	for _, matchQuery := range matchQuerys {
		//matchQuerys 中的规则全部匹配才算匹配成功
		if !match {
			return false, nil
		}
		//jsonPath查找
		matchResult := gjson.GetBytes(data, matchQuery.JSONPath)
		if !matchResult.Exists() {
			//未匹配
			return false, nil
		}
		switch matchResult.Type {
		case gjson.Null:
			//没有结果
			return false, nil
		case gjson.String, gjson.Number, gjson.True, gjson.False:
			if strings.ToLower(matchQuery.Operation) == OperationIn {
				for _, v := range matchQuery.Value {
					if v == matchResult.String() {
						match = true
						break
					}
					match = false
				}
			} else {
				for _, v := range matchQuery.Value {
					if v == matchResult.String() {
						match = false
						break
					}
					match = true
				}
			}
		case gjson.JSON:
			if !matchResult.IsArray() {
				blog.Warnf("暂不支持非复杂类型的规则匹配")
				return false, nil
			}
			//array匹配，默认都是in的情况
			if strings.ToLower(matchQuery.Operation) == OperationIn {
			matchInArrayLoop:
				for _, matchRule := range matchQuery.Value {
					for _, result := range matchResult.Array() {
						if matchRule == result.String() {
							match = true
							//有交集，说明当前匹配成功
							break matchInArrayLoop
						}
						match = false
					}
				}
			} else {
			matchNotInArrayLoop:
				for _, matchRule := range matchQuery.Value {
					for _, result := range matchResult.Array() {
						if matchRule == result.String() {
							match = false
							//有交集，说明当前匹配成功
							break matchNotInArrayLoop
						}
						match = true
					}
				}
			}
		}
	}
	return match, nil
}
