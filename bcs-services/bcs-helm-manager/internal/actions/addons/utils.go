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

package addons

import (
	"errors"
	"time"
)

var errorAddonsNotFound = errors.New("addons not found")

// DefaultArgs is the default args for helm install
var defaultArgs = []string{"--wait=true", "--create-namespace=true"}

const (
	releaseDefaultTimeout = time.Hour
)

// 某些组件不需要真实安装，直接返回安装成功，因此 chart 为空
// db-privilege 组件不需要真实安装
func isFakeChart(chart string) bool {
	return chart == "" || chart == "db-privilege"
}
