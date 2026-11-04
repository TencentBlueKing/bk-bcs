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
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/stretchr/testify/require"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	storeutils "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

func TestToInstalledMeshItemKeepsRunningRemoteClusters(t *testing.T) {
	item := toInstalledMeshItem(&entity.MeshIstio{
		MeshID:          "mesh-1",
		Name:            "demo",
		ProjectCode:     "proj",
		Version:         "1.18.0",
		ChartVersion:    "1.18.0-bcs.1",
		Status:          common.IstioStatusRunning,
		StatusMessage:   "",
		CreateTime:      100,
		PrimaryClusters: []string{"BCS-K8S-1"},
		RemoteClusters: []*entity.RemoteCluster{
			{ClusterID: "BCS-K8S-2", JoinTime: 200, Status: common.RemoteClusterStatusRunning},
			{ClusterID: "BCS-K8S-3", JoinTime: 300, Status: common.RemoteClusterStatusInstalling},
			{ClusterID: "BCS-K8S-4", JoinTime: 400, Status: common.RemoteClusterStatusInstallFailed},
			nil,
		},
	})

	require.Equal(t, "mesh-1", item.MeshID)
	require.Equal(t, []string{"BCS-K8S-1"}, item.PrimaryClusters)
	require.Len(t, item.RemoteClusters, 1)
	require.Equal(t, "BCS-K8S-2", item.RemoteClusters[0].ClusterID)
	require.Equal(t, int64(200), item.RemoteClusters[0].JoinTime)
	require.Equal(t, common.RemoteClusterStatusRunning, item.RemoteClusters[0].Status)
}

func TestListInstalledMeshesQueriesActiveStatuses(t *testing.T) {
	fake := &fakeMeshModel{
		meshes: []*entity.MeshIstio{
			{
				MeshID:          "mesh-1",
				ProjectCode:     "proj",
				Status:          common.IstioStatusRunning,
				PrimaryClusters: []string{"BCS-K8S-1"},
			},
		},
	}
	action := NewListInstalledMeshesAction(fake)
	resp := &meshmanager.ListInstalledMeshesResponse{}

	err := action.Handle(context.Background(), &meshmanager.ListInstalledMeshesRequest{}, resp)
	require.NoError(t, err)
	require.Equal(t, common.SuccessCode, resp.Code)
	require.Equal(t, int32(1), resp.Data.Total)
	require.Equal(t, "mesh-1", resp.Data.Items[0].MeshID)

	require.Equal(t, operator.In, fake.cond.Op)
	got, ok := fake.cond.Value.(operator.M)[entity.FieldKeyStatus].([]string)
	require.True(t, ok)
	require.Equal(t, installedMeshStatuses, got)
	require.Zero(t, fake.opt.Page)
	require.Zero(t, fake.opt.Size)
}

type fakeMeshModel struct {
	meshes []*entity.MeshIstio
	cond   *operator.Condition
	opt    *storeutils.ListOption
}

func (f *fakeMeshModel) Create(context.Context, *entity.MeshIstio) error { return nil }

func (f *fakeMeshModel) Update(context.Context, string, entity.M) error { return nil }

func (f *fakeMeshModel) SoftDelete(context.Context, string) error { return nil }

func (f *fakeMeshModel) Delete(context.Context, string) error { return nil }

func (f *fakeMeshModel) List(
	_ context.Context, cond *operator.Condition, opt *storeutils.ListOption,
) (int64, []*entity.MeshIstio, error) {
	f.cond = cond
	f.opt = opt
	return int64(len(f.meshes)), f.meshes, nil
}

func (f *fakeMeshModel) Get(context.Context, *operator.Condition) (*entity.MeshIstio, error) {
	return nil, nil
}
