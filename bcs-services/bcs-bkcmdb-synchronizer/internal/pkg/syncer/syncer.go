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
// NOCC:tosa/comment_ratio(ignore),gofmt/notformat(ignore)
package syncer

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	bkcmdbkube "configcenter/src/kube/types" // nolint

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage"
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage/tkex/gamedeployment/v1alpha1"
	gsv1alpha1 "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/storage/tkex/gamestatefulset/v1alpha1"
	"gorm.io/gorm"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client"
	bsc "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/bcsstorage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/cmdb"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/store/db/sqlite"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/store/model"
)

// Syncer the syncer
type Syncer struct {
	BkcmdbSynchronizerOption *option.BkcmdbSynchronizerOption
	ClientTls                *tls.Config
	// Rabbit                   mq.MQ
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
		// static.ClientCertPwd)
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
// nolint (error) is always nil
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
	bkCluster, err := s.GetBkCluster(cluster, nil, false)
	clusterType := "INDEPENDENT_CLUSTER"
	if cluster.IsShared {
		clusterType = "SHARE_CLUSTER"
	}
	path := "/data/bcs/bcs-bkcmdb-synchronizer/db/" + cluster.ClusterID + ".db"
	db := sqlite.Open(path)
	if db == nil {
		blog.Errorf("open db failed, path: %s", path)
		return fmt.Errorf("open db failed, path: %s", path)
	}
	s.SyncStoreMigrate(db)
	if err != nil { // nolint
		if err.Error() == "cluster not found" { // nolint
			var clusterBkBizID int64
			if bkBizID == 0 {
				clusterBkBizID, err = strconv.ParseInt(cluster.BusinessID, 10, 64)
				if err != nil {
					blog.Errorf("An error occurred: %s\n", err)
				} else {
					blog.Infof("Successfully converted string to int64: %d\n", clusterBkBizID)
				}
			} else {
				clusterBkBizID = bkBizID
			}
			// CreateBcsCluster creates a new BCS cluster with the given request.
			_, err = s.CMDBClient.CreateBcsCluster(&client.CreateBcsClusterRequest{
				BKBizID:          &clusterBkBizID,
				Name:             &cluster.ClusterName,
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
			}, db)
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
				Name:        &cluster.ClusterName,
				Version:     &cluster.ClusterBasicSettings.Version,
				NetworkType: &cluster.NetworkType,
				Region:      &cluster.Region,
				Network:     &clusterNetwork,
				Environment: &cluster.Environment,
			},
		}, db)
		if err != nil {
			blog.Errorf("update bcs cluster failed, err: %s", err.Error())
		}
		if *bkCluster.Type != clusterType {
			// UpdateBcsClusterType updates the BCS cluster type with the given request.
			err = s.CMDBClient.UpdateBcsClusterType(&client.UpdateBcsClusterTypeRequest{
				BKBizID: &bkCluster.BizID,
				ID:      &bkCluster.ID,
				Type:    &clusterType,
			}, db)
			if err != nil {
				blog.Errorf("update bcs cluster type failed, err: %s", err.Error())
			}
		}
	}
	return nil
}

// SyncNodes sync nodes
func (s *Syncer) SyncNodes(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
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
	}, true, db)

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
			blog.Infof("nodeToDelete: %s+%s", bkCluster.Uid, *v.Name)
		}
	}

	for k, v := range nodeMap {
		if _, ok := bkNodeMap[k]; ok {
			// CompareNode compare bknode and k8snode
			needToUpdate, updateData := s.CompareNode(bkNodeMap[k], v)

			if needToUpdate {
				nodeToUpdate[bkNodeMap[k].ID] = updateData
				blog.Infof("nodeToUpdate: %s+%s", bkCluster.Uid, v.Data.Name)
			}
		} else {
			// GenerateBkNodeData generate bknode data from k8snode
			nodeData, errN := s.GenerateBkNodeData(bkCluster, v)
			if errN == nil {
				nodeToAdd = append(nodeToAdd, nodeData)
				blog.Infof("nodeToAdd: %s+%s", bkCluster.Uid, v.Data.Name)
			}

		}
	}

	s.CreateBkNodes(bkCluster, &nodeToAdd, db)
	s.DeleteBkNodes(bkCluster, &nodeToDelete, db) // nolint  not checked
	s.UpdateBkNodes(bkCluster, &nodeToUpdate, db)

	return err
}

func (s *Syncer) getBizidByProjectCode(projectCode string) (int64, error) {
	// GetProjectManagerGrpcGwClient is a function that returns a project manager gRPC gateway client.
	pmCli, err := s.GetProjectManagerGrpcGwClient()
	if err != nil {
		blog.Errorf("get project manager grpc gw client failed, err: %s", err.Error())
		return 0, err
	}
	gpr := pmp.GetProjectRequest{
		ProjectIDOrCode: projectCode,
	}
	project, errP := pmCli.Cli.GetProject(pmCli.Ctx, &gpr)
	if errP != nil {
		blog.Errorf("get project failed, err: %s", errP.Error())
		return 0, errP
	}

	if project != nil && project.Data != nil && project.Data.BusinessID != "" {
		bizid, errPP := strconv.ParseInt(project.Data.BusinessID, 10, 64)
		if errPP != nil {
			blog.Errorf("projectcode parse string err: %v", errPP)
			return 0, errPP
		}

		blog.Infof("projectcode: %s, bizid: %d", projectCode, bizid)
		return bizid, nil
	}
	return 0, errors.New("projectcode not found")
}

// SyncNamespaces sync namespaces
// nolint funlen
func (s *Syncer) SyncNamespaces(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	// GetBcsStorageClient is a function that returns a BCS storage client.
	storageCli, err := s.GetBcsStorageClient()
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
		return err
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
	}, true, db)
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	blog.Infof("get bcs namespace success, namespaces: %d", len(*bkNamespaceList))

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
			blog.Infof("nsToDelete: %s+%s", bkCluster.Uid, v.Name)
		}
	}

	for k, v := range nsMap {
		nsbizid := bkCluster.BizID
		if projectCode, ok := v.Data.Annotations[s.BkcmdbSynchronizerOption.SharedCluster.AnnotationKeyProjCode]; ok {
			bizid, errP := s.getBizidByProjectCode(projectCode)
			if errP != nil {
				blog.Errorf("get bizid by projectcode err: %v", errP)
			} else {
				nsbizid = bizid
			}
		}

		if _, ok := bkNsMap[k]; ok {
			if bkNsMap[k].BizID != nsbizid {
				err = s.DeleteAllByClusterAndNamespace(bkCluster, bkNsMap[k], db)
				if err != nil {
					blog.Errorf("delete namespace error: %v", err)
					continue
				}
				if _, ok = nsToAdd[nsbizid]; ok {
					nsToAdd[nsbizid] = append(nsToAdd[nsbizid], s.GenerateBkNsData(bkCluster, v))
					blog.Infof("nsToAdd: %s+%s", bkCluster.Uid, v.Data.Name)
				} else {
					nsToAdd[nsbizid] = []bkcmdbkube.Namespace{s.GenerateBkNsData(bkCluster, v)}
					blog.Infof("nsToAdd: %s+%s", bkCluster.Uid, v.Data.Name)
				}
				continue
			}
			// CompareNamespace compare bkns and k8sns
			needToUpdate, updateData := s.CompareNamespace(bkNsMap[k], v)
			if needToUpdate {
				nsToUpdate[bkNsMap[k].ID] = updateData
				blog.Infof("nsToUpdate: %s+%s", bkCluster.Uid, v.Data.Name)
			}
		} else {
			if _, ok = nsToAdd[nsbizid]; ok {
				nsToAdd[nsbizid] = append(nsToAdd[nsbizid], s.GenerateBkNsData(bkCluster, v))
				blog.Infof("nsToAdd: %s+%s", bkCluster.Uid, v.Data.Name)
			} else {
				nsToAdd[nsbizid] = []bkcmdbkube.Namespace{s.GenerateBkNsData(bkCluster, v)}
				blog.Infof("nsToAdd: %s+%s", bkCluster.Uid, v.Data.Name)
			}
		}
	}

	s.DeleteBkNamespaces(bkCluster, &nsToDelete, db) // nolint  not checked
	s.CreateBkNamespaces(bkCluster, nsToAdd, db)
	s.UpdateBkNamespaces(bkCluster, &nsToUpdate, db)

	return nil
}

// SyncWorkloads sync workloads
func (s *Syncer) SyncWorkloads(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	// syncDeployments sync deployments
	err := s.syncDeployments(cluster, bkCluster, db)
	if err != nil {
		blog.Errorf("sync deployment failed, err: %s", err.Error())
	}

	// syncStatefulSets sync statefulsets
	err = s.syncStatefulSets(cluster, bkCluster, db)
	if err != nil {
		blog.Errorf("sync statefulset failed, err: %s", err.Error())
	}

	// syncDaemonSets sync daemonsets
	err = s.syncDaemonSets(cluster, bkCluster, db)
	if err != nil {
		blog.Errorf("sync daemonset failed, err: %s", err.Error())
	}

	// syncGameDeployments sync gamedeployments
	err = s.syncGameDeployments(cluster, bkCluster, db)
	if err != nil {
		blog.Errorf("sync gamedeployment failed, err: %s", err.Error())
	}

	// syncGameStatefulSets sync gamestatefulsets
	err = s.syncGameStatefulSets(cluster, bkCluster, db)
	if err != nil {
		blog.Errorf("sync gamestatefulset failed, err: %s", err.Error())
	}

	// syncWorkloadPods sync workloadPods
	err = s.syncWorkloadPods(cluster, bkCluster, db)
	if err != nil {
		blog.Errorf("sync workload pods failed, err: %s", err.Error())
	}

	return err
}

// syncDeployments sync deployments
// nolint funlen
func (s *Syncer) syncDeployments(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
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
	}, true, db)
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
	}, true, db)
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
			blog.Infof("deploymentToDelete: %s+%s+%s", bkCluster.Uid, v.Namespace, v.Name)
		}
	}

	for k, v := range deploymentMap {
		if _, ok := bkDeploymentMap[k]; !ok {
			// GenerateBkDeployment generate bkdeployment from k8sdeployment
			toAddData := s.GenerateBkDeployment(bkNamespaceMap[v.Data.Namespace], v)

			if toAddData != nil {
				if _, ok = deploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID]; ok {
					deploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = append(deploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID], *toAddData)
					blog.Infof("deploymentToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
				} else {
					deploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
					blog.Infof("deploymentToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
				}
			}
		} else {
			// CompareDeployment compare bkdeployment and k8sdeployment
			needToUpdate, updateData := s.CompareDeployment(bkDeploymentMap[k], v)

			if needToUpdate {
				deploymentToUpdate[bkDeploymentMap[k].ID] = updateData
				blog.Infof("deploymentToUpdate: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
			}
		}
	}

	s.DeleteBkWorkloads(bkCluster, kind, &deploymentToDelete, db) // nolint  not checked
	s.CreateBkWorkloads(bkCluster, kind, deploymentToAdd, db)
	s.UpdateBkWorkloads(bkCluster, kind, &deploymentToUpdate, db)

	return nil
}

// syncStatefulSets sync statefulsets
// nolint funlen
func (s *Syncer) syncStatefulSets(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
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
	}, true, db)
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
	}, true, db)
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
			blog.Infof("statefulSetToDelete: %s+%s+%s", bkCluster.Uid, v.Namespace, v.Name)
		}
	}

	for k, v := range statefulSetMap {
		if _, ok := bkStatefulSetMap[k]; !ok {
			toAddData := s.GenerateBkStatefulSet(bkNamespaceMap[v.Data.Namespace], v)

			if toAddData != nil {
				if _, ok = statefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID]; ok {
					statefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = append(statefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID], *toAddData)
					blog.Infof("statefulSetToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
				} else {
					statefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
					blog.Infof("statefulSetToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
				}
			}
		} else {
			needToUpdate, updateData := s.CompareStatefulSet(bkStatefulSetMap[k], v)

			if needToUpdate {
				statefulSetToUpdate[bkStatefulSetMap[k].ID] = updateData
				blog.Infof("statefulSetToUpdate: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
			}
		}
	}

	s.DeleteBkWorkloads(bkCluster, kind, &statefulSetToDelete, db) // nolint  not checked
	s.CreateBkWorkloads(bkCluster, kind, statefulSetToAdd, db)
	s.UpdateBkWorkloads(bkCluster, kind, &statefulSetToUpdate, db)

	return nil
}

// syncDaemonSets sync daemonsets
// nolint funlen
func (s *Syncer) syncDaemonSets(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
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
	}, true, db)
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	// 将bk命名空间列表转换为map，便于后续查找
	bkNamespaceMap := make(map[string]*bkcmdbkube.Namespace)
	for k, v := range *bkNamespaceList {
		bkNamespaceMap[v.Name] = &(*bkNamespaceList)[k]
	}

	// 初始化daemonSet列表
	daemonSetList := make([]*storage.DaemonSet, 0)
	bkDaemonSetList := make([]bkcmdbkube.DaemonSet, 0)

	// 遍历bk命名空间列表，查询每个命名空间下的daemonSet
	for _, ns := range *bkNamespaceList {
		daemonSets, err := storageCli.QueryK8SDaemonSet(cluster.ClusterID, ns.Name)
		if err != nil {
			blog.Errorf("query k8s daemonset failed, err: %s", err.Error())
			return err
		}
		daemonSetList = append(daemonSetList, daemonSets...)
	}
	blog.Infof("get daemonset list success, len: %d", len(daemonSetList))

	// 获取bk的daemonSet列表
	bkDaemonSets, err := s.GetBkWorkloads(bkCluster.BizID, "daemonSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		blog.Errorf("get bk daemonset failed, err: %s", err.Error())
		return err
	}

	// 将bk的daemonSet列表转换为结构体列表
	for _, bkDaemonSet := range *bkDaemonSets {
		b := bkcmdbkube.DaemonSet{}
		err := common.InterfaceToStruct(bkDaemonSet, &b)
		if err != nil {
			blog.Errorf("convert bk daemonset failed, err: %s", err.Error())
			return err
		}

		bkDaemonSetList = append(bkDaemonSetList, b)
	}

	// 初始化待添加、更新、删除的daemonSet列表
	daemonSetToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	daemonSetToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
	daemonSetToDelete := make([]int64, 0)

	// 将daemonSet列表转换为map，便于后续查找
	daemonSetMap := make(map[string]*storage.DaemonSet)
	for _, v := range daemonSetList {
		daemonSetMap[v.Data.Namespace+v.Data.Name] = v
	}

	// 将bkDaemonSet列表转换为map，便于后续查找
	bkDaemonSetMap := make(map[string]*bkcmdbkube.DaemonSet)
	for k, v := range bkDaemonSetList {
		bkDaemonSetMap[v.Namespace+v.Name] = &bkDaemonSetList[k]

		// 如果在daemonSetMap中找不到对应的daemonSet，则标记为待删除
		if _, ok := daemonSetMap[v.Namespace+v.Name]; !ok {
			daemonSetToDelete = append(daemonSetToDelete, v.ID)
			blog.Infof("daemonSetToDelete: %s+%s+%s", bkCluster.Uid, v.Namespace, v.Name)
		}
	}

	// 遍历daemonSetMap，检查是否需要添加或更新daemonSet
	for k, v := range daemonSetMap {
		if _, ok := bkDaemonSetMap[k]; !ok {
			// 如果在bkDaemonSetMap中找不到对应的daemonSet，则生成待添加的数据
			toAddData := s.GenerateBkDaemonSet(bkNamespaceMap[v.Data.Namespace], v)

			if toAddData != nil {
				if _, ok = daemonSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID]; ok {
					daemonSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] =
						append(daemonSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID], *toAddData)
					blog.Infof("daemonSetToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
				} else {
					daemonSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] =
						[]client.CreateBcsWorkloadRequestData{*toAddData}
					blog.Infof("daemonSetToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
				}
			}
		} else {
			// 如果存在对应的daemonSet，则比较是否需要更新
			needToUpdate, updateData := s.CompareDaemonSet(bkDaemonSetMap[k], v)

			if needToUpdate {
				daemonSetToUpdate[bkDaemonSetMap[k].ID] = updateData
				blog.Infof("daemonSetToUpdate: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
			}
		}
	}

	// 执行删除、添加、更新操作
	s.DeleteBkWorkloads(bkCluster, kind, &daemonSetToDelete, db) // nolint  not checked
	s.CreateBkWorkloads(bkCluster, kind, daemonSetToAdd, db)
	s.UpdateBkWorkloads(bkCluster, kind, &daemonSetToUpdate, db)

	return nil
}

// syncGameDeployments sync gamedeployments
// nolint funlen
func (s *Syncer) syncGameDeployments(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	kind := "gameDeployment"                   // 定义资源类型为游戏部署
	storageCli, err := s.GetBcsStorageClient() // 获取Bcs存储客户端
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}

	// 获取bk命名空间列表
	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	// 将命名空间列表转换为map，便于后续查找
	bkNamespaceMap := make(map[string]*bkcmdbkube.Namespace)
	for k, v := range *bkNamespaceList {
		bkNamespaceMap[v.Name] = &(*bkNamespaceList)[k]
	}

	gameDeploymentList := make([]*storage.GameDeployment, 0)     // 存储从K8S查询到的游戏部署列表
	bkGameDeploymentList := make([]bkcmdbkube.GameDeployment, 0) // 存储从bk查询到的游戏部署列表

	// 遍历命名空间列表，查询每个命名空间下的游戏部署
	for _, ns := range *bkNamespaceList {
		gameDeployments, err := storageCli.QueryK8SGameDeployment(cluster.ClusterID, ns.Name) // nolint
		if err != nil {
			blog.Errorf("query k8s gamedeployment failed, err: %s", err.Error())
			return err
		}
		gameDeploymentList = append(gameDeploymentList, gameDeployments...)
	}
	blog.Infof("game deployment list: %v", gameDeploymentList)

	// 获取bk游戏部署列表
	bkGameDeployments, err := s.GetBkWorkloads(bkCluster.BizID, "gameDeployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		blog.Errorf("get bk gamedeployment failed, err: %s", err.Error())
		return err
	}

	// 将bk游戏部署列表转换为结构体列表
	for _, bkGameDeployment := range *bkGameDeployments {
		b := bkcmdbkube.GameDeployment{}
		err := common.InterfaceToStruct(bkGameDeployment, &b)
		if err != nil {
			blog.Errorf("convert bk gamedeployment failed, err: %s", err.Error())
			return err
		}

		bkGameDeploymentList = append(bkGameDeploymentList, b)
	}

	// 初始化待添加、更新、删除的游戏部署列表
	gameDeploymentToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	gameDeploymentToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
	gameDeploymentToDelete := make([]int64, 0)

	// 将游戏部署列表转换为map，便于后续查找
	gameDeploymentMap := make(map[string]*storage.GameDeployment)
	for _, v := range gameDeploymentList {
		gameDeploymentMap[v.Data.Namespace+v.Data.Name] = v
	}

	bkGameDeploymentMap := make(map[string]*bkcmdbkube.GameDeployment)
	for k, v := range bkGameDeploymentList {
		bkGameDeploymentMap[v.Namespace+v.Name] = &bkGameDeploymentList[k]

		// 如果在K8S中不存在该游戏部署，则标记为待删除
		if _, ok := gameDeploymentMap[v.Namespace+v.Name]; !ok {
			gameDeploymentToDelete = append(gameDeploymentToDelete, v.ID)
			blog.Infof("gameDeploymentToDelete: %s+%s+%s", bkCluster.Uid, v.Namespace, v.Name)
		}
	}

	// 遍历K8S中的游戏部署列表，判断是否需要添加或更新
	for k, v := range gameDeploymentMap {
		if _, ok := bkGameDeploymentMap[k]; !ok {
			// 如果在bk中不存在该游戏部署，则标记为待添加
			toAddData := s.GenerateBkGameDeployment(bkNamespaceMap[v.Data.Namespace], v)
			if toAddData != nil {
				if _, ok = gameDeploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID]; ok {
					gameDeploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID] =
						append(gameDeploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID], *toAddData)
					blog.Infof("gameDeploymentToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
				} else {
					gameDeploymentToAdd[bkNamespaceMap[v.Data.Namespace].BizID] =
						[]client.CreateBcsWorkloadRequestData{*toAddData}
					blog.Infof("gameDeploymentToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
				}
			}
		} else {
			// 如果在bk中存在该游戏部署，但内容不同，则标记为待更新
			needToUpdate, updateData := s.CompareGameDeployment(bkGameDeploymentMap[k], v)
			if needToUpdate {
				gameDeploymentToUpdate[bkGameDeploymentMap[k].ID] = updateData
				blog.Infof("gameDeploymentToUpdate: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
			}
		}
	}

	// 执行删除、添加、更新操作
	s.DeleteBkWorkloads(bkCluster, kind, &gameDeploymentToDelete, db) // nolint  not checked
	s.CreateBkWorkloads(bkCluster, kind, gameDeploymentToAdd, db)
	s.UpdateBkWorkloads(bkCluster, kind, &gameDeploymentToUpdate, db)

	return nil
}

