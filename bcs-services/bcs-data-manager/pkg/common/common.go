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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/prom"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	"github.com/patrickmn/go-cache"
)

// GetterInterface interface of getter
type GetterInterface interface {
	// GetProjectIDList get project id list
	GetProjectIDList(ctx context.Context, client cm.ClusterManagerClient) ([]string, error)
	// GetClusterIDList get cluster id list
	GetClusterIDList(ctx context.Context, client cm.ClusterManagerClient) ([]*types.ClusterMeta, error)
	// GetNamespaceList get namespace list
	GetNamespaceList(ctx context.Context, cmCli cm.ClusterManagerClient,
		k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*types.NamespaceMeta, error)
	// GetNamespaceListByCluster get namespace list by cluster
	GetNamespaceListByCluster(clusterMeta *types.ClusterMeta,
		k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*types.NamespaceMeta, error)
	// GetK8sWorkloadList get k8s workload list by namespace
	GetK8sWorkloadList(namespace []*types.NamespaceMeta,
		k8sStorageCli bcsapi.Storage) ([]*types.WorkloadMeta, error)
	// GetMesosWorkloadList get mesos workload list by cluster
	GetMesosWorkloadList(cluster *types.ClusterMeta, mesosStorageCli bcsapi.Storage) ([]*types.WorkloadMeta, error)
}

// ResourceGetter common resource getter
type ResourceGetter struct {
	needFilter bool
	clusterIDs map[string]bool
	env        string
	cache      *cache.Cache
}

// NewGetter new common resource getter
func NewGetter(needFilter bool, clusterIds []string, env string) GetterInterface {
	clusterMap := make(map[string]bool, len(clusterIds))
	for index := range clusterIds {
		clusterMap[clusterIds[index]] = true
	}
	return &ResourceGetter{
		needFilter: needFilter,
		clusterIDs: clusterMap,
		env:        env,
		cache:      cache.New(time.Minute*10, time.Minute*60),
	}
}

// GetProjectIDList get project id list
func (g *ResourceGetter) GetProjectIDList(ctx context.Context, cmCli cm.ClusterManagerClient) ([]string, error) {
	projectList := make([]string, 0)
	projectMap := make(map[string]bool)
	clusterList, err := g.GetClusterIDList(ctx, cmCli)
	if err != nil {
		return nil, fmt.Errorf("get cluster list err: %v", err)
	}

	for _, cluster := range clusterList {
		if !g.needFilter || g.clusterIDs[cluster.ClusterID] {
			projectMap[cluster.ProjectID] = true
		}
	}
	for projectID := range projectMap {
		projectList = append(projectList, projectID)
	}
	return projectList, nil
}

// GetClusterIDList get cluster id list
func (g *ResourceGetter) GetClusterIDList(ctx context.Context,
	cmCli cm.ClusterManagerClient) ([]*types.ClusterMeta, error) {
	// get cluster list from cache first.
	// If found, return. Otherwise, call cluster manager api to get and set in cache.
	cacheClusterMetaList, found := g.cache.Get("clusterList")
	if found {
		return cacheClusterMetaList.([]*types.ClusterMeta), nil
	}
	blog.Infof("get cluster list from cache failed.")
	start := time.Now()
	clusterList, err := cmCli.ListCluster(ctx, &cm.ListClusterReq{Environment: g.env})
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsClusterManager, "ListCluster", "GET", err, start)
		return nil, fmt.Errorf("get cluster list err: %v", err)
	}
	prom.ReportLibRequestMetric(prom.BkBcsClusterManager, "ListCluster", "GET", err, start)
	clusterMetaList := make([]*types.ClusterMeta, 0)
	for _, cluster := range clusterList.Data {
		if (!g.needFilter || g.clusterIDs[cluster.ClusterID]) && cluster.Status != "DELETED" {
			clusterMeta := &types.ClusterMeta{
				ProjectID:   cluster.ProjectID,
				ClusterID:   cluster.ClusterID,
				ClusterType: cluster.EngineType,
			}
			clusterMetaList = append(clusterMetaList, clusterMeta)
		}
	}
	g.cache.Set("clusterList", clusterMetaList, 15*time.Minute)
	return clusterMetaList, nil
}

