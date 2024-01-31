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

// Package syncer define syncer methods
package syncer

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	bkcmdbkube "configcenter/src/kube/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
	bsc "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/bcsstorage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/cmdb"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
)

// Syncer the syncer
type Syncer struct {
	BkcmdbSynchronizerOption *option.BkcmdbSynchronizerOption
	ClientTls                *tls.Config
	//Rabbit                   mq.MQ
	CMDBClient client.CMDBClient
}

var bkBizID = int64(0)
var hostID = int64(0)

// NewSyncer create a new Syncer
func NewSyncer(bkcmdbSynchronizerOption *option.BkcmdbSynchronizerOption) *Syncer {
	return &Syncer{
		BkcmdbSynchronizerOption: bkcmdbSynchronizerOption,
	}
}

// Init init the synchronizer
func (s *Syncer) Init() {
	blog.InitLogs(s.BkcmdbSynchronizerOption.Bcslog)
	err := s.initTlsConfig()
	if err != nil {
		blog.Errorf("init tls config failed, err: %s", err.Error())
	}

	err = s.initCMDBClient()
	if err != nil {
		blog.Errorf("init cmdb client failed, err: %s", err.Error())
	}
	bkBizID = s.BkcmdbSynchronizerOption.Synchronizer.BkBizID
	hostID = s.BkcmdbSynchronizerOption.Synchronizer.HostID
}

// init Tls Config
func (s *Syncer) initTlsConfig() error {
	if len(s.BkcmdbSynchronizerOption.Client.ClientCrt) != 0 &&
		len(s.BkcmdbSynchronizerOption.Client.ClientKey) != 0 &&
		len(s.BkcmdbSynchronizerOption.Client.ClientCa) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(
			s.BkcmdbSynchronizerOption.Client.ClientCa,
			s.BkcmdbSynchronizerOption.Client.ClientCrt,
			s.BkcmdbSynchronizerOption.Client.ClientKey,
			s.BkcmdbSynchronizerOption.Client.ClientCrtPwd,
		)
		//static.ClientCertPwd)
		if err != nil {
			blog.Errorf("init tls config failed, err: %s", err.Error())
			return err
		}
		s.ClientTls = tlsConfig
		blog.Infof("init tls config success")

	}
	return nil
}

// init CMDB Client
func (s *Syncer) initCMDBClient() error {
	blog.Infof("init cmdb client")
	cmdbCli := cmdb.NewCmdbClient(&cmdb.Options{
		AppCode:    s.BkcmdbSynchronizerOption.CMDB.AppCode,
		AppSecret:  s.BkcmdbSynchronizerOption.CMDB.AppSecret,
		BKUserName: s.BkcmdbSynchronizerOption.CMDB.BKUserName,
		Server:     s.BkcmdbSynchronizerOption.CMDB.Server,
		Debug:      s.BkcmdbSynchronizerOption.CMDB.Debug,
	})

	// GetCMDBClient get cmdb client
	cli, err := cmdbCli.GetCMDBClient()
	if err != nil {
		blog.Errorf("get cmdb client failed: %s", err.Error())
	}
	s.CMDBClient = cli
	return nil
}

// SyncCluster  sync cluster
func (s *Syncer) SyncCluster(cluster *cmp.Cluster) error {
	blog.Infof("sync cluster: %s", cluster.ClusterID)
	clusterNetwork := make([]string, 0)
	if cluster.NetworkSettings != nil {
		clusterNetwork = append(clusterNetwork, cluster.NetworkSettings.ClusterIPv4CIDR)
		clusterNetwork = append(clusterNetwork, cluster.NetworkSettings.ServiceIPv4CIDR)
		clusterNetwork = append(clusterNetwork, cluster.NetworkSettings.MultiClusterCIDR...)
	}

	// GetBkCluster get bkcluster
	bkCluster, err := s.GetBkCluster(cluster)
	clusterType := "INDEPENDENT_CLUSTER"
	if cluster.IsShared {
		clusterType = "SHARE_CLUSTER"
	}

	if err != nil {
		if err.Error() == "cluster not found" {
			var clusterBkBizID int64
			if bkBizID == 0 {
				clusterBkBizID, err = strconv.ParseInt(cluster.BusinessID, 10, 64)
				if err != nil {
					blog.Errorf("An error occurred: %s\n", err)
				} else {
					blog.Infof("Successfully converted string to int64: %d\n", bkBizID)
				}
			} else {
				clusterBkBizID = bkBizID
			}
			// CreateBcsCluster creates a new BCS cluster with the given request.
			_, err := s.CMDBClient.CreateBcsCluster(&client.CreateBcsClusterRequest{
				BKBizID:          &clusterBkBizID,
				Name:             &cluster.ClusterID,
				SchedulingEngine: &cluster.EngineType,
				UID:              &cluster.ClusterID,
				XID:              &cluster.SystemID,
				Version:          &cluster.ClusterBasicSettings.Version,
				NetworkType:      &cluster.NetworkType,
				Region:           &cluster.Region,
				Vpc:              &cluster.VpcID,
				Network:          &clusterNetwork,
				Type:             &clusterType,
				Environment:      &cluster.Environment,
			})
			if err != nil {
				blog.Errorf("create bcs cluster failed, err: %s", err.Error())
			}
		} else {
			blog.Errorf("get bcs cluster failed, err: %s", err.Error())
			return err
		}
	}

	blog.Infof("get bcs cluster success, cluster: %v", bkCluster)

	if bkCluster != nil {
		// UpdateBcsCluster updates the BCS cluster with the given request.
		err = s.CMDBClient.UpdateBcsCluster(&client.UpdateBcsClusterRequest{
			BKBizID: &bkCluster.BizID,
			IDs:     &[]int64{bkCluster.ID},
			Data: &client.UpdateBcsClusterRequestData{
				Version:     &cluster.ClusterBasicSettings.Version,
				NetworkType: &cluster.NetworkType,
				Region:      &cluster.Region,
				Network:     &clusterNetwork,
				Environment: &cluster.Environment,
			},
		})
		if err != nil {
			blog.Errorf("update bcs cluster failed, err: %s", err.Error())
		}

		if *bkCluster.Type != clusterType {
			// UpdateBcsClusterType updates the BCS cluster type with the given request.
			err = s.CMDBClient.UpdateBcsClusterType(&client.UpdateBcsClusterTypeRequest{
				BKBizID: &bkCluster.BizID,
				ID:      &bkCluster.ID,
				Type:    &clusterType,
			})
			if err != nil {
				blog.Errorf("update bcs cluster type failed, err: %s", err.Error())
			}
		}

	}

	return nil
}

// SyncNodes sync nodes
func (s *Syncer) SyncNodes(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) error {
	// GetBcsStorageClient is a function that returns a BCS storage client.
	storageCli, err := s.GetBcsStorageClient()
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}
	// get node
	nodeList, err := storageCli.QueryK8SNode(cluster.ClusterID)
	if err != nil {
		blog.Errorf("query k8s node failed, err: %s", err.Error())
		return err
	}
	blog.Infof("query k8s node success, nodes: %v", nodeList)

	// GetBkNodes get bknodes
	bkNodeList, err := s.GetBkNodes(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})

	if err != nil {
		blog.Errorf("get bk node failed, err: %s", err.Error())
		return err
	}

	nodeToAdd := make([]client.CreateBcsNodeRequestData, 0)
	nodeToUpdate := make(map[int64]*client.UpdateBcsNodeRequestData, 0)
	nodeToDelete := make([]int64, 0)

	nodeMap := make(map[string]*storage.K8sNode)
	for _, node := range nodeList {
		nodeMap[node.Data.Name] = node
	}

	bkNodeMap := make(map[string]*bkcmdbkube.Node)
	for k, v := range *bkNodeList {
		bkNodeMap[*v.Name] = &(*bkNodeList)[k]

		if _, ok := nodeMap[*v.Name]; !ok {
			nodeToDelete = append(nodeToDelete, v.ID)
		}
	}

	for k, v := range nodeMap {
		if _, ok := bkNodeMap[k]; ok {
			// CompareNode compare bknode and k8snode
			needToUpdate, updateData := s.CompareNode(bkNodeMap[k], v)

			if needToUpdate {
				nodeToUpdate[bkNodeMap[k].ID] = updateData
			}
		} else {
			// GenerateBkNodeData generate bknode data from k8snode
			nodeToAdd = append(nodeToAdd, s.GenerateBkNodeData(bkCluster, v))
		}
	}

	s.CreateBkNodes(bkCluster, &nodeToAdd)
	s.DeleteBkNodes(bkCluster, &nodeToDelete)
	s.UpdateBkNodes(bkCluster, &nodeToUpdate)

	return err
}

// SyncNamespaces sync namespaces
func (s *Syncer) SyncNamespaces(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) error {
	// GetBcsStorageClient is a function that returns a BCS storage client.
	storageCli, err := s.GetBcsStorageClient()
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}

	// get namespace
	nsList, err := storageCli.QueryK8SNamespace(cluster.ClusterID)
	if err != nil {
		blog.Errorf("query k8s namespace failed, err: %s", err.Error())
		return err
	}
	blog.Infof("query k8s namespace success, namespaces: %v", nsList)

	// GetBkNamespaces get bknamespaces
	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	blog.Infof("get bcs namespace success, namespaces: %v", bkNamespaceList)

	nsToAdd := make(map[int64][]bkcmdbkube.Namespace, 0)
	nsToUpdate := make(map[int64]*client.UpdateBcsNamespaceRequestData, 0)
	nsToDelete := make([]int64, 0)

	nsMap := make(map[string]*storage.Namespace)
	for _, ns := range nsList {
		nsMap[ns.Data.Name] = ns
	}

	bkNsMap := make(map[string]*bkcmdbkube.Namespace)
	for k, v := range *bkNamespaceList {
		bkNsMap[v.Name] = &(*bkNamespaceList)[k]

		if _, ok := nsMap[v.Name]; !ok {
			nsToDelete = append(nsToDelete, v.ID)
		}
	}

	// GetProjectManagerGrpcGwClient is a function that returns a project manager gRPC gateway client.
	pmCli, err := s.GetProjectManagerGrpcGwClient()
	if err != nil {
		blog.Errorf("get project manager grpc gw client failed, err: %s", err.Error())
		return nil
	}

	for k, v := range nsMap {
		if _, ok := bkNsMap[k]; ok {
			// CompareNamespace compare bkns and k8sns
			needToUpdate, updateData := s.CompareNamespace(bkNsMap[k], v)
			if needToUpdate {
				nsToUpdate[bkNsMap[k].ID] = updateData
			}
		} else {
			bizid := bkCluster.BizID
			if projectCode, ok := v.Data.Annotations["io.tencent.bcs.projectcode"]; ok {
				gpr := pmp.GetProjectRequest{
					ProjectIDOrCode: projectCode,
				}

				if project, err := pmCli.Cli.GetProject(pmCli.Ctx, &gpr); err == nil {
					if project != nil && project.Data != nil && project.Data.BusinessID != "" {
						bizid, err = strconv.ParseInt(project.Data.BusinessID, 10, 64)
						if err != nil {
							blog.Errorf("parse string err: %v", err)
						}
					}
				} else {
					blog.Errorf("get project error : %v", err)
				}
			}

			if bizid != bkCluster.BizID {
				bizid = int64(71)
			}

			if _, ok = nsToAdd[bizid]; ok {
				nsToAdd[bizid] = append(nsToAdd[bizid], s.GenerateBkNsData(bkCluster, v))
			} else {
				nsToAdd[bizid] = []bkcmdbkube.Namespace{s.GenerateBkNsData(bkCluster, v)}
			}
		}
	}

	s.DeleteBkNamespaces(bkCluster, &nsToDelete)
	s.CreateBkNamespaces(bkCluster, nsToAdd)
	s.UpdateBkNamespaces(bkCluster, &nsToUpdate)

	return nil
}

// SyncWorkloads sync workloads
func (s *Syncer) SyncWorkloads(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) error {
	// syncDeployments sync deployments
	err := s.syncDeployments(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync deployment failed, err: %s", err.Error())
	}

	// syncStatefulSets sync statefulsets
	err = s.syncStatefulSets(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync statefulset failed, err: %s", err.Error())
	}

	// syncDaemonSets sync daemonsets
	err = s.syncDaemonSets(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync daemonset failed, err: %s", err.Error())
	}

	// syncGameDeployments sync gamedeployments
	err = s.syncGameDeployments(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync gamedeployment failed, err: %s", err.Error())
	}

	// syncGameStatefulSets sync gamestatefulsets
	err = s.syncGameStatefulSets(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync gamestatefulset failed, err: %s", err.Error())
	}

	// syncWorkloadPods sync workloadPods
	err = s.syncWorkloadPods(cluster, bkCluster)
	if err != nil {
		blog.Errorf("sync workload pods failed, err: %s", err.Error())
	}

	return err
}

