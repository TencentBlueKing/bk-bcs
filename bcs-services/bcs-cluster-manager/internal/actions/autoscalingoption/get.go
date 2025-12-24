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

package autoscalingoption

import (
	"context"
	"errors"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// GetAction action for getting cluster credential
type GetAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.GetAutoScalingOptionRequest
	resp  *cmproto.GetAutoScalingOptionResponse
}

// NewGetAction create get action for online cluster credential
func NewGetAction(model store.ClusterManagerModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

func (ga *GetAction) getOption() error {
	option, err := ga.model.GetAutoScalingOption(ga.ctx, ga.req.ClusterID)
	if err == nil {
		// default parameter ExpendablePodsPriorityCutoff=-10
		if option.ExpendablePodsPriorityCutoff == 0 {
			option.ExpendablePodsPriorityCutoff = -10
		}
		ga.resp.Data = option
		return nil
	}
	// if not found, create a default option
	if errors.Is(err, drivers.ErrTableRecordNotFound) {
		return ga.setDefaultOption()
	}
	return err
}

func (ga *GetAction) setDefaultOption() error {
	cluster, err := ga.model.GetCluster(ga.ctx, ga.req.ClusterID)
	if err != nil {
		return err
	}

	provider := cluster.Provider
	if ga.req.Provider != "" {
		provider = ga.req.Provider
	}

	option := &cmproto.ClusterAutoScalingOption{
		ProjectID:  cluster.ProjectID,
		ClusterID:  cluster.ClusterID,
		Creator:    "admin",
		Updater:    "admin",
		CreateTime: time.Now().UTC().Format(time.RFC3339),
		UpdateTime: time.Now().UTC().Format(time.RFC3339),
		Provider:   provider,

		EnableAutoscale:               false,
		IsScaleDownEnable:             true,
		Expander:                      "random",
		MaxEmptyBulkDelete:            10,
		ScaleDownDelay:                600,
		ScaleDownUnneededTime:         600,
		ScaleDownUtilizationThreahold: 50,

		SkipNodesWithSystemPods:   true,
		SkipNodesWithLocalStorage: true,

		IgnoreDaemonSetsUtilization: true,

		OkTotalUnreadyCount:       3,
		MaxTotalUnreadyPercentage: 45,

		ScaleDownUnreadyTime:             1200,
		ScaleDownGpuUtilizationThreshold: 50,

		BufferResourceRatio:       100,
		MaxGracefulTerminationSec: 600,
		ScanInterval:              10,

		// max-node-startup-time / max-node-start-schedule-time
		MaxNodeProvisionTime: 900,

		ScaleUpFromZero:            true,
		ScaleDownDelayAfterAdd:     1200,
		ScaleDownDelayAfterDelete:  0,
		ScaleDownDelayAfterFailure: 180,
		Status:                     common.StatusAutoScalingOptionStopped,
		BufferResourceCpuRatio:     100,
		BufferResourceMemRatio:     100,
		Module:                     &cmproto.ModuleInfo{},
		Webhook:                    &cmproto.WebhookMode{},

		ExpendablePodsPriorityCutoff: -2147483648,
		NewPodScaleUpDelay:           0,
	}
	err = ga.model.CreateAutoScalingOption(ga.ctx, option)
	if err != nil {
		return err
	}
	ga.resp.Data = option
	return nil
}

func (ga *GetAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle get cluster credential
func (ga *GetAction) Handle(
	ctx context.Context, req *cmproto.GetAutoScalingOptionRequest, resp *cmproto.GetAutoScalingOptionResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get ClusterAutoScalingOption failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := req.Validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ga.getOption(); err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