// GetNamespaceList get namespace list
func (g *ResourceGetter) GetNamespaceList(ctx context.Context, cmCli cm.ClusterManagerClient,
	k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*types.NamespaceMeta, error) {
	// get namespace list from cache first
	// if found, return. Otherwise, call bcs storage api to get and set in cache.
	cacheList, found := g.cache.Get("namespaceList")
	if found {
		return cacheList.([]*types.NamespaceMeta), nil
	}
	blog.Infof("get namespace list from cache failed.")
	namespaceMetaList := make([]*types.NamespaceMeta, 0)
	clusterList, err := g.GetClusterIDList(ctx, cmCli)
	if err != nil {
		return nil, fmt.Errorf("get cluster list err: %v", err)
	}
	chPool := make(chan struct{}, 30)
	wg := sync.WaitGroup{}
	lock := &sync.Mutex{}
	for _, cluster := range clusterList {
		if !g.needFilter || g.clusterIDs[cluster.ClusterID] {
			wg.Add(1)
			chPool <- struct{}{}
			switch cluster.ClusterType {
			case types.Kubernetes:
				go func(cluster *types.ClusterMeta) {
					defer wg.Done()
					namespaces := GetK8sNamespaceList(cluster.ClusterID, cluster.ProjectID, k8sStorageCli)
					lock.Lock()
					namespaceMetaList = append(namespaceMetaList, namespaces...)
					lock.Unlock()
					<-chPool
				}(cluster)
			case types.Mesos:
				go func(cluster *types.ClusterMeta) {
					defer wg.Done()
					namespaces := GetMesosNamespaceList(cluster.ClusterID, cluster.ProjectID, mesosStorageCli)
					lock.Lock()
					namespaceMetaList = append(namespaceMetaList, namespaces...)
					lock.Unlock()
					<-chPool
				}(cluster)
			default:
				wg.Done()
				<-chPool
				return nil, fmt.Errorf("wrong cluster engine type : %s", cluster.ClusterType)
			}
		}
	}
	wg.Wait()
	g.cache.Set("namespaceList", namespaceMetaList, 15*time.Minute)
	return namespaceMetaList, err
}

// GetNamespaceListByCluster get namespace list by cluster
func (g *ResourceGetter) GetNamespaceListByCluster(clusterMeta *types.ClusterMeta,
	k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*types.NamespaceMeta, error) {
	// get from cache first
	cacheList, found := g.cache.Get(fmt.Sprintf("%s-ns", clusterMeta.ClusterID))
	if found {
		return cacheList.([]*types.NamespaceMeta), nil
	}
	blog.Infof("get namespace list by cluster id from cache failed.")
	switch clusterMeta.ClusterType {
	case types.Kubernetes:
		namespaceList := GetK8sNamespaceList(clusterMeta.ClusterID, clusterMeta.ProjectID, k8sStorageCli)
		g.cache.Set(fmt.Sprintf("%s-ns", clusterMeta.ClusterID), namespaceList, 15*time.Minute)
		return namespaceList, nil
	case types.Mesos:
		namespaceList := GetMesosNamespaceList(clusterMeta.ClusterID, clusterMeta.ProjectID, mesosStorageCli)
		g.cache.Set(fmt.Sprintf("%s-ns", clusterMeta.ClusterID), namespaceList, 15*time.Minute)
		return namespaceList, nil
	default:
		return nil, fmt.Errorf("wrong cluster engine type : %s", clusterMeta.ClusterType)
	}
}

