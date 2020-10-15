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

package delete

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	v1 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/logmanager/v1"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/logmanager"
)

func deleteLogCollectionTask(c *utils.ClientContext) error {
	err := c.MustSpecified(utils.OptionClusterID, utils.OptionNamespace, utils.OptionName)
	if err != nil {
		return err
	}
	manager, err := v1.NewLogManager(context.Background(), utils.GetClientOption())
	if err != nil {
		return err
	}
	var req proto.DeleteLogCollectionTaskReq
	req.ClusterIDs = c.ClusterID()
	req.ConfigNamespace = c.Namespace()
	req.ConfigName = c.String(utils.OptionName)
	err = manager.DeleteLogCollectionTask(&req)
	if err != nil {
		return err
	}
	fmt.Printf("success to delete log collection task\n")
	return nil
}
