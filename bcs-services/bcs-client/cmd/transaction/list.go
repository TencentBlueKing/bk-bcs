/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package transaction

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
)

func listTransaction(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}

	objKind := c.String("objkind")
	objName := c.String("objname")
	ns := c.String("namespace")

	scheduler := v4.NewBcsScheduler(utils.GetClientOption())

	transList, err := scheduler.ListTransaction(c.ClusterID(), ns, objKind, objName)
	if err != nil {
		return fmt.Errorf("failed to list transaction: %v", err)
	}

	return printListTransaction(transList)
}

func printListTransaction(transList []*types.Transaction) error {
	base := "%-30s %-30s %-15s %-50s %-15s %-15s"
	columns := []interface{}{"NAMESPACE", "TRANS_ID", "OBJ_KIND", "OBJ_NAME", "TRANS_TYPE", "TRANS_STATUS"}

	fmt.Printf(base+"\n", columns...)
	for _, trans := range transList {
		fmt.Printf(base+"\n", trans.Namespace, trans.TransactionID, trans.ObjectKind,
			trans.ObjectName, trans.CurOp.OpType, trans.Status)
	}
	return nil
}
