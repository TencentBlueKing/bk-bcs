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

package business

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// GetTopologyAction action for get business topology
type GetTopologyAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.GetBusinessTopologyRequest
	resp  *proto.GetBusinessTopologyResponse
}

// NewGetTopologyAction new get business topology action
func NewGetTopologyAction(model store.ProjectModel) *GetTopologyAction {
	return &GetTopologyAction{
		model: model,
	}
}

// Do get business topology
func (ga *GetTopologyAction) Do(ctx context.Context,
	req *proto.GetBusinessTopologyRequest, resp *proto.GetBusinessTopologyResponse) error {
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	p, err := ga.model.GetProject(ctx, req.GetProjectCode())
	if err != nil {
		return errorx.NewDBErr(err.Error())
	}

	if p.BusinessID == "" || p.BusinessID == "0" {
		return errorx.NewReadableErr(errorx.ParamErr, "project businessID is empty")
	}

	topologyDatas, err := cmdb.GetBusinessTopology(ctx, p.BusinessID)
	if err != nil {
		return errorx.NewRequestCMDBErr(err.Error())
	}

	retDatas := []*proto.TopologyData{}
	for _, topologyData := range topologyDatas {
		retData := topologyData.TransferToProto()
		retDatas = append(retDatas, retData)
	}
	resp.Data = retDatas
	return nil
}
