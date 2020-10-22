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

package meshmanager

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
)

type MeshManager interface {
	//create meshcluster crd and install istio service
	CreateMeshCluster(req *meshmanager.CreateMeshClusterReq) (*meshmanager.CreateMeshClusterResp, error)
	//delete meshcluster crd and uninstall istio service
	DeleteMeshCluster(req *meshmanager.DeleteMeshClusterReq) (*meshmanager.DeleteMeshClusterResp, error)
	//list meshcluster crds, contains istio components service status
	ListMeshCluster(req *meshmanager.ListMeshClusterReq) (*meshmanager.ListMeshClusterResp, error)
}
