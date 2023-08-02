/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nodegroup

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/taskserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CreateAction action for create nodeGroup
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateNodeGroupRequest
	resp  *cmproto.CreateNodeGroupResponse

	group   *cmproto.NodeGroup
	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
}

// NewCreateAction create namespace action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) getRelativeResource() error {
	//get relative cluster for information injection
	cluster, err := ca.model.GetCluster(ca.ctx, ca.req.ClusterID)
	if err != nil {
		blog.Errorf("can not get relative Cluster %s when create NodeGroup", ca.req.ClusterID)
		return fmt.Errorf("get relative cluster %s info err, %s", ca.req.ClusterID, err.Error())
	}
	ca.cluster = cluster

	// clusterManager and nodeGroup is different, clusterManager tencent_cloud, nodeGroup other provider
	// if nodeGroup provider is null, use cluster provider
	if len(ca.req.Provider) == 0 {
		ca.req.Provider = cluster.Provider
	}

	cloud, err := actions.GetCloudByCloudID(ca.model, ca.req.Provider)
	if err != nil {
		blog.Errorf("can not get relative Cloud %s when create NodeGroup for Cluster %s, %s",
			ca.req.Provider, ca.req.ClusterID, err.Error(),
		)
		return err
	}
	ca.cloud = cloud

	return nil
}

func (ca *CreateAction) constructNodeGroup() *cmproto.NodeGroup {
	timeStr := time.Now().Format(time.RFC3339)
	group := &cmproto.NodeGroup{
		NodeGroupID:     ca.generateNodeGroupID(),
		Name:            ca.req.Name,
		ClusterID:       ca.req.ClusterID,
		Region:          ca.req.Region,
		ProjectID:       ca.cluster.ProjectID,
		EnableAutoscale: ca.req.EnableAutoscale,
		AutoScaling:     ca.req.AutoScaling,
		LaunchTemplate:  ca.req.LaunchTemplate,
		NodeTemplate:    ca.req.NodeTemplate,
		Labels:          ca.req.Labels,
		Taints:          ca.req.Taints,
		Tags:            ca.req.Tags,
		NodeOS:          ca.req.NodeOS,
		Provider:        ca.req.Provider,
		Status:          common.StatusCreateNodeGroupCreating,
		ConsumerID:      ca.req.ConsumerID,
		Creator:         ca.req.Creator,
		CreateTime:      timeStr,
		UpdateTime:      timeStr,
		Area: &cmproto.CloudArea{
			BkCloudID:   ca.req.BkCloudID,
			BkCloudName: ca.req.CloudAreaName,
		},
		NodeGroupType: func() string {
			// default cloud nodePool
			_, ok := common.NodeGroupTypeMap[common.NodeGroupType(ca.req.NodeGroupType)]
			if ok {
				return ca.req.NodeGroupType
			}
			return common.Normal.String()
		}(),
	}
	// set default parameters
	if group.Region == "" {
		group.Region = ca.cluster.Region
	}
	if group.ProjectID == "" {
		group.ProjectID = ca.cluster.ProjectID
	}
	if group.Provider == "" {
		group.Provider = ca.cluster.Provider
	}

	// base64 encode secret file
	if group.LaunchTemplate != nil && group.LaunchTemplate.KeyPair != nil &&
		len(group.LaunchTemplate.KeyPair.KeySecret) > 0 {
		group.LaunchTemplate.KeyPair.KeySecret = utils.Base64Encode(group.LaunchTemplate.KeyPair.KeySecret)
	}

	// base64 encode script file
	if group.NodeTemplate != nil {
		group.NodeTemplate.UserScript = utils.Base64Encode(group.NodeTemplate.UserScript)
		group.NodeTemplate.PreStartUserScript = utils.Base64Encode(group.NodeTemplate.PreStartUserScript)
		group.NodeTemplate.ScaleInPreScript = utils.Base64Encode(group.NodeTemplate.ScaleInPreScript)
		group.NodeTemplate.ScaleInPostScript = utils.Base64Encode(group.NodeTemplate.ScaleInPostScript)
	}

	return group
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ca *CreateAction) generateNodeGroupID() string {
	str := utils.RandomString(8)
	return fmt.Sprintf("BCS-ng-%s", str)
}

func (ca *CreateAction) validate() error {
	if err := ca.req.Validate(); err != nil {
		return err
	}
	if ca.req.ClusterID == "" {
		return fmt.Errorf("clusterID is empty")
	}
	if ca.req.AutoScaling == nil {
		return fmt.Errorf("autoScaling is empty")
	}
	if ca.req.LaunchTemplate == nil {
		return fmt.Errorf("launchTemplate is empty")
	}
	if ca.req.NodeTemplate == nil {
		return fmt.Errorf("nodeTemplate is empty")
	}
	// not external nodeGroup check launchTemplate
	if ca.req.NodeGroupType != common.External.String() {
		err := validateLaunchTemplate(ca.req.LaunchTemplate)
		if err != nil {
			return err
		}
	}

	// cloud special validate info
	cloudValidate, err := cloudprovider.GetCloudValidateMgr(ca.cloud.CloudProvider)
	if err != nil {
		return err
	}
	// first, get cloud credentialInfo from project; second, get from cloud provider when failed to obtain
	cOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ca.cloud,
		AccountID: ca.cluster.GetCloudAccountID(),
	})
	if err != nil {
		blog.Errorf("Get Credential failed from Cloud %s: %s",
			ca.cloud.CloudID, err.Error())
		return err
	}
	cOption.Region = ca.cluster.Region

	err = cloudValidate.CreateNodeGroupValidate(ca.req, cOption)
	if err != nil {
		return err
	}

	return nil
}