// syncDeployments sync deployments
func (s *Syncer) syncDeployments(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) error {
	kind := "deployment"
	// GetBcsStorageClient is a function that returns a BCS storage client.
	storageCli, err := s.GetBcsStorageClient()
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}

	// GetBkNamespaces get bknamespaces
	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	bkNamespaceMap := make(map[string]*bkcmdbkube.Namespace)

	for k, v := range *bkNamespaceList {
		bkNamespaceMap[v.Name] = &(*bkNamespaceList)[k]
	}

	//get workload
	deploymentList := make([]*storage.Deployment, 0)
	bkDeploymentList := make([]bkcmdbkube.Deployment, 0)

	for _, ns := range *bkNamespaceList {
		deployments, err := storageCli.QueryK8SDeployment(cluster.ClusterID, ns.Name)
		if err != nil {
			blog.Errorf("query k8s deployment failed, err: %s", err.Error())
			return err
		}
		deploymentList = append(deploymentList, deployments...)
	}
	blog.Infof("query k8s deployment success, deployments: %v", deploymentList)

	// GetBkWorkloads get bkworkloads
	bkDeployments, err := s.GetBkWorkloads(bkCluster.BizID, "deployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk deployment failed, err: %s", err.Error())
		return err
	}

	for _, bkDeployment := range *bkDeployments {
		b := bkcmdbkube.Deployment{}
		err := common.InterfaceToStruct(bkDeployment, &b)
		if err != nil {
			blog.Errorf("convert bk deployment failed, err: %s", err.Error())
			return err
		}

		bkDeploymentList = append(bkDeploymentList, b)
	}

	deploymentToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	deploymentToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
	deploymentToDelete := make([]int64, 0)

	deploymentMap := make(map[string]*storage.Deployment)
	for _, v := range deploymentList {
		deploymentMap[v.Data.Namespace+v.Data.Name] = v
	}

	bkDeploymentMap := make(map[string]*bkcmdbkube.Deployment)
	for k, v := range bkDeploymentList {
		bkDeploymentMap[v.Namespace+v.Name] = &bkDeploymentList[k]

		if _, ok := deploymentMap[v.Namespace+v.Name]; !ok {
			deploymentToDelete = append(deploymentToDelete, v.ID)
		}
	}

	for k, v := range deploymentMap {
		if _, ok := bkDeploymentMap[k]; !ok {
			// GenerateBkDeployment generate bkdeployment from k8sdeployment
			toAddData := s.GenerateBkDeployment(bkNamespaceMap[v.Data.Namespace], v)

			if toAddData != nil {
				if _, ok = deploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID]; ok {
					deploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = append(deploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID], *toAddData)
				} else {
					deploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
				}
			}
		} else {
			// CompareDeployment compare bkdeployment and k8sdeployment
			needToUpdate, updateData := s.CompareDeployment(bkDeploymentMap[k], v)

			if needToUpdate {
				deploymentToUpdate[bkDeploymentMap[k].ID] = updateData
			}
		}
	}

	s.DeleteBkWorkloads(bkCluster, kind, &deploymentToDelete)
	s.CreateBkWorkloads(bkCluster, kind, deploymentToAdd)
	s.UpdateBkWorkloads(bkCluster, kind, &deploymentToUpdate)

	return nil
}

// syncStatefulSets sync statefulsets
func (s *Syncer) syncStatefulSets(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) error {
	kind := "statefulSet"
	// GetBcsStorageClient is a function that returns a BCS storage client.
	storageCli, err := s.GetBcsStorageClient()
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}

	// GetBkNamespaces get bknamespaces
	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	bkNamespaceMap := make(map[string]*bkcmdbkube.Namespace)

	for k, v := range *bkNamespaceList {
		bkNamespaceMap[v.Name] = &(*bkNamespaceList)[k]
	}

	statefulSetList := make([]*storage.StatefulSet, 0)
	bkStatefulSetList := make([]bkcmdbkube.StatefulSet, 0)

	for _, ns := range *bkNamespaceList {
		statefulSets, err := storageCli.QueryK8SStatefulSet(cluster.ClusterID, ns.Name)
		if err != nil {
			blog.Errorf("query k8s statefulset failed, err: %s", err.Error())
			return err
		}
		statefulSetList = append(statefulSetList, statefulSets...)
	}
	blog.Infof("get statefulset list success, len: %d", len(statefulSetList))

	// GetBkWorkloads get bkworkloads
	bkStatefulSets, err := s.GetBkWorkloads(bkCluster.BizID, "statefulSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk statefulset failed, err: %s", err.Error())
		return err
	}

	for _, bkStatefulSet := range *bkStatefulSets {
		b := bkcmdbkube.StatefulSet{}
		err := common.InterfaceToStruct(bkStatefulSet, &b)
		if err != nil {
			blog.Errorf("convert bk statefulset failed, err: %s", err.Error())
			return err
		}

		bkStatefulSetList = append(bkStatefulSetList, b)
	}

	statefulSetToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	statefulSetToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
	statefulSetToDelete := make([]int64, 0)

	statefulSetMap := make(map[string]*storage.StatefulSet)
	for _, v := range statefulSetList {
		statefulSetMap[v.Data.Namespace+v.Data.Name] = v
	}

	bkStatefulSetMap := make(map[string]*bkcmdbkube.StatefulSet)
	for k, v := range bkStatefulSetList {
		bkStatefulSetMap[v.Namespace+v.Name] = &bkStatefulSetList[k]

		if _, ok := statefulSetMap[v.Namespace+v.Name]; !ok {
			statefulSetToDelete = append(statefulSetToDelete, v.ID)
		}
	}

	for k, v := range statefulSetMap {
		if _, ok := bkStatefulSetMap[k]; !ok {
			toAddData := s.GenerateBkStatefulSet(bkNamespaceMap[v.Data.Namespace], v)

			if toAddData != nil {
				if _, ok = statefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID]; ok {
					statefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = append(statefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID], *toAddData)
				} else {
					statefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
				}
			}
		} else {
			needToUpdate, updateData := s.CompareStatefulSet(bkStatefulSetMap[k], v)

			if needToUpdate {
				statefulSetToUpdate[bkStatefulSetMap[k].ID] = updateData
			}
		}
	}

	s.DeleteBkWorkloads(bkCluster, kind, &statefulSetToDelete)
	s.CreateBkWorkloads(bkCluster, kind, statefulSetToAdd)
	s.UpdateBkWorkloads(bkCluster, kind, &statefulSetToUpdate)

	return nil
}

// syncDaemonSets sync daemonsets
func (s *Syncer) syncDaemonSets(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) error {
	kind := "daemonSet"
	// GetBcsStorageClient is a function that returns a BCS storage client.
	storageCli, err := s.GetBcsStorageClient()
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}

	// GetBkNamespaces get bknamespaces
	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	bkNamespaceMap := make(map[string]*bkcmdbkube.Namespace)

	for k, v := range *bkNamespaceList {
		bkNamespaceMap[v.Name] = &(*bkNamespaceList)[k]
	}

	daemonSetList := make([]*storage.DaemonSet, 0)
	bkDaemonSetList := make([]bkcmdbkube.DaemonSet, 0)

	for _, ns := range *bkNamespaceList {
		daemonSets, err := storageCli.QueryK8SDaemonSet(cluster.ClusterID, ns.Name)
		if err != nil {
			blog.Errorf("query k8s daemonset failed, err: %s", err.Error())
			return err
		}
		daemonSetList = append(daemonSetList, daemonSets...)
	}
	blog.Infof("get daemonset list success, len: %d", len(daemonSetList))

	bkDaemonSets, err := s.GetBkWorkloads(bkCluster.BizID, "daemonSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk daemonset failed, err: %s", err.Error())
		return err
	}

	for _, bkDaemonSet := range *bkDaemonSets {
		b := bkcmdbkube.DaemonSet{}
		err := common.InterfaceToStruct(bkDaemonSet, &b)
		if err != nil {
			blog.Errorf("convert bk daemonset failed, err: %s", err.Error())
			return err
		}

		bkDaemonSetList = append(bkDaemonSetList, b)
	}

	daemonSetToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	daemonSetToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
	daemonSetToDelete := make([]int64, 0)

	daemonSetMap := make(map[string]*storage.DaemonSet)
	for _, v := range daemonSetList {
		daemonSetMap[v.Data.Namespace+v.Data.Name] = v
	}

	bkDaemonSetMap := make(map[string]*bkcmdbkube.DaemonSet)
	for k, v := range bkDaemonSetList {
		bkDaemonSetMap[v.Namespace+v.Name] = &bkDaemonSetList[k]

		if _, ok := daemonSetMap[v.Namespace+v.Name]; !ok {
			daemonSetToDelete = append(daemonSetToDelete, v.ID)
		}
	}

	for k, v := range daemonSetMap {
		if _, ok := bkDaemonSetMap[k]; !ok {
			toAddData := s.GenerateBkDaemonSet(bkNamespaceMap[v.Data.Namespace], v)

			if toAddData != nil {
				if _, ok = daemonSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID]; ok {
					daemonSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = append(daemonSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID], *toAddData)
				} else {
					daemonSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
				}
			}
		} else {
			needToUpdate, updateData := s.CompareDaemonSet(bkDaemonSetMap[k], v)

			if needToUpdate {
				daemonSetToUpdate[bkDaemonSetMap[k].ID] = updateData
			}
		}
	}

	s.DeleteBkWorkloads(bkCluster, kind, &daemonSetToDelete)
	s.CreateBkWorkloads(bkCluster, kind, daemonSetToAdd)
	s.UpdateBkWorkloads(bkCluster, kind, &daemonSetToUpdate)

	return nil
}

// syncGameDeployments sync gamedeployments
func (s *Syncer) syncGameDeployments(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) error {
	kind := "gameDeployment"
	storageCli, err := s.GetBcsStorageClient()
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}

	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	bkNamespaceMap := make(map[string]*bkcmdbkube.Namespace)

	for k, v := range *bkNamespaceList {
		bkNamespaceMap[v.Name] = &(*bkNamespaceList)[k]
	}

	gameDeploymentList := make([]*storage.GameDeployment, 0)
	bkGameDeploymentList := make([]bkcmdbkube.GameDeployment, 0)

	for _, ns := range *bkNamespaceList {
		gameDeployments, err := storageCli.QueryK8SGameDeployment(cluster.ClusterID, ns.Name)
		if err != nil {
			blog.Errorf("query k8s gamedeployment failed, err: %s", err.Error())
			return err
		}
		gameDeploymentList = append(gameDeploymentList, gameDeployments...)
	}
	blog.Infof("game deployment list: %v", gameDeploymentList)

	bkGameDeployments, err := s.GetBkWorkloads(bkCluster.BizID, "gameDeployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk gamedeployment failed, err: %s", err.Error())
		return err
	}

	for _, bkGameDeployment := range *bkGameDeployments {
		b := bkcmdbkube.GameDeployment{}
		err := common.InterfaceToStruct(bkGameDeployment, &b)
		if err != nil {
			blog.Errorf("convert bk gamedeployment failed, err: %s", err.Error())
			return err
		}

		bkGameDeploymentList = append(bkGameDeploymentList, b)
	}

	gameDeploymentToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	gameDeploymentToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
	gameDeploymentToDelete := make([]int64, 0)

	gameDeploymentMap := make(map[string]*storage.GameDeployment)
	for _, v := range gameDeploymentList {
		gameDeploymentMap[v.Data.Namespace+v.Data.Name] = v
	}

	bkGameDeploymentMap := make(map[string]*bkcmdbkube.GameDeployment)
	for k, v := range bkGameDeploymentList {
		bkGameDeploymentMap[v.Namespace+v.Name] = &bkGameDeploymentList[k]

		if _, ok := gameDeploymentMap[v.Namespace+v.Name]; !ok {
			gameDeploymentToDelete = append(gameDeploymentToDelete, v.ID)
		}
	}

	for k, v := range gameDeploymentMap {
		if _, ok := bkGameDeploymentMap[k]; !ok {
			toAddData := s.GenerateBkGameDeployment(bkNamespaceMap[v.Data.Namespace], v)
			if toAddData != nil {
				if _, ok = gameDeploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID]; ok {
					gameDeploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = append(gameDeploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID], *toAddData)
				} else {
					gameDeploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
				}
			}
		} else {
			needToUpdate, updateData := s.CompareGameDeployment(bkGameDeploymentMap[k], v)
			if needToUpdate {
				gameDeploymentToUpdate[bkGameDeploymentMap[k].ID] = updateData
			}
		}
	}

	s.DeleteBkWorkloads(bkCluster, kind, &gameDeploymentToDelete)
	s.CreateBkWorkloads(bkCluster, kind, gameDeploymentToAdd)
	s.UpdateBkWorkloads(bkCluster, kind, &gameDeploymentToUpdate)

	return nil
}

