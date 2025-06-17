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

// Package handler is a package for handling requests
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedtasks "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/tasks"
)

const (
	startLoopTickTime          = 3 * time.Minute
	syncNamespaceQuotaTickTime = 5 * time.Minute
)

// Controller is an interface for a controller that syncs namespace quota
type Controller interface {
	Start(ctx context.Context)
	Stop()
}

// SyncNamespaceQuotaController is a struct that implements the Controller interface
type SyncNamespaceQuotaController struct {
	fedClusterID  string
	hostClusterID string

	cancelFunc context.CancelFunc

	taskmanager *task.TaskManager
	clusterCli  cluster.Client
	store       store.FederationMangerModel
}

// NewSyncNamespaceQuotaController creates a new SyncNamespaceQuotaController
func NewSyncNamespaceQuotaController() *SyncNamespaceQuotaController {
	return &SyncNamespaceQuotaController{}
}

// Start starts the SyncNamespaceQuotaController
func (s *SyncNamespaceQuotaController) Start(ctx context.Context) {
	blog.Infof("syncNamespaceQuotaController is running, fedClusterID: %s, hostClusterID: %s",
		s.fedClusterID, s.hostClusterID)

	if ctx == nil {
		ctx = context.Background()
	}

	ticker := time.NewTicker(syncNamespaceQuotaTickTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processNamespaces(ctx)
		case <-ctx.Done():
			blog.Errorf("syncNamespaceQuotaController has been stopped, fedClusterID: %s, hostClusterID: %s",
				s.fedClusterID, s.hostClusterID)
			return
		}
	}
}

// processNamespaces 处理所有命名空间的同步逻辑
func (s *SyncNamespaceQuotaController) processNamespaces(ctx context.Context) {
	namespaceList, err := s.clusterCli.ListNamespace(s.hostClusterID)
	if err != nil {
		blog.Errorf("list namespaces failed, hostClusterID %s, err %s", s.hostClusterID, err.Error())
		return
	}

	if len(namespaceList) == 0 {
		blog.Errorf("namespaceList is empty for hostClusterID [%s]", s.hostClusterID)
		return
	}

	for _, ns := range namespaceList {
		s.processSingleNamespace(ctx, ns)
	}
}

// processSingleNamespace 处理单个命名空间的同步逻辑
func (s *SyncNamespaceQuotaController) processSingleNamespace(ctx context.Context, ns v1.Namespace) {
	blog.Infof("sync namespace quota controller for hostClusterID [%s], namespace [%s]",
		s.hostClusterID, ns.Name)

	if !s.shouldProcessNamespace(ctx, ns) {
		return
	}

	subClusterIDs := s.extractSubClusterIDs(ns)
	validSubClusterIDs, err := s.getValidSubClusterIDs(ctx, s.fedClusterID, subClusterIDs)
	if err != nil {
		blog.Errorf("getValidSubClusterIDs failed, fedClusterID %s, err %s", s.fedClusterID, err.Error())
		return
	}

	if len(validSubClusterIDs) == 0 {
		blog.Errorf("validSubClusterIDs is empty for fedClusterID [%s]", s.fedClusterID)
		return
	}

	s.syncNamespaceToSubClusters(ns, validSubClusterIDs)
}

// shouldProcessNamespace 检查是否需要处理该命名空间
func (s *SyncNamespaceQuotaController) shouldProcessNamespace(ctx context.Context, ns v1.Namespace) bool {
	if ns.Annotations[cluster.CreateNamespaceTaskId] == "" {
		return true
	}

	taskWithID, err := s.taskmanager.GetTaskWithID(ctx, ns.Annotations[cluster.CreateNamespaceTaskId])
	blog.Infof("taskWithID %+v , hostClusterID %s, taskID %s",
		taskWithID, s.hostClusterID, ns.Annotations[cluster.CreateNamespaceTaskId])
	if err != nil {
		blog.Errorf("getTaskWithID failed, hostClusterID %s, err %s", s.hostClusterID, err.Error())
		return true
	}

	if taskWithID != nil && (taskWithID.Status == cluster.TaskStatusRUNNING ||
		taskWithID.Status == cluster.TaskStatusINITIALIZING) {
		blog.Infof("taskWithID.Status is running or initialization, hostClusterID %s, taskID %s",
			s.hostClusterID, ns.Annotations[cluster.CreateNamespaceTaskId])
		return false
	}

	return true
}

