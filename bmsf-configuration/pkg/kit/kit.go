/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package kit

import (
	"context"
	"errors"
)

// Kit is common request context kit.
type Kit struct {
	// Ctx is request context.
	Ctx context.Context

	// User is user name.
	User string

	// Rid is request id.
	Rid string

	// Method is rpc method name.
	Method string

	// AppCode is app code.
	AppCode string

	// Authorization is internal authorization info.
	// Only used for inner modules to authorize.
	Authorization string
}

// Validate checks request kit fields.
func (k Kit) Validate() error {
	if k.Ctx == nil {
		return errors.New("context is required")
	}

	if len(k.User) == 0 {
		return errors.New("user is required")
	}

	if len(k.Rid) == 0 {
		return errors.New("rid is required")
	}

	if len(k.AppCode) == 0 {
		return errors.New("app_code is required")
	}

	return nil
}