// syncGameStatefulSets sync gamestatefulsets
// nolint funlen
func (s *Syncer) syncGameStatefulSets(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	kind := "gameStatefulSet"                  // 定义资源类型为游戏状态集
	storageCli, err := s.GetBcsStorageClient() // 获取Bcs存储客户端
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}

	// 获取bk命名空间列表
	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	// 将bk命名空间列表转换为map，便于后续查找
	bkNamespaceMap := make(map[string]*bkcmdbkube.Namespace)
	for k, v := range *bkNamespaceList {
		bkNamespaceMap[v.Name] = &(*bkNamespaceList)[k]
	}

	gameStatefulSetList := make([]*storage.GameStatefulSet, 0)     // 存储从K8S查询到的游戏状态集列表
	bkGameStatefulSetList := make([]bkcmdbkube.GameStatefulSet, 0) // 存储从bk查询到的游戏状态集列表

	// 遍历bk命名空间列表，查询每个命名空间下的游戏状态集
	for _, ns := range *bkNamespaceList {
		gameStatefulSets, err := storageCli.QueryK8SGameStatefulSet(cluster.ClusterID, ns.Name)
		if err != nil {
			blog.Errorf("query k8s gameStatefulSets failed, err: %s", err.Error())
			return err
		}
		gameStatefulSetList = append(gameStatefulSetList, gameStatefulSets...)
	}
	blog.Infof("gamestatefulset list: %v", gameStatefulSetList)

	// 从bk获取游戏状态集列表
	bkGameStatefulSets, err :=
		s.GetBkWorkloads(bkCluster.BizID, "gameStatefulSet", &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "cluster_uid",
					Operator: "in",
					Value:    []string{bkCluster.Uid},
				},
			},
		}, true, db)
	if err != nil {
		blog.Errorf("get bk gamestatefulset failed, err: %s", err.Error())
		return err
	}

	// 将bk游戏状态集列表转换为结构体列表
	for _, bkGameStatefulSet := range *bkGameStatefulSets {
		b := bkcmdbkube.GameStatefulSet{}
		err := common.InterfaceToStruct(bkGameStatefulSet, &b)
		if err != nil {
			blog.Errorf("convert bk gamestatefulset failed, err: %s", err.Error())
			return err
		}

		bkGameStatefulSetList = append(bkGameStatefulSetList, b)
	}

	// 初始化待添加、待更新、待删除的游戏状态集列表
	gameStatefulSetToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	gameStatefulSetToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
	gameStatefulSetToDelete := make([]int64, 0)

	// 将K8S游戏状态集列表转换为map，便于后续查找
	gameStatefulSetMap := make(map[string]*storage.GameStatefulSet)
	for _, v := range gameStatefulSetList {
		gameStatefulSetMap[v.Data.Namespace+v.Data.Name] = v
	}

	// 将bk游戏状态集列表转换为map，便于后续查找
	bkGameStatefulSetMap := make(map[string]*bkcmdbkube.GameStatefulSet)
	for k, v := range bkGameStatefulSetList {
		bkGameStatefulSetMap[v.Namespace+v.Name] = &bkGameStatefulSetList[k]

		// 如果K8S中不存在该游戏状态集，则标记为待删除
		if _, ok := gameStatefulSetMap[v.Namespace+v.Name]; !ok {
			gameStatefulSetToDelete = append(gameStatefulSetToDelete, v.ID)
			blog.Infof("gameStatefulSetToDelete: %s+%s+%s", bkCluster.Uid, v.Namespace, v.Name)
		}
	}

	// 遍历K8S游戏状态集map，判断是否需要添加或更新
	for k, v := range gameStatefulSetMap {
		if _, ok := bkGameStatefulSetMap[k]; !ok {
			// 如果bk中不存在该游戏状态集，则标记为待添加
			toAddData := s.GenerateBkGameStatefulSet(bkNamespaceMap[v.Data.Namespace], v)

			if toAddData != nil {
				if _, ok = gameStatefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID]; ok {
					gameStatefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] =
						append(gameStatefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID], *toAddData)
					blog.Infof("gameStatefulSetToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
				} else {
					gameStatefulSetToAdd[bkNamespaceMap[v.Data.Namespace].BizID] =
						[]client.CreateBcsWorkloadRequestData{*toAddData}
					blog.Infof("gameStatefulSetToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
				}
			}
		} else {
			// 如果bk中存在该游戏状态集，但内容不同，则标记为待更新
			needToUpdate, updateData := s.CompareGameStatefulSet(bkGameStatefulSetMap[k], v)
			if needToUpdate {
				gameStatefulSetToUpdate[bkGameStatefulSetMap[k].ID] = updateData
				blog.Infof("gameStatefulSetToUpdate: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
			}
		}
	}

	// 执行删除、添加、更新操作
	s.DeleteBkWorkloads(bkCluster, kind, &gameStatefulSetToDelete, db) // nolint  not checked
	s.CreateBkWorkloads(bkCluster, kind, gameStatefulSetToAdd, db)
	s.UpdateBkWorkloads(bkCluster, kind, &gameStatefulSetToUpdate, db)

	return nil
}

// syncWorkloadPods sync workloadPods
// nolint
func (s *Syncer) syncWorkloadPods(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	kind := "pods" // nolint

	// 获取bk集群中的命名空间列表
	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return err
	}

	// 将命名空间列表转换为map，便于后续查找
	bkNamespaceMap := make(map[string]*bkcmdbkube.Namespace)
	for k, v := range *bkNamespaceList {
		bkNamespaceMap[v.Name] = &(*bkNamespaceList)[k]
	}

	// 初始化一个空的Pods工作负载列表
	bkWorkloadPodsList := make([]bkcmdbkube.PodsWorkload, 0)

	// 获取bk集群中的工作负载Pods
	bkWorkloadPods, err := s.GetBkWorkloads(bkCluster.BizID, "pods", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		blog.Errorf("get bk workload pods failed, err: %s", err.Error())
		return err
	}

	// 将获取到的工作负载Pods转换为结构体列表
	for _, workloadPods := range *bkWorkloadPods {
		p := bkcmdbkube.PodsWorkload{}
		err := common.InterfaceToStruct(workloadPods, &p)
		if err != nil {
			blog.Errorf("convert bk workload pods failed, err: %s", err.Error())
			return err
		}
		bkWorkloadPodsList = append(bkWorkloadPodsList, p)
	}

	// 初始化添加、更新、删除工作负载Pods的数据结构
	workloadPodsToAdd := make(map[int64][]client.CreateBcsWorkloadRequestData, 0)
	workloadPodsToUpdate := make(map[int64]*client.UpdateBcsWorkloadRequestData, 0)
	workloadPodsToDelete := make([]int64, 0)

	// 将工作负载Pods列表转换为map，便于后续查找
	bkWorkloadPodsMap := make(map[string]*bkcmdbkube.PodsWorkload)
	for k, v := range bkWorkloadPodsList {
		bkWorkloadPodsMap[v.Namespace+v.Name] = &bkWorkloadPodsList[k]
	}

	// 遍历命名空间map，判断是否需要添加或更新工作负载Pods
	for k, v := range bkNamespaceMap {
		if _, ok := bkWorkloadPodsMap[k+"pods"]; !ok {
			// 如果不存在，则需要添加
			toAddData := s.GenerateBkWorkloadPods(v)
			if toAddData != nil {
				if _, ok = workloadPodsToAdd[v.BizID]; ok {
					workloadPodsToAdd[v.BizID] = append(workloadPodsToAdd[v.BizID], *toAddData)
					blog.Infof("workloadPodsToAdd: %s+%s+%s", bkCluster.Uid, v.Name, "Pods")
				} else {
					workloadPodsToAdd[v.BizID] = []client.CreateBcsWorkloadRequestData{*toAddData}
				}
			}
		} else {
			// 如果存在，则比较是否需要更新
			needToUpdate, updateData := s.CompareBkWorkloadPods(bkWorkloadPodsMap[k+"pods"])
			if needToUpdate {
				workloadPodsToUpdate[bkWorkloadPodsMap[k+"pods"].ID] = updateData
				blog.Infof("workloadPodsToUpdate: %s+%s+%s", bkCluster.Uid, v.Name, "Pods")
			}
		}
	}

	// 删除不需要的Pods工作负载
	s.DeleteBkWorkloads(bkCluster, kind, &workloadPodsToDelete, db) // nolint  not checked
	// 添加新的Pods工作负载
	s.CreateBkWorkloads(bkCluster, kind, workloadPodsToAdd, db)
	// 更新已有的Pods工作负载
	s.UpdateBkWorkloads(bkCluster, kind, &workloadPodsToUpdate, db)

	return nil
}

// nolint
// getBkNsMap 获取与特定集群关联的所有命名空间的映射
func (s *Syncer) getBkNsMap(
	cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) (map[string]*bkcmdbkube.Namespace, error) {
	// 调用GetBkNamespaces方法获取与bkCluster.Uid关联的命名空间列表
	bkNamespaceList, err := s.GetBkNamespaces(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",           // 过滤条件字段为集群UID
				Operator: "in",                    // 使用"in"操作符
				Value:    []string{bkCluster.Uid}, // 匹配的UID值
			},
		},
	}, true, db)
	if err != nil {
		// 如果获取命名空间列表失败，则记录错误日志并返回错误
		blog.Errorf("get bk namespace failed, err: %s", err.Error())
		return nil, err
	}

	// 创建一个新的映射，用于存储命名空间名称到命名空间对象的映射
	bkNsMap := make(map[string]*bkcmdbkube.Namespace)
	// 遍历命名空间列表，填充映射
	for k, v := range *bkNamespaceList {
		bkNsMap[v.Name] = &(*bkNamespaceList)[k]
	}

	// 返回填充好的命名空间映射
	return bkNsMap, nil
}

// nolint
// getBkDeploymentMap 获取与特定集群相关的部署信息的映射
func (s *Syncer) getBkDeploymentMap(
	cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) (map[string]*bkcmdbkube.Deployment, error) {

	// 初始化一个空的部署列表
	bkDeploymentList := make([]bkcmdbkube.Deployment, 0)

	// 调用GetBkWorkloads函数获取特定业务ID和集群UID下的部署列表
	bkDeployments, err := s.GetBkWorkloads(bkCluster.BizID, "deployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid", // 过滤条件：集群UID
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		blog.Errorf("get bk deployment failed, err: %s", err.Error())
		return nil, err // 如果获取部署列表失败，则返回错误
	}

	// 遍历获取到的部署列表，将每个部署项转换为bkcmdbkube.Deployment结构体
	for _, bkDeployment := range *bkDeployments {
		b := bkcmdbkube.Deployment{}
		err := common.InterfaceToStruct(bkDeployment, &b)
		if err != nil {
			blog.Errorf("convert bk deployment failed, err: %s", err.Error())
			return nil, err // 如果转换部署项失败，则返回错误
		}

		// 将转换后的部署项添加到部署列表中
		bkDeploymentList = append(bkDeploymentList, b)
	}

	// 初始化一个空的部署映射
	bkDeploymentMap := make(map[string]*bkcmdbkube.Deployment)
	// 遍历部署列表，将每个部署项添加到映射中，键为命名空间和名称的组合
	for k, v := range bkDeploymentList {
		bkDeploymentMap[v.Namespace+v.Name] = &bkDeploymentList[k]
	}

	// 返回部署映射
	return bkDeploymentMap, nil
}

// nolint
// getBkStatefulSetMap 获取bkcmdbkube中的StatefulSet列表，并将其转换为map
// 参数:
//
//	cluster: cmp集群对象
//	bkCluster: bkcmdbkube集群对象
//	db: gorm数据库连接对象
//
// 返回值:
//
//	map[string]*bkcmdbkube.StatefulSet: StatefulSet对象的map，键为命名空间+名称
//	error: 如果发生错误，返回错误信息
func (s *Syncer) getBkStatefulSetMap(
	cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) (map[string]*bkcmdbkube.StatefulSet, error) {
	// 初始化StatefulSet列表
	bkStatefulSetList := make([]bkcmdbkube.StatefulSet, 0)

	// 调用GetBkWorkloads方法获取StatefulSet列表
	bkStatefulSets, err := s.GetBkWorkloads(bkCluster.BizID, "statefulSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		// 如果获取StatefulSet列表失败，记录错误日志并返回错误
		blog.Errorf("get bk statefulset failed, err: %s", err.Error())
		return nil, err
	}

	// 遍历获取到的StatefulSet列表，并将其转换为bkcmdbkube.StatefulSet结构体
	for _, bkStatefulSet := range *bkStatefulSets {
		b := bkcmdbkube.StatefulSet{}
		err := common.InterfaceToStruct(bkStatefulSet, &b)
		if err != nil {
			// 如果转换失败，记录错误日志并返回错误
			blog.Errorf("convert bk statefulset failed, err: %s", err.Error())
			return nil, err
		}

		// 将转换后的StatefulSet对象添加到列表中
		bkStatefulSetList = append(bkStatefulSetList, b)
	}

	// 初始化StatefulSet对象的map
	bkStatefulSetMap := make(map[string]*bkcmdbkube.StatefulSet)
	// 遍历StatefulSet列表，将其转换为map，键为命名空间+名称
	for k, v := range bkStatefulSetList {
		bkStatefulSetMap[v.Namespace+v.Name] = &bkStatefulSetList[k]
	}
	// 返回StatefulSet对象的map
	return bkStatefulSetMap, nil
}

// nolint
// getBkDaemonSetMap 获取指定集群的DaemonSet映射
func (s *Syncer) getBkDaemonSetMap(
	cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) (map[string]*bkcmdbkube.DaemonSet, error) {
	// 初始化DaemonSet列表
	bkDaemonSetList := make([]bkcmdbkube.DaemonSet, 0)

	// 调用GetBkWorkloads方法获取DaemonSet列表
	bkDaemonSets, err := s.GetBkWorkloads(bkCluster.BizID, "daemonSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		// 如果获取DaemonSet失败，记录错误并返回
		blog.Errorf("get bk daemonset failed, err: %s", err.Error())
		return nil, err
	}

	// 遍历获取到的DaemonSet列表，并转换为结构体
	for _, bkDaemonSet := range *bkDaemonSets {
		b := bkcmdbkube.DaemonSet{}
		err := common.InterfaceToStruct(bkDaemonSet, &b)
		if err != nil {
			// 如果转换失败，记录错误并返回
			blog.Errorf("convert bk daemonset failed, err: %s", err.Error())
			return nil, err
		}

		// 将转换后的DaemonSet添加到列表中
		bkDaemonSetList = append(bkDaemonSetList, b)
	}

	// 初始化DaemonSet映射
	bkDaemonSetMap := make(map[string]*bkcmdbkube.DaemonSet)
	// 遍历DaemonSet列表，生成以命名空间和名称为键的映射
	for k, v := range bkDaemonSetList {
		bkDaemonSetMap[v.Namespace+v.Name] = &bkDaemonSetList[k]
	}
	// 返回DaemonSet映射
	return bkDaemonSetMap, nil
}