// GetK8sWorkloadList get workload list by namespace
func (g *ResourceGetter) GetK8sWorkloadList(namespace []*types.NamespaceMeta,
	k8sStorageCli bcsapi.Storage) ([]*types.WorkloadMeta, error) {
	workloadList := make([]*types.WorkloadMeta, 0)
	for _, namespaceMeta := range namespace {
		workloads := GetK8sWorkloadList(namespaceMeta.ClusterID, namespaceMeta.ProjectID, namespaceMeta.Name,
			k8sStorageCli)
		workloadList = append(workloadList, workloads...)
	}
	return workloadList, nil
}

// GetMesosWorkloadList get workload list by cluster
func (g *ResourceGetter) GetMesosWorkloadList(cluster *types.ClusterMeta,
	mesosStorageCli bcsapi.Storage) ([]*types.WorkloadMeta, error) {
	workloadList := make([]*types.WorkloadMeta, 0)
	applications, err := mesosStorageCli.QueryMesosApplication(cluster.ClusterID)
	if err != nil {
		return workloadList, fmt.Errorf("get cluster %s application list error: %v", cluster.ClusterID, err)
	}
	for _, application := range applications {
		workloadMeta := generateCommonWorkloadList(cluster.ClusterID, cluster.ProjectID,
			types.Mesos, types.MesosApplicationType, application.CommonDataHeader)
		workloadList = append(workloadList, workloadMeta)
	}
	deployments, err := mesosStorageCli.QueryMesosDeployment(cluster.ClusterID)
	if err != nil {
		return workloadList, fmt.Errorf("get cluster %s deployment list error: %v", cluster.ClusterID, err)
	}
	for _, deployment := range deployments {
		workloadMeta := generateCommonWorkloadList(cluster.ClusterID, cluster.ProjectID, types.Mesos, types.MesosDeployment,
			deployment.CommonDataHeader)
		workloadList = append(workloadList, workloadMeta)
	}
	return workloadList, nil
}

