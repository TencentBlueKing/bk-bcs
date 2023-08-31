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

package health

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// HealthAction action for service health
type HealthAction struct {
	ctx  context.Context
	req  *cmproto.HealthRequest
	resp *cmproto.HealthResponse
}

// NewHealthAction create health action for service health check
func NewHealthAction() *HealthAction {
	return &HealthAction{}
}

func (ha *HealthAction) validate() error {
	if err := ha.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (ha *HealthAction) setResp(code uint32, msg string) {
	ha.resp.Code = code
	ha.resp.Message = msg
	ha.resp.Available = "false"
	if code == common.BcsErrClusterManagerSuccess {
		ha.resp.Available = "true"
	}
}

// Handle handle health check
func (ha *HealthAction) Handle(
	ctx context.Context, req *cmproto.HealthRequest, resp *cmproto.HealthResponse) {
	if req == nil || resp == nil {
		blog.Errorf("health check failed, req or resp is empty")
		return
	}
	ha.ctx = ctx
	ha.req = req
	ha.resp = resp

	if err := ha.validate(); err != nil {
		ha.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	ha.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
