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

// Package utils xxx
package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	// NodeTemplate node template
	NodeTemplate = "nt"
	// GroupTemplate node group template
	GroupTemplate = "ng"
	// NotifyTemplate notify template
	NotifyTemplate = "nf"
)

// CheckClusterConnection check cluster connection when delete cluster or other scenes
func CheckClusterConnection(operator *clusterops.K8SOperator, clusterID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	err := operator.CheckClusterConnection(ctx, clusterID)
	if err != nil {
		blog.Errorf("CheckClusterConnection[%s] failed: %v", clusterID, err)
		return false
	}

	blog.Infof("CheckClusterConnection[%s] success", clusterID)
	return true
}

// GetCloudZones get cloud region zones
func GetCloudZones(cls *proto.Cluster, cloud *proto.Cloud) ([]*proto.ZoneInfo, error) {
	nodeMgr, err := cloudprovider.GetNodeMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s NodeManager getCloudZones failed, %s", cloud.CloudProvider, err.Error())
		return nil, err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cls.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s getCloudZones failed, %s",
			cloud.CloudID, cloud.CloudProvider, err.Error())
		return nil, err
	}
	cmOption.Region = cls.Region

	return nodeMgr.GetZoneList(&cloudprovider.GetZoneListOption{CommonOption: *cmOption})
}

// CheckIfGetNodesFromCluster check if get k8s nodes from cluster
func CheckIfGetNodesFromCluster(cls *proto.Cluster, cloud *proto.Cloud, nodes []*proto.ClusterNode) bool {
	clsMgr, err := cloudprovider.GetClusterMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("CheckIfGetNodesFromCluster[%s] failed: %v", cls.ClusterID, err)
		return false
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cls.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s getCloudInstanceList failed, %s",
			cloud.CloudID, cloud.CloudProvider, err.Error())
		return false
	}
	cmOption.Region = cls.Region

	return clsMgr.CheckIfGetNodesFromCluster(context.Background(), cls, nodes)
}

// UpdateClusterCloudInfo update cloud cluster info
func UpdateClusterCloudInfo(cls *proto.Cluster, cloud *proto.Cloud) error {
	cloudMgr, err := cloudprovider.GetCloudInfoMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("UpdateClusterCloudInfo[%s] failed: %v", cls.ClusterID, err)
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cls.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s getCloudInstanceList failed, %s",
			cloud.CloudID, cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = cls.Region

	return cloudMgr.UpdateClusterCloudInfo(cls)
}

// GetCloudInstanceList get cloud instances info
func GetCloudInstanceList(ips []string, cls *proto.Cluster, cloud *proto.Cloud) ([]*proto.Node, error) {
	nodeMgr, err := cloudprovider.GetNodeMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s NodeManager getCloudInstanceList failed, %s", cloud.CloudProvider, err.Error())
		return nil, err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cls.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s getCloudInstanceList failed, %s",
			cloud.CloudID, cloud.CloudProvider, err.Error())
		return nil, err
	}
	cmOption.Region = cls.Region

	return nodeMgr.ListNodesByIP(ips, &cloudprovider.ListNodesOption{Common: cmOption})
}

// FormatTaskTime format task time
func FormatTaskTime(t *proto.Task) {
	if t.Start != "" {
		t.Start = utils.TransTimeFormat(t.Start)
	}
	if t.End != "" {
		t.End = utils.TransTimeFormat(t.End)
	}
	for i := range t.Steps {
		if t.Steps[i].Start != "" {
			t.Steps[i].Start = utils.TransTimeFormat(t.Steps[i].Start)
		}
		if t.Steps[i].End != "" {
			t.Steps[i].End = utils.TransTimeFormat(t.Steps[i].End)
		}
	}
}

// Passwd flag
var Passwd = []string{"password", "passwd"}

// HandleTaskStepData handle task step data(hidden passwd, step name)
func HandleTaskStepData(ctx context.Context, task *proto.Task) {
	if task != nil && len(task.Steps) > 0 {
		for i := range task.Steps {

			task.Steps[i].TaskName = Translate(ctx, task.Steps[i].TaskMethod,
				task.Steps[i].TaskName, task.Steps[i].Translate)

			for k := range task.Steps[i].Params {
				if utils.StringInSlice(k, []string{cloudprovider.BkSopsTaskUrlKey.String(),
					cloudprovider.ShowSopsUrlKey.String(), cloudprovider.ConnectClusterKey.String(),
					cloudprovider.InstallGseAgentKey.String()}) {
					continue
				}
				delete(task.Steps[i].Params, k)
			}
		}
	}

	if task != nil && len(task.CommonParams) > 0 {
		for k, v := range task.CommonParams {
			if utils.StringInSlice(strings.ToLower(k), Passwd) || utils.StringContainInSlice(v, Passwd) ||
				utils.StringInSlice(k, []string{cloudprovider.DynamicClusterKubeConfigKey.String()}) {
				delete(task.CommonParams, k)
			}
		}
	}
}

// GenerateTemplateID generate random templateID
func GenerateTemplateID(templateType string) string {
	randomStr := utils.RandomString(8)

	return fmt.Sprintf("BCS-%s-%s", templateType, randomStr)
}

// IsKubeConfigImportCluster kubeconfig cluster
func IsKubeConfigImportCluster(cls *proto.Cluster) bool {
	if cls.GetClusterCategory() == common.Importer && cls.GetImportCategory() == common.KubeConfigImport {
		return true
	}

	return false
}

// IsCloudImportCluster cloud cluster
func IsCloudImportCluster(cls *proto.Cluster) bool {
	if cls.GetClusterCategory() == common.Importer && cls.GetImportCategory() == common.CloudImport {
		return true
	}

	return false
}

// CheckClusterNodeNum check managed cluster nodes num
func CheckClusterNodeNum(model store.ClusterManagerModel, cls *proto.Cluster) (bool, error) {
	nodeStatus := []string{common.StatusRunning, common.StatusInitialization}
	nodes, err := GetClusterStatusNodes(model, cls, nodeStatus)
	if err != nil {
		blog.Errorf("checkManagedClusterNodeNum[%s] GetClusterStatusNodes failed: %v", cls.ClusterID, err)
		return false, err
	}

	blog.Infof("checkManagedClusterNodeNum[%s] GetClusterStatusNodes[%v]", cls.ClusterID, len(nodes))

	if len(nodes) > 0 {
		return false, nil
	}

	return true, nil
}

// GetClusterStatusNodes get cluster status nodes
func GetClusterStatusNodes(
	store store.ClusterManagerModel, cls *proto.Cluster, status []string) ([]*proto.Node, error) {
	clusterCond := operator.NewLeafCondition(operator.Eq, operator.M{"clusterid": cls.ClusterID})
	statusCond := operator.NewLeafCondition(operator.In, operator.M{"status": status})
	cond := operator.NewBranchCondition(operator.And, clusterCond, statusCond)

	nodes, err := store.ListNode(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("get Cluster %s all Nodes failed when AddNodesToCluster, %s", cls.ClusterID, err.Error())
		return nil, err
	}

	return nodes, nil
}