// GetK8sWorkloadList get k8s workload list
func GetK8sWorkloadList(clusterID, projectID, namespace string, storageCli bcsapi.Storage) []*types.WorkloadMeta {
	workloadList := make([]*types.WorkloadMeta, 0)
	start := time.Now()
	deployments, err := storageCli.QueryK8SDeployment(clusterID, namespace)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetDeployment", "GET", err, start)
		blog.Errorf("get cluster %s deployment list error: %v", clusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetDeployment", "GET", err, start)
		for _, deployment := range deployments {
			workloadMeta := generateCommonWorkloadList(clusterID, projectID, types.Kubernetes, types.DeploymentType,
				deployment.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	start = time.Now()
	statefulSets, err := storageCli.QueryK8SStatefulSet(clusterID, namespace)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetStatefulSet", "GET", err, start)
		blog.Errorf("get cluster %s statefulSet list error: %v", clusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetStatefulSet", "GET", err, start)
		for _, statefulSet := range statefulSets {
			workloadMeta := generateCommonWorkloadList(clusterID, projectID, types.Kubernetes, types.StatefulSetType,
				statefulSet.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	start = time.Now()
	daemonSets, err := storageCli.QueryK8SDaemonSet(clusterID, namespace)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetDaemonSet", "GET", err, start)
		blog.Errorf("get cluster %s daemonSet list error: %v", clusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetDaemonSet", "GET", err, start)
		for _, daemonSet := range daemonSets {
			workloadMeta := generateCommonWorkloadList(clusterID, projectID, types.Kubernetes, types.DaemonSetType,
				daemonSet.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	start = time.Now()
	gameDeployments, err := storageCli.QueryK8SGameDeployment(clusterID, namespace)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetGameDeployment", "GET", err, start)
		blog.Errorf("get cluster %s game deployment list error: %v", clusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetGameDeployment", "GET", err, start)
		for _, gameDeployment := range gameDeployments {
			workloadMeta := generateCommonWorkloadList(clusterID, projectID, types.Kubernetes, types.GameDeploymentType,
				gameDeployment.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	start = time.Now()
	gameStatefulSets, err := storageCli.QueryK8SGameStatefulSet(clusterID, namespace)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetGameStatefulSet", "GET", err, start)
		blog.Errorf("get cluster %s game stateful set list error: %v", clusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetGameStatefulSet", "GET", err, start)
		for _, gameStatefulSet := range gameStatefulSets {
			workloadMeta := generateCommonWorkloadList(clusterID, projectID, types.Kubernetes, types.GameStatefulSetType,
				gameStatefulSet.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	return workloadList
}

// GetMesosWorkloadList get mesos workload list
func GetMesosWorkloadList(clusterID, projectID string, storageCli bcsapi.Storage) []*types.WorkloadMeta {
	workloadList := make([]*types.WorkloadMeta, 0)
	start := time.Now()
	applications, err := storageCli.QueryMesosApplication(clusterID)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosApp", "GET", err, start)
		blog.Errorf("get cluster %s application list error: %v", clusterID, err)
		return workloadList
	}
	prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosApp", "GET", err, start)
	for _, application := range applications {
		workloadMeta := generateCommonWorkloadList(clusterID, projectID, types.Mesos, types.MesosApplicationType,
			application.CommonDataHeader)
		workloadList = append(workloadList, workloadMeta)
	}
	start = time.Now()
	deployments, err := storageCli.QueryMesosDeployment(clusterID)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosDeploy", "GET", err, start)
		blog.Errorf("get cluster %s deployment list error: %v", clusterID, err)
		return workloadList
	}
	prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosDeploy", "GET", err, start)
	for _, deployment := range deployments {
		workloadMeta := generateCommonWorkloadList(clusterID, projectID, types.Mesos, types.MesosDeployment,
			deployment.CommonDataHeader)
		workloadList = append(workloadList, workloadMeta)
	}
	return workloadList
}

// GetK8sNamespaceList get k8s namespace list
func GetK8sNamespaceList(clusterID, projectID string, storageCli bcsapi.Storage) []*types.NamespaceMeta {
	start := time.Now()
	namespaces, err := storageCli.QueryK8SNamespace(clusterID)
	namespaceList := make([]*types.NamespaceMeta, 0)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetK8sNamespace", "GET", err, start)
		blog.Errorf("get cluster %s namespace list error :%v", clusterID, err)
		return namespaceList
	}
	prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetK8sNamespace", "GET", err, start)
	for _, namespace := range namespaces {
		namespaceMeta := &types.NamespaceMeta{
			ProjectID:   projectID,
			ClusterID:   clusterID,
			ClusterType: types.Kubernetes,
			Name:        namespace.ResourceName,
		}
		namespaceList = append(namespaceList, namespaceMeta)
	}
	return namespaceList
}

// GetMesosNamespaceList get mesos namespace list
func GetMesosNamespaceList(clusterID, projectID string, storageCli bcsapi.Storage) []*types.NamespaceMeta {
	namespaceList := make([]*types.NamespaceMeta, 0)
	start := time.Now()
	namespaces, err := storageCli.QueryMesosNamespace(clusterID)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosNamespace", "GET", err, start)
		blog.Errorf("get cluster %s namespace list error :%v", clusterID, err)
		return namespaceList
	}
	prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosNamespace", "GET", err, start)
	for _, namespace := range namespaces {
		namespaceMeta := &types.NamespaceMeta{
			ProjectID:   projectID,
			ClusterID:   clusterID,
			ClusterType: types.Mesos,
			Name:        string(*namespace),
		}
		namespaceList = append(namespaceList, namespaceMeta)
	}
	return namespaceList
}

func generateCommonWorkloadList(clusterID, projectID, clusterType, workloadType string,
	commonHeader storage.CommonDataHeader) *types.WorkloadMeta {
	workloadMeta := &types.WorkloadMeta{
		ProjectID:    projectID,
		ClusterID:    clusterID,
		ClusterType:  clusterType,
		Namespace:    commonHeader.Namespace,
		ResourceType: workloadType,
		Name:         commonHeader.ResourceName,
	}
	return workloadMeta
}
