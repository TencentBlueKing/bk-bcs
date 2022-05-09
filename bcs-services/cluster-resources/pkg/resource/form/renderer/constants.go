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

package renderer

import (
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
)

// RecursionMaxNums 模板 include 嵌套最大层数
const RecursionMaxNums = 100

// TmplRandomNameLength 模板随机名称长度
const TmplRandomNameLength = 12

// FormRenderSupportedResAPIVersion 支持表单化的资源版本
var FormRenderSupportedResAPIVersion = map[string][]string{
	res.Deploy: {"apps/v1", "extensions/v1", "extensions/v1beta1"},
	// TODO 补充其他资源类型
}