// syncGameStatefulSets sync gamestatefulsets
func (s *Syncer) syncGameStatefulSets(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) error {
	kind := "gameStatefulSet"
	storageCli, err := s.GetBcsStorageClient()
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}

	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	bkNamespaceMap := make(map[string]*bkcmdbkube.Namespace)

	for k, v := range *bkNamespaceList {
		bkNamespaceMap[v.Name] = &(*bkNamespaceList)[k]
	}

	gameStatefulSetList := make([]*storage.GameStatefulSet, 0)
	bkGameStatefulSetList := make([]bkcmdbkube.GameStatefulSet, 0)

	for _, ns := range *bkNamespaceList {
		gameStatefulSets, err := storageCli.QueryK8SGameStatefulSet(cluster.ClusterID, ns.Name)
		if err != nil {
			blog.Errorf("query k8s gameStatefulSets failed, err: %s", err.Error())
			return err
		}
		gameStatefulSetList = append(gameStatefulSetList, gameStatefulSets...)
	}
	blog.Infof("gamestatefulset list: %v", gameStatefulSetList)

	bkGameStatefulSets, err := s.GetBkWorkloads(bkCluster.BizID, "gameStatefulSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk gamestatefulset failed, err: %s", err.Error())
		return err
	}

	for _, bkGameStatefulSet := range *bkGameStatefulSets {
		b := bkcmdbkube.GameStatefulSet{}
		err := common.InterfaceToStruct(bkGameStatefulSet, &b)
		if err != nil {
			blog.Errorf("convert bk gamestatefulset failed, err: %s", err.Error())
			return err
		}

		bkGameStatefulSetList = append(bkGameStatefulSetList, b)
	}

	gameStatefulSetToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	gameStatefulSetToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
	gameStatefulSetToDelete := make([]int64, 0)

	gameStatefulSetMap := make(map[string]*storage.GameStatefulSet)
	for _, v := range gameStatefulSetList {
		gameStatefulSetMap[v.Data.Namespace+v.Data.Name] = v
	}

	bkGameStatefulSetMap := make(map[string]*bkcmdbkube.GameStatefulSet)
	for k, v := range bkGameStatefulSetList {
		bkGameStatefulSetMap[v.Namespace+v.Name] = &bkGameStatefulSetList[k]

		if _, ok := gameStatefulSetMap[v.Namespace+v.Name]; !ok {
			gameStatefulSetToDelete = append(gameStatefulSetToDelete, v.ID)
		}
	}

	for k, v := range gameStatefulSetMap {
		if _, ok := bkGameStatefulSetMap[k]; !ok {
			toAddData := s.GenerateBkGameStatefulSet(bkNamespaceMap[v.Data.Namespace], v)

			if toAddData != nil {
				if _, ok = gameStatefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID]; ok {
					gameStatefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = append(gameStatefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID], *toAddData)
				} else {
					gameStatefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
				}
			}
		} else {
			needToUpdate, updateData := s.CompareGameStatefulSet(bkGameStatefulSetMap[k], v)
			if needToUpdate {
				gameStatefulSetToUpdate[bkGameStatefulSetMap[k].ID] = updateData
			}
		}
	}

	s.DeleteBkWorkloads(bkCluster, kind, &gameStatefulSetToDelete)
	s.CreateBkWorkloads(bkCluster, kind, gameStatefulSetToAdd)
	s.UpdateBkWorkloads(bkCluster, kind, &gameStatefulSetToUpdate)

	return nil
}

// syncWorkloadPods sync workloadPods
func (s *Syncer) syncWorkloadPods(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) error {
	kind := "pods" // nolint
	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	bkNamespaceMap := make(map[string]*bkcmdbkube.Namespace)

	for k, v := range *bkNamespaceList {
		bkNamespaceMap[v.Name] = &(*bkNamespaceList)[k]
	}

	bkWorkloadPodsList := make([]bkcmdbkube.PodsWorkload, 0)

	bkWorkloadPods, err := s.GetBkWorkloads(bkCluster.BizID, "pods", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk workload pods failed, err: %s", err.Error())
		return err
	}

	for _, workloadPods := range *bkWorkloadPods {
		p := bkcmdbkube.PodsWorkload{}
		err := common.InterfaceToStruct(workloadPods, &p)
		if err != nil {
			blog.Errorf("convert bk workload pods failed, err: %s", err.Error())
			return err
		}

		bkWorkloadPodsList = append(bkWorkloadPodsList, p)
	}

	workloadPodsToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	workloadPodsToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
	workloadPodsToDelete := make([]int64, 0)

	bkWorkloadPodsMap := make(map[string]*bkcmdbkube.PodsWorkload)
	for k, v := range bkWorkloadPodsList {
		bkWorkloadPodsMap[v.Namespace+v.Name] = &bkWorkloadPodsList[k]
	}

	for k, v := range bkNamespaceMap {
		if _, ok := bkWorkloadPodsMap[k+"pods"]; !ok {
			toAddData := s.GenerateBkWorkloadPods(v)
			if toAddData != nil {
				if _, ok = workloadPodsToAdd[v.BizID]; ok {
					workloadPodsToAdd[v.BizID] = append(workloadPodsToAdd[v.BizID], *toAddData)
				} else {
					workloadPodsToAdd[v.BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
				}
			}
		} else {
			needToUpdate, updateData := s.CompareBkWorkloadPods(bkWorkloadPodsMap[k+"pods"])
			if needToUpdate {
				workloadPodsToUpdate[bkWorkloadPodsMap[k+"pods"].ID] = updateData
			}
		}
	}

	s.DeleteBkWorkloads(bkCluster, kind, &workloadPodsToDelete)
	s.CreateBkWorkloads(bkCluster, kind, workloadPodsToAdd)
	s.UpdateBkWorkloads(bkCluster, kind, &workloadPodsToUpdate)

	return nil
}

func (s *Syncer) getBkNsMap(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) (map[string]*bkcmdbkube.Namespace, error) {
	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return nil, err
	}

	bkNsMap := make(map[string]*bkcmdbkube.Namespace)
	for k, v := range *bkNamespaceList {
		bkNsMap[v.Name] = &(*bkNamespaceList)[k]
	}

	return bkNsMap, nil
}

func (s *Syncer) getBkDeploymentMap(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) (map[string]*bkcmdbkube.Deployment, error) {
	bkDeploymentList := make([]bkcmdbkube.Deployment, 0)

	bkDeployments, err := s.GetBkWorkloads(bkCluster.BizID, "deployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk deployment failed, err: %s", err.Error())
		return nil, err
	}

	for _, bkDeployment := range *bkDeployments {
		b := bkcmdbkube.Deployment{}
		err := common.InterfaceToStruct(bkDeployment, &b)
		if err != nil {
			blog.Errorf("convert bk deployment failed, err: %s", err.Error())
			return nil, err
		}

		bkDeploymentList = append(bkDeploymentList, b)
	}

	bkDeploymentMap := make(map[string]*bkcmdbkube.Deployment)
	for k, v := range bkDeploymentList {
		bkDeploymentMap[v.Namespace+v.Name] = &bkDeploymentList[k]
	}

	return bkDeploymentMap, nil
}

func (s *Syncer) getBkStatefulSetMap(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) (map[string]*bkcmdbkube.StatefulSet, error) {
	bkStatefulSetList := make([]bkcmdbkube.StatefulSet, 0)

	bkStatefulSets, err := s.GetBkWorkloads(bkCluster.BizID, "statefulSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk statefulset failed, err: %s", err.Error())
		return nil, err
	}

	for _, bkStatefulSet := range *bkStatefulSets {
		b := bkcmdbkube.StatefulSet{}
		err := common.InterfaceToStruct(bkStatefulSet, &b)
		if err != nil {
			blog.Errorf("convert bk statefulset failed, err: %s", err.Error())
			return nil, err
		}

		bkStatefulSetList = append(bkStatefulSetList, b)
	}

	bkStatefulSetMap := make(map[string]*bkcmdbkube.StatefulSet)
	for k, v := range bkStatefulSetList {
		bkStatefulSetMap[v.Namespace+v.Name] = &bkStatefulSetList[k]
	}
	return bkStatefulSetMap, nil
}

func (s *Syncer) getBkDaemonSetMap(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) (map[string]*bkcmdbkube.DaemonSet, error) {
	bkDaemonSetList := make([]bkcmdbkube.DaemonSet, 0)

	bkDaemonSets, err := s.GetBkWorkloads(bkCluster.BizID, "daemonSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk daemonset failed, err: %s", err.Error())
		return nil, err
	}

	for _, bkDaemonSet := range *bkDaemonSets {
		b := bkcmdbkube.DaemonSet{}
		err := common.InterfaceToStruct(bkDaemonSet, &b)
		if err != nil {
			blog.Errorf("convert bk daemonset failed, err: %s", err.Error())
			return nil, err
		}

		bkDaemonSetList = append(bkDaemonSetList, b)
	}

	bkDaemonSetMap := make(map[string]*bkcmdbkube.DaemonSet)
	for k, v := range bkDaemonSetList {
		bkDaemonSetMap[v.Namespace+v.Name] = &bkDaemonSetList[k]
	}
	return bkDaemonSetMap, nil
}

func (s *Syncer) getBkGameDeploymentMap(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) (map[string]*bkcmdbkube.GameDeployment, error) {
	bkGameDeploymentList := make([]bkcmdbkube.GameDeployment, 0)

	bkGameDeployments, err := s.GetBkWorkloads(bkCluster.BizID, "gameDeployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk gamedeployment failed, err: %s", err.Error())
		return nil, err
	}

	for _, bkGameDeployment := range *bkGameDeployments {
		b := bkcmdbkube.GameDeployment{}
		err := common.InterfaceToStruct(bkGameDeployment, &b)
		if err != nil {
			blog.Errorf("convert bk gamedeployment failed, err: %s", err.Error())
			return nil, err
		}

		bkGameDeploymentList = append(bkGameDeploymentList, b)
	}

	bkGameDeploymentMap := make(map[string]*bkcmdbkube.GameDeployment)
	for k, v := range bkGameDeploymentList {
		bkGameDeploymentMap[v.Namespace+v.Name] = &bkGameDeploymentList[k]
	}
	return bkGameDeploymentMap, nil
}

func (s *Syncer) getBkGameStatefulSetMap(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) (map[string]*bkcmdbkube.GameStatefulSet, error) {
	bkGameStatefulSetList := make([]bkcmdbkube.GameStatefulSet, 0)

	bkGameStatefulSets, err := s.GetBkWorkloads(bkCluster.BizID, "gameStatefulSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk gamestatefulset failed, err: %s", err.Error())
		return nil, err
	}

	for _, bkGameStatefulSet := range *bkGameStatefulSets {
		b := bkcmdbkube.GameStatefulSet{}
		err := common.InterfaceToStruct(bkGameStatefulSet, &b)
		if err != nil {
			blog.Errorf("convert bk gamestatefulset failed, err: %s", err.Error())
			return nil, err
		}

		bkGameStatefulSetList = append(bkGameStatefulSetList, b)
	}

	bkGameStatefulSetMap := make(map[string]*bkcmdbkube.GameStatefulSet)
	for k, v := range bkGameStatefulSetList {
		bkGameStatefulSetMap[v.Namespace+v.Name] = &bkGameStatefulSetList[k]
	}
	return bkGameStatefulSetMap, nil
}

func (s *Syncer) getBkNodeMap(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) (map[string]*bkcmdbkube.Node, error) {
	bkNodeList, err := s.GetBkNodes(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk node failed, err: %s", err.Error())
		return nil, err
	}

	bkNodeMap := make(map[string]*bkcmdbkube.Node)
	for k, v := range *bkNodeList {
		bkNodeMap[*v.Name] = &(*bkNodeList)[k]
	}
	return bkNodeMap, nil
}

func (s *Syncer) getBkWorkloadPods(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, pod *storage.Pod) (*bkcmdbkube.PodsWorkload, error) {
	bkWorkloadPods, err := s.GetBkWorkloads(bkCluster.BizID, "pods", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "namespace",
				Operator: "in",
				Value:    []string{pod.Data.Namespace},
			},
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{pod.ClusterID},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk workload pods failed, err: %s", err.Error())
		return nil, err
	}

	if len(*bkWorkloadPods) != 1 {
		blog.Errorf("get bk workload pods len is %d", len(*bkWorkloadPods))
		return nil, err
	}

	p := bkcmdbkube.PodsWorkload{}
	err = common.InterfaceToStruct((*bkWorkloadPods)[0], &p)
	if err != nil {
		blog.Errorf("convert bk workload pods failed, err: %s", err.Error())
		return nil, err
	}

	return &p, nil
}

// nolint
func (s *Syncer) getPodOperator(cluster *cmp.Cluster, workloadLabels, nsLabels *map[string]string) []string {
	var operator []string
	if workloadLabels != nil {
		if creator, creatorOk := (*workloadLabels)["io.tencent.paas.creator"]; creatorOk && (creator != "") {
			operator = append(operator, creator)
		} else if creator, creatorOk = (*workloadLabels)["io．tencent．paas．creator"]; creatorOk && (creator != "") {
			operator = append(operator, creator)
		} else if updater, updaterOk := (*workloadLabels)["io.tencent.paas.updater"]; updaterOk && (updater != "") {
			operator = append(operator, updater)
		} else if updater, updaterOk = (*workloadLabels)["io．tencent．paas．updator"]; updaterOk && (updater != "") {
			operator = append(operator, updater)
		}
	}

	if len(operator) == 0 && (nsLabels != nil) {
		if creator, creatorOk := (*nsLabels)["io.tencent.paas.creator"]; creatorOk && (creator != "") {
			operator = append(operator, creator)
		} else if creator, creatorOk = (*nsLabels)["io．tencent．paas．creator"]; creatorOk && (creator != "") {
			operator = append(operator, creator)
		} else if updater, updaterOk := (*nsLabels)["io.tencent.paas.updater"]; updaterOk && (updater != "") {
			operator = append(operator, updater)
		} else if updater, updaterOk = (*nsLabels)["io．tencent．paas．updator"]; updaterOk && (updater != "") {
			operator = append(operator, updater)
		}
	}

	if len(operator) == 0 {
		if cluster.Creator != "" {
			operator = append(operator, cluster.Creator)
		} else if cluster.Updater != "" {
			operator = append(operator, cluster.Updater)
		}
	}

	if len(operator) == 0 {
		operator = append(operator, "")
	}
	return operator
}

