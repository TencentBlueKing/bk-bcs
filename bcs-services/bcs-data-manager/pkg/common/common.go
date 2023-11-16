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

// Package common xxx
package common

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	pm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	"github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/prom"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

// GetterInterface interface of getter
type GetterInterface interface {
	// GetProjectIDList get project id list
	GetProjectIDList(ctx context.Context, client cm.ClusterManagerClient) ([]*types.ProjectMeta, error)
	// GetProjectInfo get project info from bcs project or cache
	GetProjectInfo(ctx context.Context, projectId, projectCode string,
		pmCli *bcsproject.BcsProjectClientWithHeader) (*pm.Project, error)
	// GetClusterIDList get cluster id list
	GetClusterIDList(ctx context.Context, client cm.ClusterManagerClient) ([]*types.ClusterMeta, error)
	// GetNamespaceList get namespace list
	GetNamespaceList(ctx context.Context, cmCli cm.ClusterManagerClient,
		k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*types.NamespaceMeta, error)
	// GetNamespaceListByCluster get namespace list by cluster
	GetNamespaceListByCluster(ctx context.Context, clusterMeta *types.ClusterMeta,
		k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*types.NamespaceMeta, error)
	// GetK8sWorkloadList get k8s workload list by namespace
	GetK8sWorkloadList(namespace []*types.NamespaceMeta,
		k8sStorageCli bcsapi.Storage) ([]*types.WorkloadMeta, error)
	// GetMesosWorkloadList get mesos workload list by cluster
	GetMesosWorkloadList(cluster *types.ClusterMeta, mesosStorageCli bcsapi.Storage) ([]*types.WorkloadMeta, error)
	// GetPodAutoscalerList get podAutoscaler list by namespace
	GetPodAutoscalerList(podAutoscalerType string, namespace []*types.NamespaceMeta,
		k8sStorageCli bcsapi.Storage) ([]*types.PodAutoscalerMeta, error)
}

// ResourceGetter common resource getter
type ResourceGetter struct {
	needFilter     bool
	clusterIDs     map[string]bool
	env            string
	cache          *cache.Cache
	projectManager bcsproject.BcsProjectManagerClient
	bcsMonitorCli  bcsmonitor.ClientInterface
}

// NewGetter new common resource getter
func NewGetter(needFilter bool, clusterIds []string, env string,
	pmClient bcsproject.BcsProjectManagerClient, bcsMonitorCli bcsmonitor.ClientInterface) GetterInterface {
	clusterMap := make(map[string]bool, len(clusterIds))
	for index := range clusterIds {
		clusterMap[clusterIds[index]] = true
	}
	return &ResourceGetter{
		needFilter:     needFilter,
		clusterIDs:     clusterMap,
		env:            env,
		cache:          cache.New(time.Minute*10, time.Minute*60),
		projectManager: pmClient,
		bcsMonitorCli:  bcsMonitorCli,
	}
}

// GetProjectIDList get project id list
func (g *ResourceGetter) GetProjectIDList(ctx context.Context,
	cmCli cm.ClusterManagerClient) ([]*types.ProjectMeta, error) {
	projectList := make([]*types.ProjectMeta, 0)
	projectMap := make(map[string]*types.ProjectMeta)
	clusterList, err := g.GetClusterIDList(ctx, cmCli)
	if err != nil {
		return nil, fmt.Errorf("get cluster list err: %v", err)
	}
	pmConn, err := g.projectManager.GetBcsProjectManagerConn()
	if err != nil {
		blog.Errorf("get pm conn error:%v", err)
		return nil, err
	}
	defer pmConn.Close() // nolint
	pmCli := g.projectManager.NewGrpcClientWithHeader(ctx, pmConn)
	for _, cluster := range clusterList {
		// if needFilter, just handle particular cluster list
		if !g.needFilter || g.clusterIDs[cluster.ClusterID] {
			project, err := g.GetProjectInfo(ctx, cluster.ProjectID, "", pmCli)
			if err != nil {
				blog.Errorf("get project info err:%v", err)
				return nil, err
			}
			if project == nil {
				blog.Errorf("project info is nil, projectID:%s", cluster.ProjectID)
				continue
			}
			projectMap[cluster.ProjectID] = &types.ProjectMeta{
				ProjectID:   project.ProjectID,
				ProjectCode: project.ProjectCode,
				BusinessID:  project.BusinessID,
				Label: map[string]string{
					"BGName":     project.BGName,
					"BGID":       project.BGID,
					"deptName":   project.DeptName,
					"deptID":     project.DeptID,
					"centerName": project.CenterName,
					"centerID":   project.CenterID,
				},
			}
		}
	}
	for _, project := range projectMap {
		projectList = append(projectList, project)
	}
	return projectList, nil
}

