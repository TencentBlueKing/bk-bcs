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

package util

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
)

// CreateResponeData create response data
func CreateResponeData(err error, code int, msg string, data interface{}) []byte {
	var rpyErr error
	if err != nil {
		rpyErr = bhttp.InternalError(code, msg)
	} else {
		rpyErr = fmt.Errorf(bhttp.GetRespone(0, "successful ", data))
	}

	blog.V(3).Infof("createRespone: %s", rpyErr.Error())

	return []byte(rpyErr.Error())
}