// extractSubClusterIDs 从命名空间注解中提取子集群ID
func (s *SyncNamespaceQuotaController) extractSubClusterIDs(ns v1.Namespace) []string {
	subClusterIDs := make([]string, 0)
	if clusterRangeStr, ok := ns.Annotations[cluster.FedNamespaceClusterRangeKey]; ok {
		if len(clusterRangeStr) != 0 {
			for _, sc := range strings.Split(clusterRangeStr, ",") {
				subClusterIDs = append(subClusterIDs, strings.ToUpper(sc))
			}
		}
	}
	return subClusterIDs
}

// syncNamespaceToSubClusters 将命名空间同步到子集群
func (s *SyncNamespaceQuotaController) syncNamespaceToSubClusters(ns v1.Namespace, subClusterIDs []string) {
	for _, subClusterID := range subClusterIDs {
		if err := s.checkSubClusterNamespace(subClusterID, ns.Name); err != nil {
			blog.Errorf("checkSubClusterNamespace failed, hostClusterID %s, subClusterID %s, err %s",
				s.hostClusterID, subClusterID, err.Error())
			continue
		}

		taskID, err := s.getManagedClusterAndBuildTask(ns.Name, subClusterID)
		if err != nil {
			blog.Errorf("getManagedClusterAndBuildTask failed, hostClusterID %s, subClusterID %s, "+
				"namespace %s, err %s", s.hostClusterID, subClusterID, ns.Name, err.Error())
			continue
		}

		if taskID != "" {
			if err := s.updateNamespace(taskID, ns); err != nil {
				blog.Errorf("update namespace failed, hostClusterID %s, subClusterID %s, "+
					"namespace %s, err %s", s.hostClusterID, subClusterID, ns.Name, err.Error())
			}
		}
	}
}

// checkSubClusterNamespace 检测子集群是否有这个命名空间，避免联邦的命名空间在子集群（作为独立集群工作时）和联邦有冲突。
func (s *SyncNamespaceQuotaController) checkSubClusterNamespace(subClusterId, namespace string) error {
	subClusterNamespace, err := s.clusterCli.GetNamespace(subClusterId, namespace)
	if err != nil && !errors.IsNotFound(err) {
		blog.Errorf("checkSubClusterNamespace failed, subClusterId: %s, namespace: %s, err: %s",
			subClusterId, namespace, err.Error())
		return err
	}

	if subClusterNamespace != nil {
		blog.Errorf("checkSubClusterNamespace failed the namespace already exist in subCluster "+
			"subClusterNamespace: %+v", subClusterNamespace)
		return fmt.Errorf("the subCluster [%s] namespace [%s] already exist", subClusterId, namespace)
	}

	return nil
}

func (s *SyncNamespaceQuotaController) updateNamespace(taskID string, ns v1.Namespace) error {

	ns.Annotations[cluster.CreateNamespaceTaskId] = taskID
	uerr := s.clusterCli.UpdateNamespace(s.hostClusterID, &ns)
	if uerr != nil {
		blog.Errorf("updateNamespace failed, hostClusterID %s, namespace %s, err %s",
			s.hostClusterID, ns.Name, uerr.Error())
		return uerr
	}
	blog.Infof("updateNamespace success, hostClusterID %s, namespace %s",
		s.hostClusterID, ns.Name)

	return nil
}