func (s *Syncer) getPodWordloadInfo(
	cluster *cmp.Cluster,
	pod *storage.Pod,
	bkWorkloadPods *bkcmdbkube.PodsWorkload,
	bkGameDeploymentMap map[string]*bkcmdbkube.GameDeployment,
	bkGameStatefulSetMap map[string]*bkcmdbkube.GameStatefulSet,
	bkStatefulSetMap map[string]*bkcmdbkube.StatefulSet,
	bkDaemonSetMap map[string]*bkcmdbkube.DaemonSet,
	storageCli bcsapi.Storage,
	bkDeploymentMap map[string]*bkcmdbkube.Deployment) (
	workloadKind, workloadName string, workloadID int64, labels *map[string]string) {

	workloadKind = "pods" // nolint
	workloadName = "pods" // nolint
	workloadID = bkWorkloadPods.ID

	if len(pod.Data.OwnerReferences) == 1 {
		ownerRef := pod.Data.OwnerReferences[0]
		switch ownerRef.Kind {
		case "GameDeployment":
			workloadKind = "gameDeployment"
			workloadName = ownerRef.Name
			if _, ok := bkGameDeploymentMap[pod.Data.Namespace+workloadName]; ok {
				workloadID = bkGameDeploymentMap[pod.Data.Namespace+workloadName].ID
				labels = bkGameDeploymentMap[pod.Data.Namespace+workloadName].Labels
			}
		case "GameStatefulSet":
			workloadKind = "gameStatefulSet"
			workloadName = ownerRef.Name
			if _, ok := bkGameStatefulSetMap[pod.Data.Namespace+workloadName]; ok {
				workloadID = bkGameStatefulSetMap[pod.Data.Namespace+workloadName].ID
				labels = bkGameStatefulSetMap[pod.Data.Namespace+workloadName].Labels
			}
		case "StatefulSet":
			workloadKind = "statefulSet"
			workloadName = ownerRef.Name

			if _, ok := bkStatefulSetMap[pod.Data.Namespace+workloadName]; ok {
				workloadID = bkStatefulSetMap[pod.Data.Namespace+workloadName].ID
				labels = bkStatefulSetMap[pod.Data.Namespace+workloadName].Labels
			}
		case "DaemonSet":
			workloadKind = "daemonSet"
			workloadName = ownerRef.Name

			if _, ok := bkDaemonSetMap[pod.Data.Namespace+workloadName]; ok {
				workloadID = bkDaemonSetMap[pod.Data.Namespace+workloadName].ID
				labels = bkDaemonSetMap[pod.Data.Namespace+workloadName].Labels
			}
		case "ReplicaSet":
			rsList, err := storageCli.QueryK8sReplicaSet(cluster.ClusterID, pod.Data.Namespace, ownerRef.Name)
			if err != nil {
				blog.Errorf("query replicaSet %s failed, err: %s", ownerRef.Name, err.Error())
				break
			}
			if len(rsList) != 1 {
				blog.Errorf("replicaSet %s not found", ownerRef.Name)
				break
			}
			rs := rsList[0]

			if len(rs.Data.OwnerReferences) == 0 {
				break
			}
			rsOwnerRef := rs.Data.OwnerReferences[0]
			switch rsOwnerRef.Kind {
			case "Deployment":
				workloadKind = "deployment"
				workloadName = rsOwnerRef.Name
				if _, ok := bkDeploymentMap[rs.Namespace+workloadName]; ok {
					workloadID = bkDeploymentMap[rs.Namespace+workloadName].ID
					labels = bkDeploymentMap[pod.Data.Namespace+workloadName].Labels
				}
			}
		}
	}

	return workloadKind, workloadName, workloadID, labels
}

// SyncPods sync pods
func (s *Syncer) SyncPods(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster) error {
	storageCli, err := s.GetBcsStorageClient()
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}

	bkNsMap, err := s.getBkNsMap(cluster, bkCluster)
	if err != nil {
		return err
	}

	podList := make([]*storage.Pod, 0)

	for _, ns := range bkNsMap {
		pods, err := storageCli.QueryK8SPod(cluster.ClusterID, ns.Name)
		if err != nil {
			blog.Errorf("query k8s pod failed, err: %s", err.Error())
			return err
		}
		podList = append(podList, pods...)
	}

	bkDeploymentMap, err := s.getBkDeploymentMap(cluster, bkCluster)
	if err != nil {
		return err
	}

	bkStatefulSetMap, err := s.getBkStatefulSetMap(cluster, bkCluster)
	if err != nil {
		return err
	}

	bkDaemonSetMap, err := s.getBkDaemonSetMap(cluster, bkCluster)
	if err != nil {
		return err
	}

	bkGameDeploymentMap, err := s.getBkGameDeploymentMap(cluster, bkCluster)
	if err != nil {
		return err
	}

	bkGameStatefulSetMap, err := s.getBkGameStatefulSetMap(cluster, bkCluster)
	if err != nil {
		return err
	}

	bkNodeMap, err := s.getBkNodeMap(cluster, bkCluster)
	if err != nil {
		return err
	}

	blog.Infof("query k8s pod success, pods: %v", podList)

	bkPodList, err := s.GetBkPods(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	})
	if err != nil {
		blog.Errorf("get bk pod failed, err: %s", err.Error())
		return err
	}

	podToAdd := make(map[int64][]client.CreateBcsPodRequestDataPod, 0)
	podToDelete := make([]int64, 0)

	podMap := make(map[string]*storage.Pod)
	for _, v := range podList {
		podMap[v.Data.Namespace+v.Data.Name] = v
	}

	bkPodMap := make(map[string]*bkcmdbkube.Pod)
	for k, v := range *bkPodList {
		bkPodMap[v.NameSpace+*v.Name] = &(*bkPodList)[k]

		if _, ok := podMap[v.NameSpace+*v.Name]; !ok {
			podToDelete = append(podToDelete, v.ID)
		}
	}

	for k, v := range podMap {
		//var operator []string
		if _, ok := bkPodMap[k]; !ok {
			bkWorkloadPods, podsErr := s.getBkWorkloadPods(cluster, bkCluster, v)
			if podsErr != nil {
				continue
			}

			workloadKind, workloadName, workloadID, workloadLabels := s.getPodWordloadInfo(
				cluster, v, bkWorkloadPods, bkGameDeploymentMap, bkGameStatefulSetMap,
				bkStatefulSetMap, bkDaemonSetMap, storageCli, bkDeploymentMap)

			operator := s.getPodOperator(cluster, workloadLabels, bkNsMap[v.Data.Namespace].Labels)

			var nodeID int64
			if _, okk := bkNodeMap[v.Data.Spec.NodeName]; !okk {
				continue
			}

			nodeID = bkNodeMap[v.Data.Spec.NodeName].ID
			hostID = bkNodeMap[v.Data.Spec.NodeName].HostID

			podIPs := make([]bkcmdbkube.PodIP, 0)
			for _, ip := range v.Data.Status.PodIPs {
				podIPs = append(podIPs, bkcmdbkube.PodIP{
					IP: ip.IP,
				})
			}

			containerStatusMap := make(map[string]corev1.ContainerStatus)

			for _, containerStatus := range v.Data.Status.ContainerStatuses {
				containerStatusMap[containerStatus.Name] = containerStatus
			}

			containers := make([]bkcmdbkube.ContainerBaseFields, 0)
			for _, container := range v.Data.Spec.Containers {

				ports := make([]bkcmdbkube.ContainerPort, 0)

				for _, port := range container.Ports {
					ports = append(ports, bkcmdbkube.ContainerPort{
						Name:          port.Name,
						HostPort:      port.HostPort,
						ContainerPort: port.ContainerPort,
						Protocol:      bkcmdbkube.Protocol(port.Protocol),
						HostIP:        port.HostIP,
					})
				}

				env := make([]bkcmdbkube.EnvVar, 0)

				for _, envVar := range container.Env {
					env = append(env, bkcmdbkube.EnvVar{
						Name:  envVar.Name,
						Value: envVar.Value,
					})
				}

				mounts := make([]bkcmdbkube.VolumeMount, 0)

				for _, mount := range container.VolumeMounts {
					mounts = append(mounts, bkcmdbkube.VolumeMount{
						Name:        mount.Name,
						MountPath:   mount.MountPath,
						SubPath:     mount.SubPath,
						ReadOnly:    mount.ReadOnly,
						SubPathExpr: mount.SubPathExpr,
					})
				}

				containerID := containerStatusMap[container.Name].ContainerID

				if containerID == "" {
					continue
				}

				containers = append(containers, bkcmdbkube.ContainerBaseFields{
					Name:        &container.Name,
					Image:       &container.Image,
					ContainerID: &containerID,
					Ports:       &ports,
					Args:        &container.Args,
					Environment: &env,
					Mounts:      &mounts,
				})
			}

			if _, ok = podToAdd[bkNsMap[v.Data.Namespace].BizID]; ok {
				podToAdd[bkNsMap[v.Data.Namespace].BizID] = append(podToAdd[bkNsMap[v.Data.Namespace].BizID], client.CreateBcsPodRequestDataPod{
					Spec: &client.CreateBcsPodRequestPodSpec{
						ClusterID:    &bkCluster.ID,
						NameSpaceID:  &bkNsMap[v.Data.Namespace].ID,
						WorkloadKind: &workloadKind,
						WorkloadID:   &workloadID,
						NodeID:       &nodeID,
						Ref: &bkcmdbkube.Reference{
							Kind: bkcmdbkube.WorkloadType(workloadKind),
							Name: workloadName,
							ID:   workloadID,
						},
					},

					Name:       &v.Data.Name,
					HostID:     &hostID,
					Priority:   v.Data.Spec.Priority,
					Labels:     &v.Data.Labels,
					IP:         &v.Data.Status.PodIP,
					IPs:        &podIPs,
					Containers: &containers,
					Operator:   &operator,
				})
			} else {
				podToAdd[bkNsMap[v.Data.Namespace].BizID] = []client.CreateBcsPodRequestDataPod{
					client.CreateBcsPodRequestDataPod{
						Spec: &client.CreateBcsPodRequestPodSpec{
							ClusterID:    &bkCluster.ID,
							NameSpaceID:  &bkNsMap[v.Data.Namespace].ID,
							WorkloadKind: &workloadKind,
							WorkloadID:   &workloadID,
							NodeID:       &nodeID,
							Ref: &bkcmdbkube.Reference{
								Kind: bkcmdbkube.WorkloadType(workloadKind),
								Name: workloadName,
								ID:   workloadID,
							},
						},

						Name:       &v.Data.Name,
						HostID:     &hostID,
						Priority:   v.Data.Spec.Priority,
						Labels:     &v.Data.Labels,
						IP:         &v.Data.Status.PodIP,
						IPs:        &podIPs,
						Containers: &containers,
						Operator:   &operator,
					},
				}
			}
		}
	}

	s.CreateBkPods(bkCluster, podToAdd)
	s.DeleteBkPods(bkCluster, &podToDelete)

	return err
}

// GetBcsStorageClient is a function that returns a BCS storage client.
func (s *Syncer) GetBcsStorageClient() (bcsapi.Storage, error) {
	// Create a BCS API configuration with the given options.
	config := &bcsapi.Config{
		Hosts:     []string{s.BkcmdbSynchronizerOption.Bcsapi.HttpAddr},
		AuthToken: s.BkcmdbSynchronizerOption.Bcsapi.BearerToken,
		TLSConfig: s.ClientTls,
		Gateway:   true,
	}

	// Create a new BCS storage client with the configuration.
	cli := bsc.NewStorageClient(config)

	// Get the storage client from the BCS storage client.
	storageCli, err := cli.GetStorageClient()

	// Return the storage client and any error that occurred.
	return storageCli, err
}

// GetProjectManagerGrpcGwClient is a function that returns a project manager gRPC gateway client.
func (s *Syncer) GetProjectManagerGrpcGwClient() (pmCli *client.ProjectManagerClientWithHeader, err error) {
	// Create a project manager gRPC gateway client configuration with the given options.
	opts := &pm.Options{
		Module:          pm.ModuleProjectManager,
		Address:         s.BkcmdbSynchronizerOption.Bcsapi.GrpcAddr,
		EtcdRegistry:    nil,
		ClientTLSConfig: s.ClientTls,
		AuthToken:       s.BkcmdbSynchronizerOption.Bcsapi.BearerToken,
	}

	// Create a new project manager gRPC gateway client with the configuration.
	pmCli, err = pm.NewProjectManagerGrpcGwClient(opts)

	// Return the project manager gRPC gateway client and any error that occurred.
	return pmCli, err
}

// GetBkCluster get bkcluster
func (s *Syncer) GetBkCluster(cluster *cmp.Cluster) (*bkcmdbkube.Cluster, error) {
	var clusterBkBizID int64
	if bkBizID == 0 {
		bizid, err := strconv.ParseInt(cluster.BusinessID, 10, 64)
		if err != nil {
			blog.Errorf("An error occurred: %s\n", err)
		} else {
			blog.Infof("Successfully converted string to int64: %d\n", clusterBkBizID)
		}
		clusterBkBizID = bizid
	} else {
		clusterBkBizID = bkBizID
	}

	bkClusterList, err := s.CMDBClient.GetBcsCluster(&client.GetBcsClusterRequest{
		CommonRequest: client.CommonRequest{
			BKBizID: clusterBkBizID,
			Page: client.Page{
				Limit: 10,
				Start: 0,
			},
			Fields: []string{},
			Filter: &client.PropertyFilter{
				Condition: "AND",
				Rules: []client.Rule{
					{
						Field:    "uid",
						Operator: "in",
						Value:    []string{cluster.ClusterID},
					},
				},
			},
		},
	})

	if err != nil {
		blog.Errorf("get bcs cluster failed, err: %s", err.Error())
		return nil, err
	}

	if len(*bkClusterList) == 0 {
		return nil, fmt.Errorf("cluster not found")
	}

	return &(*bkClusterList)[0], nil
}

