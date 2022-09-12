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
 *
 */

package node

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// DrainNodeAction action for drain node
type DrainNodeAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.DrainNodeRequest
	resp  *cmproto.DrainNodeResponse
	k8sOp *clusterops.K8SOperator

	failed []string
}

// NewDrainNodeAction create update action
func NewDrainNodeAction(model store.ClusterManagerModel, k8sOp *clusterops.K8SOperator) *DrainNodeAction {
	return &DrainNodeAction{
		model:  model,
		k8sOp:  k8sOp,
		failed: make([]string, 0),
	}
}

func (ua *DrainNodeAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (ua *DrainNodeAction) drainClusterNodes() error {
	drainer := clusterops.DrainHelper{
		Force:                           ua.req.Force,
		GracePeriodSeconds:              int(ua.req.GracePeriodSeconds),
		IgnoreAllDaemonSets:             ua.req.IgnoreAllDaemonSets,
		Timeout:                         int(ua.req.Timeout),
		DeleteLocalData:                 ua.req.DeleteLocalData,
		Selector:                        ua.req.Selector,
		PodSelector:                     ua.req.PodSelector,
		DisableEviction:                 ua.req.DisableEviction,
		DryRun:                          ua.req.DryRun,
		SkipWaitForDeleteTimeoutSeconds: int(ua.req.SkipWaitForDeleteTimeoutSeconds),
	}
	for _, ip := range ua.req.InnerIPs {
		err := ua.k8sOp.ClusterUpdateScheduleNode(ua.ctx, clusterops.NodeInfo{
			ClusterID: ua.req.ClusterID,
			NodeIP:    ip,
			Desired:   true,
		})
		if err != nil {
			blog.Errorf("drainClusterNodes[%s] failed: %+v", ip, err)
			ua.failed = append(ua.failed, ip)
			continue
		}
		err = ua.k8sOp.DrainNode(ua.ctx, ua.req.ClusterID, ip, drainer)
		if err != nil {
			blog.Errorf("drainClusterNodes[%s] failed: %+v", ip, err)
			ua.failed = append(ua.failed, ip)
			continue
		}
	}

	return nil
}

func (ua *DrainNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ua.resp.Fail = ua.failed
}

// Handle handles node drain
func (ua *DrainNodeAction) Handle(ctx context.Context, req *cmproto.DrainNodeRequest, resp *cmproto.DrainNodeResponse) {
	if req == nil || resp == nil {
		blog.Errorf("drain cluster node failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.drainClusterNodes(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
