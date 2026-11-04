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

package istio

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	storeutils "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// installedMeshStatuses 控制面仍在使用的网格状态。
var installedMeshStatuses = []string{
	common.IstioStatusRunning,
	common.IstioStatusInstalled,
	common.IstioStatusUpdating,
	common.IstioStatusUpdateFailed,
	common.IstioStatusUninstallingFailed,
}

// ListInstalledMeshesAction 查询当前仍在使用的网格。
type ListInstalledMeshesAction struct {
	model store.MeshManagerModel
	req   *meshmanager.ListInstalledMeshesRequest
	resp  *meshmanager.ListInstalledMeshesResponse
}

// NewListInstalledMeshesAction create list installed meshes action.
func NewListInstalledMeshesAction(model store.MeshManagerModel) *ListInstalledMeshesAction {
	return &ListInstalledMeshesAction{model: model}
}

// Handle 返回未删除且仍在使用的网格，从集群只保留已安装成功的记录。
func (l *ListInstalledMeshesAction) Handle(
	ctx context.Context,
	req *meshmanager.ListInstalledMeshesRequest,
	resp *meshmanager.ListInstalledMeshesResponse,
) error {
	l.req = req
	l.resp = resp

	result, err := l.list(ctx)
	if err != nil {
		blog.Errorf("list installed meshes failed, %s", err.Error())
		l.setResp(common.DBErrorCode, err.Error(), nil)
		return nil
	}
	l.setResp(common.SuccessCode, "list installed meshes success", result)
	return nil
}

func (l *ListInstalledMeshesAction) setResp(code uint32, message string, data *meshmanager.ListInstalledMeshesData) {
	l.resp.Code = code
	l.resp.Message = message
	l.resp.Data = data
}

func (l *ListInstalledMeshesAction) list(ctx context.Context) (*meshmanager.ListInstalledMeshesData, error) {
	cond := operator.NewLeafCondition(operator.In, operator.M{
		entity.FieldKeyStatus: installedMeshStatuses,
	})
	total, meshes, err := l.model.List(ctx, cond, &storeutils.ListOption{
		Sort: map[string]int{
			entity.FieldKeyCreateTime: -1,
		},
	})
	if err != nil {
		return nil, err
	}

	items := make([]*meshmanager.InstalledMeshItem, 0, len(meshes))
	for _, mesh := range meshes {
		if mesh == nil {
			return nil, fmt.Errorf("data error, mesh is nil")
		}
		items = append(items, toInstalledMeshItem(mesh))
	}
	return &meshmanager.ListInstalledMeshesData{
		Total: int32(total),
		Items: items,
	}, nil
}

func toInstalledMeshItem(mesh *entity.MeshIstio) *meshmanager.InstalledMeshItem {
	remoteClusters := make([]*meshmanager.InstalledRemoteCluster, 0, len(mesh.RemoteClusters))
	for _, cluster := range mesh.RemoteClusters {
		if cluster == nil || cluster.Status != common.RemoteClusterStatusRunning {
			continue
		}
		remoteClusters = append(remoteClusters, &meshmanager.InstalledRemoteCluster{
			ClusterID: cluster.ClusterID,
			JoinTime:  cluster.JoinTime,
			Status:    cluster.Status,
		})
	}
	return &meshmanager.InstalledMeshItem{
		MeshID:          mesh.MeshID,
		Name:            mesh.Name,
		ProjectCode:     mesh.ProjectCode,
		Version:         mesh.Version,
		ChartVersion:    mesh.ChartVersion,
		Status:          mesh.Status,
		StatusMessage:   mesh.StatusMessage,
		CreateTime:      mesh.CreateTime,
		PrimaryClusters: append([]string(nil), mesh.PrimaryClusters...),
		RemoteClusters:  remoteClusters,
	}
}
