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

package errorx

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/stringx"
)

// PermissionDeniedError 权限异常内容
type PermissionDeniedError struct {
	code     uint32
	msg      string
	applyUrl string
	actionID string
	hasPerm  bool
}

// Error ...
func (e *PermissionDeniedError) Error() string {
	return e.msg
}

// Code ...
func (e *PermissionDeniedError) Code() uint32 {
	return e.code
}

// ApplyUrl ...
func (e *PermissionDeniedError) ApplyUrl() string {
	return e.applyUrl
}

// ActionID ...
func (e *PermissionDeniedError) ActionID() string {
	return e.actionID
}

// HasPerm ...
func (e *PermissionDeniedError) HasPerm() bool {
	return e.hasPerm
}

// NewPermDeniedErr ...
func NewPermDeniedErr(applyUrl string, actionID string, hasPerm bool, msg ...string) *PermissionDeniedError {
	return &PermissionDeniedError{
		code:     PermDeniedErr,
		msg:      fmt.Sprintf("%s,%s", PermDeniedErrMsg, stringx.JoinString(msg...)),
		applyUrl: applyUrl,
		actionID: actionID,
		hasPerm:  hasPerm,
	}
}