// nolint
// getBkGameDeploymentMap 获取与特定集群相关的游戏部署映射
func (s *Syncer) getBkGameDeploymentMap(
	cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) (map[string]*bkcmdbkube.GameDeployment, error) {

	// 初始化游戏部署列表
	bkGameDeploymentList := make([]bkcmdbkube.GameDeployment, 0)

	// 调用GetBkWorkloads方法获取特定业务ID和类型为"gameDeployment"的工作负载列表，
	// 并且过滤条件是集群UID匹配
	bkGameDeployments, err := s.GetBkWorkloads(bkCluster.BizID, "gameDeployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
	if err != nil {
		// 如果获取工作负载失败，记录错误日志并返回错误
		blog.Errorf("get bk gamedeployment failed, err: %s", err.Error())
		return nil, err
	}

	// 遍历获取到的游戏部署列表
	for _, bkGameDeployment := range *bkGameDeployments {
		// 创建一个新的GameDeployment结构体实例
		b := bkcmdbkube.GameDeployment{}
		// 将接口类型的数据转换为具体的GameDeployment结构体
		err := common.InterfaceToStruct(bkGameDeployment, &b)
		if err != nil {
			// 如果转换失败，记录错误日志并返回错误
			blog.Errorf("convert bk gamedeployment failed, err: %s", err.Error())
			return nil, err
		}

		// 将转换后的GameDeployment添加到列表中
		bkGameDeploymentList = append(bkGameDeploymentList, b)
	}

	// 初始化游戏部署映射
	bkGameDeploymentMap := make(map[string]*bkcmdbkube.GameDeployment)
	// 遍历游戏部署列表，构建以命名空间和名称组合为键的映射
	for k, v := range bkGameDeploymentList {
		bkGameDeploymentMap[v.Namespace+v.Name] = &bkGameDeploymentList[k]
	}
	// 返回构建好的游戏部署映射
	return bkGameDeploymentMap, nil
}

// nolint
// getBkGameStatefulSetMap 获取bkcmdbkube中的GameStatefulSet映射
// 参数:
//
//	cluster: cmp集群对象
//	bkCluster: bkcmdbkube集群对象
//	db: gorm数据库连接对象
//
// 返回值:
//
//	map[string]*bkcmdbkube.GameStatefulSet: 游戏有状态集的映射，键为命名空间加名称
//	error: 如果操作过程中发生错误，则返回错误信息
func (s *Syncer) getBkGameStatefulSetMap(
	cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) (map[string]*bkcmdbkube.GameStatefulSet, error) {
	// 初始化一个空的GameStatefulSet列表
	bkGameStatefulSetList := make([]bkcmdbkube.GameStatefulSet, 0)

	// 调用GetBkWorkloads方法获取bkcmdbkube中的GameStatefulSet列表
	// 参数包括业务ID、资源类型、过滤条件等
	bkGameStatefulSets, err :=
		s.GetBkWorkloads(bkCluster.BizID, "gameStatefulSet", &client.PropertyFilter{
			Condition: "AND",
			Rules: []client.Rule{
				{
					Field:    "cluster_uid",           // 过滤条件：集群UID
					Operator: "in",                    // 操作符：在列表中
					Value:    []string{bkCluster.Uid}, // 值：当前集群的UID
				},
			},
		}, true, db)
	if err != nil {
		// 如果获取过程中发生错误，记录错误日志并返回错误
		blog.Errorf("get bk gamestatefulset failed, err: %s", err.Error())
		return nil, err
	}

	// 遍历获取到的GameStatefulSet列表
	for _, bkGameStatefulSet := range *bkGameStatefulSets {
		// 初始化一个新的GameStatefulSet结构体
		b := bkcmdbkube.GameStatefulSet{}
		// 将接口类型的数据转换为结构体类型
		err := common.InterfaceToStruct(bkGameStatefulSet, &b)
		if err != nil {
			// 如果转换过程中发生错误，记录错误日志并返回错误
			blog.Errorf("convert bk gamestatefulset failed, err: %s", err.Error())
			return nil, err
		}

		// 将转换后的结构体添加到列表中
		bkGameStatefulSetList = append(bkGameStatefulSetList, b)
	}

	// 初始化一个空的映射，用于存储命名空间加名称到GameStatefulSet的映射
	bkGameStatefulSetMap := make(map[string]*bkcmdbkube.GameStatefulSet)
	// 遍历列表，构建映射
	for k, v := range bkGameStatefulSetList {
		// 使用命名空间加名称作为键，GameStatefulSet的指针作为值
		bkGameStatefulSetMap[v.Namespace+v.Name] = &bkGameStatefulSetList[k]
	}
	// 返回构建好的映射
	return bkGameStatefulSetMap, nil
}

// nolint
func (s *Syncer) getBkNodeMap(
	cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) (map[string]*bkcmdbkube.Node, error) {
	bkNodeList, err := s.GetBkNodes(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, true, db)
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

// nolint
func (s *Syncer) getBkWorkloadPods(
	cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster,
	pod *storage.Pod, db *gorm.DB) (*bkcmdbkube.PodsWorkload, error) {
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
	}, true, db)
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
		s.getPodOperatorWorkload(workloadLabels, &operator)
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

func (s *Syncer) getPodOperatorWorkload(workloadLabels *map[string]string, operator *[]string) {
	if creator, creatorOk := (*workloadLabels)["io.tencent.paas.creator"]; creatorOk && (creator != "") {
		*operator = append(*operator, creator)
	} else if creator, creatorOk = (*workloadLabels)["io．tencent．paas．creator"]; creatorOk && (creator != "") {
		*operator = append(*operator, creator)
	} else if updater, updaterOk := (*workloadLabels)["io.tencent.paas.updater"]; updaterOk && (updater != "") {
		*operator = append(*operator, updater)
	} else if updater, updaterOk = (*workloadLabels)["io．tencent．paas．updator"]; updaterOk && (updater != "") {
		*operator = append(*operator, updater)
	}
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
	if bkWorkloadPods != nil {
		workloadID = bkWorkloadPods.ID
	}

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
			switch rsOwnerRef.Kind { // nolint
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
// nolint
func (s *Syncer) SyncPods(cluster *cmp.Cluster, bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	storageCli, err := s.GetBcsStorageClient()
	if err != nil {
		blog.Errorf("get bcs storage client failed, err: %s", err.Error())
	}

	bkNsMap, err := s.getBkNsMap(cluster, bkCluster, db)
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

	bkDeploymentMap, err := s.getBkDeploymentMap(cluster, bkCluster, db)
	if err != nil {
		return err
	}

	bkStatefulSetMap, err := s.getBkStatefulSetMap(cluster, bkCluster, db)
	if err != nil {
		return err
	}

	bkDaemonSetMap, err := s.getBkDaemonSetMap(cluster, bkCluster, db)
	if err != nil {
		return err
	}

	bkGameDeploymentMap, err := s.getBkGameDeploymentMap(cluster, bkCluster, db)
	if err != nil {
		return err
	}

	bkGameStatefulSetMap, err := s.getBkGameStatefulSetMap(cluster, bkCluster, db)
	if err != nil {
		return err
	}

	bkNodeMap, err := s.getBkNodeMap(cluster, bkCluster, db)
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
	}, true, db)
	//blog.Infof("bkPodList: %v, len: %d", *bkPodList, len(*bkPodList))
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
	//blog.Infof("podMap: %v", podMap)

	bkPodMap := make(map[string]*bkcmdbkube.Pod)
	for k, v := range *bkPodList {
		bkPodMap[v.NameSpace+*v.Name] = &(*bkPodList)[k]

		if _, ok := podMap[v.NameSpace+*v.Name]; !ok {
			podToDelete = append(podToDelete, v.ID)
			blog.Infof("podToDelete: %s+%s", v.NameSpace, *v.Name)
		} else {
			if s.ComparePod(&v, podMap[v.NameSpace+*v.Name], db) {
				podToDelete = append(podToDelete, v.ID)
				blog.Infof("podToDelete: %s+%s", v.NameSpace, *v.Name)
				delete(bkPodMap, v.NameSpace+*v.Name)
			}
		}
	}

	s.syncPodsCheck(podMap, bkPodMap,
		bkDeploymentMap, bkStatefulSetMap,
		bkDaemonSetMap, bkGameDeploymentMap,
		bkGameStatefulSetMap, bkNsMap,
		bkNodeMap, &podToDelete,
		podToAdd, cluster,
		bkCluster, db, storageCli)

	//blog.Infof("podToDelete: %v", podToDelete)
	s.DeleteBkPods(bkCluster, &podToDelete, db) // nolint  not checked
	s.CreateBkPods(bkCluster, podToAdd, db)

	return err
}

// nolint
func (s *Syncer) syncPodsCheck(podMap map[string]*storage.Pod, bkPodMap map[string]*bkcmdbkube.Pod,
	bkDeploymentMap map[string]*bkcmdbkube.Deployment, bkStatefulSetMap map[string]*bkcmdbkube.StatefulSet,
	bkDaemonSetMap map[string]*bkcmdbkube.DaemonSet, bkGameDeploymentMap map[string]*bkcmdbkube.GameDeployment,
	bkGameStatefulSetMap map[string]*bkcmdbkube.GameStatefulSet, bkNsMap map[string]*bkcmdbkube.Namespace,
	bkNodeMap map[string]*bkcmdbkube.Node, podToDelete *[]int64,
	podToAdd map[int64][]client.CreateBcsPodRequestDataPod, cluster *cmp.Cluster,
	bkCluster *bkcmdbkube.Cluster, db *gorm.DB, storageCli bcsapi.Storage) {

	for k, v := range podMap {
		// var operator []string
		if _, ok := bkPodMap[k]; !ok {
			if v.Data.Status.Phase != corev1.PodRunning {
				continue
			}
			bkWorkloadPods, podsErr := s.getBkWorkloadPods(cluster, bkCluster, v, db)
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

			tnodeID := bkNodeMap[v.Data.Spec.NodeName].ID
			thostID := bkNodeMap[v.Data.Spec.NodeName].HostID

			blog.Infof("NodeName: %s, nodeID: %d, hostID: %d", v.Data.Spec.NodeName, nodeID, hostID)

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

				cName := container.Name
				cImage := container.Image
				cArgs := container.Args

				containers = append(containers, bkcmdbkube.ContainerBaseFields{
					Name:        &cName,
					Image:       &cImage,
					ContainerID: &containerID,
					Ports:       &ports,
					Args:        &cArgs,
					Environment: &env,
					Mounts:      &mounts,
				})
			}

			if _, ok = podToAdd[bkNsMap[v.Data.Namespace].BizID]; ok {
				podToAdd[bkNsMap[v.Data.Namespace].BizID] =
					append(podToAdd[bkNsMap[v.Data.Namespace].BizID], client.CreateBcsPodRequestDataPod{
						Spec: &client.CreateBcsPodRequestPodSpec{
							ClusterID:    &bkCluster.ID,
							NameSpaceID:  &bkNsMap[v.Data.Namespace].ID,
							WorkloadKind: &workloadKind,
							WorkloadID:   &workloadID,
							NodeID:       &tnodeID,
							Ref: &bkcmdbkube.Reference{
								Kind: bkcmdbkube.WorkloadType(workloadKind),
								Name: workloadName,
								ID:   workloadID,
							},
						},

						Name:       &v.Data.Name,
						HostID:     &thostID,
						Priority:   v.Data.Spec.Priority,
						Labels:     &v.Data.Labels,
						IP:         &v.Data.Status.PodIP,
						IPs:        &podIPs,
						Containers: &containers,
						Operator:   &operator,
					})
				blog.Infof("podToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
			} else {
				podToAdd[bkNsMap[v.Data.Namespace].BizID] = []client.CreateBcsPodRequestDataPod{
					client.CreateBcsPodRequestDataPod{
						Spec: &client.CreateBcsPodRequestPodSpec{
							ClusterID:    &bkCluster.ID,
							NameSpaceID:  &bkNsMap[v.Data.Namespace].ID,
							WorkloadKind: &workloadKind,
							WorkloadID:   &workloadID,
							NodeID:       &tnodeID,
							Ref: &bkcmdbkube.Reference{
								Kind: bkcmdbkube.WorkloadType(workloadKind),
								Name: workloadName,
								ID:   workloadID,
							},
						},

						Name:       &v.Data.Name,
						HostID:     &thostID,
						Priority:   v.Data.Spec.Priority,
						Labels:     &v.Data.Labels,
						IP:         &v.Data.Status.PodIP,
						IPs:        &podIPs,
						Containers: &containers,
						Operator:   &operator,
					},
				}
				blog.Infof("podToAdd: %s+%s+%s", bkCluster.Uid, v.Data.Namespace, v.Data.Name)
			}

		} else {
			if v.Data.Status.Phase != corev1.PodRunning {
				*podToDelete = append(*podToDelete, bkPodMap[k].ID)
				blog.Infof("podToDelete: %s+%s", bkPodMap[k].NameSpace, *bkPodMap[k].Name)
			}
		}
	}
}

// SyncStore sync store
func (s *Syncer) SyncStore(bkCluster *bkcmdbkube.Cluster, force bool) error {
	path := "/data/bcs/bcs-bkcmdb-synchronizer/db/" + bkCluster.Uid + ".db"

	db := sqlite.Open(path)
	if db == nil {
		blog.Errorf("open db failed, path: %s", path)
		return fmt.Errorf("open db failed, path: %s", path)
	}
	s.SyncStoreMigrate(db)

	err := s.SyncStoreNode(bkCluster, db)
	if err != nil {
		blog.Errorf("sync store node failed, err: %s", err.Error())
	}

	if !force {
		if s.SyncStoreSynced(db) {
			blog.Infof("SyncStore synced skip.")
			return nil
		}
	}

	err = s.SyncStoreCluster(bkCluster, db)
	if err != nil {
		blog.Errorf("sync store cluster failed, err: %s", err.Error())
	}

	err = s.SyncStoreNamespace(bkCluster, db)
	if err != nil {
		blog.Errorf("sync store namespace failed, err: %s", err.Error())
	}

	err = s.SyncStorePod(bkCluster, db)
	if err != nil {
		blog.Errorf("sync store pod failed, err: %s", err.Error())
	}

	err = s.SyncStoreContainer(bkCluster, db)
	if err != nil {
		blog.Errorf("sync store container failed, err: %s", err.Error())
	}

	err = s.SyncStoreNode(bkCluster, db)
	if err != nil {
		blog.Errorf("sync store node failed, err: %s", err.Error())
	}

	err = s.SyncStoreWorkload(bkCluster, db)
	if err != nil {
		blog.Errorf("sync store workload failed, err: %s", err.Error())
	}

	err = s.SyncDone(bkCluster, db)
	if err != nil {
		blog.Errorf("sync done failed, err: %s", err.Error())
	}

	return nil
}

// SyncDone sync done
func (s *Syncer) SyncDone(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.SyncMigrate(db)
	if err != nil {
		blog.Errorf("migrate sync failed, err: %s", err.Error())
		return fmt.Errorf("migrate sync failed, err: %s", err.Error())
	}
	err = db.Create(&model.Sync{
		FullSyncTime: time.Now().Unix(),
	}).Error
	if err != nil {
		blog.Errorf("create sync failed, err: %s", err.Error())
		return fmt.Errorf("create sync failed, err: %s", err.Error())
	}
	return nil
}

// SyncStoreSynced sync store synced
func (s *Syncer) SyncStoreSynced(db *gorm.DB) bool {
	if err := db.Where("full_sync_time > 0").First(&model.Sync{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
		blog.Errorf("get sync failed, err: %s", err.Error())
		return false
	}
	return true
}

// SyncStoreMigrate sync store migrate
func (s *Syncer) SyncStoreMigrate(db *gorm.DB) {
	_ = model.ClusterMigrate(db)
	_ = model.SyncMigrate(db)
	_ = model.NodeMigrate(db)
	_ = model.NamespaceMigrate(db)
	_ = model.DeploymentMigrate(db)
	_ = model.StatefulSetMigrate(db)
	_ = model.DaemonSetMigrate(db)
	_ = model.GameDeploymentMigrate(db)
	_ = model.GameStatefulSetMigrate(db)
	_ = model.PodsWorkloadMigrate(db)
	_ = model.PodMigrate(db)
	_ = model.ContainerMigrate(db)
}

// SyncStoreCluster sync store cluster
func (s *Syncer) SyncStoreCluster(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.ClusterMigrate(db)
	if err != nil {
		blog.Errorf("migrate cluster failed, err: %s", err.Error())
		return fmt.Errorf("migrate cluster failed, err: %s", err.Error())
	}
	clusterMarshal, err := json.Marshal(bkCluster)
	if err != nil {
		blog.Errorf("marshal cluster failed, err: %s", err.Error())
		return fmt.Errorf("marshal cluster failed, err: %s", err.Error())
	}
	var cluster model.Cluster
	err = json.Unmarshal(clusterMarshal, &cluster)
	if err != nil {
		blog.Errorf("unmarshal cluster failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal cluster failed, err: %s", err.Error())
	}

	var existingCluster model.Cluster
	err = db.Where("id = ?", bkCluster.ID).First(&existingCluster).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = db.Create(&cluster).Error
			if err != nil {
				blog.Errorf("create cluster failed, err: %s", err.Error())
				return fmt.Errorf("create cluster failed, err: %s", err.Error())
			}
		} else {
			blog.Errorf("get cluster failed, err: %s", err.Error())
		}
	} else {
		err = db.Save(&cluster).Error
		if err != nil {
			blog.Errorf("update cluster failed, err: %s", err.Error())
			return fmt.Errorf("update cluster failed, err: %s", err.Error())
		}
		blog.Infof("update cluster success")
	}
	return nil
}

// SyncStoreNamespace sync store namespace
func (s *Syncer) SyncStoreNamespace(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.NamespaceMigrate(db)
	if err != nil {
		blog.Errorf("migrate namespace failed, err: %s", err.Error())
		return fmt.Errorf("migrate namespace failed, err: %s", err.Error())
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
	}, false, nil)

	if err != nil {
		blog.Errorf("get bk namespaces failed, err: %s", err.Error())
		return fmt.Errorf("get bk namespaces failed, err: %s", err.Error())
	}

	namespaceMarshal, err := json.Marshal(bkNamespaceList)
	if err != nil {
		blog.Errorf("marshal namespace failed, err: %s", err.Error())
		return fmt.Errorf("marshal namespace failed, err: %s", err.Error())
	}
	var ns []model.Namespace
	err = json.Unmarshal(namespaceMarshal, &ns)
	if err != nil {
		blog.Errorf("unmarshal namespace failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal namespace failed, err: %s", err.Error())
	}
	for _, n := range ns {
		var existingNamespace model.Namespace
		err = db.Where("id = ?", n.ID).First(&existingNamespace).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = db.Create(&n).Error
				if err != nil {
					blog.Errorf("syncStore create namespace failed, err: %s", err.Error())
				}
			} else {
				blog.Errorf("syncStore get namespace failed, err: %s", err.Error())
			}
		} else {
			err = db.Save(&n).Error
			if err != nil {
				blog.Errorf("syncStore update namespace failed, err: %s", err.Error())
			} else {
				blog.Infof("syncStore update namespace success")
			}
		}
	}

	return nil
}

// SyncStorePod 同步 Pod 信息到数据库
func (s *Syncer) SyncStorePod(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.PodMigrate(db)
	if err != nil {
		blog.Errorf("migrate pod failed, err: %s", err.Error())
		return fmt.Errorf("migrate pod failed, err: %s", err.Error())
	}
	// err = model.ContainerMigrate(db)
	// if err != nil {
	//	 blog.Errorf("migrate container failed, err: %s", err.Error())
	//	 return fmt.Errorf("migrate container failed, err: %s", err.Error())
	// }

	bkPodList, err := s.GetBkPods(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, false, nil)
	// blog.Infof("bkPodList: %v, len: %d", *bkPodList, len(*bkPodList))
	if err != nil {
		blog.Errorf("get bk pod failed, err: %s", err.Error())
	}

	podMarshal, err := json.Marshal(bkPodList)
	if err != nil {
		blog.Errorf("marshal bk pod failed, err: %s", err.Error())
		return fmt.Errorf("marshal bk pod failed, err: %s", err.Error())
	}
	var pods []model.Pod

	err = json.Unmarshal(podMarshal, &pods)
	if err != nil {
		blog.Errorf("unmarshal bk pod failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal bk pod failed, err: %s", err.Error())
	}

	for _, pod := range pods {
		var existingPod model.Pod
		if errP := db.Where("id = ?", pod.ID).First(&existingPod).Error; errP != nil {
			if errors.Is(errP, gorm.ErrRecordNotFound) {
				if errPC := db.Create(&pod).Error; errPC != nil {
					blog.Errorf("syncStore create pod failed, err: %s", errPC.Error())
				}
			} else {
				blog.Errorf("syncStore get pod failed, err: %s", errP.Error())
			}
		} else {
			if errPS := db.Save(&pod).Error; errPS != nil {
				blog.Errorf("syncStore update pod failed, err: %s", errPS.Error())
			} else {
				blog.Infof("syncStore update pod success, pod: %d", pod.ID)
			}
		}
		// bkContainers, err := s.Syncer.CMDBClient.GetBcsContainer(&client.GetBcsContainerRequest{
		//	CommonRequest: client.CommonRequest{
		//		BKBizID: pod.BizID,
		//		Page: client.Page{
		//			Limit: 200,
		//			Start: 0,
		//		},
		//	},
		//	BkPodID: pod.ID,
		// })
		//
		// if err != nil {
		//	blog.Errorf("syncStore GetBcsContainer err: %v", err)
		// } else {
		//	cs, err := json.Marshal(bkContainers)
		//	if err != nil {
		//		blog.Errorf("syncStore json marshal err: %v", err)
		//		continue
		//	}
		//	var containers []model.Container
		//	err = json.Unmarshal(cs, &containers)
		//	if err != nil {
		//		blog.Errorf("syncStore json unmarshal err: %v", err)
		//		continue
		//	}
		//	for _, container := range containers {
		//		db.Create(&container)
		//	}
		// }
	}
	return nil
}

