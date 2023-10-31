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

// Package thirdparty xxx
package thirdparty

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ListCCTopologyAction action for list cc topology
type ListCCTopologyAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListCCTopologyRequest
	resp  *cmproto.CommonResp
}

// NewListCCTopologyAction create action
func NewListCCTopologyAction(model store.ClusterManagerModel) *ListCCTopologyAction {
	return &ListCCTopologyAction{
		model: model,
	}
}

func (la *ListCCTopologyAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (la *ListCCTopologyAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (la *ListCCTopologyAction) filterInter() bool {
	if la.req.FilterInter != nil {
		return la.req.FilterInter.GetValue()
	}

	return false
}

func (la *ListCCTopologyAction) listTopology() error {
	cluster, err := la.model.GetCluster(la.ctx, la.req.ClusterID)
	if err != nil {
		blog.Errorf("GetBizInternalModule get cluster failed, err: %s", err.Error())
		return fmt.Errorf("get cluster failed, err: %s", err.Error())
	}
	bizID := cluster.BusinessID
	if len(la.req.BizID) > 0 {
		bizID = la.req.BizID
	}
	bkBizID, err := strconv.Atoi(bizID)
	if err != nil {
		blog.Errorf("GetBizInternalModule get cluster bkBizID failed, err: %s", err.Error())
		return fmt.Errorf("get cluster bkBizID failed, err: %s", err.Error())
	}
	cli := cmdb.GetCmdbClient()
	internalModules, err := cli.ListTopology(la.ctx, bkBizID, la.filterInter(), false)
	if err != nil {
		blog.Errorf("GetBizInternalModule failed, err %s", err.Error())
		return err
	}

	result, err := utils.MarshalInterfaceToValue(internalModules)
	if err != nil {
		blog.Errorf("marshal modules err, %s", err.Error())
		la.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return err
	}
	la.resp.Data = result
	return nil
}

// Handle handles list cc topology
func (la *ListCCTopologyAction) Handle(ctx context.Context, req *cmproto.ListCCTopologyRequest,
	resp *cmproto.CommonResp) {
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.listTopology(); err != nil {
		la.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	blog.Infof("list cc topology successfully")
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
