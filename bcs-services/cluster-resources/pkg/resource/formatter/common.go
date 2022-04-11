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

package formatter

import (
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/timex"
)

// CommonFormatRes 通用资源格式化
func CommonFormatRes(manifest map[string]interface{}) map[string]interface{} {
	rawCreateTime, _ := mapx.GetItems(manifest, "metadata.creationTimestamp")
	createTime, _ := timex.NormalizeDatetime(rawCreateTime.(string))
	ret := map[string]interface{}{
		"age":        timex.CalcAge(rawCreateTime.(string)),
		"createTime": createTime,
	}
	return ret
}

// GetFormatFunc 获取资源对应 FormatFunc
func GetFormatFunc(kind string) func(manifest map[string]interface{}) map[string]interface{} {
	formatFunc, ok := Kind2FormatFuncMap[kind]
	if !ok {
		// 若指定资源类型没有对应的，则当作自定义资源处理
		return FormatCObj
	}
	return formatFunc
}