// GetProjectInfo get project info by projectId or projectCode
func (g *ResourceGetter) GetProjectInfo(ctx context.Context, projectId, projectCode string,
	pmCli *bcsproject.BcsProjectClientWithHeader) (*pm.Project, error) {
	if pmCli == nil {
		pmConn, err := g.projectManager.GetBcsProjectManagerConn()
		if err != nil {
			blog.Errorf("get pm conn error:%v", err)
			return nil, err
		}
		defer pmConn.Close() // nolint
		pmCli = g.projectManager.NewGrpcClientWithHeader(ctx, pmConn)
	}
	if projectId == "" && projectCode == "" {
		return nil, fmt.Errorf("projectId and projectCode is empty")
	}

	var project *pm.Project
	if projectId != "" {
		if projectInfo, ok := g.cache.Get(projectId); !ok {
			projectResponse, err := pmCli.Cli.GetProject(pmCli.Ctx, &pm.GetProjectRequest{ProjectIDOrCode: projectId})
			if err != nil || projectResponse.Code != 0 {
				blog.Errorf("get project from bcs project err. err:%v, message:%s, projectId:%s, projectCode:%s",
					err, projectResponse.Message, projectId, projectCode)
				return nil, fmt.Errorf("get project from bcs project err. err:%s", projectResponse.Message)
			}
			project = projectResponse.Data
			g.cache.Set(projectId, project, 1*time.Hour)
		} else {
			project = projectInfo.(*pm.Project)
		}
		return project, nil
	}
	if projectInfo, ok := g.cache.Get(projectCode); !ok {
		projectResponse, err := pmCli.Cli.GetProject(pmCli.Ctx, &pm.GetProjectRequest{ProjectIDOrCode: projectCode})
		if err != nil || projectResponse.Code != 0 {
			blog.Errorf("get project from bcs project err. err:%v, message:%s, projectId:%s, projectCode:%s",
				err, projectResponse.Message, projectId, projectCode)
			return nil, err
		}
		project = projectResponse.Data
		g.cache.Set(projectId, project, 1*time.Hour)
	} else {
		project = projectInfo.(*pm.Project)
	}
	return project, nil
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
	// public cluster is duplicate in cluster list
	uniqueClusterList := removeDuplicateCluster(clusterList.Data)
	clusterMetaList := make([]*types.ClusterMeta, 0)
	pmConn, err := g.projectManager.GetBcsProjectManagerConn()
	if err != nil {
		blog.Errorf("get pm conn error:%v", err)
		return nil, err
	}
	defer pmConn.Close() // nolint
	pmCli := g.projectManager.NewGrpcClientWithHeader(ctx, pmConn)
	for _, cluster := range uniqueClusterList {
		if (!g.needFilter || g.clusterIDs[cluster.ClusterID]) && cluster.Status != "DELETED" {
			project, err := g.GetProjectInfo(ctx, cluster.ProjectID, "", pmCli)
			if err != nil {
				blog.Errorf("get project info err:%v", err)
				continue
			}
			projectCode := ""
			if project != nil {
				projectCode = project.ProjectCode
			}
			var isBKMonitor bool
			if result, err := g.bcsMonitorCli.CheckIfBKMonitor(cluster.ClusterID); err != nil {
				blog.Errorf("check cluster[%s] if bk monitor error:%s", cluster.ClusterID, err.Error())
			} else {
				isBKMonitor = result
			}
			clusterMeta := &types.ClusterMeta{
				ProjectID:   cluster.ProjectID,
				ProjectCode: projectCode,
				BusinessID:  cluster.BusinessID,
				ClusterID:   cluster.ClusterID,
				ClusterType: cluster.EngineType,
				Label:       map[string]string{"isShared": strconv.FormatBool(cluster.IsShared)},
				IsBKMonitor: isBKMonitor,
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
					namespaces := g.GetK8sNamespaceList(ctx, cluster, k8sStorageCli)
					lock.Lock()
					namespaceMetaList = append(namespaceMetaList, namespaces...)
					lock.Unlock()
					<-chPool
				}(cluster)
			case types.Mesos:
				go func(cluster *types.ClusterMeta) {
					defer wg.Done()
					namespaces := g.GetMesosNamespaceList(cluster, mesosStorageCli)
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
func (g *ResourceGetter) GetNamespaceListByCluster(ctx context.Context, clusterMeta *types.ClusterMeta,
	k8sStorageCli, mesosStorageCli bcsapi.Storage) ([]*types.NamespaceMeta, error) {
	// get from cache first
	cacheList, found := g.cache.Get(fmt.Sprintf("%s-ns", clusterMeta.ClusterID))
	if found {
		return cacheList.([]*types.NamespaceMeta), nil
	}
	blog.Infof("get namespace list by cluster id from cache failed.")
	switch clusterMeta.ClusterType {
	case types.Kubernetes:
		namespaceList := g.GetK8sNamespaceList(ctx, clusterMeta, k8sStorageCli)
		g.cache.Set(fmt.Sprintf("%s-ns", clusterMeta.ClusterID), namespaceList, 15*time.Minute)
		return namespaceList, nil
	case types.Mesos:
		namespaceList := g.GetMesosNamespaceList(clusterMeta, mesosStorageCli)
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
		workloads := GetK8sWorkloadList(namespaceMeta, k8sStorageCli)
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
		workloadMeta := generateMesosWorkloadList(cluster, types.MesosApplicationType, application.CommonDataHeader)
		workloadList = append(workloadList, workloadMeta)
	}
	deployments, err := mesosStorageCli.QueryMesosDeployment(cluster.ClusterID)
	if err != nil {
		return workloadList, fmt.Errorf("get cluster %s deployment list error: %v", cluster.ClusterID, err)
	}
	for _, deployment := range deployments {
		workloadMeta := generateMesosWorkloadList(cluster, types.MesosDeployment,
			deployment.CommonDataHeader)
		workloadList = append(workloadList, workloadMeta)
	}
	return workloadList, nil
}

// GetK8sWorkloadList get k8s workload list
// deployment, daemonset, statefulSet, gameDeployment, gameStatefulset
func GetK8sWorkloadList(namespaceMeta *types.NamespaceMeta, storageCli bcsapi.Storage) []*types.WorkloadMeta {
	workloadList := make([]*types.WorkloadMeta, 0)
	start := time.Now()
	deployments, err := storageCli.QueryK8SDeployment(namespaceMeta.ClusterID, namespaceMeta.Name)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetDeployment", "GET", err, start)
		blog.Errorf("get cluster %s deployment list error: %v", namespaceMeta.ClusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetDeployment", "GET", err, start)
		for _, deployment := range deployments {
			workloadMeta := generateK8sWorkloadList(namespaceMeta, types.DeploymentType,
				deployment.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	start = time.Now()
	statefulSets, err := storageCli.QueryK8SStatefulSet(namespaceMeta.ClusterID, namespaceMeta.Name)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetStatefulSet", "GET", err, start)
		blog.Errorf("get cluster %s statefulSet list error: %v", namespaceMeta.ClusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetStatefulSet", "GET", err, start)
		for _, statefulSet := range statefulSets {
			workloadMeta := generateK8sWorkloadList(namespaceMeta, types.StatefulSetType,
				statefulSet.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	start = time.Now()
	daemonSets, err := storageCli.QueryK8SDaemonSet(namespaceMeta.ClusterID, namespaceMeta.Name)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetDaemonSet", "GET", err, start)
		blog.Errorf("get cluster %s daemonSet list error: %v", namespaceMeta.ClusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetDaemonSet", "GET", err, start)
		for _, daemonSet := range daemonSets {
			workloadMeta := generateK8sWorkloadList(namespaceMeta, types.DaemonSetType, daemonSet.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	start = time.Now()
	gameDeployments, err := storageCli.QueryK8SGameDeployment(namespaceMeta.ClusterID, namespaceMeta.Name)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetGameDeployment", "GET", err, start)
		blog.Errorf("get cluster %s game deployment list error: %v", namespaceMeta.ClusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetGameDeployment", "GET", err, start)
		for _, gameDeployment := range gameDeployments {
			workloadMeta := generateK8sWorkloadList(namespaceMeta, types.GameDeploymentType,
				gameDeployment.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	start = time.Now()
	gameStatefulSets, err := storageCli.QueryK8SGameStatefulSet(namespaceMeta.ClusterID, namespaceMeta.Name)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetGameStatefulSet", "GET", err, start)
		blog.Errorf("get cluster %s game stateful set list error: %v", namespaceMeta.ClusterID, err)
	} else {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetGameStatefulSet", "GET", err, start)
		for _, gameStatefulSet := range gameStatefulSets {
			workloadMeta := generateK8sWorkloadList(namespaceMeta, types.GameStatefulSetType,
				gameStatefulSet.CommonDataHeader)
			workloadList = append(workloadList, workloadMeta)
		}
	}
	return workloadList
}

// GetK8sNamespaceList get k8s namespace list
func (g *ResourceGetter) GetK8sNamespaceList(ctx context.Context, clusterMeta *types.ClusterMeta,
	storageCli bcsapi.Storage) []*types.NamespaceMeta {
	start := time.Now()
	namespaces, err := storageCli.QueryK8SNamespace(clusterMeta.ClusterID)
	namespaceList := make([]*types.NamespaceMeta, 0)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetK8sNamespace", "GET", err, start)
		blog.Errorf("get cluster %s namespace list error :%v", clusterMeta.ClusterID, err)
		return namespaceList
	}
	prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetK8sNamespace", "GET", err, start)
	clusterLabel := clusterMeta.Label
	namespaceProjectID := clusterMeta.ProjectID
	namespaceBusinessID := clusterMeta.BusinessID
	namespaceProjectCode := clusterMeta.ProjectCode
	pmConn, err := g.projectManager.GetBcsProjectManagerConn()
	if err != nil {
		blog.Errorf("get pm conn error:%v", err)
		return namespaceList
	}
	defer pmConn.Close() // nolint
	pmCli := g.projectManager.NewGrpcClientWithHeader(ctx, pmConn)
	for _, namespace := range namespaces {
		if clusterLabel != nil && clusterLabel["isShared"] == "true" {
			nsAnnotation := namespace.Data.Annotations
			if projectCode, ok := nsAnnotation["io.tencent.bcs.projectcode"]; ok {
				namespaceProjectCode = projectCode
				project, err := g.GetProjectInfo(ctx, "", namespaceProjectCode, pmCli)
				if err != nil {
					blog.Errorf("get project info err:%v", err)
				} else if project != nil {
					namespaceBusinessID = project.BusinessID
					namespaceProjectID = project.ProjectID
				}
			} else {
				namespaceProjectCode = clusterMeta.ProjectCode
				namespaceBusinessID = clusterMeta.BusinessID
				namespaceProjectID = clusterMeta.ProjectID
			}
		}
		namespaceMeta := &types.NamespaceMeta{
			ProjectID:   namespaceProjectID,
			ProjectCode: namespaceProjectCode,
			BusinessID:  namespaceBusinessID,
			ClusterID:   clusterMeta.ClusterID,
			ClusterType: types.Kubernetes,
			Name:        namespace.ResourceName,
			IsBKMonitor: clusterMeta.IsBKMonitor,
		}
		namespaceList = append(namespaceList, namespaceMeta)
	}
	return namespaceList
}

// GetMesosNamespaceList get mesos namespace list
func (g *ResourceGetter) GetMesosNamespaceList(clusterMeta *types.ClusterMeta,
	storageCli bcsapi.Storage) []*types.NamespaceMeta {
	namespaceList := make([]*types.NamespaceMeta, 0)
	start := time.Now()
	namespaces, err := storageCli.QueryMesosNamespace(clusterMeta.ClusterID)
	if err != nil {
		prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosNamespace", "GET", err, start)
		blog.Errorf("get cluster %s namespace list error :%v", clusterMeta.ClusterID, err)
		return namespaceList
	}
	prom.ReportLibRequestMetric(prom.BkBcsStorage, "GetMesosNamespace", "GET", err, start)
	for _, namespace := range namespaces {
		// generate namespace metadata
		namespaceMeta := &types.NamespaceMeta{
			ProjectID:   clusterMeta.ProjectID,
			ProjectCode: clusterMeta.ProjectCode,
			BusinessID:  clusterMeta.BusinessID,
			ClusterID:   clusterMeta.ClusterID,
			ClusterType: types.Mesos,
			Name:        string(*namespace),
			IsBKMonitor: clusterMeta.IsBKMonitor,
		}
		namespaceList = append(namespaceList, namespaceMeta)
	}
	return namespaceList
}

// GetPodAutoscalerList get podAutoscaler list by namespace
// gpa, hpa
func (g *ResourceGetter) GetPodAutoscalerList(podAutoscalerType string, namespaces []*types.NamespaceMeta,
	k8sStorageCli bcsapi.Storage) ([]*types.PodAutoscalerMeta, error) {
	autoscalerList := make([]*types.PodAutoscalerMeta, 0)
	switch podAutoscalerType {
	case types.HPAType:
		for _, namespace := range namespaces {
			startTime := time.Now()
			hpaList, err := k8sStorageCli.QueryK8sHPA(namespace.ClusterID, namespace.Name)
			if err != nil {
				prom.ReportLibRequestMetric(prom.BkBcsStorage, "QueryK8sHPA", "GET", err, startTime)
				blog.Errorf("get hpa list error, cluster:%s, namespace:%s, error:%v",
					namespace.ClusterID, namespace.Name, err)
				return autoscalerList, err
			}
			prom.ReportLibRequestMetric(prom.BkBcsStorage, "QueryK8sHPA", "GET", err, startTime)
			// generate hpa metadata
			for _, hpa := range hpaList {
				hpaMeta := &types.PodAutoscalerMeta{
					ProjectID:          namespace.ProjectID,
					ProjectCode:        namespace.ProjectCode,
					ClusterID:          namespace.ClusterID,
					BusinessID:         namespace.BusinessID,
					ClusterType:        namespace.ClusterType,
					Namespace:          namespace.Name,
					TargetResourceType: hpa.Data.Spec.ScaleTargetRef.Kind,
					TargetWorkloadName: hpa.Data.Spec.ScaleTargetRef.Name,
					PodAutoscaler:      hpa.Data.Name,
				}
				autoscalerList = append(autoscalerList, hpaMeta)
			}
		}
	case types.GPAType:
		for _, namespace := range namespaces {
			startTime := time.Now()
			gpaList, err := k8sStorageCli.QueryK8sGPA(namespace.ClusterID, namespace.Name)
			if err != nil {
				prom.ReportLibRequestMetric(prom.BkBcsStorage, "QueryK8sGPA", "GET", err, startTime)
				blog.Errorf("get gpa list error, cluster:%s, namespace:%s, error:%v",
					namespace.ClusterID, namespace.Name, err)
				return autoscalerList, err
			}
			prom.ReportLibRequestMetric(prom.BkBcsStorage, "QueryK8sGPA", "GET", err, startTime)
			for _, gpa := range gpaList {
				// generate gpa metadata
				gpaMeta := &types.PodAutoscalerMeta{
					ProjectID:          namespace.ProjectID,
					ProjectCode:        namespace.ProjectCode,
					ClusterID:          namespace.ClusterID,
					BusinessID:         namespace.BusinessID,
					ClusterType:        namespace.ClusterType,
					Namespace:          namespace.Name,
					TargetResourceType: gpa.Data.Spec.ScaleTargetRef.Kind,
					TargetWorkloadName: gpa.Data.Spec.ScaleTargetRef.Name,
					PodAutoscaler:      gpa.Data.Name,
				}
				autoscalerList = append(autoscalerList, gpaMeta)
			}
		}
	}
	return autoscalerList, nil
}

// generateK8sWorkloadList generate k8s workload metadata list
func generateK8sWorkloadList(namespaceMeta *types.NamespaceMeta, workloadType string,
	commonHeader storage.CommonDataHeader) *types.WorkloadMeta {
	workloadMeta := &types.WorkloadMeta{
		ProjectID:    namespaceMeta.ProjectID,
		ProjectCode:  namespaceMeta.ProjectCode,
		BusinessID:   namespaceMeta.BusinessID,
		ClusterID:    namespaceMeta.ClusterID,
		ClusterType:  namespaceMeta.ClusterType,
		Namespace:    commonHeader.Namespace,
		ResourceType: workloadType,
		Name:         commonHeader.ResourceName,
		IsBKMonitor:  namespaceMeta.IsBKMonitor,
	}
	return workloadMeta
}

// generateMesosWorkloadList generate mesos workload metadata list
func generateMesosWorkloadList(cluster *types.ClusterMeta, workloadType string,
	commonHeader storage.CommonDataHeader) *types.WorkloadMeta {
	workloadMeta := &types.WorkloadMeta{
		ProjectID:    cluster.ProjectID,
		BusinessID:   cluster.BusinessID,
		ClusterID:    cluster.ClusterID,
		ClusterType:  cluster.ClusterType,
		Namespace:    commonHeader.Namespace,
		ResourceType: workloadType,
		Name:         commonHeader.ResourceName,
		IsBKMonitor:  cluster.IsBKMonitor,
	}
	return workloadMeta
}

// removeDuplicateCluster use map remove duplicate public cluster
func removeDuplicateCluster(clusterList []*cm.Cluster) []*cm.Cluster {
	clusterMap := make(map[string]struct{})
	result := make([]*cm.Cluster, 0)
	for _, cluster := range clusterList {
		if cluster.IsShared {
			clusterMap[cluster.ClusterID] = struct{}{}
			result = append(result, cluster)
		}
	}
	for _, cluster := range clusterList {
		if _, ok := clusterMap[cluster.ClusterID]; !ok {
			clusterMap[cluster.ClusterID] = struct{}{}
			result = append(result, cluster)
		}
	}
	return result
}
