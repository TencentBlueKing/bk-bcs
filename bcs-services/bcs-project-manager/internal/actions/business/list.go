/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// ListAction action for get business
type ListAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.ListBusinessRequest
}

// NewListAction new get business action
func NewListAction(model store.ProjectModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

// Do get business info
func (ga *ListAction) Do(ctx context.Context, req *proto.ListBusinessRequest) ([]*proto.BusinessData, error) {
	ga.ctx = ctx
	ga.req = req

	authUser, err := middleware.GetUserFromContext(ctx)
	if err != nil || authUser.Username == "" {
		return nil, errorx.NewReadableErr(errorx.ParamErr, "username is empty")
	}
	searchData, err := cmdb.SearchBusiness(authUser.Username, "")
	if err != nil {
		return nil, errorx.NewRequestCMDBErr(err.Error())
	}

	retDatas := []*proto.BusinessData{}

	for _, business := range searchData.Info {
		retData := &proto.BusinessData{
			BusinessID: uint32(business.BKBizID),
			Name:       business.BKBizName,
			Maintainer: stringx.SplitString(business.BKBizMaintainer),
		}
		retDatas = append(retDatas, retData)
	}

	return retDatas, nil
}
