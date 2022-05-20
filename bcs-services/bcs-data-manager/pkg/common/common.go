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
	GetClusterIDList(ctx context.Context, client cm.ClusterManagerClient) ([]*ClusterMeta, error)
	// GetNamespaceList get namespace list
	GetNamespaceList(ctx context.Context, cmCli cm.ClusterManagerClient,
		k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*NamespaceMeta, error)
	// GetNamespaceListByCluster get namespace list by cluster
	GetNamespaceListByCluster(clusterMeta *ClusterMeta,
		k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*NamespaceMeta, error)
	// GetK8sWorkloadList get k8s workload list by namespace
	GetK8sWorkloadList(namespace []*NamespaceMeta,
		k8sStorageCli bcsapi.Storage) ([]*WorkloadMeta, error)
	// GetMesosWorkloadList get mesos workload list by cluster
	GetMesosWorkloadList(cluster *ClusterMeta, mesosStorageCli bcsapi.Storage) ([]*WorkloadMeta, error)
}

// ResourceGetter common resource getter
type ResourceGetter struct {
	needFilter bool
	clusterIDs map[string]bool
	cache      *cache.Cache
}

// NewGetter new common resource getter
func NewGetter(needFilter bool, clusterIds []string) GetterInterface {
	clusterMap := make(map[string]bool, len(clusterIds))
	for index := range clusterIds {
		clusterMap[clusterIds[index]] = true
	}
	return &ResourceGetter{
		needFilter: needFilter,
		clusterIDs: clusterMap,
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
func (g *ResourceGetter) GetClusterIDList(ctx context.Context, cmCli cm.ClusterManagerClient) ([]*ClusterMeta, error) {
	// get cluster list from cache first.
	// If found, return. Otherwise, call cluster manager api to get and set in cache.
	cacheClusterMetaList, found := g.cache.Get("clusterList")
	if found {
		return cacheClusterMetaList.([]*ClusterMeta), nil
	}
	blog.Infof("get cluster list from cache failed.")
	start := time.Now()
	clusterList, err := cmCli.ListCluster(ctx, &cm.ListClusterReq{})
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsClusterManager, "ListCluster", "GET", err, start)
		return nil, fmt.Errorf("get cluster list err: %v", err)
	}
	prom.ReportLibRequestMetric(prom.BkBcsClusterManager, "ListCluster", "GET", err, start)
	clusterMetaList := make([]*ClusterMeta, 0)
	for _, cluster := range clusterList.Data {
		if (!g.needFilter || g.clusterIDs[cluster.ClusterID]) && cluster.Status != "DELETED" {
			clusterMeta := &ClusterMeta{
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
	k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*NamespaceMeta, error) {
	// get namespace list from cache first
	// if found, return. Otherwise, call bcs storage api to get and set in cache.
	cacheList, found := g.cache.Get("namespaceList")
	if found {
		return cacheList.([]*NamespaceMeta), nil
	}
	blog.Infof("get namespace list from cache failed.")
	namespaceMetaList := make([]*NamespaceMeta, 0)
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
			case Kubernetes:
				go func(cluster *ClusterMeta) {
					defer wg.Done()
					namespaces := GetK8sNamespaceList(cluster.ClusterID, cluster.ProjectID, k8sStorageCli)
					lock.Lock()
					namespaceMetaList = append(namespaceMetaList, namespaces...)
					lock.Unlock()
					<-chPool
				}(cluster)
			case Mesos:
				go func(cluster *ClusterMeta) {
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
func (g *ResourceGetter) GetNamespaceListByCluster(clusterMeta *ClusterMeta,
	k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*NamespaceMeta, error) {
	// get from cache first
	cacheList, found := g.cache.Get(fmt.Sprintf("%s-ns", clusterMeta.ClusterID))
	if found {
		return cacheList.([]*NamespaceMeta), nil
	}
	blog.Infof("get namespace list by cluster id from cache failed.")
	switch clusterMeta.ClusterType {
	case Kubernetes:
		namespaceList := GetK8sNamespaceList(clusterMeta.ClusterID, clusterMeta.ProjectID, k8sStorageCli)
		g.cache.Set(fmt.Sprintf("%s-ns", clusterMeta.ClusterID), namespaceList, 15*time.Minute)
		return namespaceList, nil
	case Mesos:
		namespaceList := GetMesosNamespaceList(clusterMeta.ClusterID, clusterMeta.ProjectID, mesosStorageCli)
		g.cache.Set(fmt.Sprintf("%s-ns", clusterMeta.ClusterID), namespaceList, 15*time.Minute)
		return namespaceList, nil
	default:
		return nil, fmt.Errorf("wrong cluster engine type : %s", clusterMeta.ClusterType)
	}
}

// GetK8sWorkloadList get workload list by namespace
func (g *ResourceGetter) GetK8sWorkloadList(namespace []*NamespaceMeta,
	k8sStorageCli bcsapi.Storage) ([]*WorkloadMeta, error) {
	workloadList := make([]*WorkloadMeta, 0)
	for _, namespaceMeta := range namespace {
		workloads := GetK8sWorkloadList(namespaceMeta.ClusterID, namespaceMeta.ProjectID, namespaceMeta.Name,
			k8sStorageCli)
		workloadList = append(workloadList, workloads...)
	}
	return workloadList, nil
}

// GetMesosWorkloadList get workload list by cluster
func (g *ResourceGetter) GetMesosWorkloadList(cluster *ClusterMeta,
	mesosStorageCli bcsapi.Storage) ([]*WorkloadMeta, error) {
	workloadList := make([]*WorkloadMeta, 0)
	applications, err := mesosStorageCli.QueryMesosApplication(cluster.ClusterID)
	if err != nil {
		return workloadList, fmt.Errorf("get cluster %s application list error: %v", cluster.ClusterID, err)
	}
	for _, application := range applications {
		workloadMeta := generateCommonWorkloadList(cluster.ClusterID, cluster.ProjectID, Mesos, MesosApplicationType,
			application.CommonDataHeader)
		workloadList = append(workloadList, workloadMeta)
	}
	deployments, err := mesosStorageCli.QueryMesosDeployment(cluster.ClusterID)
	if err != nil {
		return workloadList, fmt.Errorf("get cluster %s deployment list error: %v", cluster.ClusterID, err)
	}
	for _, deployment := range deployments {
		workloadMeta := generateCommonWorkloadList(cluster.ClusterID, cluster.ProjectID, Mesos, MesosDeployment,
			deployment.CommonDataHeader)
		workloadList = append(workloadList, workloadMeta)
	}
	return workloadList, nil
}

// GetK8sWorkloadList get k8s workload list
func GetK8sWorkloadList(clusterID, projectID, namespace string, storageCli bcsapi.Storage) []*WorkloadMeta {
	workloadList := make([]*WorkloadMeta, 0)
	start := time.Now()
	deployments, err := storageCli.QueryK8SDeployment(clusterID, namespace)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetDeployment", "GET", err, start)
		blog.Errorf("get cluster %s deployment list error: %v", clusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetDeployment", "GET", err, start)
		for _, deployment := range deployments {
			workloadMeta := generateCommonWorkloadList(clusterID, projectID, Kubernetes, DeploymentType,
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
			workloadMeta := generateCommonWorkloadList(clusterID, projectID, Kubernetes, StatefulSetType,
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
			workloadMeta := generateCommonWorkloadList(clusterID, projectID, Kubernetes, DaemonSetType,
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
			workloadMeta := generateCommonWorkloadList(clusterID, projectID, Kubernetes, GameDeploymentType,
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
			workloadMeta := generateCommonWorkloadList(clusterID, projectID, Kubernetes, GameStatefulSetType,
				gameStatefulSet.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	return workloadList
}

// GetMesosWorkloadList get mesos workload list
func GetMesosWorkloadList(clusterID, projectID string, storageCli bcsapi.Storage) []*WorkloadMeta {
	workloadList := make([]*WorkloadMeta, 0)
	start := time.Now()
	applications, err := storageCli.QueryMesosApplication(clusterID)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosApp", "GET", err, start)
		blog.Errorf("get cluster %s application list error: %v", clusterID, err)
		return workloadList
	}
	prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosApp", "GET", err, start)
	for _, application := range applications {
		workloadMeta := generateCommonWorkloadList(clusterID, projectID, Mesos, MesosApplicationType,
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
		workloadMeta := generateCommonWorkloadList(clusterID, projectID, Mesos, MesosDeployment,
			deployment.CommonDataHeader)
		workloadList = append(workloadList, workloadMeta)
	}
	return workloadList
}

// GetK8sNamespaceList get k8s namespace list
func GetK8sNamespaceList(clusterID, projectID string, storageCli bcsapi.Storage) []*NamespaceMeta {
	start := time.Now()
	namespaces, err := storageCli.QueryK8SNamespace(clusterID)
	namespaceList := make([]*NamespaceMeta, 0)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetK8sNamespace", "GET", err, start)
		blog.Errorf("get cluster %s namespace list error :%v", clusterID, err)
		return namespaceList
	}
	prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetK8sNamespace", "GET", err, start)
	for _, namespace := range namespaces {
		namespaceMeta := &NamespaceMeta{
			ProjectID:   projectID,
			ClusterID:   clusterID,
			ClusterType: Kubernetes,
			Name:        namespace.ResourceName,
		}
		namespaceList = append(namespaceList, namespaceMeta)
	}
	return namespaceList
}

// GetMesosNamespaceList get mesos namespace list
func GetMesosNamespaceList(clusterID, projectID string, storageCli bcsapi.Storage) []*NamespaceMeta {
	namespaceList := make([]*NamespaceMeta, 0)
	start := time.Now()
	namespaces, err := storageCli.QueryMesosNamespace(clusterID)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosNamespace", "GET", err, start)
		blog.Errorf("get cluster %s namespace list error :%v", clusterID, err)
		return namespaceList
	}
	prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosNamespace", "GET", err, start)
	for _, namespace := range namespaces {
		namespaceMeta := &NamespaceMeta{
			ProjectID:   projectID,
			ClusterID:   clusterID,
			ClusterType: Mesos,
			Name:        string(*namespace),
		}
		namespaceList = append(namespaceList, namespaceMeta)
	}
	return namespaceList
}

func generateCommonWorkloadList(clusterID, projectID, clusterType, workloadType string,
	commonHeader storage.CommonDataHeader) *WorkloadMeta {
	workloadMeta := &WorkloadMeta{
		ProjectID:    projectID,
		ClusterID:    clusterID,
		ClusterType:  clusterType,
		Namespace:    commonHeader.Namespace,
		ResourceType: workloadType,
		Name:         commonHeader.ResourceName,
	}
	return workloadMeta
}

func formatTimeIgnoreSec(originalTime time.Time) time.Time {
	local := time.Local
	formatString, err := time.ParseInLocation(MinuteTimeFormat, originalTime.Format(MinuteTimeFormat), local)
	if err != nil {
		blog.Errorf("format time ignore second error :%v", err)
		return originalTime
	}
	return formatString
}

func formatTimeIgnoreMin(originalTime time.Time) time.Time {
	local := time.Local
	formatString, err := time.ParseInLocation(HourTimeFormat, originalTime.Format(HourTimeFormat), local)
	if err != nil {
		blog.Errorf("format time ignore minute error :%v", err)
		return originalTime
	}
	return formatString
}

func formatTimeIgnoreHour(originalTime time.Time) time.Time {
	local := time.Local
	formatString, err := time.ParseInLocation(DayTimeFormat, originalTime.Format(DayTimeFormat), local)
	if err != nil {
		blog.Errorf("format time ignore day error :%v", err)
		return originalTime
	}
	return formatString
}

// FormatTime format time
func FormatTime(originalTime time.Time, dimension string) time.Time {
	switch dimension {
	case DimensionDay:
		return formatTimeIgnoreHour(originalTime)
	case DimensionHour:
		return formatTimeIgnoreMin(originalTime)
	case DimensionMinute:
		return formatTimeIgnoreSec(originalTime)
	default:
		return originalTime
	}
}

// GetBucketTime get bucket time
func GetBucketTime(currentTime time.Time, dimension string) (string, error) {
	switch dimension {
	case DimensionDay:
		return currentTime.Format(MonthTimeFormat), nil
	case DimensionHour:
		return currentTime.Format(DayTimeFormat), nil
	case DimensionMinute:
		return currentTime.Format(HourTimeFormat), nil
	default:
		return "", fmt.Errorf("wrong dimension :%s", dimension)
	}
}

// GetIndex get metric index
func GetIndex(currentTime time.Time, dimension string) int {
	switch dimension {
	case DimensionDay:
		return currentTime.Day()
	case DimensionHour:
		return currentTime.Hour()
	case DimensionMinute:
		return currentTime.Minute()
	default:
		return 0
	}
}
