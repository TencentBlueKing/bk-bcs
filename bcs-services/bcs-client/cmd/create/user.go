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

package create

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/usermanager/v1"
)

func createUser(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionUserName, utils.OptionUserType); err != nil {
		return err
	}

	userManager := v1.NewBcsUserManager(utils.GetClientOption())
	user, err := userManager.CreateOrGetUser(c.String(utils.OptionUserType), c.String(utils.OptionUserName),
		http.MethodPost)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return printResult(user)
}

func printResult(single interface{}) error {
	fmt.Printf("%s\n", utils.TryIndent(single))
	return nil
}