// SyncStoreContainer 同步容器信息到数据库
func (s *Syncer) SyncStoreContainer(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.ContainerMigrate(db)
	if err != nil {
		blog.Errorf("migrate container failed, err: %s", err.Error())
		return fmt.Errorf("migrate container failed, err: %s", err.Error())
	}

	bkContainerList, err := s.GetBkContainers(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "bk_cluster_id",
				Operator: "in",
				Value:    []int64{bkCluster.ID},
			},
		},
	}, false, nil)
	blog.Infof("bkContainerList: %v, len: %d", *bkContainerList, len(*bkContainerList))
	if err != nil {
		blog.Errorf("get bk container failed, err: %s", err.Error())
	}

	containerMarshal, err := json.Marshal(bkContainerList)
	if err != nil {
		blog.Errorf("marshal bk container failed, err: %s", err.Error())
		return fmt.Errorf("marshal bk container failed, err: %s", err.Error())
	}
	var containers []model.Container

	err = json.Unmarshal(containerMarshal, &containers)
	if err != nil {
		blog.Errorf("unmarshal bk container failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal bk container failed, err: %s", err.Error())
	}

	for _, c := range containers {
		var existingContainer model.Container
		if errC := db.Where("id = ?", c.ID).First(&existingContainer).Error; errC != nil {
			if errors.Is(errC, gorm.ErrRecordNotFound) {
				if errCC := db.Create(&c).Error; errCC != nil {
					blog.Errorf("syncStore create container failed, err: %s", errCC.Error())
				}
			} else {
				blog.Errorf("syncStore get container failed, err: %s", errC.Error())
			}
		} else {
			if errCS := db.Save(&c).Error; errCS != nil {
				blog.Errorf("syncStore update container failed, err: %s", errCS.Error())
			} else {
				blog.Infof("syncStore update container success, pod: %d", c.ID)
			}
		}
	}
	return nil
}

// SyncStoreNode 同步节点信息到数据库
func (s *Syncer) SyncStoreNode(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.NodeMigrate(db)
	if err != nil {
		blog.Errorf("migrate node failed, err: %s", err.Error())
		return fmt.Errorf("migrate node failed, err: %s", err.Error())
	}

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
	}, false, nil)

	if err != nil {
		blog.Errorf("get bk node failed, err: %s", err.Error())
		return fmt.Errorf("get bk node failed, err: %s", err.Error())
	}

	nodeMarshal, err := json.Marshal(bkNodeList)
	if err != nil {
		blog.Errorf("marshal bk node failed, err: %s", err.Error())
		return fmt.Errorf("marshal bk node failed, err: %s", err.Error())
	}

	var nodes []model.Node

	err = json.Unmarshal(nodeMarshal, &nodes)
	if err != nil {
		blog.Errorf("unmarshal bk node failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal bk node failed, err: %s", err.Error())
	}

	var nodeIDs []int64
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.ID)
		var existNode model.Node
		if errN := db.Where("id = ?", node.ID).First(&existNode).Error; errN != nil {
			if errors.Is(errN, gorm.ErrRecordNotFound) {
				if errNC := db.Create(&node).Error; errNC != nil {
					blog.Errorf("syncStore create node err: %v", errNC)
				}
			} else {
				blog.Errorf("syncStore get node err: %v", errN)
			}
		} else {
			if errNS := db.Save(&node).Error; errNS != nil {
				blog.Errorf("syncStore update node err: %v", errNS)
			} else {
				blog.Infof("syncStore update node success, node: %d", node.ID)
			}
		}
	}

	// blog.Infof("SyncStoreNode nodeIDs: %v", nodeIDs)
	var dbNodeList []model.Node
	if errDB := db.Find(&dbNodeList).Error; errDB != nil {
		blog.Errorf("find node err: %v", errDB)
	} else {
		for _, node := range dbNodeList {
			if exist, _ := common.InArray(node.ID, nodeIDs); !exist {
				// blog.Infof("SyncStoreNode delete node success, node: %d", node.ID)
				if errN := db.Delete(&node).Error; errN != nil {
					blog.Errorf("syncStore delete node err: %v", errN)
				}
			}
		}
	}

	return nil
}

// SyncStoreWorkload 同步工作负载信息到数据库
func (s *Syncer) SyncStoreWorkload(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := s.SyncStoreDeployment(bkCluster, db)
	if err != nil {
		blog.Errorf("syncStore Deployment err: %v", err)
	}
	err = s.SyncStoreStatefulSet(bkCluster, db)
	if err != nil {
		blog.Errorf("syncStore StatefulSet err: %v", err)
	}
	err = s.SyncStoreDaemonSet(bkCluster, db)
	if err != nil {
		blog.Errorf("syncStore DaemonSet err: %v", err)
	}
	err = s.SyncStoreGameDeployment(bkCluster, db)
	if err != nil {
		blog.Errorf("syncStore GameDeployment err: %v", err)
	}
	err = s.SyncStoreGameStatefulSet(bkCluster, db)
	if err != nil {
		blog.Errorf("syncStore GameStatefulSet err: %v", err)
	}
	err = s.SyncStorePodsWorkload(bkCluster, db)
	if err != nil {
		blog.Errorf("syncStore PodsWordload err: %v", err)
	}
	return nil
}

// SyncStoreDeployment 同步 Deployment 信息到数据库
func (s *Syncer) SyncStoreDeployment(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.DeploymentMigrate(db)
	if err != nil {
		blog.Errorf("migrate deployment failed, err: %s", err.Error())
		return fmt.Errorf("migrate deployment failed, err: %s", err.Error())
	}

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
	}, false, nil)
	if err != nil {
		blog.Errorf("get bk deployment failed, err: %s", err.Error())
		return fmt.Errorf("get bk deployment failed, err: %s", err.Error())
	}
	deploymentMarshal, err := json.Marshal(bkDeployments)
	if err != nil {
		blog.Errorf("marshal bk deployment failed, err: %s", err.Error())
	}
	var deployments []model.Deployment
	err = json.Unmarshal(deploymentMarshal, &deployments)
	if err != nil {
		blog.Errorf("unmarshal bk deployment failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal bk deployment failed, err: %s", err.Error())
	}
	for _, deployment := range deployments {
		var existDeployment model.Deployment
		if errD := db.Where("id = ?", deployment.ID).First(&existDeployment).Error; errD != nil {
			if errors.Is(errD, gorm.ErrRecordNotFound) {
				if errDC := db.Create(&deployment).Error; errDC != nil {
					blog.Errorf("syncStore create deployment err: %v", errDC)
				}
			} else {
				blog.Errorf("syncStore get deployment err: %v", errD)
			}
		} else {
			if errDS := db.Save(&deployment).Error; errDS != nil {
				blog.Errorf("syncStore update deployment err: %v", errDS)
			} else {
				blog.Infof("syncStore update deployment success, deployment: %d", deployment.ID)
			}
		}
	}
	return nil
}

// SyncStoreStatefulSet 同步 StatefulSet 信息到数据库
func (s *Syncer) SyncStoreStatefulSet(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.StatefulSetMigrate(db)
	if err != nil {
		blog.Errorf("migrate statefulset failed, err: %s", err.Error())
		return fmt.Errorf("migrate statefulset failed, err: %s", err.Error())
	}

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
	}, false, nil)
	if err != nil {
		blog.Errorf("get bk statefulSet failed, err: %s", err.Error())
		return fmt.Errorf("get bk statefulSet failed, err: %s", err.Error())
	}

	statefulSetMarshal, err := json.Marshal(bkStatefulSets)
	if err != nil {
		blog.Errorf("marshal bk statefulSet failed, err: %s", err.Error())
	}
	var statefulSets []model.StatefulSet
	err = json.Unmarshal(statefulSetMarshal, &statefulSets)
	if err != nil {
		blog.Errorf("unmarshal bk statefulSet failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal bk statefulSet failed, err: %s", err.Error())
	}
	for _, statefulSet := range statefulSets {
		var existStatefulSet model.StatefulSet
		if errS := db.Where("id = ?", statefulSet.ID).First(&existStatefulSet).Error; errS != nil {
			if errors.Is(errS, gorm.ErrRecordNotFound) {
				if errSC := db.Create(&statefulSet).Error; errSC != nil {
					blog.Errorf("syncStore create statefulSet err: %v", errSC)
				}
			} else {
				blog.Errorf("syncStore get statefulSet err: %v", errS)
			}
		} else {
			if errSS := db.Save(&statefulSet).Error; errSS != nil {
				blog.Errorf("syncStore update statefulSet err: %v", errSS)
			} else {
				blog.Infof("syncStore update statefulSet success, statefulSet: %d", statefulSet.ID)
			}
		}
	}
	return nil
}

// SyncStoreDaemonSet 同步 DaemonSet 信息到数据库
func (s *Syncer) SyncStoreDaemonSet(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.DaemonSetMigrate(db)
	if err != nil {
		blog.Errorf("migrate daemonset failed, err: %s", err.Error())
		return fmt.Errorf("migrate daemonset failed, err: %s", err.Error())
	}

	// GetBkWorkloads get bkworkloads
	bkDaemonSets, err := s.GetBkWorkloads(bkCluster.BizID, "daemonSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, false, nil)
	if err != nil {
		blog.Errorf("get bk daemonSet failed, err: %s", err.Error())
		return fmt.Errorf("get bk daemonSet failed, err: %s", err.Error())
	}

	daemonSetMarshal, err := json.Marshal(bkDaemonSets)
	if err != nil {
		blog.Errorf("marshal bk daemonSet failed, err: %s", err.Error())
	}
	var daemonSets []model.DaemonSet
	err = json.Unmarshal(daemonSetMarshal, &daemonSets)
	if err != nil {
		blog.Errorf("unmarshal bk daemonSet failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal bk daemonSet failed, err: %s", err.Error())
	}
	for _, daemonSet := range daemonSets {
		var existDaemonSet model.DaemonSet
		if errD := db.Where("id = ?", daemonSet.ID).First(&existDaemonSet).Error; errD != nil {
			if errors.Is(errD, gorm.ErrRecordNotFound) {
				if errDC := db.Create(&daemonSet).Error; errDC != nil {
					blog.Errorf("syncStore create daemonSet err: %v", errDC)
				}
			} else {
				blog.Errorf("syncStore get daemonSet err: %v", errD)
			}
		} else {
			if errDS := db.Save(&daemonSet).Error; errDS != nil {
				blog.Errorf("syncStore update daemonSet err: %v", errDS)
			} else {
				blog.Infof("syncStore update daemonSet success, daemonSet: %d", daemonSet.ID)
			}
		}
	}
	return nil
}

// SyncStoreGameDeployment 同步 GameDeployment 信息到数据库
func (s *Syncer) SyncStoreGameDeployment(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.GameDeploymentMigrate(db)
	if err != nil {
		blog.Errorf("migrate gamedeployment failed, err: %s", err.Error())
		return fmt.Errorf("migrate gamedeployment failed, err: %s", err.Error())
	}

	// GetBkWorkloads get bkworkloads
	bkGameDeployments, err := s.GetBkWorkloads(bkCluster.BizID, "gameDeployment", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, false, nil)
	if err != nil {
		blog.Errorf("get bk gameDeployment failed, err: %s", err.Error())
		return fmt.Errorf("get bk gameDeployment failed, err: %s", err.Error())
	}

	gameDeploymentMarshal, err := json.Marshal(bkGameDeployments)
	if err != nil {
		blog.Errorf("marshal bk gameDeployment failed, err: %s", err.Error())
		return fmt.Errorf("marshal bk gameDeployment failed, err: %s", err.Error())
	}
	var gameDeployments []model.GameDeployment
	err = json.Unmarshal(gameDeploymentMarshal, &gameDeployments)
	if err != nil {
		blog.Errorf("unmarshal bk gameDeployment failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal bk gameDeployment failed, err: %s", err.Error())
	}
	for _, gameDeployment := range gameDeployments {
		var existGameDeployment model.GameDeployment
		if errG := db.Where("id = ?", gameDeployment.ID).First(&existGameDeployment).Error; errG != nil {
			if errors.Is(errG, gorm.ErrRecordNotFound) {
				if errGC := db.Create(&gameDeployment).Error; errGC != nil {
					blog.Errorf("syncStore create gameDeployment err: %v", errGC)
				}
			} else {
				blog.Errorf("syncStore get gameDeployment err: %v", errG)
			}
		} else {
			if errGS := db.Save(&gameDeployment).Error; errGS != nil {
				blog.Errorf("syncStore update gameDeployment err: %v", errGS)
			} else {
				blog.Infof("syncStore update gameDeployment success, gameDeployment: %d", gameDeployment.ID)
			}
		}
	}
	return nil
}

// SyncStoreGameStatefulSet 同步 GameStatefulSet 信息到数据库
func (s *Syncer) SyncStoreGameStatefulSet(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.GameStatefulSetMigrate(db)
	if err != nil {
		blog.Errorf("migrate gameStatefulSet failed, err: %s", err.Error())
		return fmt.Errorf("migrate gameStatefulSet failed, err: %s", err.Error())
	}

	// GetBkWorkloads get bkworkloads
	bkGameStatefulSets, err := s.GetBkWorkloads(bkCluster.BizID, "gameStatefulSet", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, false, nil)
	if err != nil {
		blog.Errorf("get bk gameStatefulSet failed, err: %s", err.Error())
		return fmt.Errorf("get bk gameStatefulSet failed, err: %s", err.Error())
	}

	gameStatefulSetMarshal, err := json.Marshal(bkGameStatefulSets)
	if err != nil {
		blog.Errorf("marshal bk gameStatefulSet failed, err: %s", err.Error())
		return fmt.Errorf("marshal bk gameStatefulSet failed, err: %s", err.Error())
	}
	var gameStatefulSets []model.GameStatefulSet
	err = json.Unmarshal(gameStatefulSetMarshal, &gameStatefulSets)
	if err != nil {
		blog.Errorf("unmarshal bk gameStatefulSet failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal bk gameStatefulSet failed, err: %s", err.Error())
	}
	for _, gameStatefulSet := range gameStatefulSets {
		var existGameStatefulSet model.GameStatefulSet
		if errG := db.Where("id = ?", gameStatefulSet.ID).First(&existGameStatefulSet).Error; errG != nil {
			if errors.Is(errG, gorm.ErrRecordNotFound) {
				if errGC := db.Create(&gameStatefulSet).Error; errGC != nil {
					blog.Errorf("syncStore create gameStatefulSet err: %v", errGC)
				}
			} else {
				blog.Errorf("syncStore get gameStatefulSet err: %v", errG)
			}
		} else {
			if errGS := db.Save(&gameStatefulSet).Error; errGS != nil {
				blog.Errorf("syncStore update gameStatefulSet err: %v", errGS)
			} else {
				blog.Infof("syncStore update gameStatefulSet success, gameStatefulSet: %d", gameStatefulSet.ID)
			}
		}
	}
	return nil
}

// SyncStorePodsWorkload 同步 PodsWorkload 信息到数据库
func (s *Syncer) SyncStorePodsWorkload(bkCluster *bkcmdbkube.Cluster, db *gorm.DB) error {
	err := model.PodsWorkloadMigrate(db)
	if err != nil {
		blog.Errorf("migrate podsworkload failed, err: %s", err.Error())
		return fmt.Errorf("migrate podsworkload failed, err: %s", err.Error())
	}

	// GetBkWorkloads get bkworkloads
	bkPodsWorkloads, err := s.GetBkWorkloads(bkCluster.BizID, "pods", &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "cluster_uid",
				Operator: "in",
				Value:    []string{bkCluster.Uid},
			},
		},
	}, false, nil)
	if err != nil {
		blog.Errorf("get bk podsworkload failed, err: %s", err.Error())
		return fmt.Errorf("get bk podsworkload failed, err: %s", err.Error())
	}

	podsWorkloadMarshal, err := json.Marshal(bkPodsWorkloads)
	if err != nil {
		blog.Errorf("marshal bk podsworkload failed, err: %s", err.Error())
	}
	var podsWorkloads []model.PodsWorkload
	err = json.Unmarshal(podsWorkloadMarshal, &podsWorkloads)
	if err != nil {
		blog.Errorf("unmarshal bk podsworkload failed, err: %s", err.Error())
		return fmt.Errorf("unmarshal bk podsworkload failed, err: %s", err.Error())
	}
	for _, podsWorkload := range podsWorkloads {
		var existPodsWorkload model.PodsWorkload
		if errP := db.Where("id = ?", podsWorkload.ID).First(&existPodsWorkload).Error; errP != nil {
			if errors.Is(errP, gorm.ErrRecordNotFound) {
				if errPC := db.Create(&podsWorkload).Error; errPC != nil {
					blog.Errorf("syncStore create podsworkload err: %v", errPC)
				}
			} else {
				blog.Errorf("syncStore get podsworkload err: %v", errP)
			}
		} else {
			if errPS := db.Save(&podsWorkload).Error; errPS != nil {
				blog.Errorf("syncStore update podsworkload err: %v", errPS)
			} else {
				blog.Infof("syncStore update podsworkload success, podsworkload: %d", podsWorkload.ID)
			}
		}
	}
	return nil
}

// GetBcsStorageClient is a function that returns a BCS storage client.
func (s *Syncer) GetBcsStorageClient() (bcsapi.Storage, error) {
	// Create a BCS API configuration with the given options.
	config := &bcsapi.Config{
		Hosts:     []string{s.BkcmdbSynchronizerOption.Bcsapi.HttpAddr},
		AuthToken: s.BkcmdbSynchronizerOption.Bcsapi.BearerToken,
		TLSConfig: s.ClientTls,
		Gateway:   strings.Contains(s.BkcmdbSynchronizerOption.Bcsapi.HttpAddr, "gateway"),
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
func (s *Syncer) GetBkCluster(cluster *cmp.Cluster, db *gorm.DB, withDB bool) (*bkcmdbkube.Cluster, error) {
	// Check if cluster is nil
	if cluster == nil {
		blog.Errorf("cluster is nil in GetBkCluster")
		return nil, errors.New("cluster is nil")
	}

	var clusterBkBizID int64
	if bkBizID == 0 {
		bizid, err := strconv.ParseInt(cluster.BusinessID, 10, 64)
		if err != nil {
			blog.Errorf("An error occurred: %s\n", err)
		} else {
			blog.Infof("Successfully converted string to int64: %d\n", bizid)
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
	}, db, withDB)

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
func (s *Syncer) GetBkNodes(
	bkBizID int64, filter *client.PropertyFilter, withDB bool, db *gorm.DB) (*[]bkcmdbkube.Node, error) {
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
		}, db, withDB)

		if err != nil {
			blog.Errorf("get bcs node failed, err: %s", err.Error())
			return nil, err
		}
		bkNodeList = append(bkNodeList, *bkNodes...)

		if len(*bkNodes) < 100 {
			break
		}

		if withDB {
			break
		}

		pageStart++
	}

	return &bkNodeList, nil
}