// GetBkNodes get bknodes
func (s *Syncer) GetBkNodes(bkBizID int64, filter *client.PropertyFilter) (*[]bkcmdbkube.Node, error) {
	bkNodeList := make([]bkcmdbkube.Node, 0)

	pageStart := 0
	for {
		bkNodes, err := s.CMDBClient.GetBcsNode(&client.GetBcsNodeRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkBizID,
				Page: client.Page{
					Limit: 100,
					Start: 100 * pageStart,
				},
				Fields: []string{},
				Filter: filter,
			},
		})

		if err != nil {
			blog.Errorf("get bcs node failed, err: %s", err.Error())
			return nil, err
		}
		bkNodeList = append(bkNodeList, *bkNodes...)

		if len(*bkNodes) < 100 {
			break
		}
		pageStart++
	}

	return &bkNodeList, nil
}

// CreateBkNodes create bknodes
func (s *Syncer) CreateBkNodes(bkCluster *bkcmdbkube.Cluster, toCreate *[]client.CreateBcsNodeRequestData) {
	if len(*toCreate) > 0 {
		_, err := s.CMDBClient.CreateBcsNode(&client.CreateBcsNodeRequest{
			BKBizID: &bkCluster.BizID,
			Data:    toCreate,
		})
		if err != nil {
			for i := 0; i < len(*toCreate); i++ {
				var section []client.CreateBcsNodeRequestData
				if i+1 > len(*toCreate) {
					section = (*toCreate)[i:]
				} else {
					section = (*toCreate)[i : i+1]
				}

				_, err = s.CMDBClient.CreateBcsNode(&client.CreateBcsNodeRequest{
					BKBizID: &bkCluster.BizID,
					Data:    &section,
				})

				if err != nil {
					blog.Errorf("create node failed, err: %s", err.Error())
				} else {
					blog.Infof("create node success, %v", section)
				}
			}
		} else {
			blog.Infof("create node to cmdb success, nodes: %v", toCreate)
		}
	}
}

// CompareNode compare bknode and k8snode
func (s *Syncer) CompareNode(bkNode *bkcmdbkube.Node, k8sNode *storage.K8sNode) (needToUpdate bool, updateData *client.UpdateBcsNodeRequestData) {
	updateData = &client.UpdateBcsNodeRequestData{}
	needToUpdate = false
	labelsEmpty := map[string]string{}

	//var updateDataIDs []int64
	//updateData.IDs = &updateDataIDs
	//var updateDataNode client.UpdateBcsNodeRequestDataNode
	//updateData.Node = &updateDataNode

	taints := make(map[string]string)
	for _, taint := range k8sNode.Data.Spec.Taints {
		taints[taint.Key] = taint.Value
	}

	if taints == nil {
		if bkNode.Taints != nil {
			updateData.Taints = &taints
			needToUpdate = true
		}
	} else if bkNode.Taints == nil || fmt.Sprint(*bkNode.Taints) != fmt.Sprint(taints) {
		updateData.Taints = &taints
		needToUpdate = true
	}

	if k8sNode.Data.Labels == nil {
		if bkNode.Labels != nil {
			updateData.Labels = &labelsEmpty
			needToUpdate = true
		}
	} else if bkNode.Labels == nil || fmt.Sprint(*bkNode.Labels) != fmt.Sprint(k8sNode.Data.Labels) {
		updateData.Labels = &k8sNode.Data.Labels
		needToUpdate = true
	}

	if bkNode.Unschedulable == nil {
		updateData.Unschedulable = &k8sNode.Data.Spec.Unschedulable
		needToUpdate = true
	} else {
		if *bkNode.Unschedulable != k8sNode.Data.Spec.Unschedulable {
			updateData.Unschedulable = &k8sNode.Data.Spec.Unschedulable
			needToUpdate = true
		}
	}

	if bkNode.PodCidr == nil {
		updateData.PodCidr = &k8sNode.Data.Spec.PodCIDR
		needToUpdate = true
	} else {
		if *bkNode.PodCidr != k8sNode.Data.Spec.PodCIDR {
			updateData.PodCidr = &k8sNode.Data.Spec.PodCIDR
			needToUpdate = true
		}
	}

	if *bkNode.RuntimeComponent != k8sNode.Data.Status.NodeInfo.ContainerRuntimeVersion {
		updateData.RuntimeComponent = &k8sNode.Data.Status.NodeInfo.ContainerRuntimeVersion
		needToUpdate = true
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// GenerateBkNodeData generate bknode data from k8snode
func (s *Syncer) GenerateBkNodeData(bkCluster *bkcmdbkube.Cluster, k8sNode *storage.K8sNode) client.CreateBcsNodeRequestData {
	taints := make(map[string]string)
	for _, taint := range k8sNode.Data.Spec.Taints {
		taints[taint.Key] = taint.Value
	}

	internalIP := make([]string, 0)
	externalIP := make([]string, 0)
	hostName := k8sNode.Data.Name
	for _, address := range k8sNode.Data.Status.Addresses {
		if address.Type == "InternalIP" {
			internalIP = append(internalIP, address.Address)
		}
		if address.Type == "ExternalIP" {
			externalIP = append(externalIP, address.Address)
		}
		if address.Type == "Hostname" {
			hostName = address.Address
		}
	}
	var theHostID int64
	if hostID == 0 {
		hosts, err := s.CMDBClient.GetHostInfo(internalIP)
		if err != nil {
			return client.CreateBcsNodeRequestData{}
		}
		if len(*hosts) > 0 {
			theHostID = (*hosts)[0].BkHostId
		}
	} else {
		theHostID = hostID
	}

	return client.CreateBcsNodeRequestData{
		HostID:           &theHostID,
		ClusterID:        &bkCluster.ID,
		Name:             &k8sNode.Data.Name,
		Labels:           &k8sNode.Data.Labels,
		Taints:           &taints,
		Unschedulable:    &k8sNode.Data.Spec.Unschedulable,
		InternalIP:       &internalIP,
		ExternalIP:       &externalIP,
		HostName:         &hostName,
		RuntimeComponent: &k8sNode.Data.Status.NodeInfo.ContainerRuntimeVersion,
		KubeProxyMode:    nil,
		PodCidr:          &k8sNode.Data.Spec.PodCIDR,
	}
}

// UpdateBkNodes update bknodes
// nolint
func (s *Syncer) UpdateBkNodes(bkCluster *bkcmdbkube.Cluster, toUpdate *map[int64]*client.UpdateBcsNodeRequestData) {
	if toUpdate == nil {
		return
	}

	for k, v := range *toUpdate {
		err := s.CMDBClient.UpdateBcsNode(&client.UpdateBcsNodeRequest{
			BKBizID: &bkCluster.BizID,
			IDs:     &[]int64{k},
			Data:    v,
		})

		if err != nil {
			blog.Errorf("update node failed, err: %s", err.Error())
		}
	}
}

// DeleteBkNodes delete bknodes
func (s *Syncer) DeleteBkNodes(bkCluster *bkcmdbkube.Cluster, toDelete *[]int64) error {
	if len(*toDelete) > 0 {
		err := s.CMDBClient.DeleteBcsNode(&client.DeleteBcsNodeRequest{
			BKBizID: &bkCluster.BizID,
			IDs:     toDelete,
		})
		if err != nil {
			for i := 0; i < len(*toDelete); i++ {
				var section []int64
				if i+1 > len(*toDelete) {
					section = (*toDelete)[i:]
				} else {
					section = (*toDelete)[i : i+1]
				}

				err = s.CMDBClient.DeleteBcsNode(&client.DeleteBcsNodeRequest{
					BKBizID: &bkCluster.BizID,
					IDs:     &section,
				})

				if err != nil {
					blog.Errorf("delete node failed, err: %s", err.Error())
					return err
				}
			}
		}
		blog.Infof("delete node from cmdb success, ids: %v", toDelete)
	}
	return nil
}

// GetBkNamespaces get bknamespaces
func (s *Syncer) GetBkNamespaces(bkBizID int64, filter *client.PropertyFilter) (*[]bkcmdbkube.Namespace, error) {
	bkNamespaceList := make([]bkcmdbkube.Namespace, 0)

	pageStart := 0
	for {
		bkNamespaces, err := s.CMDBClient.GetBcsNamespace(&client.GetBcsNamespaceRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkBizID,
				Page: client.Page{
					Limit: 100,
					Start: 100 * pageStart,
				},
				Fields: []string{},
				Filter: filter,
			},
		})

		if err != nil {
			blog.Errorf("get bcs namespace failed, err: %s", err.Error())
			return nil, err
		}
		bkNamespaceList = append(bkNamespaceList, *bkNamespaces...)

		if len(*bkNamespaces) < 100 {
			break
		}
		pageStart++
	}

	return &bkNamespaceList, nil
}

// CompareNamespace compare bkns and k8sns
func (s *Syncer) CompareNamespace(bkNs *bkcmdbkube.Namespace, k8sNs *storage.Namespace) (needToUpdate bool, updateData *client.UpdateBcsNamespaceRequestData) {
	updateData = &client.UpdateBcsNamespaceRequestData{}
	needToUpdate = false
	labelsEmpty := map[string]string{}

	//var updateDataIDs []int64
	//updateData.IDs = &updateDataIDs
	//var updateDataInfo client.UpdateBcsNamespaceRequestDataInfo
	//updateData.Info = &updateDataInfo

	if k8sNs.Data.Labels == nil {
		if bkNs.Labels != nil {
			updateData.Labels = &labelsEmpty
			needToUpdate = true
		}
	} else if bkNs.Labels == nil || fmt.Sprint(*bkNs.Labels) != fmt.Sprint(k8sNs.Data.Labels) {
		updateData.Labels = &k8sNs.Data.Labels
		needToUpdate = true
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// GenerateBkNsData generate bknsdata from k8sns
func (s *Syncer) GenerateBkNsData(bkCluster *bkcmdbkube.Cluster, k8sNs *storage.Namespace) bkcmdbkube.Namespace {
	var labels *map[string]string
	if k8sNs.Data.Labels == nil {
		labels = nil
	} else {
		labels = &k8sNs.Data.Labels
	}

	return bkcmdbkube.Namespace{
		ClusterSpec: bkcmdbkube.ClusterSpec{
			ClusterID: bkCluster.ID,
		},
		Name:   k8sNs.Data.Name,
		Labels: labels,
		//ResourceQuotas:
	}
}

// CreateBkNamespaces create bknamespaces
func (s *Syncer) CreateBkNamespaces(bkCluster *bkcmdbkube.Cluster, toCreate map[int64][]bkcmdbkube.Namespace) {
	if len(toCreate) > 0 {
		for bizid, bkNsList := range toCreate {
			if len(bkNsList) > 0 {
				for i := 0; i < len(bkNsList); i++ {
					var section []bkcmdbkube.Namespace
					if i+1 > len(bkNsList) {
						section = (bkNsList)[i:]
					} else {
						section = (bkNsList)[i : i+1]
					}

					bkNsIDs, err := s.CMDBClient.CreateBcsNamespace(&client.CreateBcsNamespaceRequest{
						BKBizID: &bizid,
						Data:    &section,
					})

					if err != nil {
						blog.Errorf("create namespace failed, err: %s", err.Error())
						return
					}

					podsKind := "pods" // nolint
					podsName := "pods" // nolint
					_, err = s.CMDBClient.CreateBcsWorkload(&client.CreateBcsWorkloadRequest{
						BKBizID: &bizid,
						Kind:    &podsKind,
						Data: &[]client.CreateBcsWorkloadRequestData{
							{
								NamespaceID: &(*bkNsIDs)[0],
								Name:        &podsName,
							},
						},
					})

					if err != nil {
						blog.Errorf("create workload pods failed, err: %s", err.Error())
					}
				}
			}
		}
	}
}

// UpdateBkNamespaces update bknamespaces
// nolint
func (s *Syncer) UpdateBkNamespaces(bkCluster *bkcmdbkube.Cluster, toUpdate *map[int64]*client.UpdateBcsNamespaceRequestData) {
	if toUpdate == nil {
		return
	}

	for k, v := range *toUpdate {
		err := s.CMDBClient.UpdateBcsNamespace(&client.UpdateBcsNamespaceRequest{
			BKBizID: &bkCluster.BizID,
			IDs:     &[]int64{k},
			Data:    v,
		})

		if err != nil {
			blog.Errorf("update namespace failed, err: %s", err.Error())
		}
	}
}

// DeleteBkNamespaces delete bknamespaces
func (s *Syncer) DeleteBkNamespaces(bkCluster *bkcmdbkube.Cluster, toDelete *[]int64) error {
	if len(*toDelete) > 0 {
		for i := 0; i < len(*toDelete); i++ {
			var section []int64
			if i+1 > len(*toDelete) {
				section = (*toDelete)[i:]
			} else {
				section = (*toDelete)[i : i+1]
			}

			bkWorkloadPods, err := s.GetBkWorkloads(bkCluster.BizID, "pods", &client.PropertyFilter{
				Condition: "AND",
				Rules: []client.Rule{
					{
						Field:    "bk_namespace_id",
						Operator: "in",
						Value:    section,
					},
				},
			})
			if err != nil {
				blog.Errorf("get bk workload pods failed, err: %s", err.Error())
				return errors.New("get bk workload pods failed")
			}

			if len(*bkWorkloadPods) != 1 {
				blog.Errorf("get bk workload pods len is %d", len(*bkWorkloadPods))
				return errors.New("len(bkWorkloadPods) should be 1")
			}

			p := bkcmdbkube.PodsWorkload{}
			err = common.InterfaceToStruct((*bkWorkloadPods)[0], &p)
			if err != nil {
				blog.Errorf("convert bk workload pods failed, err: %s", err.Error())
				return err
			}

			podsKind := "pods" // nolint

			err = s.CMDBClient.DeleteBcsWorkload(&client.DeleteBcsWorkloadRequest{
				BKBizID: &bkCluster.BizID,
				Kind:    &podsKind,
				IDs:     &[]int64{p.ID},
			})

			if err != nil {
				blog.Errorf("delete bk workload pods failed, err: %s", err.Error())
				return err
			}

			err = s.CMDBClient.DeleteBcsNamespace(&client.DeleteBcsNamespaceRequest{
				BKBizID: &bkCluster.BizID,
				IDs:     &section,
			})

			if err != nil {
				blog.Errorf("delete namespace failed, err: %s", err.Error())
			}
		}
	}
	return nil
}

// GetBkWorkloads get bkworkloads
func (s *Syncer) GetBkWorkloads(bkBizID int64, workloadType string, filter *client.PropertyFilter) (*[]interface{}, error) {
	bkWorkloadList := make([]interface{}, 0)

	pageStart := 0
	for {
		bkWorkloads, err := s.CMDBClient.GetBcsWorkload(&client.GetBcsWorkloadRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkBizID,
				Page: client.Page{
					Limit: 100,
					Start: 100 * pageStart,
				},
				Filter: filter,
				//Filter: &client.PropertyFilter{
				//	Condition: "OR",
				//	Rules: []client.Rule{
				//		{
				//			Field:    "cluster_uid",
				//			Operator: "in",
				//			Value:    []string{bkCluster.Uid},
				//		},
				//	},
				//},
			},
			//ClusterUID: bkCluster.Uid,
			Kind: workloadType,
		})

		if err != nil {
			blog.Errorf("get bcs workload failed, err: %s", err.Error())
			return nil, err
		}
		bkWorkloadList = append(bkWorkloadList, *bkWorkloads...)

		if len(*bkWorkloads) < 100 {
			break
		}
		pageStart++
	}

	return &bkWorkloadList, nil
}

