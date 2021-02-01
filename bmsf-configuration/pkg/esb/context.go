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

package esb

import (
	"errors"
)

// Context is esb context.
type Context struct {
	// AppCode is bk app code.
	AppCode string

	// AppSecret is bk app secret.
	AppSecret string

	// User is bk app user.
	User string
}

// Validate validates context.
func (ctx *Context) Validate() error {
	if ctx == nil {
		return errors.New("empty ctx")
	}

	if len(ctx.AppCode) == 0 {
		return errors.New("empty app code")
	}

	if len(ctx.AppSecret) == 0 {
		return errors.New("empty app secret")
	}

	if len(ctx.User) == 0 {
		return errors.New("empty user")
	}
	return nil
}
