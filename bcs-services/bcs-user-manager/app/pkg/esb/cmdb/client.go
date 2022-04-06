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

package cmdb

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb/cmdbv3"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/options"
)

// CMDBClient is the CMDB client
var CMDBClient *cmdbv3.Client

// InitCMDBClient init CMDB client
func InitCMDBClient(op *options.UserManagerOptions) error {
	if !op.Cmdb.Enable {
		return nil
	}
	appSecret, err := encrypt.DesDecryptFromBase([]byte(op.Cmdb.AppSecret))
	if err != nil {
		return fmt.Errorf("error decrypting cmdb app secret, %s", err.Error())
	}
	cli := cmdbv3.NewClientInterface(op.Cmdb.Host, nil)
	cli.SetCommonReq(map[string]interface{}{
		"bk_app_code":   op.Cmdb.AppCode,
		"bk_app_secret": string(appSecret),
		"bk_username":   op.Cmdb.BkUserName,
	})
	CMDBClient = cli
	return nil
}
