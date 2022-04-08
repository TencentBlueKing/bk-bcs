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

package permission

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	userV1 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/usermanager/v1"
	v1http "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/permission"
)

func getPermission(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionUserName, utils.OptionResourceType); err != nil {
		return err
	}

	userManager := userV1.NewBcsUserManager(utils.GetClientOption())
	pf := v1http.GetPermissionForm{
		UserName:     c.String(utils.OptionUserName),
		ResourceType: c.String(utils.OptionResourceType),
	}
	data, err := json.Marshal(pf)
	if err != nil {
		return err
	}
	permissions, err := userManager.GetPermission(http.MethodGet, data)
	if err != nil {
		return fmt.Errorf("failed to grant permission: %v", err)
	}

	return printGet(permissions)
}

func printGet(single interface{}) error {
	fmt.Printf("%s\n", utils.TryIndent(single))
	return nil
}
