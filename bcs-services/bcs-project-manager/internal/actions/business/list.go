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
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
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

	// list all business that enable bcs
	if req.UseBCS {
		return ga.listBusinessEnabledBCS()
	}
	// list business that user has business maintainer permission
	authUser, err := middleware.GetUserFromContext(ctx)
	if err != nil || authUser.Username == "" {
		return nil, errorx.NewReadableErr(errorx.ParamErr, "username is empty")
	}
	searchData, err := cmdb.SearchBusiness(ctx, authUser.Username, "")
	if err != nil {
		return nil, errorx.NewRequestCMDBErr(err.Error())
	}

	retDatas := []*proto.BusinessData{}

	for _, business := range searchData.Info {
		retData := &proto.BusinessData{
			BusinessID: strconv.Itoa(int(business.BKBizID)),
			Name:       business.BKBizName,
			Maintainer: stringx.SplitString(business.BKBizMaintainer),
		}
		retDatas = append(retDatas, retData)
	}
	return retDatas, nil
}

func (ga *ListAction) listBusinessEnabledBCS() ([]*proto.BusinessData, error) {
	searchData, err := cmdb.SearchBusiness(ga.ctx, "", "")
	if err != nil {
		return nil, errorx.NewRequestCMDBErr(err.Error())
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"kind": "k8s",
	})

	if tenant.IsMultiTenantEnabled() {
		tenantCond := operator.NewLeafCondition(operator.Eq, operator.M{"tenantID": tenant.GetTenantIdFromContext(ga.ctx)})
		cond = operator.NewBranchCondition(operator.And, cond, tenantCond)
	}

	// 查询所有开启了容器服务的项目
	projects, _, err := ga.model.ListProjects(ga.ctx, cond, &page.Pagination{All: true})
	if err != nil {
		return nil, err
	}
	businessUsed := map[string]bool{}
	for _, project := range projects {
		if project.BusinessID != "0" && project.BusinessID != "" {
			businessUsed[project.BusinessID] = true
		}
	}

	retDatas := []*proto.BusinessData{}

	for _, business := range searchData.Info {
		if businessUsed[strconv.Itoa(int(business.BKBizID))] {
			retData := &proto.BusinessData{
				BusinessID: strconv.Itoa(int(business.BKBizID)),
				Name:       business.BKBizName,
				Maintainer: stringx.SplitString(business.BKBizMaintainer),
			}
			retDatas = append(retDatas, retData)
		}
	}
	return retDatas, nil
}