// CreateBkNodes create bknodes
func (s *Syncer) CreateBkNodes(
	bkCluster *bkcmdbkube.Cluster, toCreate *[]client.CreateBcsNodeRequestData, db *gorm.DB) {
	if len(*toCreate) > 0 {
		_, err := s.CMDBClient.CreateBcsNode(&client.CreateBcsNodeRequest{
			BKBizID: &bkCluster.BizID,
			Data:    toCreate,
		}, db)
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
				}, db)

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

// ComparePod compare pod
func (s *Syncer) ComparePod(bkPod *bkcmdbkube.Pod, k8sPod *storage.Pod, db *gorm.DB) (needToRecreate bool) {
	needToRecreate = false
	// bkContainers, err := s.CMDBClient.GetBcsContainer(&client.GetBcsContainerRequest{
	//	CommonRequest: client.CommonRequest{
	//		BKBizID: bkPod.BizID,
	//		Page: client.Page{
	//			Limit: 200,
	//			Start: 0,
	//		},
	//	},
	//	BkPodID: bkPod.ID,
	// }, nil, false)

	bkContainers, err := s.GetBkContainers(bkPod.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "bk_pod_id",
				Operator: "in",
				Value:    []int64{bkPod.ID},
			},
		},
	}, true, db)

	if err != nil {
		blog.Errorf("ComparePod GetBcsContainer err: %v", err)
		return needToRecreate
	}

	bkContainerMap := make(map[string]client.Container)

	for k, v := range *bkContainers {
		bkContainerMap[*v.ContainerID] = (*bkContainers)[k]
	}

	k8sContainerMap := make(map[string]corev1.Container)

	for k, v := range k8sPod.Data.Status.ContainerStatuses {
		k8sContainerMap[v.ContainerID] = k8sPod.Data.Spec.Containers[k]
	}

	if len(k8sContainerMap) != len(bkContainerMap) {
		needToRecreate = true
	}

	for k := range bkContainerMap {
		if _, ok := k8sContainerMap[k]; !ok {
			blog.Infof("Compare Pod: %v, %v", *bkPod, *k8sPod)
			needToRecreate = true
		}
		// else {
		//	if *v.Name != k8sContainerMap[k].Name || *v.Image != k8sContainerMap[k].Image {
		//		blog.Infof("Compare Pod: name %s vs %s, image %s vs %s",
		//		*v.Name,k8sContainerMap[k].Name,*v.Image , k8sContainerMap[k].Image)
		//		needToRecreate = true
		//	}
		// }
	}

	return needToRecreate
}

func sortedMapString(m map[string]string) string { // nolint
	// 获取所有的键
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// 对键进行排序
	sort.Strings(keys)

	// 按键的顺序构建字符串
	s := "{"
	for i, k := range keys {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%q:%v", k, m[k])
	}
	s += "}"
	return s
}

func mapToString(m map[string]string) string {
	b := new(strings.Builder)
	for k, v := range m {
		fmt.Fprintf(b, "%s:%s,", k, v)
	}
	result := b.String()
	return strings.TrimRight(result, ",")
}

func stringToMap(s string) map[string]string {
	resultMap := make(map[string]string)
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, ":")
		if len(kv) == 2 {
			resultMap[kv[0]] = kv[1]
		}
	}
	return resultMap
}

func mapsEqual(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if v2, ok := m2[k]; !ok || v != v2 {
			return false
		}
	}
	return true
}

// CompareNode compare bknode and k8snode
// nolint funlen
func (s *Syncer) CompareNode(bkNode *bkcmdbkube.Node, k8sNode *storage.K8sNode) (
	needToUpdate bool, updateData *client.UpdateBcsNodeRequestData) {
	updateData = &client.UpdateBcsNodeRequestData{}
	needToUpdate = false
	labelsEmpty := map[string]string{}

	s.compareNode(bkNode, k8sNode, &needToUpdate, updateData)

	if k8sNode.Data.Labels == nil {
		if bkNode.Labels != nil {
			updateData.Labels = &labelsEmpty
			needToUpdate = true
		}
	} else if bkNode.Labels == nil ||
		!mapsEqual(stringToMap(strings.ReplaceAll(mapToString(*bkNode.Labels), "．", ".")),
			stringToMap(strings.ReplaceAll(mapToString(k8sNode.Data.Labels), "．", "."))) {
		if len(k8sNode.Data.Labels) != 0 {
			updateData.Labels = &k8sNode.Data.Labels
			needToUpdate = true
		}

		if bkNode.Labels == nil {
			blog.Infof("CompareNode Labels: bknode: nil, k8snode: %v",
				strings.ReplaceAll(mapToString(k8sNode.Data.Labels), "．", "."))
		} else {
			blog.Infof("CompareNode Labels: bknode: %v\n k8snode: %v",
				strings.ReplaceAll(mapToString(*bkNode.Labels), "．", "."),
				strings.ReplaceAll(mapToString(k8sNode.Data.Labels), "．", "."))
		}
	}

	if bkNode.Unschedulable == nil {
		if k8sNode.Data.Spec.Unschedulable {
			updateData.Unschedulable = &k8sNode.Data.Spec.Unschedulable
			needToUpdate = true
			blog.Infof("CompareNode Unschedulable nil")
			blog.Infof("bkNode: %v", *bkNode)
		}
	} else if *bkNode.Unschedulable != k8sNode.Data.Spec.Unschedulable {
		updateData.Unschedulable = &k8sNode.Data.Spec.Unschedulable
		blog.Infof("CompareNode Unschedulable: %v, %v", *bkNode.Unschedulable, k8sNode.Data.Spec.Unschedulable)
		needToUpdate = true
	}

	if bkNode.PodCidr == nil {
		if k8sNode.Data.Spec.PodCIDR != "" {
			updateData.PodCidr = &k8sNode.Data.Spec.PodCIDR
			blog.Infof("CompareNode PodCIDR nil")
			blog.Infof("bkNode: %v", *bkNode)
			needToUpdate = true
		}
	} else if *bkNode.PodCidr != k8sNode.Data.Spec.PodCIDR {
		updateData.PodCidr = &k8sNode.Data.Spec.PodCIDR
		blog.Infof("CompareNode PodCIDR: %v, %v", *bkNode.PodCidr, k8sNode.Data.Spec.PodCIDR)
		needToUpdate = true
	}

	if bkNode.RuntimeComponent != nil {
		if *bkNode.RuntimeComponent != k8sNode.Data.Status.NodeInfo.ContainerRuntimeVersion {
			updateData.RuntimeComponent = &k8sNode.Data.Status.NodeInfo.ContainerRuntimeVersion
			blog.Infof("CompareNode RuntimeComponent")
			needToUpdate = true
		}
	} else {
		if k8sNode.Data.Status.NodeInfo.ContainerRuntimeVersion != "" {
			updateData.RuntimeComponent = &k8sNode.Data.Status.NodeInfo.ContainerRuntimeVersion
			blog.Infof("CompareNode RuntimeComponent")
			needToUpdate = true
		}
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

func (s *Syncer) compareNode(bkNode *bkcmdbkube.Node, k8sNode *storage.K8sNode,
	needToUpdate *bool, updateData *client.UpdateBcsNodeRequestData) {
	taints := make(map[string]string) // nolint
	for _, taint := range k8sNode.Data.Spec.Taints {
		taints[taint.Key] = taint.Value
	}

	blog.Infof("bkNode.Taints: %v, k8sNode.Taints: %v", bkNode.Taints, k8sNode.Data.Spec.Taints)

	if taints == nil { // nolint this nil check is never true
		if bkNode.Taints != nil {
			updateData.Taints = &taints
			*needToUpdate = true
		}
	} else if bkNode.Taints == nil ||
		!mapsEqual(stringToMap(strings.ReplaceAll(mapToString(*bkNode.Taints), "．", ".")),
			stringToMap(strings.ReplaceAll(mapToString(taints), "．", "."))) {
		if len(taints) != 0 {
			updateData.Taints = &taints
			*needToUpdate = true
		}

		if bkNode.Taints == nil {
			blog.Infof("CompareNode Taints: bknode: nil, k8snode: %v",
				stringToMap(strings.ReplaceAll(mapToString(taints), "．", ".")))
		} else {
			blog.Infof("CompareNode Taints: bknode: %v\n, k8snode: %v",
				stringToMap(strings.ReplaceAll(mapToString(*bkNode.Taints), "．", ".")),
				stringToMap(strings.ReplaceAll(mapToString(taints), "．", ".")))
		}
	}
}

// GenerateBkNodeData generate bknode data from k8snode
func (s *Syncer) GenerateBkNodeData(
	bkCluster *bkcmdbkube.Cluster, k8sNode *storage.K8sNode) (data client.CreateBcsNodeRequestData, err error) {
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
		hosts, err := s.CMDBClient.GetHostsByBiz(bkCluster.BizID, internalIP)
		if err != nil {
			blog.Errorf("GetHostsByBiz error: %v", err)
			return client.CreateBcsNodeRequestData{}, err
		}
		if len(*hosts) > 0 {
			theHostID = (*hosts)[0].BkHostId
		}
		if len(*hosts) == 0 {
			blog.Errorf("GetHostsByBiz can not find host: %s in biz %d", internalIP, bkCluster.BizID)
			return client.CreateBcsNodeRequestData{},
				fmt.Errorf("GetHostsByBiz can not find host: %s in biz %d", internalIP, bkCluster.BizID)
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
	}, nil
}

// UpdateBkNodes update bknodes
// nolint
func (s *Syncer) UpdateBkNodes(
	bkCluster *bkcmdbkube.Cluster, toUpdate *map[int64]*client.UpdateBcsNodeRequestData, db *gorm.DB) {
	if toUpdate == nil {
		return
	}

	for k, v := range *toUpdate {
		err := s.CMDBClient.UpdateBcsNode(&client.UpdateBcsNodeRequest{
			BKBizID: &bkCluster.BizID,
			IDs:     &[]int64{k},
			Data:    v,
		}, db)

		if err != nil {
			blog.Errorf("update node failed, err: %s", err.Error())
		}
	}
}

// DeleteBkNodes delete bknodes
func (s *Syncer) DeleteBkNodes(bkCluster *bkcmdbkube.Cluster, toDelete *[]int64, db *gorm.DB) error {
	if len(*toDelete) == 0 {
		return nil
	}

	bkPodList, err := s.GetBkPods(bkCluster.BizID, &client.PropertyFilter{
		Condition: "AND",
		Rules: []client.Rule{
			{
				Field:    "bk_node_id",
				Operator: "in",
				Value:    *toDelete,
			},
		},
	}, true, db)
	// blog.Infof("bkPodList: %v, len: %d", *bkPodList, len(*bkPodList))
	if err != nil {
		blog.Errorf("get bk pod failed, err: %s", err.Error())
	} else if len(*bkPodList) > 0 {
		podToDelete := make([]int64, 0)
		for _, pod := range *bkPodList {
			podToDelete = append(podToDelete, pod.ID)
			blog.Infof("podToDelete: %s+%s", pod.NameSpace, *pod.Name)
		}
		_ = s.DeleteBkPods(bkCluster, &podToDelete, db)
	}

	err = s.CMDBClient.DeleteBcsNode(&client.DeleteBcsNodeRequest{
		BKBizID: &bkCluster.BizID,
		IDs:     toDelete,
	}, nil)
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
			}, nil)

			if err != nil {
				blog.Errorf("delete node failed, err: %s", err.Error())
			}
		}
	}
	blog.Infof("delete node from cmdb success, ids: %v", toDelete)
	return nil
}

// GetBkNamespaces get bknamespaces
func (s *Syncer) GetBkNamespaces(
	bkBizID int64, filter *client.PropertyFilter, withDB bool, db *gorm.DB) (*[]bkcmdbkube.Namespace, error) {
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
		}, db, withDB)

		if err != nil {
			blog.Errorf("get bcs namespace failed, err: %s", err.Error())
			return nil, err
		}
		bkNamespaceList = append(bkNamespaceList, *bkNamespaces...)

		if len(*bkNamespaces) < 100 {
			break
		}

		if withDB {
			break
		}
		pageStart++
	}

	return &bkNamespaceList, nil
}

// CompareNamespace compare bkns and k8sns
func (s *Syncer) CompareNamespace(bkNs *bkcmdbkube.Namespace, k8sNs *storage.Namespace) (
	needToUpdate bool, updateData *client.UpdateBcsNamespaceRequestData) {
	updateData = &client.UpdateBcsNamespaceRequestData{}
	needToUpdate = false
	labelsEmpty := map[string]string{}

	// var updateDataIDs []int64
	// updateData.IDs = &updateDataIDs
	// var updateDataInfo client.UpdateBcsNamespaceRequestDataInfo
	// updateData.Info = &updateDataInfo

	if k8sNs.Data.Labels == nil {
		if bkNs.Labels != nil {
			updateData.Labels = &labelsEmpty
			needToUpdate = true
		}
	} else if bkNs.Labels == nil ||
		!mapsEqual(stringToMap(strings.ReplaceAll(mapToString(*bkNs.Labels), "．", ".")),
			stringToMap(strings.ReplaceAll(mapToString(k8sNs.Data.Labels), "．", "."))) {
		if len(k8sNs.Data.Labels) != 0 {
			updateData.Labels = &k8sNs.Data.Labels
			blog.Infof("CompareNamespace labels: %v, %v", bkNs.Labels, fmt.Sprint(k8sNs.Data.Labels))
			needToUpdate = true
		}
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
func (s *Syncer) CreateBkNamespaces(
	bkCluster *bkcmdbkube.Cluster, toCreate map[int64][]bkcmdbkube.Namespace, db *gorm.DB) {
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
					}, db)

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
					}, db)

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
func (s *Syncer) UpdateBkNamespaces(
	bkCluster *bkcmdbkube.Cluster, toUpdate *map[int64]*client.UpdateBcsNamespaceRequestData, db *gorm.DB) {
	if toUpdate == nil {
		return
	}

	for k, v := range *toUpdate {
		err := s.CMDBClient.UpdateBcsNamespace(&client.UpdateBcsNamespaceRequest{
			BKBizID: &bkCluster.BizID,
			IDs:     &[]int64{k},
			Data:    v,
		}, db)

		if err != nil {
			blog.Errorf("update namespace failed, err: %s", err.Error())
		}
	}
}

// DeleteBkNamespaces delete bknamespaces
func (s *Syncer) DeleteBkNamespaces(bkCluster *bkcmdbkube.Cluster, toDelete *[]int64, db *gorm.DB) error {
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
			}, false, nil)
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
			}, db)

			if err != nil {
				blog.Errorf("delete bk workload pods failed, err: %s", err.Error())
				return err
			}

			err = s.CMDBClient.DeleteBcsNamespace(&client.DeleteBcsNamespaceRequest{
				BKBizID: &bkCluster.BizID,
				IDs:     &section,
			}, db)

			if err != nil {
				blog.Errorf("delete namespace failed, err: %s", err.Error())
			}
		}
	}
	return nil
}

// GetBkWorkloads get bkworkloads
func (s *Syncer) GetBkWorkloads(
	bkBizID int64, workloadType string,
	filter *client.PropertyFilter, withDB bool, db *gorm.DB) (*[]interface{}, error) {
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
				// Filter: &client.PropertyFilter{
				//	Condition: "OR",
				//	Rules: []client.Rule{
				//		{
				//			Field:    "cluster_uid",
				//			Operator: "in",
				//			Value:    []string{bkCluster.Uid},
				//		},
				//	},
				// },
			},
			// ClusterUID: bkCluster.Uid,
			Kind: workloadType,
		}, db, withDB)

		if err != nil {
			blog.Errorf("get bcs workload failed, err: %s", err.Error())
			return nil, err
		}
		bkWorkloadList = append(bkWorkloadList, *bkWorkloads...)

		if len(*bkWorkloads) < 100 {
			break
		}
		if withDB {
			break
		}
		pageStart++
	}

	return &bkWorkloadList, nil
}

