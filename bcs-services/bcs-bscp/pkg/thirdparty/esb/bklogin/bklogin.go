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

// Package bklogin NOTES
package bklogin

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

// Client is an esb client to request bkLogin.
type Client interface {
	// IsLogin check user is login
	IsLogin(ctx context.Context, bkToken string) (string, error)
}

// NewClient initialize a new bklogin client
func NewClient(client rest.ClientInterface) Client {
	return &bklogin{
		client: client,
	}
}

// bklogin is an esb client to request bklogin comp.
type bklogin struct {
	client rest.ClientInterface
}

// IsLogin check user is login
func (c *bklogin) IsLogin(ctx context.Context, bkToken string) (string, error) {
	resp := new(IsLoginResp)

	err := c.client.Get().
		SubResourcef("/bk_login/is_login/").
		WithContext(ctx).
		WithParam("bk_token", bkToken).
		Do().Into(resp)
	if err != nil {
		return "", err
	}

	if !resp.Result || resp.Code != 0 {
		// 无权限处理
		if resp.Code == codeNotHasPerm {
			return "", errors.Wrap(errf.ErrPermissionDenied, resp.Message)
		}

		return "", fmt.Errorf("user not login, code: %d, msg: %s, rid: %s", resp.Code, resp.Message, resp.Rid)
	}

	return resp.BKUsername, nil
}
