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

package autoscalingoption

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// CreateAction action for create namespace
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateAutoScalingOptionRequest
	resp  *cmproto.CreateAutoScalingOptionResponse

	//inner data for creation
	cluster *cmproto.Cluster
	project *cmproto.Project
	//cloud for implementation
	cloud *cmproto.Cloud
}

// NewCreateAction create namespace action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) generateClusterAutoScalingOption() *cmproto.ClusterAutoScalingOption {
	timeStr := time.Now().Format(time.RFC3339)
	return &cmproto.ClusterAutoScalingOption{
		ClusterID:                     ca.req.ClusterID,
		ProjectID:                     ca.cluster.ProjectID,
		IsScaleDwonEnable:             ca.req.IsScaleDwonEnable,
		Expander:                      ca.req.Expander,
		MaxEmptyBulkDelete:            ca.req.MaxEmptyBulkDelete,
		ScaleDownDelay:                ca.req.ScaleDownDelay,
		ScaleDownUnneededTime:         ca.req.ScaleDownUnneededTime,
		ScaleDownUtilizationThreahold: ca.req.ScaleDownUtilizationThreahold,
		SkipNodesWithLocalStorage:     ca.req.SkipNodesWithLocalStorage,
		SkipNodesWithSystemPods:       ca.req.SkipNodesWithSystemPods,
		IgnoreDaemonSetsUtilization:   ca.req.IgnoreDaemonSetsUtilization,
		OkTotalUnreadyCount:           ca.req.OkTotalUnreadyCount,
		MaxTotalUnreadyPercentage:     ca.req.MaxTotalUnreadyPercentage,
		ScaleDownUnreadyTime:          ca.req.ScaleDownUnreadyTime,
		UnregisteredNodeRemovalTime:   ca.req.UnregisteredNodeRemovalTime,
		Creator:                       ca.req.Creator,
		Provider:                      ca.req.Provider,
		CreateTime:                    timeStr,
		UpdateTime:                    timeStr,
	}
}

func (ca *CreateAction) createAutoScalingOption() error {
	option := ca.generateClusterAutoScalingOption()
	//check default value
	if len(option.Expander) == 0 {
		option.Expander = "random"
	}
	if option.MaxEmptyBulkDelete == 0 {
		option.MaxEmptyBulkDelete = 10
	}
	if option.ScaleDownDelay == 0 {
		option.ScaleDownDelay = 10
	}
	if option.ScaleDownUnneededTime == 0 {
		option.ScaleDownUnneededTime = 10
	}
	if option.ScaleDownUtilizationThreahold == 0 {
		option.ScaleDownUtilizationThreahold = 50
	}
	if option.OkTotalUnreadyCount == 0 {
		option.OkTotalUnreadyCount = 3
	}
	if option.MaxTotalUnreadyPercentage == 0 {
		option.MaxTotalUnreadyPercentage = 45
	}
	if option.ScaleDownUnreadyTime == 0 {
		option.ScaleDownUnreadyTime = 20
	}
	if option.UnregisteredNodeRemovalTime == 0 {
		option.UnregisteredNodeRemovalTime = 30
	}

	// create dependency system implementation according cloudprovider
	mgr, err := cloudprovider.GetNodeGroupMgr(ca.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s implementation for AutoScalingOption %s failed, %s",
			ca.cloud.CloudProvider, ca.cluster.ClusterID, err.Error(),
		)
		return err
	}
	if err = ca.model.CreateAutoScalingOption(ca.ctx, option); err != nil {
		blog.Errorf("AutoScalingOption %s store to DB failed, %s", option.ClusterID, err.Error())
		return err
	}
	coption, err := cloudprovider.GetCredential(ca.project, ca.cloud)
	if err != nil {
		blog.Errorf("get credential for AutoScalingOption %s failed, %s",
			option.ClusterID, err.Error(),
		)
		return err
	}

	// generate task to call helm interface for creating autoscaler
	task, err := mgr.CreateAutoScalingOption(option, &cloudprovider.CreateScalingOption{
		CommonOption: *coption,
	})
	if err != nil {
		blog.Errorf("create AutoScalingOption %s in Cloud %s failed, %s",
			option.ClusterID, ca.cloud.CloudProvider, err.Error(),
		)
		return err
	}
	blog.Infof("create AutoScalingOption %s with Configuration %s in Cloud %s successfully, %+v",
		option.ClusterID, ca.cloud.CloudID, ca.cloud.CloudProvider, task,
	)
	return nil
}

func (ca *CreateAction) validate() error {
	var err error
	if err = ca.req.Validate(); err != nil {
		return err
	}

	// validate cluster information & project information
	ca.cluster, err = ca.model.GetCluster(ca.ctx, ca.req.ClusterID)
	if err != nil {
		blog.Errorf("Get ClusterAutoScalingOption relative cluster %s failed, %s", ca.req.ClusterID, err.Error())
		return fmt.Errorf("get relative cluster failed, %s", err.Error())
	}

	// get cloud information for cloudprovider
	ca.cloud, err = ca.model.GetCloud(ca.ctx, ca.req.Provider)
	if err != nil {
		blog.Errorf("Get %s ClusterAutoScalingOption relative Cloud %s failed, %s",
			ca.req.ClusterID, ca.req.Provider, err.Error(),
		)
		return fmt.Errorf("get relative cloud err, %s", err.Error())
	}

	// get cloud information for cloudprovider
	ca.project, err = ca.model.GetProject(ca.ctx, ca.cluster.ProjectID)
	if err != nil {
		blog.Errorf("Get %s ClusterAutoScalingOption relative Project %s failed, %s",
			ca.req.ClusterID, ca.cluster.ProjectID, err.Error(),
		)
		return fmt.Errorf("get relative project err, %s", err.Error())
	}
	return nil
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create namespace request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateAutoScalingOptionRequest, resp *cmproto.CreateAutoScalingOptionResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create ClusterAutoScalingOption failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.createAutoScalingOption(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