// CreateBkWorkloads create bkworkloads
func (s *Syncer) CreateBkWorkloads(bkCluster *bkcmdbkube.Cluster, kind string, toCreate map[int64][]client.CreateBcsWorkloadRequestData) {
	if len(toCreate) > 0 {
		for bizid, workloads := range toCreate {
			if len(workloads) > 0 {
				_, err := s.CMDBClient.CreateBcsWorkload(&client.CreateBcsWorkloadRequest{
					BKBizID: &bizid,
					Kind:    &kind,
					Data:    &workloads,
				})
				if err != nil {
					for i := 0; i < len(workloads); i++ {
						var section []client.CreateBcsWorkloadRequestData
						if i+1 > len(workloads) {
							section = (workloads)[i:]
						} else {
							section = (workloads)[i : i+1]
						}
						_, err = s.CMDBClient.CreateBcsWorkload(&client.CreateBcsWorkloadRequest{
							BKBizID: &bkCluster.BizID,
							Kind:    &kind,
							Data:    &section,
						})

						if err != nil {
							blog.Errorf("create workload %s failed, err: %s", kind, err.Error())
						} else {
							blog.Infof("create workload %s success, %v", kind, section)
						}
					}
				} else {
					blog.Infof("create workload %s success, %v", kind, toCreate)
				}
			}
		}
	}
}

// CompareDeployment compare bkdeployment and k8sdeployment
func (s *Syncer) CompareDeployment(bkDeployment *bkcmdbkube.Deployment, k8sDeployment *storage.Deployment) (needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	updateData = &client.UpdateBcsWorkloadRequestData{}
	needToUpdate = false
	labelsEmpty := map[string]string{}

	//var updateDataIDs []int64
	//updateData.IDs = &updateDataIDs
	//var updateDataInfo client.UpdateBcsWorkloadRequestDataInfo
	//updateData.Info = &updateDataInfo

	if k8sDeployment.Data.Labels == nil {
		if bkDeployment.Labels != nil {
			updateData.Labels = &labelsEmpty
			needToUpdate = true
		}
	} else if bkDeployment.Labels == nil || fmt.Sprint(k8sDeployment.Data.Labels) != fmt.Sprint(*bkDeployment.Labels) {
		updateData.Labels = &k8sDeployment.Data.Labels
		needToUpdate = true
	}

	if k8sDeployment.Data.Spec.Selector == nil {
		if bkDeployment.Selector != nil {
			updateData.Selector = nil
			needToUpdate = true
		}
	} else if bkDeployment.Selector == nil || fmt.Sprint(k8sDeployment.Data.Spec.Selector) != fmt.Sprint(*bkDeployment.Selector) {
		me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
		for _, m := range k8sDeployment.Data.Spec.Selector.MatchExpressions {
			me = append(me, bkcmdbkube.LabelSelectorRequirement{
				Key:      m.Key,
				Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
				Values:   m.Values,
			})
		}

		updateData.Selector = &bkcmdbkube.LabelSelector{
			MatchLabels:      k8sDeployment.Data.Spec.Selector.MatchLabels,
			MatchExpressions: me,
		}

		needToUpdate = true
	}

	if *bkDeployment.Replicas != int64(*k8sDeployment.Data.Spec.Replicas) {
		replicas := int64(*k8sDeployment.Data.Spec.Replicas)
		updateData.Replicas = &replicas
		needToUpdate = true
	}

	if *bkDeployment.MinReadySeconds != int64(k8sDeployment.Data.Spec.MinReadySeconds) {
		minReadySeconds := int64(k8sDeployment.Data.Spec.MinReadySeconds)
		updateData.MinReadySeconds = &minReadySeconds
		needToUpdate = true
	}

	if *bkDeployment.StrategyType != bkcmdbkube.DeploymentStrategyType(k8sDeployment.Data.Spec.Strategy.Type) {
		strategyType := string(k8sDeployment.Data.Spec.Strategy.Type)
		updateData.StrategyType = &strategyType
		needToUpdate = true
	}

	rusEmpty := map[string]interface{}{}

	if k8sDeployment.Data.Spec.Strategy.RollingUpdate == nil {
		if bkDeployment.RollingUpdateStrategy != nil {
			updateData.RollingUpdateStrategy = &rusEmpty
			needToUpdate = true
		}
	} else if fmt.Sprint(k8sDeployment.Data.Spec.Strategy.RollingUpdate) != fmt.Sprint(*bkDeployment.RollingUpdateStrategy) {
		rud := bkcmdbkube.RollingUpdateDeployment{}

		if k8sDeployment.Data.Spec.Strategy.RollingUpdate != nil {
			if k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				rud.MaxUnavailable = &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxUnavailable.Type),
					IntVal: k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxUnavailable.IntVal,
					StrVal: k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxUnavailable.StrVal,
				}
			}

			if k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxSurge != nil {
				rud.MaxSurge = &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxSurge.Type),
					IntVal: k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxSurge.IntVal,
					StrVal: k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxSurge.StrVal,
				}
			}
		}

		jsonBytes, err := json.Marshal(rud)
		if err != nil {
			blog.Errorf("marshal rolling update deployment failed, err: %s", err.Error())
			return false, nil
		}

		rudMap := make(map[string]interface{})
		err = json.Unmarshal(jsonBytes, &rudMap)

		updateData.RollingUpdateStrategy = &rudMap
		needToUpdate = true
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// GenerateBkDeployment generate bkdeployment from k8sdeployment
func (s *Syncer) GenerateBkDeployment(bkNs *bkcmdbkube.Namespace, k8sDeployment *storage.Deployment) *client.CreateBcsWorkloadRequestData {
	me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
	for _, m := range k8sDeployment.Data.Spec.Selector.MatchExpressions {
		me = append(me, bkcmdbkube.LabelSelectorRequirement{
			Key:      m.Key,
			Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
			Values:   m.Values,
		})
	}

	rud := bkcmdbkube.RollingUpdateDeployment{}

	if k8sDeployment.Data.Spec.Strategy.RollingUpdate != nil {
		if k8sDeployment.Data.Spec.Strategy.RollingUpdate != nil {
			if k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxUnavailable != nil {
				rud.MaxUnavailable = &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxUnavailable.Type),
					IntVal: k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxUnavailable.IntVal,
					StrVal: k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxUnavailable.StrVal,
				}
			}

			if k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxSurge != nil {
				rud.MaxSurge = &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxSurge.Type),
					IntVal: k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxSurge.IntVal,
					StrVal: k8sDeployment.Data.Spec.Strategy.RollingUpdate.MaxSurge.StrVal,
				}
			}
		}
	}

	jsonBytes, err := json.Marshal(rud)
	if err != nil {
		blog.Errorf("marshal rolling update deployment failed, err: %s", err.Error())
		return nil
	}

	rudMap := make(map[string]interface{})
	err = json.Unmarshal(jsonBytes, &rudMap)
	replicas := int64(*k8sDeployment.Data.Spec.Replicas)
	minReadySeconds := int64(k8sDeployment.Data.Spec.MinReadySeconds)
	strategyType := string(k8sDeployment.Data.Spec.Strategy.Type)

	return &client.CreateBcsWorkloadRequestData{
		NamespaceID: &bkNs.ID,
		Name:        &k8sDeployment.Data.Name,
		Labels:      &k8sDeployment.Data.Labels,
		Selector: &bkcmdbkube.LabelSelector{
			MatchLabels:      k8sDeployment.Data.Spec.Selector.MatchLabels,
			MatchExpressions: me,
		},
		Replicas:              &replicas,
		MinReadySeconds:       &minReadySeconds,
		StrategyType:          &strategyType,
		RollingUpdateStrategy: &rudMap,
	}
}

// CompareStatefulSet compare bkstatefulset and k8sstatefulset
func (s *Syncer) CompareStatefulSet(bkStatefulSet *bkcmdbkube.StatefulSet, k8sStatefulSet *storage.StatefulSet) (needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	needToUpdate = false
	updateData = &client.UpdateBcsWorkloadRequestData{}
	labelsEmpty := map[string]string{}

	//var updateDataIDs []int64
	//updateData.IDs = &updateDataIDs
	//var updateDataInfo client.UpdateBcsWorkloadRequestDataInfo
	//updateData.Info = &updateDataInfo

	if k8sStatefulSet.Data.Labels == nil {
		if bkStatefulSet.Labels != nil {
			updateData.Labels = &labelsEmpty
			needToUpdate = true
		}
	} else if bkStatefulSet.Labels == nil || fmt.Sprint(k8sStatefulSet.Data.Labels) != fmt.Sprint(*bkStatefulSet.Labels) {
		updateData.Labels = &k8sStatefulSet.Data.Labels
		needToUpdate = true
	}

	if k8sStatefulSet.Data.Spec.Selector == nil {
		if bkStatefulSet.Selector != nil {
			updateData.Selector = nil
			needToUpdate = true
		}
	} else if bkStatefulSet.Selector == nil || fmt.Sprint(k8sStatefulSet.Data.Spec.Selector) != fmt.Sprint(*bkStatefulSet.Selector) {
		me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
		for _, m := range k8sStatefulSet.Data.Spec.Selector.MatchExpressions {
			me = append(me, bkcmdbkube.LabelSelectorRequirement{
				Key:      m.Key,
				Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
				Values:   m.Values,
			})
		}

		updateData.Selector = &bkcmdbkube.LabelSelector{
			MatchLabels:      k8sStatefulSet.Data.Spec.Selector.MatchLabels,
			MatchExpressions: me,
		}

		needToUpdate = true
	}

	if *bkStatefulSet.Replicas != int64(*k8sStatefulSet.Data.Spec.Replicas) {
		replicas := int64(*k8sStatefulSet.Data.Spec.Replicas)
		updateData.Replicas = &replicas
		needToUpdate = true
	}

	if *bkStatefulSet.MinReadySeconds != int64(k8sStatefulSet.Data.Spec.MinReadySeconds) {
		minReadySeconds := int64(k8sStatefulSet.Data.Spec.MinReadySeconds)
		updateData.MinReadySeconds = &minReadySeconds
		needToUpdate = true
	}

	if *bkStatefulSet.StrategyType != bkcmdbkube.StatefulSetUpdateStrategyType(k8sStatefulSet.Data.Spec.UpdateStrategy.Type) {
		strategyType := string(k8sStatefulSet.Data.Spec.UpdateStrategy.Type)
		updateData.StrategyType = &strategyType
		needToUpdate = true
	}

	if k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate == nil {
		if bkStatefulSet.RollingUpdateStrategy != nil {
			updateData.RollingUpdateStrategy = nil
			needToUpdate = true
		}
	} else if fmt.Sprint(k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate) != fmt.Sprint(*bkStatefulSet.RollingUpdateStrategy) {
		rus := bkcmdbkube.RollingUpdateStatefulSetStrategy{}

		if k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate != nil {
			if k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil {
				rus.MaxUnavailable = &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.Type),
					IntVal: k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.IntVal,
					StrVal: k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.StrVal,
				}
			}
			rus.Partition = k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.Partition
		}

		jsonBytes, err := json.Marshal(rus)
		if err != nil {
			blog.Errorf("marshal rolling update statefulset failed, err: %s", err.Error())
			return false, nil
		}

		rusMap := make(map[string]interface{})
		err = json.Unmarshal(jsonBytes, &rusMap)

		updateData.RollingUpdateStrategy = &rusMap
		needToUpdate = true
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// GenerateBkStatefulSet generate bkstatefulset from k8sstatefulset
func (s *Syncer) GenerateBkStatefulSet(bkNs *bkcmdbkube.Namespace, k8sStatefulSet *storage.StatefulSet) *client.CreateBcsWorkloadRequestData {
	me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
	for _, m := range k8sStatefulSet.Data.Spec.Selector.MatchExpressions {
		me = append(me, bkcmdbkube.LabelSelectorRequirement{
			Key:      m.Key,
			Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
			Values:   m.Values,
		})
	}

	rus := bkcmdbkube.RollingUpdateStatefulSetStrategy{}

	if k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate != nil && k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil {
		if k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil {
			rus.MaxUnavailable = &bkcmdbkube.IntOrString{
				Type:   bkcmdbkube.Type(k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.Type),
				IntVal: k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.IntVal,
				StrVal: k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.StrVal,
			}
		}
		rus.Partition = k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.Partition
	}

	jsonBytes, err := json.Marshal(rus)
	if err != nil {
		blog.Errorf("marshal rolling update statefulset failed, err: %s", err.Error())
		return nil
	}

	rusMap := make(map[string]interface{})
	err = json.Unmarshal(jsonBytes, &rusMap)

	replicas := int64(*k8sStatefulSet.Data.Spec.Replicas)
	minReadySeconds := int64(k8sStatefulSet.Data.Spec.MinReadySeconds)
	strategyType := string(k8sStatefulSet.Data.Spec.UpdateStrategy.Type)

	return &client.CreateBcsWorkloadRequestData{
		NamespaceID: &bkNs.ID,
		Name:        &k8sStatefulSet.Data.Name,
		Labels:      &k8sStatefulSet.Data.Labels,
		Selector: &bkcmdbkube.LabelSelector{
			MatchLabels:      k8sStatefulSet.Data.Spec.Selector.MatchLabels,
			MatchExpressions: me,
		},
		Replicas:              &replicas,
		MinReadySeconds:       &minReadySeconds,
		StrategyType:          &strategyType,
		RollingUpdateStrategy: &rusMap,
	}
}