// 获取有效的子集群范围
func (s *SyncNamespaceQuotaController) getManagedClusterAndBuildTask(nsName, subClusterID string) (string, error) {

	// 获取managedCluster 对象
	managedCluster, merr := s.clusterCli.GetManagedCluster(s.hostClusterID, subClusterID)
	if merr != nil {
		blog.Errorf("getManagedCluster failed, hostClusterID %s, subClusterID %s, err %s",
			s.hostClusterID, subClusterID, merr.Error())
		return "", merr
	}
	if managedCluster == nil {
		blog.Errorf("managedCluster is nil for cluster [%s] and subCluster [%s]",
			s.hostClusterID, subClusterID)
		return "", fmt.Errorf("managedCluster is nil for cluster [%s] and subCluster [%s]", s.hostClusterID, subClusterID)
	}
	taskID, berr := s.buildSubClusterTask(nsName, subClusterID, managedCluster.Labels)
	if berr != nil {
		blog.Errorf("buildSubClusterTask failed, hostClusterID %s, subClusterID %s, namespace %s, err %s",
			s.hostClusterID, subClusterID, nsName, berr.Error())
		return "", berr
	}

	blog.Infof("buildSubClusterTask success, hostClusterID %s, subClusterID %s, namespace %s",
		s.hostClusterID, subClusterID, nsName)
	return taskID, nil
}

// 构建子集群任务
func (s *SyncNamespaceQuotaController) buildSubClusterTask(nsName, subClusterID string,
	managedClusterLabels map[string]string) (string, error) {

	blog.Infof("buildSubClusterTask for hostClusterID [%s], subClusterID [%s], namespace [%s],"+
		" managedClusterLabels %v", s.hostClusterID, subClusterID, nsName, managedClusterLabels)

	var err error
	t := &types.Task{}
	// 判断managedCluster对象的labels的标识
	switch managedClusterLabels[cluster.ManagedClusterTypeLabel] {
	case cluster.SubClusterForTaiji:
		t, err = fedtasks.NewSyncTjNamespaceQuotaTask(&fedtasks.SyncTjNamespaceQuotaOptions{
			HostClusterID: s.hostClusterID,
			Namespace:     nsName,
		}).BuildTask("admin")
		if err != nil {
			blog.Errorf("build task failed, task: %v, err: %s", t, err.Error())
			return "", err
		}
	case cluster.SubClusterForSuanli:
		t, err = fedtasks.NewSyncSlNamespaceQuotaTask(&fedtasks.SyncSlNamespaceQuotaOptions{
			HostClusterID: s.hostClusterID,
			Namespace:     nsName,
		}).BuildTask("admin")
		if err != nil {
			blog.Errorf("build task failed, task: %v, err: %s", t, err.Error())
			return "", err
		}
	case cluster.SubClusterForHunbu:
		bytes, berr := json.Marshal(managedClusterLabels)
		if berr != nil {
			blog.Errorf("json.Marshal managedClusterLabels failed, err: %s", berr.Error())
			return "", berr
		}
		t, err = fedtasks.NewSyncHbNamespaceQuotaTask(&fedtasks.SyncHbNamespaceQuotaOptions{
			Namespace:     nsName,
			HostClusterID: s.hostClusterID,
			SubClusterID:  subClusterID,
			Labels:        string(bytes),
		}).BuildTask("admin")
		if err != nil {
			blog.Errorf("build task failed, task: %v, err: %s", t, err.Error())
			return "", err
		}
	default:
		t, err = fedtasks.NewSyncNormalNamespaceQuotaTask(&fedtasks.SyncNormalNamespaceQuotaOptions{
			HostClusterID: s.hostClusterID,
			Namespace:     nsName,
			SubClusterID:  subClusterID,
		}).BuildTask("admin")
		if err != nil {
			blog.Errorf("build task failed, task: %v, err: %s", t, err.Error())
			return "", err
		}
	}

	if err = s.taskmanager.Dispatch(t); err != nil {
		blog.Errorf("run task failed, task: %v, err: %s", t, err.Error())
		return "", err
	}

	return t.TaskID, nil
}

// 获取合法的子集群范围
func (s *SyncNamespaceQuotaController) getValidSubClusterIDs(ctx context.Context, fedClusterID string, subClusterIDs []string) (
	[]string, error) {

	if len(subClusterIDs) == 0 {
		return nil, nil
	}

	subClusterMap := make(map[string]struct{}, len(subClusterIDs))
	for _, id := range subClusterIDs {
		subClusterMap[id] = struct{}{}
	}

	// 获取所有子集群
	listSubClusters, err := s.store.ListSubClusters(ctx, &store.SubClusterListOptions{
		FederationClusterID: fedClusterID,
	})
	if err != nil {
		blog.Errorf("list sub clusters failed, fedClusterID: %s, err: %s", fedClusterID, err.Error())
		return nil, fmt.Errorf("list sub clusters failed: %v", err)
	}

	if len(listSubClusters) == 0 {
		blog.Infof("no subCluster found, fedClusterID: %s, subClusterIDs: %v", fedClusterID, subClusterIDs)
		return nil, nil
	}

	var validSubClusterIDs []string
	for _, subCluster := range listSubClusters {
		if _, exists := subClusterMap[subCluster.SubClusterID]; exists {
			validSubClusterIDs = append(validSubClusterIDs, subCluster.SubClusterID)
		}
	}

	return validSubClusterIDs, nil
}

