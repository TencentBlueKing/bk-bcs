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

package resourceschema

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ListAction action for list resource schema
type ListAction struct {
	ctx context.Context

	schemaPath string
	req        *cmproto.ListResourceSchemaRequest
	resp       *cmproto.CommonListResp
}

// NewListAction create list action for resource schema
func NewListAction(schemaPath string) *ListAction {
	return &ListAction{schemaPath: schemaPath}
}

func (ga *ListAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ga *ListAction) getSchema() error {
	schemaList := []*types.ResourceSchema{}
	schemaBytes, err := ioutil.ReadFile(ga.schemaPath)
	if err != nil {
		blog.Errorf("load resource schema from file %s err: %v", ga.schemaPath, err)
		return fmt.Errorf("load resource schema from file %s err: %v", ga.schemaPath, err)
	}

	err = json.Unmarshal(schemaBytes, &schemaList)
	if err != nil {
		blog.Errorf("load resource schema from file %s Unmarshal err: %v", ga.schemaPath, err)
		return fmt.Errorf("load resource schema from file %s Unmarshal err: %v", ga.schemaPath, err)
	}

	var result []types.ResourceSchema
	for _, v := range schemaList {
		if v.CloudID == ga.req.CloudID {
			result = append(result, *v)
		}
	}

	s, err := utils.MarshalInterfaceToListValue(result)
	if err != nil {
		return fmt.Errorf("marshal schema failed, err %s", err.Error())
	}

	ga.resp.Data = s
	return nil
}

// Handle handle list resource schema
func (ga *ListAction) Handle(
	ctx context.Context, req *cmproto.ListResourceSchemaRequest, resp *cmproto.CommonListResp) {
	if req == nil || resp == nil {
		blog.Errorf("list resource schema failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := req.Validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ga.getSchema(); err != nil {
		ga.setResp(common.BcsErrClusterManagerCommonErr, err.Error())
		return
	}

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