// CompareDaemonSet compare bkdaemonset and k8sdaemonset
func (s *Syncer) CompareDaemonSet(bkDaemonSet *bkcmdbkube.DaemonSet, k8sDaemonSet *storage.DaemonSet) (needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	needToUpdate = false
	updateData = &client.UpdateBcsWorkloadRequestData{}
	labelsEmpty := map[string]string{}

	//var updateDataIDs []int64
	//updateData.IDs = &updateDataIDs
	//var updateDataInfo client.UpdateBcsWorkloadRequestDataInfo
	//updateData.Info = &updateDataInfo

	if k8sDaemonSet.Data.Labels == nil {
		if bkDaemonSet.Labels != nil {
			updateData.Labels = &labelsEmpty
			needToUpdate = true
		}
	} else if bkDaemonSet.Labels == nil || fmt.Sprint(k8sDaemonSet.Data.Labels) != fmt.Sprint(*bkDaemonSet.Labels) {
		updateData.Labels = &k8sDaemonSet.Data.Labels
		needToUpdate = true
	}

	if k8sDaemonSet.Data.Spec.Selector == nil {
		if bkDaemonSet.Selector != nil {
			updateData.Selector = nil
			needToUpdate = true
		}
	} else if bkDaemonSet.Selector == nil || fmt.Sprint(k8sDaemonSet.Data.Spec.Selector) != fmt.Sprint(*bkDaemonSet.Selector) {
		me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
		for _, m := range k8sDaemonSet.Data.Spec.Selector.MatchExpressions {
			me = append(me, bkcmdbkube.LabelSelectorRequirement{
				Key:      m.Key,
				Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
				Values:   m.Values,
			})
		}

		updateData.Selector = &bkcmdbkube.LabelSelector{
			MatchLabels:      k8sDaemonSet.Data.Spec.Selector.MatchLabels,
			MatchExpressions: me,
		}

		needToUpdate = true
	}

	if *bkDaemonSet.MinReadySeconds != int64(k8sDaemonSet.Data.Spec.MinReadySeconds) {
		minReadySeconds := int64(k8sDaemonSet.Data.Spec.MinReadySeconds)
		updateData.MinReadySeconds = &minReadySeconds
		needToUpdate = true
	}

	if *bkDaemonSet.StrategyType != bkcmdbkube.DaemonSetUpdateStrategyType(k8sDaemonSet.Data.Spec.UpdateStrategy.Type) {
		strategyType := string(k8sDaemonSet.Data.Spec.UpdateStrategy.Type)
		updateData.StrategyType = &strategyType
		needToUpdate = true
	}

	if k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate == nil {
		if bkDaemonSet.RollingUpdateStrategy != nil {
			updateData.RollingUpdateStrategy = nil
			needToUpdate = true
		}
	} else if bkDaemonSet.RollingUpdateStrategy == nil || fmt.Sprint(k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate) != fmt.Sprint(*bkDaemonSet.RollingUpdateStrategy) {
		rud := bkcmdbkube.RollingUpdateDaemonSet{}

		if k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate != nil {
			if k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil {
				rud.MaxUnavailable = &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.Type),
					IntVal: k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.IntVal,
					StrVal: k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.StrVal,
				}
			}

			if k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge != nil {
				rud.MaxSurge = &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.Type),
					IntVal: k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.IntVal,
					StrVal: k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.StrVal,
				}
			}
		}

		jsonBytes, err := json.Marshal(rud)
		if err != nil {
			blog.Errorf("marshal rolling update daemonSet failed, err: %s", err.Error())
			return false, nil
		}

		rudMap := make(map[string]interface{})
		err = json.Unmarshal(jsonBytes, &rudMap)

		updateData.RollingUpdateStrategy = &rudMap
		needToUpdate = true
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// GenerateBkDaemonSet generate bkdaemonset from k8sdaemonset
func (s *Syncer) GenerateBkDaemonSet(bkNs *bkcmdbkube.Namespace, k8sDaemonSet *storage.DaemonSet) *client.CreateBcsWorkloadRequestData {
	me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
	for _, m := range k8sDaemonSet.Data.Spec.Selector.MatchExpressions {
		me = append(me, bkcmdbkube.LabelSelectorRequirement{
			Key:      m.Key,
			Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
			Values:   m.Values,
		})
	}

	rud := bkcmdbkube.RollingUpdateDaemonSet{}

	if k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate != nil {
		if k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil {
			rud.MaxUnavailable = &bkcmdbkube.IntOrString{
				Type:   bkcmdbkube.Type(k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.Type),
				IntVal: k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.IntVal,
				StrVal: k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.StrVal,
			}
		}

		if k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge != nil {
			rud.MaxSurge = &bkcmdbkube.IntOrString{
				Type:   bkcmdbkube.Type(k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.Type),
				IntVal: k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.IntVal,
				StrVal: k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.StrVal,
			}
		}
	}

	jsonBytes, err := json.Marshal(rud)
	if err != nil {
		blog.Errorf("marshal rolling update daemonset failed, err: %s", err.Error())
		return nil
	}

	rudMap := make(map[string]interface{})
	err = json.Unmarshal(jsonBytes, &rudMap)

	minReadySeconds := int64(k8sDaemonSet.Data.Spec.MinReadySeconds)
	strategyType := string(k8sDaemonSet.Data.Spec.UpdateStrategy.Type)

	return &client.CreateBcsWorkloadRequestData{
		NamespaceID: &bkNs.ID,
		Name:        &k8sDaemonSet.Data.Name,
		Labels:      &k8sDaemonSet.Data.Labels,
		Selector: &bkcmdbkube.LabelSelector{
			MatchLabels:      k8sDaemonSet.Data.Spec.Selector.MatchLabels,
			MatchExpressions: me,
		},
		MinReadySeconds:       &minReadySeconds,
		StrategyType:          &strategyType,
		RollingUpdateStrategy: &rudMap,
	}
}

