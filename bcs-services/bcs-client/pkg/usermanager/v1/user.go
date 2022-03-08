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

package v1

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/permission"
)

type UserManager interface {
	CreateOrGetUser(userType string, userName string, method string) (*models.BcsUser, error)
	RefreshUsertoken(userType string, userName string) (*models.BcsUser, error)
	GrantOrRevokePermission(method string, data []byte) error
	GetPermission(method string, data []byte) ([]permission.PermissionsResp, error)
	AddVpcCidrs(data []byte) error
}

type bcsUserManager struct {
	bcsAPIAddress string
	requester     utils.ApiRequester
}

//NewBcsUserManager create bcs-user-manager api implemenation
func NewBcsUserManager(options types.ClientOptions) UserManager {
	return &bcsUserManager{
		bcsAPIAddress: options.BcsApiAddress,
		requester:     utils.NewApiRequester(options.ClientSSL, options.BcsToken),
	}
}

func (b *bcsUserManager) CreateOrGetUser(userType string, userName string, method string) (*models.BcsUser, error) {
	resp, err := b.requester.Do(
		fmt.Sprintf(BcsUserManagerUserURI, b.bcsAPIAddress, userType, userName),
		method,
		nil,
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("create or get %s user %s failed: %s", userType, userName, msg)
	}

	var result models.BcsUser
	err = codec.DecJson(data, &result)
	return &result, err
}

func (b *bcsUserManager) RefreshUsertoken(userType string, userName string) (*models.BcsUser, error) {
	method := http.MethodPut
	resp, err := b.requester.Do(
		fmt.Sprintf(BcsUserManagerUserRefreshURI, b.bcsAPIAddress, userType, userName),
		method,
		nil,
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("refresh usertoken for  %s user %s failed: %s", userType, userName, msg)
	}

	var result models.BcsUser
	err = codec.DecJson(data, &result)
	return &result, err
}

func (b *bcsUserManager) GrantOrRevokePermission(method string, data []byte) error {
	resp, err := b.requester.Do(
		fmt.Sprintf(BcsUserManagerPermissionURI, b.bcsAPIAddress),
		method,
		data,
	)

	if err != nil {
		return err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("failed to act permission: %s", msg)
	}

	return nil
}

func (b *bcsUserManager) GetPermission(method string, data []byte) ([]permission.PermissionsResp, error) {
	resp, err := b.requester.Do(
		fmt.Sprintf(BcsUserManagerPermissionURI, b.bcsAPIAddress),
		method,
		data,
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("failed to get permission: %s", msg)
	}

	var result []permission.PermissionsResp
	err = codec.DecJson(data, &result)
	return result, err
}

func (b *bcsUserManager) AddVpcCidrs(data []byte) error {
	resp, err := b.requester.Do(
		fmt.Sprintf(BcsUserManagerAddCidrUri, b.bcsAPIAddress),
		http.MethodPost,
		data,
	)

	if err != nil {
		return err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("failed to add cidr: %s", msg)
	}

	var result []permission.PermissionsResp
	err = codec.DecJson(data, &result)
	return err
}