// Stop 停止
func (s *SyncNamespaceQuotaController) Stop() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
}

// FedNamespaceControllerManager is a manager for managing multiple controllers
type FedNamespaceControllerManager struct {
	Controllers []Controller
}

// NewFedNamespaceControllerManager creates a new FedNamespaceControllerManager
func NewFedNamespaceControllerManager() *FedNamespaceControllerManager {
	return &FedNamespaceControllerManager{}
}

// StartLoop 启动循环
func (f *FedNamespaceControllerManager) StartLoop(ctx context.Context, st store.FederationMangerModel,
	taskmanager *task.TaskManager, clusterCli cluster.Client) error {

	if st == nil || taskmanager == nil || clusterCli == nil {
		return fmt.Errorf("store or taskmanager or clusterCli is nil")
	}

	blog.Infof("StartLoop is running...")
	// 设置n分钟检查一次的定时器
	ticker := time.NewTicker(startLoopTickTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			blog.Infof("StartLoop is starting...")
			// 获取联邦集群列表
			fedClusterList, err := st.ListFederationClusters(ctx, &store.FederationListOptions{
				Conditions: map[string]string{},
			})
			if err != nil {
				blog.Errorf("ListFederationClusters error when get fed clusters: %s", err.Error())
				continue
			}
			newFedClusterMap := make(map[string]string)
			for _, fc := range fedClusterList {
				blog.Infof("StartLoop fedCluster %+v", *fc)
				newFedClusterMap[fc.FederationClusterID] = fc.HostClusterID
			}
			blog.Infof("StartLoop newFedClusterMap %+v", newFedClusterMap)
			// 创建现有控制器映射表
			existingControllers := make(map[string]Controller)
			for _, c := range f.Controllers {
				if rc, ok := c.(*SyncNamespaceQuotaController); ok {
					existingControllers[rc.fedClusterID] = c
				}
			}
			// 对比联邦集群列表和现有控制器
			// 清理不需要的控制器
			var newSyncControllers []Controller
			for _, controller := range f.Controllers {
				if rc, ok := controller.(*SyncNamespaceQuotaController); ok {
					if _, found := newFedClusterMap[rc.fedClusterID]; found {
						blog.Infof("StartLoop new fedClusterID %+v", rc.fedClusterID)
						newSyncControllers = append(newSyncControllers, controller)
					} else {
						blog.Infof("StartLoop stop fedClusterID %+v", rc.fedClusterID)
						// 停止并移除不存在的集群控制器
						controller.Stop()
					}
				}
			}
			blog.Infof("StartLoop newSyncControllers %+v", newSyncControllers)
			f.Controllers = newSyncControllers
			// 添加新的集群控制器
			for fedClusterID, hostClusterID := range newFedClusterMap {
				if _, exists := existingControllers[fedClusterID]; !exists {
					nc := NewSyncNamespaceQuotaController()
					nc.clusterCli = clusterCli
					nc.taskmanager = taskmanager
					nc.fedClusterID = fedClusterID
					nc.store = st
					nc.hostClusterID = hostClusterID
					// 初始化context和cancelFunc
					childCtx, cancelFunc := context.WithCancel(ctx)
					nc.cancelFunc = cancelFunc
					go nc.Start(childCtx)
					f.Controllers = append(f.Controllers, nc)
				}
			}
		case <-ctx.Done():
			// 停止所有控制器
			blog.Infof("StartLoop end stop controllers %+v", f.Controllers)
			for _, c := range f.Controllers {
				c.Stop()
			}
			return fmt.Errorf("startLoop context is canceled")
		}
	}
}