// CompareGameDeployment compare bkgamedeployment and k8sgamedeployment
func (s *Syncer) CompareGameDeployment(bkGameDeployment *bkcmdbkube.GameDeployment, k8sGameDeployment *storage.GameDeployment) (needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	needToUpdate = false
	updateData = &client.UpdateBcsWorkloadRequestData{}
	labelsEmpty := map[string]string{}

	//var updateDataIDs []int64
	//updateData.IDs = &updateDataIDs
	//var updateDataInfo client.UpdateBcsWorkloadRequestDataInfo
	//updateData.Info = &updateDataInfo

	if k8sGameDeployment.Data.Labels == nil {
		if bkGameDeployment.Labels != nil {
			updateData.Labels = &labelsEmpty
			needToUpdate = true
		}
	} else if bkGameDeployment.Labels == nil || fmt.Sprint(k8sGameDeployment.Data.Labels) != fmt.Sprint(*bkGameDeployment.Labels) {
		updateData.Labels = &k8sGameDeployment.Data.Labels
		needToUpdate = true
	}

	if k8sGameDeployment.Data.Spec.Selector == nil {
		if bkGameDeployment.Selector != nil {
			updateData.Selector = nil
			needToUpdate = true
		}
	} else if bkGameDeployment.Selector == nil || fmt.Sprint(k8sGameDeployment.Data.Spec.Selector) != fmt.Sprint(*bkGameDeployment.Selector) {
		me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
		for _, m := range k8sGameDeployment.Data.Spec.Selector.MatchExpressions {
			me = append(me, bkcmdbkube.LabelSelectorRequirement{
				Key:      m.Key,
				Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
				Values:   m.Values,
			})
		}

		updateData.Selector = &bkcmdbkube.LabelSelector{
			MatchLabels:      k8sGameDeployment.Data.Spec.Selector.MatchLabels,
			MatchExpressions: me,
		}

		needToUpdate = true
	}

	if *bkGameDeployment.Replicas != int64(*k8sGameDeployment.Data.Spec.Replicas) {
		replicas := int64(*k8sGameDeployment.Data.Spec.Replicas)
		updateData.Replicas = &replicas
		needToUpdate = true
	}

	if *bkGameDeployment.MinReadySeconds != int64(k8sGameDeployment.Data.Spec.MinReadySeconds) {
		minReadySeconds := int64(k8sGameDeployment.Data.Spec.MinReadySeconds)
		updateData.MinReadySeconds = &minReadySeconds
		needToUpdate = true
	}

	if *bkGameDeployment.StrategyType != bkcmdbkube.GameDeploymentUpdateStrategyType(k8sGameDeployment.Data.Spec.UpdateStrategy.Type) {
		strategyType := string(k8sGameDeployment.Data.Spec.UpdateStrategy.Type)
		updateData.StrategyType = &strategyType
		needToUpdate = true
	}

	rud := bkcmdbkube.RollingUpdateGameDeployment{}

	if k8sGameDeployment.Data.Spec.UpdateStrategy.MaxUnavailable != nil && k8sGameDeployment.Data.Spec.UpdateStrategy.MaxSurge != nil {
		rud = bkcmdbkube.RollingUpdateGameDeployment{
			MaxUnavailable: &bkcmdbkube.IntOrString{
				Type:   bkcmdbkube.Type(k8sGameDeployment.Data.Spec.UpdateStrategy.MaxUnavailable.Type),
				IntVal: k8sGameDeployment.Data.Spec.UpdateStrategy.MaxUnavailable.IntVal,
				StrVal: k8sGameDeployment.Data.Spec.UpdateStrategy.MaxUnavailable.StrVal,
			},
			MaxSurge: &bkcmdbkube.IntOrString{
				Type:   bkcmdbkube.Type(k8sGameDeployment.Data.Spec.UpdateStrategy.MaxSurge.Type),
				IntVal: k8sGameDeployment.Data.Spec.UpdateStrategy.MaxSurge.IntVal,
				StrVal: k8sGameDeployment.Data.Spec.UpdateStrategy.MaxSurge.StrVal,
			},
		}

		jsonBytes, err := json.Marshal(rud)
		if err != nil {
			blog.Errorf("marshal rolling update gamedeployment failed, err: %s", err.Error())
			return false, nil
		}

		rudMap := make(map[string]interface{})
		err = json.Unmarshal(jsonBytes, &rudMap)

		updateData.RollingUpdateStrategy = &rudMap
		needToUpdate = true

	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// GenerateBkGameDeployment generate bkgamedeployment from k8sgamedeployment
func (s *Syncer) GenerateBkGameDeployment(bkNs *bkcmdbkube.Namespace, k8sGameDeployment *storage.GameDeployment) *client.CreateBcsWorkloadRequestData {
	me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
	for _, m := range k8sGameDeployment.Data.Spec.Selector.MatchExpressions {
		me = append(me, bkcmdbkube.LabelSelectorRequirement{
			Key:      m.Key,
			Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
			Values:   m.Values,
		})
	}

	rud := bkcmdbkube.RollingUpdateGameDeployment{}

	if k8sGameDeployment.Data.Spec.UpdateStrategy.MaxUnavailable != nil && k8sGameDeployment.Data.Spec.UpdateStrategy.MaxSurge != nil {
		rud = bkcmdbkube.RollingUpdateGameDeployment{
			MaxUnavailable: &bkcmdbkube.IntOrString{
				Type:   bkcmdbkube.Type(k8sGameDeployment.Data.Spec.UpdateStrategy.MaxUnavailable.Type),
				IntVal: k8sGameDeployment.Data.Spec.UpdateStrategy.MaxUnavailable.IntVal,
				StrVal: k8sGameDeployment.Data.Spec.UpdateStrategy.MaxUnavailable.StrVal,
			},
			MaxSurge: &bkcmdbkube.IntOrString{
				Type:   bkcmdbkube.Type(k8sGameDeployment.Data.Spec.UpdateStrategy.MaxSurge.Type),
				IntVal: k8sGameDeployment.Data.Spec.UpdateStrategy.MaxSurge.IntVal,
				StrVal: k8sGameDeployment.Data.Spec.UpdateStrategy.MaxSurge.StrVal,
			},
		}
	}

	jsonBytes, err := json.Marshal(rud)
	if err != nil {
		blog.Errorf("marshal rolling update gamedeployment failed, err: %s", err.Error())
		return nil
	}

	rudMap := make(map[string]interface{})
	err = json.Unmarshal(jsonBytes, &rudMap)

	replicas := int64(*k8sGameDeployment.Data.Spec.Replicas)
	minReadySeconds := int64(k8sGameDeployment.Data.Spec.MinReadySeconds)
	strategyType := string(k8sGameDeployment.Data.Spec.UpdateStrategy.Type)

	return &client.CreateBcsWorkloadRequestData{
		NamespaceID: &bkNs.ID,
		Name:        &k8sGameDeployment.Data.Name,
		Labels:      &k8sGameDeployment.Data.Labels,
		Selector: &bkcmdbkube.LabelSelector{
			MatchLabels:      k8sGameDeployment.Data.Spec.Selector.MatchLabels,
			MatchExpressions: me,
		},
		Replicas:              &replicas,
		MinReadySeconds:       &minReadySeconds,
		StrategyType:          &strategyType,
		RollingUpdateStrategy: &rudMap,
	}

}

// CompareGameStatefulSet compare bkgamestatefulset and k8sgamestatefulset
func (s *Syncer) CompareGameStatefulSet(bkGameStatefulSet *bkcmdbkube.GameStatefulSet, k8sGameStatefulSet *storage.GameStatefulSet) (needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	needToUpdate = false
	updateData = &client.UpdateBcsWorkloadRequestData{}
	labelsEmpty := map[string]string{}

	//var updateDataIDs []int64
	//updateData.IDs = &updateDataIDs
	//var updateDataInfo client.UpdateBcsWorkloadRequestDataInfo
	//updateData.Info = &updateDataInfo

	if k8sGameStatefulSet.Data.Labels == nil {
		if bkGameStatefulSet.Labels != nil {
			updateData.Labels = &labelsEmpty
			needToUpdate = true
		}
	} else if bkGameStatefulSet.Labels == nil || fmt.Sprint(k8sGameStatefulSet.Data.Labels) != fmt.Sprint(*bkGameStatefulSet.Labels) {
		updateData.Labels = &k8sGameStatefulSet.Data.Labels
		needToUpdate = true
	}

	if k8sGameStatefulSet.Data.Spec.Selector == nil {
		if bkGameStatefulSet.Selector != nil {
			updateData.Selector = nil
			needToUpdate = true
		}
	} else if bkGameStatefulSet.Selector == nil || fmt.Sprint(k8sGameStatefulSet.Data.Spec.Selector) != fmt.Sprint(*bkGameStatefulSet.Selector) {
		me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
		for _, m := range k8sGameStatefulSet.Data.Spec.Selector.MatchExpressions {
			me = append(me, bkcmdbkube.LabelSelectorRequirement{
				Key:      m.Key,
				Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
				Values:   m.Values,
			})
		}

		updateData.Selector = &bkcmdbkube.LabelSelector{
			MatchLabels:      k8sGameStatefulSet.Data.Spec.Selector.MatchLabels,
			MatchExpressions: me,
		}

		needToUpdate = true
	}

	if bkGameStatefulSet.Replicas != nil && k8sGameStatefulSet.Data.Spec.Replicas != nil {
		if *bkGameStatefulSet.Replicas != int64(*k8sGameStatefulSet.Data.Spec.Replicas) {
			replicas := int64(*k8sGameStatefulSet.Data.Spec.Replicas)
			updateData.Replicas = &replicas
			needToUpdate = true
		}
	}

	if *bkGameStatefulSet.StrategyType != bkcmdbkube.GameStatefulSetUpdateStrategyType(k8sGameStatefulSet.Data.Spec.UpdateStrategy.Type) {
		strategyType := string(k8sGameStatefulSet.Data.Spec.UpdateStrategy.Type)
		updateData.StrategyType = &strategyType
		needToUpdate = true
	}

	rus := bkcmdbkube.RollingUpdateGameStatefulSetStrategy{}

	if k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate != nil {
		if k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil && k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge != nil {
			blog.Infof("rolling update: %+v", k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate)
			rus = bkcmdbkube.RollingUpdateGameStatefulSetStrategy{
				MaxUnavailable: &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.Type),
					IntVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.IntVal,
					StrVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.StrVal,
				},
				MaxSurge: &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.Type),
					IntVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.IntVal,
					StrVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.StrVal,
				},
			}

			jsonBytes, err := json.Marshal(rus)
			if err != nil {
				blog.Errorf("marshal rolling update gameStatefulSets failed, err: %s", err.Error())
				return false, nil
			}

			rudMap := make(map[string]interface{})
			err = json.Unmarshal(jsonBytes, &rudMap)

			updateData.RollingUpdateStrategy = &rudMap
			needToUpdate = true
		}
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// GenerateBkGameStatefulSet generate bkgamestatefulset from k8sgamestatefulset
func (s *Syncer) GenerateBkGameStatefulSet(bkNs *bkcmdbkube.Namespace, k8sGameStatefulSet *storage.GameStatefulSet) *client.CreateBcsWorkloadRequestData {
	me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
	for _, m := range k8sGameStatefulSet.Data.Spec.Selector.MatchExpressions {
		me = append(me, bkcmdbkube.LabelSelectorRequirement{
			Key:      m.Key,
			Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
			Values:   m.Values,
		})
	}

	rus := bkcmdbkube.RollingUpdateGameStatefulSetStrategy{}

	if k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate != nil {
		if k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil && k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge != nil {
			blog.Infof("rolling update: %+v", k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate)
			rus = bkcmdbkube.RollingUpdateGameStatefulSetStrategy{
				MaxUnavailable: &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.Type),
					IntVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.IntVal,
					StrVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.StrVal,
				},
				MaxSurge: &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.Type),
					IntVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.IntVal,
					StrVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.StrVal,
				},
			}
		}
	}

	jsonBytes, err := json.Marshal(rus)
	if err != nil {
		blog.Errorf("marshal rolling update gameStatefulSets failed, err: %s", err.Error())
		return nil
	}

	rusMap := make(map[string]interface{})
	err = json.Unmarshal(jsonBytes, &rusMap)

	var replicas int64

	if k8sGameStatefulSet.Data.Spec.Replicas != nil {
		replicas = int64(*k8sGameStatefulSet.Data.Spec.Replicas)
	}

	strategyType := string(k8sGameStatefulSet.Data.Spec.UpdateStrategy.Type)

	return &client.CreateBcsWorkloadRequestData{
		NamespaceID: &bkNs.ID,
		Name:        &k8sGameStatefulSet.Data.Name,
		Labels:      &k8sGameStatefulSet.Data.Labels,
		Selector: &bkcmdbkube.LabelSelector{
			MatchLabels:      k8sGameStatefulSet.Data.Spec.Selector.MatchLabels,
			MatchExpressions: me,
		},
		Replicas:              &replicas,
		StrategyType:          &strategyType,
		RollingUpdateStrategy: &rusMap,
	}
}

// UpdateBkWorkloads update bkworkloads
// nolint
func (s *Syncer) UpdateBkWorkloads(bkCluster *bkcmdbkube.Cluster, kind string, toUpdate *map[int64]*client.UpdateBcsWorkloadRequestData) {
	if toUpdate == nil {
		return
	}

	for k, v := range *toUpdate {
		// UpdateBcsWorkload updates the BCS workload with the given request.
		err := s.CMDBClient.UpdateBcsWorkload(&client.UpdateBcsWorkloadRequest{
			BKBizID: &bkCluster.BizID,
			Kind:    &kind,
			IDs:     &[]int64{k},
			Data:    v,
		})

		if err != nil {
			blog.Errorf("update workload %s failed, err: %s", kind, err.Error())
		}
	}
}

// DeleteBkWorkloads delete bkworkloads
func (s *Syncer) DeleteBkWorkloads(bkCluster *bkcmdbkube.Cluster, kind string, toDelete *[]int64) error {
	if len(*toDelete) > 0 {
		err := s.CMDBClient.DeleteBcsWorkload(&client.DeleteBcsWorkloadRequest{
			BKBizID: &bkCluster.BizID,
			Kind:    &kind,
			IDs:     toDelete,
		})
		if err != nil {
			for i := 0; i < len(*toDelete); i++ {
				var section []int64
				if i+1 > len(*toDelete) {
					section = (*toDelete)[i:]
				} else {
					section = (*toDelete)[i : i+1]
				}
				// DeleteBcsWorkload deletes the BCS workload with the given request.
				err = s.CMDBClient.DeleteBcsWorkload(&client.DeleteBcsWorkloadRequest{
					BKBizID: &bkCluster.BizID,
					Kind:    &kind,
					IDs:     &section,
				})

				if err != nil {
					blog.Errorf("delete workload %s failed, err: %s", kind, err.Error())
					return err
				}
			}
		}
		blog.Infof("delete workload %s success, ids: %v", kind, toDelete)
	}
	return nil
}

// GetBkPods get bkpods
func (s *Syncer) GetBkPods(bkBizID int64, filter *client.PropertyFilter) (*[]bkcmdbkube.Pod, error) {
	bkPodList := make([]bkcmdbkube.Pod, 0)

	pageStart := 0
	for {
		// GetBcsPod returns the BCS pod information for the given request.
		bkPods, err := s.CMDBClient.GetBcsPod(&client.GetBcsPodRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkBizID,
				Page: client.Page{
					Limit: 100,
					Start: 100 * pageStart,
				},
				Filter: filter,
			},
		})

		if err != nil {
			blog.Errorf("get bcs pod failed, err: %s", err.Error())
			return nil, err
		}
		bkPodList = append(bkPodList, *bkPods...)

		if len(*bkPods) < 100 {
			break
		}
		pageStart++
	}

	return &bkPodList, nil
}

// CreateBkPods create bkpods
func (s *Syncer) CreateBkPods(bkCluster *bkcmdbkube.Cluster, toCreate map[int64][]client.CreateBcsPodRequestDataPod) {
	if len(toCreate) > 0 {
		for bizid, pods := range toCreate {
			if len(pods) > 0 {
				for i := 0; i < len(pods); i += 100 {
					var section []client.CreateBcsPodRequestDataPod
					if i+100 > len(pods) {
						section = (pods)[i:]
					} else {
						section = (pods)[i : i+100]
					}

					// CreateBcsPod creates a new BCS pod with the given request.
					_, err := s.CMDBClient.CreateBcsPod(&client.CreateBcsPodRequest{
						Data: &[]client.CreateBcsPodRequestData{
							{
								BizID: &bizid,
								Pods:  &section,
							},
						},
					})

					if err != nil {
						for j := 0; j < len(section); j++ {
							sec := section[j : j+1]
							_, err = s.CMDBClient.CreateBcsPod(&client.CreateBcsPodRequest{
								Data: &[]client.CreateBcsPodRequestData{
									{
										BizID: &bizid,
										Pods:  &sec,
									},
								},
							})
						}

						if err != nil {
							blog.Errorf("create pod failed, err: %s", err.Error())
						}

					} else {
						blog.Infof("create pod success, data: %v", section)
					}
				}
			}
		}
	}
}

// DeleteBkPods delete bkpods
func (s *Syncer) DeleteBkPods(bkCluster *bkcmdbkube.Cluster, toDelete *[]int64) error {
	if len(*toDelete) > 0 {
		// DeleteBcsPod deletes the BCS pod with the given request.
		err := s.CMDBClient.DeleteBcsPod(&client.DeleteBcsPodRequest{
			Data: &[]client.DeleteBcsPodRequestData{
				{
					BKBizID: &bkCluster.BizID,
					IDs:     toDelete,
				},
			},
		})
		if err != nil {
			blog.Errorf("delete pod failed, err: %s", err.Error())
			return err
		}
		blog.Infof("delete pod success, ids: %v", toDelete)
	}
	return nil
}

// GenerateBkWorkloadPods generate bkworkloadpods
func (s *Syncer) GenerateBkWorkloadPods(bkNs *bkcmdbkube.Namespace) *client.CreateBcsWorkloadRequestData {
	podsName := "pods" // nolint

	return &client.CreateBcsWorkloadRequestData{
		NamespaceID: &bkNs.ID,
		Name:        &podsName,
	}
}

// CompareBkWorkloadPods current no need to compare
func (s *Syncer) CompareBkWorkloadPods(workload *bkcmdbkube.PodsWorkload) (needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	return false, nil
}
