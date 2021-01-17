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

package lib

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// CopyMap return a new operator.M copy from provided by depth=1
func CopyMap(s operator.M) (r operator.M) {
	r = make(operator.M)
	for k, v := range s {
		r[k] = v
	}
	return
}
