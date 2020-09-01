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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	v1 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/usermanager/v1"
)

func revokePermission(c *utils.ClientContext) error {
	var data []byte
	var err error
	if !c.IsSet(utils.OptionFile) {
		//reading all data from stdin
		data, err = ioutil.ReadAll(os.Stdin)
	} else {
		data, err = c.FileData()
	}
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("failed to revoke: no available resource datas")
	}

	userManager := v1.NewBcsUserManager(utils.GetClientOption())
	err = userManager.GrantOrRevokePermission(http.MethodDelete, data)
	if err != nil {
		return fmt.Errorf("failed to revoke permission: %v", err)
	}

	fmt.Printf("success to revoke permission\n")
	return nil
}
