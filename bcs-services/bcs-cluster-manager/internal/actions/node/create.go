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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// RecordNodeDataAction action for record node
type RecordNodeDataAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.RecordNodeInfoRequest
	resp  *cmproto.CommonResp
}

// NewRecordNodeDataAction create node action
func NewRecordNodeDataAction(model store.ClusterManagerModel) *RecordNodeDataAction {
	return &RecordNodeDataAction{
		model: model,
	}
}

func (ua *RecordNodeDataAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	if len(ua.req.Nodes) == 0 {
		return fmt.Errorf("RecordNodeDataAction validate failed: body empty")
	}

	return nil
}

func (ua *RecordNodeDataAction) recordNodes() error { // nolint
	clusterIDMap := make(map[string]struct{})
	for _, node := range ua.req.Nodes {
		err := ua.model.CreateNode(context.Background(), node)
		if err != nil {
			blog.Errorf("RecordNodeDataAction recordNodes %s failed: %v", node.InnerIP, err)
		}
		clusterIDMap[node.ClusterID] = struct{}{}
	}

	for clusterID := range clusterIDMap {
		cluster, err := ua.model.GetCluster(ua.ctx, clusterID)
		if err != nil {
			continue
		}
		err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
			ResourceType: common.Cluster.String(),
			ResourceID:   clusterID,
			TaskID:       "",
			Message:      "录入节点详情信息",
			OpUser:       auth.GetUserFromCtx(ua.ctx),
			CreateTime:   time.Now().UTC().Format(time.RFC3339),
			ClusterID:    cluster.ClusterID,
			ProjectID:    cluster.ProjectID,
			ResourceName: cluster.ClusterName,
		})
		if err != nil {
			blog.Errorf("RecordNodeInfo CreateOperationLog failed: %v", err)
		}
	}

	return nil
}

func (ua *RecordNodeDataAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handles create nodes request
func (ua *RecordNodeDataAction) Handle(ctx context.Context, req *cmproto.RecordNodeInfoRequest,
	resp *cmproto.CommonResp) {
	if req == nil || resp == nil {
		blog.Errorf("record nodes failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ua.recordNodes(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	blog.Infof("RecordNodeDataAction record[%+v] success", req.Nodes)
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
