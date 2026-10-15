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

// Package business xxx
package business

import (
	"context"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// GetAction action for get business
type GetAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.GetBusinessRequest
}

// NewGetAction new get business action
func NewGetAction(model store.ProjectModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

// Do get business info
func (ga *GetAction) Do(ctx context.Context, req *proto.GetBusinessRequest) (*proto.BusinessData, error) {
	ga.ctx = ctx
	ga.req = req

	p, err := ga.model.GetProject(ctx, req.GetProjectCode())
	if err != nil {
		return nil, errorx.NewDBErr(err.Error())
	}

	if p.BusinessID == "" || p.BusinessID == "0" {
		return nil, errorx.NewReadableErr(errorx.ParamErr, "project businessID is empty")
	}

	businessInfo, err := cmdb.GetBusinessByID(ctx, p.BusinessID, false)
	if err != nil {
		return nil, errorx.NewRequestCMDBErr(err.Error())
	}

	retData := &proto.BusinessData{
		BusinessID: strconv.Itoa(int(businessInfo.BKBizID)),
		Name:       businessInfo.BKBizName,
		Maintainer: stringx.SplitString(businessInfo.BKBizMaintainer),
	}

	return retData, nil
}
