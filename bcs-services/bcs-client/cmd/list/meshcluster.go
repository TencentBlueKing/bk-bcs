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

package list

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/meshmanager/v1"
	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
)

func listMeshCluster(c *utils.ClientContext) error {
	meshManager := v1.NewMeshManager(utils.GetClientOption())
	resp, err := meshManager.ListMeshCluster(&meshmanager.ListMeshClusterReq{})
	if err != nil {
		fmt.Println("error", err.Error())
		return err
	}
	if resp.ErrCode != meshmanager.ErrCode_ERROR_OK {
		return fmt.Errorf("failed to list meshclusters: %s", resp.ErrMsg)
	}
	if len(resp.MeshClusters) == 0 {
		return fmt.Errorf("Found no meshclusters")
	}
	fmt.Printf("%-15s %-10s %-10s %-25s\n",
		"CLUSTERID",
		"VERSION",
		"STATUS",
		"MESSAGE")

	for _, mCluster := range resp.MeshClusters {
		var status, message string
		if mCluster.Deletion {
			status = "DELETING"
			message = "istio is deleting now"
		} else {
			status, message = getMeshClusterStatus(mCluster)
		}
		fmt.Printf("%-15s %-10s %-10s %-25s\n",
			mCluster.Clusterid,
			mCluster.Version,
			status,
			message)
	}
	return nil
}

//return status、message
func getMeshClusterStatus(mCluster *meshmanager.MeshCluster) (string, string) {
	var deploy, running, failed int
	var message string
	for _, component := range mCluster.Components {
		switch component.Status {
		case string(meshv1.InstallStatusNONE), string(meshv1.InstallStatusDEPLOY), string(meshv1.InstallStatusSTARTING):
			deploy++
		case string(meshv1.InstallStatusRUNNING):
			running++
		case string(meshv1.InstallStatusFAILED):
			failed++
			message = component.Message
		}
	}
	//failed
	if failed > 0 {
		return string(meshv1.InstallStatusFAILED), message
	}
	//running
	if len(mCluster.Components) == running {
		return string(meshv1.InstallStatusRUNNING), "istio is running now"
	}
	return string(meshv1.InstallStatusDEPLOY), "istio is deploying now"
}
