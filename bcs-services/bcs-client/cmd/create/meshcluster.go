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
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/meshmanager/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
)

func createMeshCluster(c *utils.ClientContext) error {
	meshManager := v1.NewMeshManager(utils.GetClientOption())
	//fetch file data
	data, err := c.FileData()
	if err != nil {
		return err
	}
	var req *meshmanager.CreateMeshClusterReq
	err = json.Unmarshal(data, &req)
	if err != nil {
		return err
	}
	resp, err := meshManager.CreateMeshCluster(req)
	if err != nil {
		return fmt.Errorf("failed to create meshcluster: %v", err)
	}
	if resp.ErrCode != meshmanager.ErrCode_ERROR_OK {
		return fmt.Errorf("failed to create meshcluster: %s", resp.ErrMsg)
	}
	fmt.Printf("success to create meshcluster\n")
	return nil
}
