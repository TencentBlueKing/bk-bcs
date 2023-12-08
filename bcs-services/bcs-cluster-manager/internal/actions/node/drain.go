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

package node

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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
	// set default GracePeriodSeconds
	if ua.req.GracePeriodSeconds == 0 {
		ua.req.GracePeriodSeconds = -1
	}

	return nil
}

func (ua *DrainNodeAction) drainClusterNodes() error {
	// new drainer
	drainer := clusterops.DrainHelper{
		Force:                           true,
		GracePeriodSeconds:              int(ua.req.GracePeriodSeconds),
		IgnoreAllDaemonSets:             true,
		Timeout:                         int(ua.req.Timeout),
		DeleteLocalData:                 true,
		Selector:                        ua.req.Selector,
		PodSelector:                     ua.req.PodSelector,
		DisableEviction:                 ua.req.DisableEviction,
		DryRun:                          ua.req.DryRun,
		SkipWaitForDeleteTimeoutSeconds: int(ua.req.SkipWaitForDeleteTimeoutSeconds),
	}

	// get node names
	if len(ua.req.Nodes) == 0 && len(ua.req.InnerIPs) > 0 {
		option := clusterops.ListNodeOption{ClusterID: ua.req.ClusterID, NodeIPs: ua.req.InnerIPs}
		nodes, err := ua.k8sOp.ListClusterNodesByIPsOrNames(ua.ctx, option)
		if err != nil {
			blog.Errorf("get nodename by ips failed in cluster %s, err %s", ua.req.ClusterID, err.Error())
			return fmt.Errorf("get nodename by ips failed in cluster %s, err %s", ua.req.ClusterID, err.Error())
		}
		for _, v := range nodes {
			ua.req.Nodes = append(ua.req.Nodes, v.Name)
		}
	}

	successCh := make(chan *cmproto.NodeOperationStatusInfo, len(ua.req.Nodes))
	failCh := make(chan *cmproto.NodeOperationStatusInfo, len(ua.req.Nodes))

	barrier := utils.NewRoutinePool(50)
	defer barrier.Close()

	for i := range ua.req.Nodes {
		barrier.Add(1)
		go func(node string) {
			defer barrier.Done()
			ctx, cancel := context.WithTimeout(context.Background(), clusterops.DefaultTimeout)
			defer cancel()
			if err := ua.k8sOp.ClusterUpdateScheduleNode(ctx, clusterops.NodeInfo{
				ClusterID: ua.req.ClusterID,
				NodeName:  node,
				Desired:   true,
			}); err != nil {
				failCh <- &cmproto.NodeOperationStatusInfo{NodeName: node, Message: err.Error()}
				blog.Errorf("drainClusterNodes[%s] failed in cluster %s, err %s", node, ua.req.ClusterID, err.Error())
				return
			}
			if err := ua.k8sOp.DrainNode(ua.ctx, ua.req.ClusterID, node, drainer); err != nil {
				failCh <- &cmproto.NodeOperationStatusInfo{NodeName: node, Message: err.Error()}
				blog.Errorf("drainClusterNodes[%s] failed in cluster %s, err %s", node, ua.req.ClusterID, err.Error())
				return
			}
			successCh <- &cmproto.NodeOperationStatusInfo{NodeName: node}
		}(ua.req.Nodes[i])
	}
	barrier.Wait()
	close(successCh)
	close(failCh)

	ua.resp.Data = &cmproto.NodeOperationStatus{
		Success: make([]*cmproto.NodeOperationStatusInfo, 0),
		Fail:    make([]*cmproto.NodeOperationStatusInfo, 0),
	}
	for v := range successCh {
		ua.resp.Data.Success = append(ua.resp.Data.Success, v)
	}
	for v := range failCh {
		ua.resp.Data.Fail = append(ua.resp.Data.Fail, v)
	}

	return nil
}

func (ua *DrainNodeAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
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
}
