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

package inspect

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/meshmanager/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
)

func inspectMeshCluster(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}
	meshManager := v1.NewMeshManager(utils.GetClientOption())
	resp, err := meshManager.ListMeshCluster(&meshmanager.ListMeshClusterReq{Clusterid: c.ClusterID()})
	if err != nil {
		return err
	}
	if resp.ErrCode != meshmanager.ErrCode_ERROR_OK {
		return fmt.Errorf("failed to inspect cluster(%s) meshcluster: %s", c.ClusterID(), resp.ErrMsg)
	}
	if len(resp.MeshClusters) == 0 {
		return fmt.Errorf("Not found cluster(%s) meshcluster", c.ClusterID())
	}
	return printInspect(resp.MeshClusters[0])
}
