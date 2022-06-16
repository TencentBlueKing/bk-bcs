/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/mock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetClusterIDList(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	getter := NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"})
	clusterList, err := getter.GetClusterIDList(ctx, cmCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(clusterList))
}

func TestGetNamespaceList(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	storageCli := mock.NewMockStorage()
	getter := NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"})
	namespaceList, err := getter.GetNamespaceList(ctx, cmCli, storageCli, storageCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 4, len(namespaceList))
}

func TestGetNamespaceListByCluster(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	storageCli := mock.NewMockStorage()
	getter := NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"})
	clusterList, err := getter.GetClusterIDList(ctx, cmCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(clusterList))
	for _, cluster := range clusterList {
		namespaceList, err := getter.GetNamespaceListByCluster(cluster, storageCli, storageCli)
		assert.Equal(t, nil, err)
		assert.NotEqual(t, 0, len(namespaceList))
	}
}

func TestGetProjectIDList(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	getter := NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"})
	projectList, err := getter.GetProjectIDList(ctx, cmCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(projectList))
}

func TestGetWorkloadList(t *testing.T) {
	storageCli := mock.NewMockStorage()
	getter := NewGetter(true, []string{"BCS-K8S-15091"})
	k8sNamespace := []*types.NamespaceMeta{{
		ProjectID:   "",
		ClusterID:   "BCS-K8S-15091",
		ClusterType: types.Kubernetes,
		Name:        "bcs-system",
	}}
	workloadList, err := getter.GetK8sWorkloadList(k8sNamespace, storageCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(workloadList))
	mesosCluster := &types.ClusterMeta{
		ProjectID:   "",
		ClusterID:   "BCS-MESOS-10039",
		ClusterType: types.Mesos,
	}
	mesosWorkloadList, err := getter.GetMesosWorkloadList(mesosCluster, storageCli)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(mesosWorkloadList))
}