// CreateBkWorkloads create bkworkloads
func (s *Syncer) CreateBkWorkloads(
	bkCluster *bkcmdbkube.Cluster, kind string, toCreate map[int64][]client.CreateBcsWorkloadRequestData, db *gorm.DB) {
	if len(toCreate) > 0 {
		for bizid, workloads := range toCreate {
			if len(workloads) > 0 {
				_, err := s.CMDBClient.CreateBcsWorkload(&client.CreateBcsWorkloadRequest{
					BKBizID: &bizid,
					Kind:    &kind,
					Data:    &workloads,
				}, db)
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
						}, db)

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

func rollingUpdateDeploymentEqual(rud *appv1.RollingUpdateDeployment, bkRud *bkcmdbkube.RollingUpdateDeployment) bool {
	if rud == nil && bkRud == nil {
		return true
	}

	if rud != nil && bkRud == nil {
		return false
	}

	if bkRud != nil && rud == nil {
		if bkRud.MaxUnavailable == nil && bkRud.MaxSurge == nil {
			return true
		}
		return false
	}

	if !rudMaxSurge(rud.MaxSurge, bkRud.MaxSurge) {
		return false
	}

	if rud.MaxUnavailable != nil && bkRud.MaxUnavailable == nil {
		return false
	}

	if rud.MaxUnavailable == nil && bkRud.MaxUnavailable != nil {
		return false
	}

	if rud.MaxUnavailable != nil && bkRud.MaxUnavailable != nil {
		if rud.MaxUnavailable.StrVal != bkRud.MaxUnavailable.StrVal {
			return false
		}
	}

	return true
}

func rudMaxSurge(a *intstr.IntOrString, b *bkcmdbkube.IntOrString) bool {
	if a != nil && b == nil {
		return false
	}
	if a == nil && b != nil {
		return false
	}
	if a != nil && b != nil {
		if a.StrVal != b.StrVal {
			return false
		}
	}
	return true
}

// CompareDeployment compare bkdeployment and k8sdeployment
// nolint funlen
func (s *Syncer) CompareDeployment(
	bkDeployment *bkcmdbkube.Deployment, k8sDeployment *storage.Deployment) (
	needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	updateData = &client.UpdateBcsWorkloadRequestData{}
	needToUpdate = false

	s.compareDeployment(bkDeployment, k8sDeployment, &needToUpdate, updateData)

	rudEmpty := map[string]interface{}{}

	if !rollingUpdateDeploymentEqual(k8sDeployment.Data.Spec.Strategy.RollingUpdate,
		bkDeployment.RollingUpdateStrategy) {
		if k8sDeployment.Data.Spec.Strategy.RollingUpdate == nil {
			updateData.RollingUpdateStrategy = &rudEmpty
			blog.Infof("CompareDeployment RollingUpdate: %v, %v",
				k8sDeployment.Data.Spec.Strategy.RollingUpdate, bkDeployment.RollingUpdateStrategy)
			needToUpdate = true
		} else {
			rud := bkcmdbkube.RollingUpdateDeployment{}
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

			jsonBytes, err := json.Marshal(rud)
			if err != nil {
				blog.Errorf("marshal rolling update deployment failed, err: %s", err.Error())
				return false, nil
			}

			rudMap := make(map[string]interface{})
			err = json.Unmarshal(jsonBytes, &rudMap)

			updateData.RollingUpdateStrategy = &rudMap
			blog.Infof("CompareDeployment RollingUpdate: %v, %v",
				fmt.Sprint(*k8sDeployment.Data.Spec.Strategy.RollingUpdate),
				fmt.Sprint(*bkDeployment.RollingUpdateStrategy))
			needToUpdate = true
		}
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// nolint funlen
// compareDeployment compares the deployment configurations between bkDeployment and k8sDeployment,
// and updates the updateData if there are differences.
func (s *Syncer) compareDeployment(bkDeployment *bkcmdbkube.Deployment, k8sDeployment *storage.Deployment,
	needToUpdate *bool, updateData *client.UpdateBcsWorkloadRequestData) {
	// Initialize an empty map for labels
	labelsEmpty := map[string]string{}

	// Compare labels
	if k8sDeployment.Data.Labels == nil {
		if bkDeployment.Labels != nil {
			updateData.Labels = &labelsEmpty
			blog.Infof("CompareDeployment labels: %v, %v", k8sDeployment.Data.Labels, bkDeployment.Labels)
			*needToUpdate = true
		}
	} else if bkDeployment.Labels == nil ||
		!mapsEqual(stringToMap(strings.ReplaceAll(mapToString(*bkDeployment.Labels), "．", ".")),
			stringToMap(strings.ReplaceAll(mapToString(k8sDeployment.Data.Labels), "．", "."))) {
		if len(k8sDeployment.Data.Labels) > 0 {
			updateData.Labels = &k8sDeployment.Data.Labels
			blog.Infof("CompareDeployment labels: %v, %v", k8sDeployment.Data.Labels, bkDeployment.Labels)
			*needToUpdate = true
		}
	}

	// Compare selectors
	if k8sDeployment.Data.Spec.Selector == nil {
		if bkDeployment.Selector != nil {
			updateData.Selector = nil
			blog.Infof("CompareDeployment Selector: %v, %v",
				k8sDeployment.Data.Spec.Selector, k8sDeployment.Data.Spec.Selector)
			*needToUpdate = true
		}
	} else if bkDeployment.Selector == nil ||
		strings.ReplaceAll(fmt.Sprint(*k8sDeployment.Data.Spec.Selector), "．", ".") !=
			strings.ReplaceAll(fmt.Sprint(*bkDeployment.Selector), "．", ".") {
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
		blog.Infof("CompareDeployment Selector: %v, %v",
			fmt.Sprint(*k8sDeployment.Data.Spec.Selector), fmt.Sprint(*bkDeployment.Selector))

		*needToUpdate = true
	}

	// Compare replicas
	if bkDeployment.Replicas != nil {
		if *bkDeployment.Replicas != int64(*k8sDeployment.Data.Spec.Replicas) {
			replicas := int64(*k8sDeployment.Data.Spec.Replicas)
			updateData.Replicas = &replicas
			blog.Infof("CompareDeployment Replicas: %v, %v",
				*bkDeployment.Replicas, *k8sDeployment.Data.Spec.Replicas)
			*needToUpdate = true
		}
	} else {
		if int64(*k8sDeployment.Data.Spec.Replicas) != 0 {
			replicas := int64(*k8sDeployment.Data.Spec.Replicas)
			updateData.Replicas = &replicas
			blog.Infof("CompareDeployment Replicas: %v, %v",
				bkDeployment.Replicas, *k8sDeployment.Data.Spec.Replicas)
			*needToUpdate = true
		}
	}

	// Compare minReadySeconds
	if bkDeployment.MinReadySeconds != nil {
		if *bkDeployment.MinReadySeconds != int64(k8sDeployment.Data.Spec.MinReadySeconds) {
			minReadySeconds := int64(k8sDeployment.Data.Spec.MinReadySeconds)
			updateData.MinReadySeconds = &minReadySeconds
			blog.Infof("CompareDeployment MinReadySeconds: %v, %v",
				*bkDeployment.MinReadySeconds, k8sDeployment.Data.Spec.MinReadySeconds)
			*needToUpdate = true
		}
	} else {
		if int64(k8sDeployment.Data.Spec.MinReadySeconds) != 0 {
			minReadySeconds := int64(k8sDeployment.Data.Spec.MinReadySeconds)
			updateData.MinReadySeconds = &minReadySeconds
			blog.Infof("CompareDeployment MinReadySeconds: %v, %v",
				bkDeployment.MinReadySeconds, k8sDeployment.Data.Spec.MinReadySeconds)
			*needToUpdate = true
		}
	}

	// Compare strategyType
	if bkDeployment.StrategyType != nil {
		if *bkDeployment.StrategyType != bkcmdbkube.DeploymentStrategyType(k8sDeployment.Data.Spec.Strategy.Type) {
			strategyType := string(k8sDeployment.Data.Spec.Strategy.Type)
			updateData.StrategyType = &strategyType
			blog.Infof("CompareDeployment StrategyType: %v, %v",
				*bkDeployment.StrategyType, k8sDeployment.Data.Spec.Strategy.Type)
			*needToUpdate = true
		}
	} else {
		if bkcmdbkube.DeploymentStrategyType(k8sDeployment.Data.Spec.Strategy.Type) != "" {
			strategyType := string(k8sDeployment.Data.Spec.Strategy.Type)
			updateData.StrategyType = &strategyType
			blog.Infof("CompareDeployment StrategyType: %v, %v",
				bkDeployment.StrategyType, k8sDeployment.Data.Spec.Strategy.Type)
			*needToUpdate = true
		}
	}
}

// GenerateBkDeployment generate bkdeployment from k8sdeployment
func (s *Syncer) GenerateBkDeployment(
	bkNs *bkcmdbkube.Namespace, k8sDeployment *storage.Deployment) *client.CreateBcsWorkloadRequestData {
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
	_ = json.Unmarshal(jsonBytes, &rudMap)
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

// nolint
func rollingUpdateStatefulSetStrategyEqual(
	russ *appv1.RollingUpdateStatefulSetStrategy, bkRuss *bkcmdbkube.RollingUpdateStatefulSetStrategy) bool {
	if russ == nil && bkRuss == nil {
		return true
	}

	if russ != nil && bkRuss == nil {
		return false
	}

	if bkRuss != nil && russ == nil {
		if bkRuss.MaxUnavailable == nil && bkRuss.Partition == nil {
			return true
		}
		return false
	}

	if !rugssMaxUnavailable(russ.MaxUnavailable, bkRuss.MaxUnavailable) {
		return false
	}

	if russ.Partition != nil && bkRuss.Partition == nil {
		return false
	}

	if russ.Partition == nil && bkRuss.Partition != nil {
		return false
	}

	if russ.Partition != nil && bkRuss.Partition != nil {
		if *russ.Partition != *bkRuss.Partition {
			return false
		}
	}

	return true
}

// CompareStatefulSet compare bkstatefulset and k8sstatefulset
// nolint funlen
func (s *Syncer) CompareStatefulSet(
	bkStatefulSet *bkcmdbkube.StatefulSet, k8sStatefulSet *storage.StatefulSet) (
	needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	needToUpdate = false
	updateData = &client.UpdateBcsWorkloadRequestData{}

	s.compareStatefulSet(bkStatefulSet, k8sStatefulSet, &needToUpdate, updateData)

	russEmpty := map[string]interface{}{}

	if !rollingUpdateStatefulSetStrategyEqual(k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate,
		bkStatefulSet.RollingUpdateStrategy) {
		if k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate == nil {
			updateData.RollingUpdateStrategy = &russEmpty
			blog.Infof("CompareStatefulSet RollingUpdate: %v, %v",
				k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate, bkStatefulSet.RollingUpdateStrategy)
			needToUpdate = true
		} else {
			russ := bkcmdbkube.RollingUpdateStatefulSetStrategy{}

			if k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil {
				russ.MaxUnavailable = &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.Type),
					IntVal: k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.IntVal,
					StrVal: k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.StrVal,
				}
			}
			russ.Partition = k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.Partition

			jsonBytes, err := json.Marshal(russ)
			if err != nil {
				blog.Errorf("marshal rolling update statefulset failed, err: %s", err.Error())
				return false, nil
			}

			rusMap := make(map[string]interface{})
			_ = json.Unmarshal(jsonBytes, &rusMap)

			updateData.RollingUpdateStrategy = &rusMap
			blog.Infof("CompareStatefulSet RollingUpdate: %v, %v",
				*k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate, *bkStatefulSet.RollingUpdateStrategy)

			needToUpdate = true
		}
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// compareStatefulSet compares the stateful set from two different sources and determines if an update is needed.
// It updates the updateData with the necessary changes and sets the needToUpdate flag if any differences are found.
// nolint funlen
func (s *Syncer) compareStatefulSet(bkStatefulSet *bkcmdbkube.StatefulSet, k8sStatefulSet *storage.StatefulSet,
	needToUpdate *bool, updateData *client.UpdateBcsWorkloadRequestData) {

	// Compare Labels
	if k8sStatefulSet.Data.Labels == nil {
		// If Kubernetes labels are nil but bkStatefulSet labels are not,
		//set updateData labels to nil and mark for update
		if bkStatefulSet.Labels != nil {
			updateData.Labels = nil
			blog.Infof("CompareStatefulSet labels: %v, %v", k8sStatefulSet.Data.Labels, bkStatefulSet.Labels)
			*needToUpdate = true
		}
	} else if bkStatefulSet.Labels == nil ||
		!mapsEqual(stringToMap(strings.ReplaceAll(mapToString(*bkStatefulSet.Labels), "．", ".")),
			stringToMap(strings.ReplaceAll(mapToString(k8sStatefulSet.Data.Labels), "．", "."))) {
		// If labels differ, update updateData with Kubernetes labels and mark for update
		if len(k8sStatefulSet.Data.Labels) > 0 {
			updateData.Labels = &k8sStatefulSet.Data.Labels
			blog.Infof("CompareStatefulSet labels: %v, %v", k8sStatefulSet.Data.Labels, bkStatefulSet.Labels)
			*needToUpdate = true
		}
	}

	// Compare Selector
	if k8sStatefulSet.Data.Spec.Selector == nil {
		// If Kubernetes selector is nil but bkStatefulSet selector is not,
		//set updateData selector to nil and mark for update
		if bkStatefulSet.Selector != nil {
			updateData.Selector = nil
			blog.Infof("CompareStatefulSet Selector: %v, %v",
				k8sStatefulSet.Data.Spec.Selector, bkStatefulSet.Selector)
			*needToUpdate = true
		}
	} else if bkStatefulSet.Selector == nil ||
		strings.ReplaceAll(fmt.Sprint(*k8sStatefulSet.Data.Spec.Selector), "．", ".") !=
			strings.ReplaceAll(fmt.Sprint(*bkStatefulSet.Selector), "．", ".") {
		// If selectors differ, update updateData with Kubernetes selector and mark for update
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

		blog.Infof("CompareStatefulSet Selector: %v, %v",
			fmt.Sprint(k8sStatefulSet.Data.Spec.Selector), fmt.Sprint(*bkStatefulSet.Selector))
		*needToUpdate = true
	}

	// Compare Replicas
	if bkStatefulSet.Replicas != nil {
		// If bkStatefulSet replicas are not nil and differ from Kubernetes replicas,
		//update updateData and mark for update
		if *bkStatefulSet.Replicas != int64(*k8sStatefulSet.Data.Spec.Replicas) {
			replicas := int64(*k8sStatefulSet.Data.Spec.Replicas)
			updateData.Replicas = &replicas
			*needToUpdate = true
		}
	} else {
		// If bkStatefulSet replicas are nil but Kubernetes replicas are not zero, update updateData and mark for update
		if int64(*k8sStatefulSet.Data.Spec.Replicas) != 0 {
			replicas := int64(*k8sStatefulSet.Data.Spec.Replicas)
			updateData.Replicas = &replicas
			*needToUpdate = true
		}
	}

	// Compare MinReadySeconds
	if bkStatefulSet.MinReadySeconds != nil {
		// If bkStatefulSet MinReadySeconds are not nil and differ from Kubernetes MinReadySeconds,
		//update updateData and mark for update
		if *bkStatefulSet.MinReadySeconds != int64(k8sStatefulSet.Data.Spec.MinReadySeconds) {
			minReadySeconds := int64(k8sStatefulSet.Data.Spec.MinReadySeconds)
			updateData.MinReadySeconds = &minReadySeconds
			*needToUpdate = true
		}
	} else {
		// If bkStatefulSet MinReadySeconds are nil but Kubernetes MinReadySeconds are not zero,
		//update updateData and mark for update
		if int64(k8sStatefulSet.Data.Spec.MinReadySeconds) != 0 {
			minReadySeconds := int64(k8sStatefulSet.Data.Spec.MinReadySeconds)
			updateData.MinReadySeconds = &minReadySeconds
			*needToUpdate = true
		}
	}

	// Compare StrategyType
	if bkStatefulSet.StrategyType != nil {
		// If bkStatefulSet StrategyType is not nil and differs from Kubernetes StrategyType,
		//update updateData and mark for update
		if *bkStatefulSet.StrategyType !=
			bkcmdbkube.StatefulSetUpdateStrategyType(k8sStatefulSet.Data.Spec.UpdateStrategy.Type) {
			strategyType := string(k8sStatefulSet.Data.Spec.UpdateStrategy.Type)
			updateData.StrategyType = &strategyType
			*needToUpdate = true
		}
	} else {
		// If bkStatefulSet StrategyType is nil but Kubernetes StrategyType is not empty,
		//update updateData and mark for update
		if bkcmdbkube.StatefulSetUpdateStrategyType(k8sStatefulSet.Data.Spec.UpdateStrategy.Type) != "" {
			strategyType := string(k8sStatefulSet.Data.Spec.UpdateStrategy.Type)
			updateData.StrategyType = &strategyType
			*needToUpdate = true
		}
	}
}

// GenerateBkStatefulSet generate bkstatefulset from k8sstatefulset
func (s *Syncer) GenerateBkStatefulSet(
	bkNs *bkcmdbkube.Namespace, k8sStatefulSet *storage.StatefulSet) *client.CreateBcsWorkloadRequestData {
	me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
	for _, m := range k8sStatefulSet.Data.Spec.Selector.MatchExpressions {
		me = append(me, bkcmdbkube.LabelSelectorRequirement{
			Key:      m.Key,
			Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
			Values:   m.Values,
		})
	}

	rus := bkcmdbkube.RollingUpdateStatefulSetStrategy{}

	if k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate != nil &&
		k8sStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil {
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
	_ = json.Unmarshal(jsonBytes, &rusMap)

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

func rollingUpdateDaemonSetEqual(rud *appv1.RollingUpdateDaemonSet, bkRud *bkcmdbkube.RollingUpdateDaemonSet) bool {
	if rud == nil && bkRud == nil {
		return true
	}

	if rud != nil && bkRud == nil {
		return false
	}

	if bkRud != nil && rud == nil {
		if bkRud.MaxUnavailable == nil && bkRud.MaxSurge == nil {
			return true
		}
		return false
	}

	if !rudMaxSurge(rud.MaxSurge, bkRud.MaxSurge) {
		return false
	}

	if rud.MaxUnavailable != nil && bkRud.MaxUnavailable == nil {
		return false
	}

	if rud.MaxUnavailable == nil && bkRud.MaxUnavailable != nil {
		return false
	}

	if rud.MaxUnavailable != nil && bkRud.MaxUnavailable != nil {
		if rud.MaxUnavailable.StrVal != bkRud.MaxUnavailable.StrVal {
			return false
		}
	}

	return true
}

// CompareDaemonSet compare bkdaemonset and k8sdaemonset
// nolint funlen
func (s *Syncer) CompareDaemonSet(
	bkDaemonSet *bkcmdbkube.DaemonSet, k8sDaemonSet *storage.DaemonSet) (
	needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	needToUpdate = false
	updateData = &client.UpdateBcsWorkloadRequestData{}

	s.compareDaemonSet(bkDaemonSet, k8sDaemonSet, &needToUpdate, updateData)

	rudsEmpty := map[string]interface{}{}

	if !rollingUpdateDaemonSetEqual(k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate,
		bkDaemonSet.RollingUpdateStrategy) {
		if k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate == nil {
			updateData.RollingUpdateStrategy = &rudsEmpty
			blog.Infof("CompareDaemonSet RollingUpdate: %v, %v",
				k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate, bkDaemonSet.RollingUpdateStrategy)
			needToUpdate = true
		} else {
			rud := bkcmdbkube.RollingUpdateDaemonSet{}
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

			jsonBytes, err := json.Marshal(rud)
			if err != nil {
				blog.Errorf("marshal rolling update daemonSet failed, err: %s", err.Error())
				return false, nil
			}

			rudMap := make(map[string]interface{})
			_ = json.Unmarshal(jsonBytes, &rudMap)

			updateData.RollingUpdateStrategy = &rudMap
			blog.Infof("CompareDaemonSet RollingUpdate: %v, %v",
				k8sDaemonSet.Data.Spec.UpdateStrategy.RollingUpdate, bkDaemonSet.RollingUpdateStrategy)
			needToUpdate = true
		}
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// compareDaemonSet compares the DaemonSet from the backend and Kubernetes,
// and determines if an update is needed. If an update is needed, it populates
// the updateData with the necessary changes.
// nolint funlen
func (s *Syncer) compareDaemonSet(bkDaemonSet *bkcmdbkube.DaemonSet, k8sDaemonSet *storage.DaemonSet,
	needToUpdate *bool, updateData *client.UpdateBcsWorkloadRequestData) {

	// Initialize an empty map for labels
	labelsEmpty := map[string]string{}

	// Compare labels from both DaemonSets
	if k8sDaemonSet.Data.Labels == nil {
		// If Kubernetes DaemonSet has no labels but backend has, mark for update
		if bkDaemonSet.Labels != nil {
			updateData.Labels = &labelsEmpty
			blog.Infof("CompareDaemonSet labels: %v, %v", k8sDaemonSet.Data.Labels, bkDaemonSet.Labels)
			*needToUpdate = true
		}
	} else if bkDaemonSet.Labels == nil ||
		!mapsEqual(stringToMap(strings.ReplaceAll(mapToString(*bkDaemonSet.Labels), "．", ".")),
			stringToMap(strings.ReplaceAll(mapToString(k8sDaemonSet.Data.Labels), "．", "."))) {
		// If labels differ, mark for update with Kubernetes labels
		if len(k8sDaemonSet.Data.Labels) != 0 {
			updateData.Labels = &k8sDaemonSet.Data.Labels
			blog.Infof("CompareDaemonSet labels: %v, %v", k8sDaemonSet.Data.Labels, bkDaemonSet.Labels)
			*needToUpdate = true
		}
	}

	// Compare selectors from both DaemonSets
	if k8sDaemonSet.Data.Spec.Selector == nil {
		// If Kubernetes DaemonSet has no selector but backend has, mark for update
		if bkDaemonSet.Selector != nil {
			updateData.Selector = nil
			blog.Infof("CompareDaemonSet Selector: %v, %v",
				k8sDaemonSet.Data.Spec.Selector, bkDaemonSet.Selector)
			*needToUpdate = true
		}
	} else if bkDaemonSet.Selector == nil ||
		strings.ReplaceAll(fmt.Sprint(*k8sDaemonSet.Data.Spec.Selector), "．", ".") !=
			strings.ReplaceAll(fmt.Sprint(*bkDaemonSet.Selector), "．", ".") {
		// If selectors differ, mark for update with Kubernetes selector
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

		blog.Infof("CompareDaemonSet Selector: %v, %v", k8sDaemonSet.Data.Spec.Selector, bkDaemonSet.Selector)
		*needToUpdate = true
	}

	// Compare MinReadySeconds from both DaemonSets
	if bkDaemonSet.MinReadySeconds != nil {
		// If MinReadySeconds differ, mark for update with Kubernetes value
		if *bkDaemonSet.MinReadySeconds != int64(k8sDaemonSet.Data.Spec.MinReadySeconds) {
			minReadySeconds := int64(k8sDaemonSet.Data.Spec.MinReadySeconds)
			updateData.MinReadySeconds = &minReadySeconds
			blog.Infof("CompareDaemonSet MinReadySeconds: %v, %v",
				bkDaemonSet.MinReadySeconds, k8sDaemonSet.Data.Spec.MinReadySeconds)
			*needToUpdate = true
		}
	} else {
		// If backend has no MinReadySeconds but Kubernetes does, mark for update
		if int64(k8sDaemonSet.Data.Spec.MinReadySeconds) != 0 {
			minReadySeconds := int64(k8sDaemonSet.Data.Spec.MinReadySeconds)
			updateData.MinReadySeconds = &minReadySeconds
			blog.Infof("CompareDaemonSet MinReadySeconds: %v, %v",
				bkDaemonSet.MinReadySeconds, k8sDaemonSet.Data.Spec.MinReadySeconds)
			*needToUpdate = true
		}
	}

	// Compare StrategyType from both DaemonSets
	if bkDaemonSet.StrategyType != nil {
		// If StrategyType differs, mark for update with Kubernetes value
		if *bkDaemonSet.StrategyType !=
			bkcmdbkube.DaemonSetUpdateStrategyType(k8sDaemonSet.Data.Spec.UpdateStrategy.Type) {
			strategyType := string(k8sDaemonSet.Data.Spec.UpdateStrategy.Type)
			updateData.StrategyType = &strategyType
			blog.Infof("CompareDaemonSet StrategyType: %v, %v",
				bkDaemonSet.StrategyType, k8sDaemonSet.Data.Spec.UpdateStrategy.Type)
			*needToUpdate = true
		}
	} else {
		// If backend has no StrategyType but Kubernetes does, mark for update
		if bkcmdbkube.DaemonSetUpdateStrategyType(k8sDaemonSet.Data.Spec.UpdateStrategy.Type) != "" {
			strategyType := string(k8sDaemonSet.Data.Spec.UpdateStrategy.Type)
			updateData.StrategyType = &strategyType
			blog.Infof("CompareDaemonSet StrategyType: %v, %v",
				bkDaemonSet.StrategyType, k8sDaemonSet.Data.Spec.UpdateStrategy.Type)
			*needToUpdate = true
		}
	}
}

// GenerateBkDaemonSet generate bkdaemonset from k8sdaemonset
func (s *Syncer) GenerateBkDaemonSet(
	bkNs *bkcmdbkube.Namespace, k8sDaemonSet *storage.DaemonSet) *client.CreateBcsWorkloadRequestData {
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
	_ = json.Unmarshal(jsonBytes, &rudMap)

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

func rollingUpdateGameDeploymentEqual(
	rug *gdv1alpha1.GameDeploymentUpdateStrategy, bkRug *bkcmdbkube.RollingUpdateGameDeployment) bool {
	if rug == nil && bkRug == nil {
		return true
	}

	if rug != nil && bkRug == nil {
		return false
	}

	if bkRug != nil && rug == nil {
		if bkRug.MaxUnavailable == nil && bkRug.MaxSurge == nil {
			return true
		}
		return false
	}

	if !rudMaxSurge(rug.MaxSurge, bkRug.MaxSurge) {
		return false
	}

	if rug.MaxUnavailable != nil && bkRug.MaxUnavailable == nil {
		return false
	}

	if rug.MaxUnavailable == nil && bkRug.MaxUnavailable != nil {
		return false
	}

	if rug.MaxUnavailable != nil && bkRug.MaxUnavailable != nil {
		if rug.MaxUnavailable.StrVal != bkRug.MaxUnavailable.StrVal {
			return false
		}
	}

	return true
}

// CompareGameDeployment compare bkgamedeployment and k8sgamedeployment
// nolint funlen
func (s *Syncer) CompareGameDeployment(
	bkGameDeployment *bkcmdbkube.GameDeployment, k8sGameDeployment *storage.GameDeployment) (
	needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	needToUpdate = false
	updateData = &client.UpdateBcsWorkloadRequestData{}

	s.compareGameDeployment(bkGameDeployment, k8sGameDeployment, &needToUpdate, updateData)

	if !rollingUpdateGameDeploymentEqual(&k8sGameDeployment.Data.Spec.UpdateStrategy,
		bkGameDeployment.RollingUpdateStrategy) {
		rud := bkcmdbkube.RollingUpdateGameDeployment{}
		if k8sGameDeployment.Data.Spec.UpdateStrategy.MaxUnavailable != nil &&
			k8sGameDeployment.Data.Spec.UpdateStrategy.MaxSurge != nil {
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
			_ = json.Unmarshal(jsonBytes, &rudMap)

			updateData.RollingUpdateStrategy = &rudMap
			blog.Infof("CompareGameDeployment RollingUpdate: %v, %v",
				k8sGameDeployment.Data.Spec.UpdateStrategy, *bkGameDeployment.RollingUpdateStrategy)
			needToUpdate = true

		}
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

// nolint funlen
func (s *Syncer) compareGameDeployment(bkGameDeployment *bkcmdbkube.GameDeployment,
	k8sGameDeployment *storage.GameDeployment, needToUpdate *bool, updateData *client.UpdateBcsWorkloadRequestData) {
	labelsEmpty := map[string]string{}

	if k8sGameDeployment.Data.Labels == nil {
		if bkGameDeployment.Labels != nil {
			updateData.Labels = &labelsEmpty
			blog.Infof("CompareGameDeployment labels: %v, %v",
				k8sGameDeployment.Data.Labels, bkGameDeployment.Labels)
			*needToUpdate = true
		}
	} else if bkGameDeployment.Labels == nil ||
		!mapsEqual(stringToMap(strings.ReplaceAll(mapToString(*bkGameDeployment.Labels), "．", ".")),
			stringToMap(strings.ReplaceAll(mapToString(k8sGameDeployment.Data.Labels), "．", "."))) {

		if len(k8sGameDeployment.Data.Labels) > 0 {
			updateData.Labels = &k8sGameDeployment.Data.Labels
			blog.Infof("CompareGameDeployment labels: %v, %v",
				k8sGameDeployment.Data.Labels, bkGameDeployment.Labels)
			*needToUpdate = true
		}
	}

	if k8sGameDeployment.Data.Spec.Selector == nil {
		if bkGameDeployment.Selector != nil {
			updateData.Selector = nil
			blog.Infof("CompareGameDeployment Selector: %v, %v",
				k8sGameDeployment.Data.Spec.Selector, bkGameDeployment.Selector)
			*needToUpdate = true
		}
	} else if bkGameDeployment.Selector == nil ||
		strings.ReplaceAll(fmt.Sprint(*k8sGameDeployment.Data.Spec.Selector), "．", ".") !=
			strings.ReplaceAll(fmt.Sprint(*bkGameDeployment.Selector), "．", ".") {
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

		blog.Infof("CompareGameDeployment Selector: %v, %v",
			k8sGameDeployment.Data.Spec.Selector, bkGameDeployment.Selector)

		*needToUpdate = true
	}

	if bkGameDeployment.Replicas != nil {
		if *bkGameDeployment.Replicas != int64(*k8sGameDeployment.Data.Spec.Replicas) {
			replicas := int64(*k8sGameDeployment.Data.Spec.Replicas)
			updateData.Replicas = &replicas
			blog.Infof("CompareGameDeployment Replicas: %v, %v",
				*bkGameDeployment.Replicas, *k8sGameDeployment.Data.Spec.Replicas)
			*needToUpdate = true
		}
	} else {
		if int64(*k8sGameDeployment.Data.Spec.Replicas) != 0 {
			replicas := int64(*k8sGameDeployment.Data.Spec.Replicas)
			updateData.Replicas = &replicas
			blog.Infof("CompareGameDeployment Replicas: %v, %v",
				bkGameDeployment.Replicas, *k8sGameDeployment.Data.Spec.Replicas)
			*needToUpdate = true
		}
	}

	if bkGameDeployment.MinReadySeconds != nil {
		if *bkGameDeployment.MinReadySeconds != int64(k8sGameDeployment.Data.Spec.MinReadySeconds) {
			minReadySeconds := int64(k8sGameDeployment.Data.Spec.MinReadySeconds)
			updateData.MinReadySeconds = &minReadySeconds
			blog.Infof("CompareGameDeployment MinReadySeconds: %v, %v",
				*bkGameDeployment.MinReadySeconds, k8sGameDeployment.Data.Spec.MinReadySeconds)
			*needToUpdate = true
		}
	} else {
		if int64(k8sGameDeployment.Data.Spec.MinReadySeconds) != 0 {
			minReadySeconds := int64(k8sGameDeployment.Data.Spec.MinReadySeconds)
			updateData.MinReadySeconds = &minReadySeconds
			blog.Infof("CompareGameDeployment MinReadySeconds: %v, %v",
				bkGameDeployment.MinReadySeconds, k8sGameDeployment.Data.Spec.MinReadySeconds)
			*needToUpdate = true
		}
	}

	if bkGameDeployment.StrategyType != nil {
		if *bkGameDeployment.StrategyType !=
			bkcmdbkube.GameDeploymentUpdateStrategyType(k8sGameDeployment.Data.Spec.UpdateStrategy.Type) {
			strategyType := string(k8sGameDeployment.Data.Spec.UpdateStrategy.Type)
			updateData.StrategyType = &strategyType
			blog.Infof("CompareGameDeployment StrategyType: %v, %v",
				*bkGameDeployment.StrategyType, k8sGameDeployment.Data.Spec.UpdateStrategy.Type)
			*needToUpdate = true
		}
	} else {
		if bkcmdbkube.GameDeploymentUpdateStrategyType(k8sGameDeployment.Data.Spec.UpdateStrategy.Type) != "" {
			strategyType := string(k8sGameDeployment.Data.Spec.UpdateStrategy.Type)
			updateData.StrategyType = &strategyType
			blog.Infof("CompareGameDeployment StrategyType: %v, %v",
				bkGameDeployment.StrategyType, k8sGameDeployment.Data.Spec.UpdateStrategy.Type)
			*needToUpdate = true
		}
	}
}

// GenerateBkGameDeployment generate bkgamedeployment from k8sgamedeployment
func (s *Syncer) GenerateBkGameDeployment(
	bkNs *bkcmdbkube.Namespace, k8sGameDeployment *storage.GameDeployment) *client.CreateBcsWorkloadRequestData {
	me := make([]bkcmdbkube.LabelSelectorRequirement, 0)
	for _, m := range k8sGameDeployment.Data.Spec.Selector.MatchExpressions {
		me = append(me, bkcmdbkube.LabelSelectorRequirement{
			Key:      m.Key,
			Operator: bkcmdbkube.LabelSelectorOperator(m.Operator),
			Values:   m.Values,
		})
	}

	// NOCC:ineffassign/assign(ignore)
	var rud bkcmdbkube.RollingUpdateGameDeployment

	if k8sGameDeployment.Data.Spec.UpdateStrategy.MaxUnavailable != nil &&
		k8sGameDeployment.Data.Spec.UpdateStrategy.MaxSurge != nil {
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
	_ = json.Unmarshal(jsonBytes, &rudMap)

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

// nolint
func rollingUpdateGameStatefulSetStrategyEqual(
	russ *gsv1alpha1.RollingUpdateStatefulSetStrategy, bkRuss *bkcmdbkube.RollingUpdateGameStatefulSetStrategy) bool {
	if russ == nil && bkRuss == nil {
		return true
	}

	if russ != nil && bkRuss == nil {
		return false
	}

	if bkRuss != nil && russ == nil {
		if bkRuss.MaxUnavailable == nil && bkRuss.Partition == nil && bkRuss.MaxSurge == nil {
			return true
		}
		return false
	}

	if !rugssMaxUnavailable(russ.MaxUnavailable, bkRuss.MaxUnavailable) {
		return false
	}

	if !rugssPartition(russ.Partition, bkRuss.Partition) {
		return false
	}

	if !rugssMaxSurge(russ.MaxSurge, bkRuss.MaxSurge) {
		return false
	}

	return true
}

func rugssMaxUnavailable(a *intstr.IntOrString, b *bkcmdbkube.IntOrString) bool {
	if a != nil && b == nil {
		return false
	}
	if a == nil && b != nil {
		return false
	}
	if a != nil && b != nil {
		if a.IntVal != b.IntVal {
			return false
		}
	}
	return true
}

func rugssPartition(a *intstr.IntOrString, b *int32) bool {
	if a != nil && b == nil {
		return false
	}
	if a == nil && b != nil {
		return false
	}
	if a != nil && b != nil {
		if a.IntVal != *b {
			return false
		}
	}
	return true
}

func rugssMaxSurge(a *intstr.IntOrString, b *bkcmdbkube.IntOrString) bool {
	if a != nil && b == nil {
		return false
	}
	if a == nil && b != nil {
		return false
	}
	if a != nil && b != nil {
		if a.StrVal != b.StrVal {
			return false
		}
	}
	return true
}

// CompareGameStatefulSet compare bkgamestatefulset and k8sgamestatefulset
// nolint
func (s *Syncer) CompareGameStatefulSet(
	bkGameStatefulSet *bkcmdbkube.GameStatefulSet, k8sGameStatefulSet *storage.GameStatefulSet) (
	needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	needToUpdate = false
	updateData = &client.UpdateBcsWorkloadRequestData{}

	s.compareGameStatefulSet(bkGameStatefulSet, k8sGameStatefulSet, &needToUpdate, updateData)

	if !rollingUpdateGameStatefulSetStrategyEqual(k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate,
		bkGameStatefulSet.RollingUpdateStrategy) {
		rus := bkcmdbkube.RollingUpdateGameStatefulSetStrategy{}
		if k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate != nil {
			if k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil &&
				k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge != nil &&
				k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
				blog.Infof("rolling update: %+v", k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate)
				rus = bkcmdbkube.RollingUpdateGameStatefulSetStrategy{
					MaxUnavailable: &bkcmdbkube.IntOrString{
						Type:   bkcmdbkube.Type(k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.Type), // nolint
						IntVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.IntVal,
						StrVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.StrVal,
					},
					MaxSurge: &bkcmdbkube.IntOrString{
						Type:   bkcmdbkube.Type(k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.Type),
						IntVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.IntVal,
						StrVal: k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge.StrVal,
					},
					Partition: &k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.Partition.IntVal,
				}

				jsonBytes, err := json.Marshal(rus)
				if err != nil {
					blog.Errorf("marshal rolling update gameStatefulSets failed, err: %s", err.Error())
					return false, nil
				}

				rudMap := make(map[string]interface{})
				_ = json.Unmarshal(jsonBytes, &rudMap)

				updateData.RollingUpdateStrategy = &rudMap
				blog.Infof("CompareGameStatefulSet RollingUpdate: %v, %v %v %v",
					k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate,
					bkGameStatefulSet.RollingUpdateStrategy.Partition,
					bkGameStatefulSet.RollingUpdateStrategy.MaxUnavailable,
					bkGameStatefulSet.RollingUpdateStrategy.MaxSurge)
				needToUpdate = true
			}
		}
	}

	if !needToUpdate {
		updateData = nil
	}

	return needToUpdate, updateData
}

func (s *Syncer) compareGameStatefulSet(bkGameStatefulSet *bkcmdbkube.GameStatefulSet,
	k8sGameStatefulSet *storage.GameStatefulSet, needToUpdate *bool, updateData *client.UpdateBcsWorkloadRequestData) {
	if k8sGameStatefulSet.Data.Labels == nil {
		if bkGameStatefulSet.Labels != nil {
			updateData.Labels = nil
			blog.Infof("CompareGameStatefulSet labels: %v, %v",
				k8sGameStatefulSet.Data.Labels, bkGameStatefulSet.Labels)
			*needToUpdate = true
		}
	} else if bkGameStatefulSet.Labels == nil ||
		!mapsEqual(stringToMap(strings.ReplaceAll(mapToString(*bkGameStatefulSet.Labels), "．", ".")),
			stringToMap(strings.ReplaceAll(mapToString(k8sGameStatefulSet.Data.Labels), "．", "."))) {
		if len(k8sGameStatefulSet.Data.Labels) > 0 {
			updateData.Labels = &k8sGameStatefulSet.Data.Labels
			blog.Infof("CompareGameStatefulSet labels: %v, %v",
				k8sGameStatefulSet.Data.Labels, bkGameStatefulSet.Labels)
			*needToUpdate = true
		}
	}

	if k8sGameStatefulSet.Data.Spec.Selector == nil {
		if bkGameStatefulSet.Selector != nil {
			updateData.Selector = nil
			blog.Infof("CompareGameStatefulSet Selector: %v, %v",
				k8sGameStatefulSet.Data.Spec.Selector, bkGameStatefulSet.Selector)
			*needToUpdate = true
		}
	} else if bkGameStatefulSet.Selector == nil ||
		strings.ReplaceAll(fmt.Sprint(*k8sGameStatefulSet.Data.Spec.Selector), "．", ".") !=
			strings.ReplaceAll(fmt.Sprint(*bkGameStatefulSet.Selector), "．", ".") {
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

		blog.Infof("CompareGameStatefulSet Selector: %v, %v",
			k8sGameStatefulSet.Data.Spec.Selector, bkGameStatefulSet.Selector)

		*needToUpdate = true
	}

	if bkGameStatefulSet.Replicas != nil {
		if *bkGameStatefulSet.Replicas != int64(*k8sGameStatefulSet.Data.Spec.Replicas) {
			replicas := int64(*k8sGameStatefulSet.Data.Spec.Replicas)
			updateData.Replicas = &replicas
			blog.Infof("CompareGameStatefulSet Replicas: %v, %v",
				*k8sGameStatefulSet.Data.Spec.Replicas, bkGameStatefulSet.Replicas)
			*needToUpdate = true
		}
	} else {
		if k8sGameStatefulSet.Data.Spec.Replicas != nil && int64(*k8sGameStatefulSet.Data.Spec.Replicas) != 0 {
			replicas := int64(*k8sGameStatefulSet.Data.Spec.Replicas)
			updateData.Replicas = &replicas
			blog.Infof("CompareGameStatefulSet Replicas: %v, %v",
				k8sGameStatefulSet.Data.Spec.Replicas, bkGameStatefulSet.Replicas)
			*needToUpdate = true
		}
	}

	if bkGameStatefulSet.StrategyType != nil {
		if *bkGameStatefulSet.StrategyType !=
			bkcmdbkube.GameStatefulSetUpdateStrategyType(k8sGameStatefulSet.Data.Spec.UpdateStrategy.Type) {
			strategyType := string(k8sGameStatefulSet.Data.Spec.UpdateStrategy.Type)
			updateData.StrategyType = &strategyType
			blog.Infof("CompareGameStatefulSet StrategyType: %v, %v",
				k8sGameStatefulSet.Data.Spec.UpdateStrategy.Type, bkGameStatefulSet.StrategyType)
			*needToUpdate = true
		}
	} else {
		if bkcmdbkube.GameStatefulSetUpdateStrategyType(k8sGameStatefulSet.Data.Spec.UpdateStrategy.Type) != "" {
			strategyType := string(k8sGameStatefulSet.Data.Spec.UpdateStrategy.Type)
			updateData.StrategyType = &strategyType
			blog.Infof("CompareGameStatefulSet StrategyType: %v, %v",
				k8sGameStatefulSet.Data.Spec.UpdateStrategy.Type, bkGameStatefulSet.StrategyType)
			*needToUpdate = true
		}
	}
}

// GenerateBkGameStatefulSet generate bkgamestatefulset from k8sgamestatefulset
func (s *Syncer) GenerateBkGameStatefulSet(
	bkNs *bkcmdbkube.Namespace, k8sGameStatefulSet *storage.GameStatefulSet) *client.CreateBcsWorkloadRequestData {
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
		if k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable != nil &&
			k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxSurge != nil {
			blog.Infof("rolling update: %+v", k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate)
			rus = bkcmdbkube.RollingUpdateGameStatefulSetStrategy{
				MaxUnavailable: &bkcmdbkube.IntOrString{
					Type:   bkcmdbkube.Type(k8sGameStatefulSet.Data.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.Type), // nolint
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
	_ = json.Unmarshal(jsonBytes, &rusMap)

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
func (s *Syncer) UpdateBkWorkloads(
	bkCluster *bkcmdbkube.Cluster, kind string, toUpdate *map[int64]*client.UpdateBcsWorkloadRequestData, db *gorm.DB) {
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
		}, db)

		if err != nil {
			blog.Errorf("update workload %s failed, err: %s", kind, err.Error())
		}
	}
}

// DeleteBkWorkloads delete bkworkloads
func (s *Syncer) DeleteBkWorkloads(bkCluster *bkcmdbkube.Cluster, kind string, toDelete *[]int64, db *gorm.DB) error {
	if len(*toDelete) > 0 {
		// DeleteBcsWorkload deletes the BCS workload with the given request.
		err := s.CMDBClient.DeleteBcsWorkload(&client.DeleteBcsWorkloadRequest{
			BKBizID: &bkCluster.BizID,
			Kind:    &kind,
			IDs:     toDelete,
		}, db)
		if err != nil {
			for i := 0; i < len(*toDelete); i++ {
				var section []int64
				if i+1 > len(*toDelete) {
					section = (*toDelete)[i:]
				} else {
					section = (*toDelete)[i : i+1]
				}
				err = s.CMDBClient.DeleteBcsWorkload(&client.DeleteBcsWorkloadRequest{
					BKBizID: &bkCluster.BizID,
					Kind:    &kind,
					IDs:     &section,
				}, db)

				if err != nil {
					blog.Errorf("delete workload %s failed, err: %s", kind, err.Error())
				}
			}
		}
		blog.Infof("delete workload %s success, ids: %v", kind, toDelete)
	}
	return nil
}

// GetBkPods get bkpods
func (s *Syncer) GetBkPods(
	bkBizID int64, filter *client.PropertyFilter, withDB bool, db *gorm.DB) (*[]bkcmdbkube.Pod, error) {
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
		}, db, withDB)

		if err != nil {
			blog.Errorf("get bcs pod failed, err: %s", err.Error())
			return nil, err
		}
		bkPodList = append(bkPodList, *bkPods...)

		if len(*bkPods) < 100 {
			break
		}
		if withDB {
			break
		}
		pageStart++
	}

	return &bkPodList, nil
}

// GetBkContainers get bkcontainers
func (s *Syncer) GetBkContainers(
	bkBizID int64, filter *client.PropertyFilter, withDB bool, db *gorm.DB) (*[]client.Container, error) {
	bkContainerList := make([]client.Container, 0)

	pageStart := 0
	for {
		// GetBcsPod returns the BCS pod information for the given request.
		bkContainers, err := s.CMDBClient.GetBcsContainer(&client.GetBcsContainerRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkBizID,
				Page: client.Page{
					Limit: 100,
					Start: 100 * pageStart,
				},
				Filter: filter,
			},
		}, db, withDB)

		if err != nil {
			blog.Errorf("get bcs container failed, err: %s", err.Error())
			return nil, err
		}
		bkContainerList = append(bkContainerList, *bkContainers...)

		if len(*bkContainers) < 100 {
			break
		}
		if withDB {
			break
		}
		pageStart++
	}

	return &bkContainerList, nil
}

// CreateBkPods create bkpods
func (s *Syncer) CreateBkPods(
	bkCluster *bkcmdbkube.Cluster, toCreate map[int64][]client.CreateBcsPodRequestDataPod, db *gorm.DB) {
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
					}, db)

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
							}, db)
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
func (s *Syncer) DeleteBkPods(bkCluster *bkcmdbkube.Cluster, toDelete *[]int64, db *gorm.DB) error {
	if len(*toDelete) > 0 {
		for i := 0; i < len(*toDelete); i += 100 {
			var ids []int64
			if i+100 > len(*toDelete) {
				ids = (*toDelete)[i:]
			} else {
				ids = (*toDelete)[i : i+100]
			}
			// DeleteBcsPod deletes the BCS pod with the given request.
			err := s.CMDBClient.DeleteBcsPod(&client.DeleteBcsPodRequest{
				Data: &[]client.DeleteBcsPodRequestData{
					{
						BKBizID: &bkCluster.BizID,
						IDs:     &ids,
					},
				},
			}, db)
			if err != nil {
				blog.Errorf("delete pod failed, err: %s", err.Error())
			}
			blog.Infof("delete pod success, ids: %v", toDelete)
		}
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
func (s *Syncer) CompareBkWorkloadPods(workload *bkcmdbkube.PodsWorkload) (
	needToUpdate bool, updateData *client.UpdateBcsWorkloadRequestData) {
	return false, nil
}

// DeleteAllByCluster clean by cluster
// nolint
func (s *Syncer) DeleteAllByCluster(bkCluster *bkcmdbkube.Cluster) error {
	blog.Infof("start delete all: %s", bkCluster.Uid)
	blog.Infof("start delete all pod: %s", bkCluster.Uid)
	for {
		got, err := s.CMDBClient.GetBcsPod(&client.GetBcsPodRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkCluster.BizID,
				Fields:  []string{"id"},
				Page: client.Page{
					Limit: 200,
					Start: 0,
				},
				Filter: &client.PropertyFilter{
					Condition: "AND",
					Rules: []client.Rule{
						{
							Field:    "cluster_uid",
							Operator: "in",
							Value:    []string{bkCluster.Uid},
						},
					},
				},
			},
		}, nil, false)
		if err != nil {
			blog.Errorf("GetBcsPod() error = %v", err)
			return fmt.Errorf("GetBcsPod() error = %v", err)
		}
		podToDelete := make([]int64, 0)
		for _, pod := range *got {
			podToDelete = append(podToDelete, pod.ID)
		}

		if len(podToDelete) == 0 {
			break
		} else {
			blog.Infof("delete pod: %v", podToDelete)
			err := s.CMDBClient.DeleteBcsPod(&client.DeleteBcsPodRequest{
				Data: &[]client.DeleteBcsPodRequestData{
					{
						BKBizID: &bkCluster.BizID,
						IDs:     &podToDelete,
					},
				},
			}, nil)
			if err != nil {
				blog.Errorf("DeleteBcsPod() error = %v", err)
				return fmt.Errorf("DeleteBcsPod() error = %v", err)
			}
		}
	}
	blog.Infof("delete all pod success: %s", bkCluster.Uid)

	blog.Infof("start delete all workload: %s", bkCluster.Uid)
	workloadTypes := []string{"deployment", "statefulSet", "daemonSet", "gameDeployment", "gameStatefulSet", "pods"}

	for _, workloadType := range workloadTypes {
		for {
			got, err := s.CMDBClient.GetBcsWorkload(&client.GetBcsWorkloadRequest{
				CommonRequest: client.CommonRequest{
					BKBizID: bkCluster.BizID,
					Fields:  []string{"id"},
					Page: client.Page{
						Limit: 200,
						Start: 0,
					},
					Filter: &client.PropertyFilter{
						Condition: "AND",
						Rules: []client.Rule{
							{
								Field:    "cluster_uid",
								Operator: "in",
								Value:    []string{bkCluster.Uid},
							},
						},
					},
				},
				Kind: workloadType,
			}, nil, false)
			if err != nil {
				blog.Errorf("GetBcsWorkload() error = %v", err)
				return fmt.Errorf("GetBcsWorkload() error = %v", err)
			}
			workloadToDelete := make([]int64, 0)
			for _, workload := range *got {
				workloadToDelete = append(workloadToDelete, (int64)(workload.(map[string]interface{})["id"].(float64)))
			}

			if len(workloadToDelete) == 0 {
				break
			} else {
				blog.Infof("delete workload: %v", workloadToDelete)
				err := s.CMDBClient.DeleteBcsWorkload(&client.DeleteBcsWorkloadRequest{
					BKBizID: &bkCluster.BizID,
					Kind:    &workloadType,
					IDs:     &workloadToDelete,
				}, nil)
				if err != nil {
					blog.Errorf("DeleteBcsWorkload() error = %v", err)
					return fmt.Errorf("DeleteBcsWorkload() error = %v", err)
				}
			}
		}
	}
	blog.Infof("delete all workload success: %s", bkCluster.Uid)

	blog.Infof("start delete all namespace: %s", bkCluster.Uid)
	if errDN := s.deleteAllByClusterNamespace(bkCluster); errDN != nil {
		return errDN
	}
	blog.Infof("delete all namespace success: %s", bkCluster.Uid)

	blog.Infof("start delete all node: %s", bkCluster.Uid)
	if errDN := s.deleteAllByClusterNode(bkCluster); errDN != nil {
		return errDN
	}
	blog.Infof("delete all node success: %s", bkCluster.Uid)

	blog.Infof("start delete all cluster: %s", bkCluster.Uid)
	if errDC := s.deleteAllByClusterCluster(bkCluster); errDC != nil {
		return errDC
	}
	blog.Infof("delete all cluster success: %s", bkCluster.Uid)
	blog.Infof("delete all success: %s", bkCluster.Uid)
	return nil
}

// DeleteAllByClusterAndNamespace clean by cluster and namespace
// nolint
func (s *Syncer) DeleteAllByClusterAndNamespace(bkCluster *bkcmdbkube.Cluster, bkNamespace *bkcmdbkube.Namespace, db *gorm.DB) error {
	blog.Infof("start delete all: %s namespace: %s", bkCluster.Uid, bkNamespace.Name)
	for {
		got, err := s.CMDBClient.GetBcsPod(&client.GetBcsPodRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkCluster.BizID,
				Fields:  []string{"id"},
				Page: client.Page{
					Limit: 200,
					Start: 0,
				},
				Filter: &client.PropertyFilter{
					Condition: "AND",
					Rules: []client.Rule{
						{
							Field:    "cluster_uid",
							Operator: "in",
							Value:    []string{bkCluster.Uid},
						},
						{
							Field:    "namespace",
							Operator: "in",
							Value:    []string{bkNamespace.Name},
						},
					},
				},
			},
		}, nil, false)
		if err != nil {
			blog.Errorf("GetBcsPod() error = %v", err)
			return fmt.Errorf("GetBcsPod() error = %v", err)
		}
		podToDelete := make([]int64, 0)
		for _, pod := range *got {
			podToDelete = append(podToDelete, pod.ID)
		}

		if len(podToDelete) == 0 {
			break
		} else {
			blog.Infof("delete pod: %v", podToDelete)
			err := s.CMDBClient.DeleteBcsPod(&client.DeleteBcsPodRequest{
				Data: &[]client.DeleteBcsPodRequestData{
					{
						BKBizID: &bkCluster.BizID,
						IDs:     &podToDelete,
					},
				},
			}, db)
			if err != nil {
				blog.Errorf("DeleteBcsPod() error = %v", err)
				return fmt.Errorf("DeleteBcsPod() error = %v", err)
			}
		}
	}
	blog.Infof("delete all pod success: %s", bkCluster.Uid)

	blog.Infof("start delete all workload: %s", bkCluster.Uid)
	workloadTypes := []string{"deployment", "statefulSet", "daemonSet", "gameDeployment", "gameStatefulSet", "pods"}

	for _, workloadType := range workloadTypes {
		for {
			got, err := s.CMDBClient.GetBcsWorkload(&client.GetBcsWorkloadRequest{
				CommonRequest: client.CommonRequest{
					BKBizID: bkCluster.BizID,
					Fields:  []string{"id"},
					Page: client.Page{
						Limit: 200,
						Start: 0,
					},
					Filter: &client.PropertyFilter{
						Condition: "AND",
						Rules: []client.Rule{
							{
								Field:    "cluster_uid",
								Operator: "in",
								Value:    []string{bkCluster.Uid},
							},
							{
								Field:    "namespace",
								Operator: "in",
								Value:    []string{bkNamespace.Name},
							},
						},
					},
				},
				Kind: workloadType,
			}, nil, false)
			if err != nil {
				blog.Errorf("GetBcsWorkload() error = %v", err)
				return fmt.Errorf("GetBcsWorkload() error = %v", err)
			}
			workloadToDelete := make([]int64, 0)
			for _, workload := range *got {
				workloadToDelete = append(workloadToDelete, (int64)(workload.(map[string]interface{})["id"].(float64)))
			}

			if len(workloadToDelete) == 0 {
				break
			} else {
				blog.Infof("delete workload: %v", workloadToDelete)
				err := s.CMDBClient.DeleteBcsWorkload(&client.DeleteBcsWorkloadRequest{
					BKBizID: &bkCluster.BizID,
					Kind:    &workloadType,
					IDs:     &workloadToDelete,
				}, db)
				if err != nil {
					blog.Errorf("DeleteBcsWorkload() error = %v", err)
					return fmt.Errorf("DeleteBcsWorkload() error = %v", err)
				}
			}
		}
	}
	blog.Infof("delete all workload success: %s", bkCluster.Uid)

	blog.Infof("start delete all namespace: %s", bkCluster.Uid)
	err := s.CMDBClient.DeleteBcsNamespace(&client.DeleteBcsNamespaceRequest{
		BKBizID: &bkCluster.BizID,
		IDs:     &[]int64{bkNamespace.ID},
	}, db)
	if err != nil {
		blog.Errorf("DeleteBcsNamespace() error = %v", err)
		return fmt.Errorf("DeleteBcsNamespace() error = %v", err)
	}
	blog.Infof("delete all cluster success: %s", bkCluster.Uid)
	blog.Infof("delete all success: %s", bkCluster.Uid)
	return nil
}

func (s *Syncer) deleteAllByClusterCluster(bkCluster *bkcmdbkube.Cluster) error {
	for {
		got, err := s.CMDBClient.GetBcsCluster(&client.GetBcsClusterRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkCluster.BizID,
				Fields:  []string{"id"},
				Page: client.Page{
					Limit: 10,
					Start: 0,
				},
				Filter: &client.PropertyFilter{
					Condition: "AND",
					Rules: []client.Rule{
						{
							Field:    "id",
							Operator: "in",
							Value:    []int64{bkCluster.ID},
						},
					},
				},
			},
		}, nil, false)
		if err != nil {
			blog.Errorf("GetBcsCluster() error = %v", err)
			return fmt.Errorf("GetBcsCluster() error = %v", err)
		}
		clusterToDelete := make([]int64, 0)
		for _, cluster := range *got {
			clusterToDelete = append(clusterToDelete, cluster.ID)
		}

		if len(clusterToDelete) == 0 {
			break
		} else {
			blog.Infof("delete cluster: %v", clusterToDelete)
			errDC := s.CMDBClient.DeleteBcsCluster(&client.DeleteBcsClusterRequest{
				BKBizID: &bkCluster.BizID,
				IDs:     &clusterToDelete,
			}, nil)
			if errDC != nil {
				blog.Errorf("DeleteBcsCluster() error = %v", errDC)
				return fmt.Errorf("DeleteBcsCluster() error = %v", errDC)
			}
		}
	}
	return nil
}

func (s *Syncer) deleteAllByClusterNode(bkCluster *bkcmdbkube.Cluster) error {
	for {
		got, err := s.CMDBClient.GetBcsNode(&client.GetBcsNodeRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkCluster.BizID,
				Page: client.Page{
					Limit: 100,
					Start: 0,
				},
				Filter: &client.PropertyFilter{
					Condition: "AND",
					Rules: []client.Rule{
						{
							Field:    "cluster_uid",
							Operator: "in",
							Value:    []string{bkCluster.Uid},
						},
					},
				},
			},
		}, nil, false)
		if err != nil {
			blog.Errorf("GetBcsNode() error = %v", err)
			return fmt.Errorf("GetBcsNode() error = %v", err)
		}
		nodeToDelete := make([]int64, 0)
		for _, node := range *got {
			nodeToDelete = append(nodeToDelete, node.ID)
		}

		if len(nodeToDelete) == 0 {
			break
		} else {
			blog.Infof("delete node: %v", nodeToDelete)
			err := s.CMDBClient.DeleteBcsNode(&client.DeleteBcsNodeRequest{
				BKBizID: &bkCluster.BizID,
				IDs:     &nodeToDelete,
			}, nil)
			if err != nil {
				blog.Errorf("DeleteBcsNode() error = %v", err)
				return fmt.Errorf("DeleteBcsNode() error = %v", err)
			}
		}
	}
	return nil
}

