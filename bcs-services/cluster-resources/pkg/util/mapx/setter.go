/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mapx

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// SetItems 对嵌套 Map 进行赋值
// paths 参数支持 []string 类型，如 []string{"metadata", "namespace"}
// 或 string 类型（以 '.' 为分隔符），如 "spec.template.spec.containers"
func SetItems(obj map[string]interface{}, paths interface{}, val interface{}) error {
	// 检查 paths 类型
	switch t := paths.(type) {
	case string:
		if err := setItems(obj, strings.Split(paths.(string), "."), val); err != nil {
			return err
		}
	case []string:
		if err := setItems(obj, paths.([]string), val); err != nil {
			return err
		}
	default:
		return errorx.New(errcode.General, "paths's type must one of (string, []string), get %v", t)
	}
	return nil
}

func setItems(obj map[string]interface{}, paths []string, val interface{}) error {
	if len(paths) == 0 {
		return errorx.New(errcode.General, "paths is empty list")
	}
	if len(paths) == 1 {
		obj[paths[0]] = val
	} else if subMap, ok := obj[paths[0]].(map[string]interface{}); ok {
		return setItems(subMap, paths[1:], val)
	} else {
		return errorx.New(errcode.General, "key %s not exists or obj[key] not map[string]interface{} type", paths[0])
	}
	return nil
}