func (ca *CreateAction) saveNodeGroup() error {
	group := ca.constructNodeGroup()

	// store NodeGroup information to DB
	if err := ca.model.CreateNodeGroup(ca.ctx, group); err != nil {
		blog.Errorf("store nodegroup %+v information to DB failed, %s", group, err.Error())
		return err
	}
	ca.group = removeSensitiveInfo(group)
	ca.resp.Data.NodeGroup = group

	blog.Infof("create nodegroup %s information for Cluster %s to DB successfully", group, ca.cluster.ClusterID)

	return nil
}

func (ca *CreateAction) createNodeGroup() error {
	// create nodegroup with cloudprovider
	mgr, err := cloudprovider.GetNodeGroupMgr(ca.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for create nodegroup in Cluster %s failed, %s",
			ca.cloud.CloudID, ca.cloud.CloudProvider, ca.cluster.ClusterID, err.Error(),
		)
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ca.cloud,
		AccountID: ca.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get Credential for Cloud %s/%s when create NodeGroup for cluster %s failed, %s",
			ca.cloud.CloudID, ca.cloud.CloudProvider, ca.cluster.ClusterID, err.Error())
		return err
	}
	cmOption.Region = ca.cluster.Region
	// cloud provider nodeGroup
	task, err := mgr.CreateNodeGroup(ca.group, &cloudprovider.CreateNodeGroupOption{
		CommonOption: *cmOption,
		Cluster:      ca.cluster,
		PoolInfo: cloudprovider.ResourcePoolData{
			Provider: func() string {
				if ca.req.Extra != nil {
					return ca.req.Extra.Provider
				}
				return ""
			}(),
			ResourcePoolID: func() string {
				if ca.req.Extra != nil {
					return ca.req.Extra.PoolID
				}
				return ""
			}(),
		},
		OnlyData: ca.req.OnlyCreateInfo,
	})
	if err != nil {
		blog.Errorf("create NodeGroup in cloudprovider %s/%s for Cluster %s failed, %s",
			ca.cloud.CloudID, ca.cloud.CloudProvider, ca.cluster.ClusterID, err.Error(),
		)
		return err
	}
	// create task and dispatch task, generate task entity need to create and dispatch task
	taskID := ""
	if task != nil {
		taskID = task.TaskID
		ca.resp.Data.Task = task
		if err = ca.model.CreateTask(ca.ctx, task); err != nil {
			blog.Errorf("save create node group task for cluster %s failed, %s",
				ca.group.ClusterID, err.Error(),
			)
			return err
		}
		if err = taskserver.GetTaskServer().Dispatch(task); err != nil {
			blog.Errorf("dispatch create node group task for cluster %s failed, %s",
				ca.group.ClusterID, err.Error(),
			)
			return err
		}
	} else {
		ca.group.Status = common.StatusRunning
		err = ca.model.UpdateNodeGroup(ca.ctx, ca.group)
		if err != nil {
			blog.Errorf("updateNodeGroup cluster[%s] %s failed, %s", ca.group.ClusterID, ca.group.Name, err.Error())
			return err
		}
	}

	// record operation log
	err = ca.model.CreateOperationLog(ca.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   ca.group.NodeGroupID,
		TaskID:       taskID,
		Message:      fmt.Sprintf("集群%s创建节点规格%s", ca.cluster.ClusterID, ca.group.NodeGroupID),
		OpUser:       ca.req.Creator,
		CreateTime:   time.Now().Format(time.RFC3339),
		ClusterID:    ca.cluster.ClusterID,
		ProjectID:    ca.cluster.ProjectID,
	})
	if err != nil {
		blog.Errorf("CreateNodeGroup[%s] CreateOperationLog failed: %v", ca.cluster.ClusterID, err)
	}
	return nil
}

// Handle create nodeGroup request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateNodeGroupRequest, resp *cmproto.CreateNodeGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create NodeGroup failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp
	ca.resp.Data = &cmproto.CreateNodeGroupResponseData{}

	// getRelativeResource get cluster / cloud provider
	if err := ca.getRelativeResource(); err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// save nodegroup to storage
	if err := ca.saveNodeGroup(); err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	// cloudprovider create nodegroup && dispatch task
	if err := ca.createNodeGroup(); err != nil {
		ca.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

func validateLaunchTemplate(launch *cmproto.LaunchConfiguration) error {
	if launch == nil {
		return fmt.Errorf("nodeGroup launchTemplate is null")
	}

	if launch.InstanceType == "" && (launch.CPU <= 0 && launch.Mem <= 0) && len(launch.Selector) == 0 {
		return fmt.Errorf("NodeGroup launchTemplate InstanceType selector is empty")
	}

	return nil
}

func validateDiskSize(disks ...*cmproto.DataDisk) error {
	for _, v := range disks {
		if v == nil {
			continue
		}

		size, _ := strconv.Atoi(v.DiskSize)
		if size < 50 || size > 32000 {
			return fmt.Errorf("disk size is invalid, it should >=50 and <=32000")
		}
	}
	return nil
}

func validateInternet(internet *cmproto.InternetAccessible) error {
	if internet == nil {
		return nil
	}
	bw, _ := strconv.Atoi(internet.InternetMaxBandwidth)
	if internet.PublicIPAssigned && bw <= 0 {
		return fmt.Errorf("InternetMaxBandwidth must be greater than 0 when PublicIPAssigned is enable")
	}
	return nil
}