func (s *Syncer) deleteAllByClusterNamespace(bkCluster *bkcmdbkube.Cluster) error {
	for {
		got, err := s.CMDBClient.GetBcsNamespace(&client.GetBcsNamespaceRequest{
			CommonRequest: client.CommonRequest{
				BKBizID: bkCluster.BizID,
				Fields:  []string{"id"},
				Page: client.Page{
					Limit: 200,
					Start: 0,
				},
				Filter: &client.PropertyFilter{
					Condition: "AND",
					Rules: []client.Rule{
						{
							Field:    "cluster_uid",
							Operator: "in",
							Value:    []string{bkCluster.Uid},
						},
					},
				},
			},
		}, nil, false)
		if err != nil {
			blog.Errorf("GetBcsNamespace() error = %v", err)
			return fmt.Errorf("GetBcsNamespace() error = %v", err)
		}
		namespaceToDelete := make([]int64, 0)
		for _, namespace := range *got {
			namespaceToDelete = append(namespaceToDelete, namespace.ID)
		}

		if len(namespaceToDelete) == 0 {
			break
		} else {
			blog.Infof("delete namespace: %v", namespaceToDelete)
			err := s.CMDBClient.DeleteBcsNamespace(&client.DeleteBcsNamespaceRequest{
				BKBizID: &bkCluster.BizID,
				IDs:     &namespaceToDelete,
			}, nil)
			if err != nil {
				blog.Errorf("DeleteBcsNamespace() error = %v", err)
				return fmt.Errorf("DeleteBcsNamespace() error = %v", err)
			}
		}
	}
	return nil
}
