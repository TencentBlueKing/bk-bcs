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

package common

/*
import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/mock"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

func TestGetClusterIDList(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	pmCli := mock.NewMockPmClient()
	getter := NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	clusterList, err := getter.GetClusterIDList(ctx, cmCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(clusterList))
}

func TestGetNamespaceList(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	storageCli := mock.NewMockStorage()
	pmCli := mock.NewMockPmClient()
	getter := NewGetter(true, []string{""}, "stag", pmCli)
	namespaceList, err := getter.GetNamespaceList(ctx, cmCli, storageCli, storageCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 4, len(namespaceList))
}

func TestGetNamespaceListByCluster(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	storageCli := mock.NewMockStorage()
	pmCli := mock.NewMockPmClient()
	getter := NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	clusterList, err := getter.GetClusterIDList(ctx, cmCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(clusterList))
	for _, cluster := range clusterList {
		namespaceList, err := getter.GetNamespaceListByCluster(ctx, cluster, storageCli, storageCli)
		assert.Equal(t, nil, err)
		assert.NotEqual(t, 0, len(namespaceList))
	}
}

func TestGetProjectIDList(t *testing.T) {
	ctx := context.Background()
	cmCli := mock.NewMockCm()
	pmCli := mock.NewMockPmClient()
	getter := NewGetter(true, []string{"BCS-MESOS-10039", "BCS-K8S-15091"}, "stag", pmCli)
	projectList, err := getter.GetProjectIDList(ctx, cmCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(projectList))
}

func TestGetWorkloadList(t *testing.T) {
	storageCli := mock.NewMockStorage()
	pmCli := mock.NewMockPmClient()
	getter := NewGetter(true, []string{"BCS-K8S-15091"}, "stag", pmCli)
	k8sNamespace := []*types.NamespaceMeta{{
		ProjectID:   "b37778ec757544868a01e1f01f07037f",
		ProjectCode: "k8stest",
		ClusterID:   "BCS-K8S-15091",
		ClusterType: types.Kubernetes,
		Name:        "bcs-system",
	}}
	workloadList, err := getter.GetK8sWorkloadList(k8sNamespace, storageCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(workloadList))
	mesosCluster := &types.ClusterMeta{
		ProjectID:   "ab2b254938e84f6b86b466cc22e730b1",
		ProjectCode: "mesostest",
		ClusterID:   "BCS-MESOS-10039",
		ClusterType: types.Mesos,
	}
	mesosWorkloadList, err := getter.GetMesosWorkloadList(mesosCluster, storageCli)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(mesosWorkloadList))
}

func TestGetPodAutoscalerList(t *testing.T) {
	storageCli := mock.NewMockStorage()
	pmCli := mock.NewMockPmClient()
	getter := NewGetter(true, []string{"BCS-K8S-15091"}, "stag", pmCli)
	k8sNamespace := []*types.NamespaceMeta{{
		ProjectID:   "",
		ClusterID:   "BCS-K8S-90000",
		ClusterType: types.Kubernetes,
		Name:        "bcs-system",
	}}
	hpaList, err := getter.GetPodAutoscalerList(types.HPAType, k8sNamespace, storageCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(hpaList))
	gpaList, err := getter.GetPodAutoscalerList(types.GPAType, k8sNamespace, storageCli)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(gpaList))
}
*/
