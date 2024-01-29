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

// Package initial NOTES
package initial

import (
	"context"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbas "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/auth-server"
)

// Initial iam init related operate.
type Initial struct {
	// iam client.
	iamSys *sys.Sys
	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
}

// NewInitial new initial.
func NewInitial(iamSys *sys.Sys, disableAuth bool) (*Initial, error) {
	if iamSys == nil {
		return nil, errf.New(errf.InvalidParameter, "iam sys is nil")
	}

	i := &Initial{
		iamSys:      iamSys,
		disableAuth: disableAuth,
	}

	return i, nil
}

// InitAuthCenter init auth center's auth model.
func (i *Initial) InitAuthCenter(ctx context.Context, req *pbas.InitAuthCenterReq) (*pbas.InitAuthCenterResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// if auth is disabled, returns error if user wants to init auth center
	// if i.disableAuth {
	// 	err := errf.RPCAborted("authorize function is disabled, can not init auth center.")
	// 	logs.Errorf("authorize function is disabled, can not init auth center, rid: %s", kt.Rid)
	// 	return nil, err
	// }

	if err := req.Validate(); err != nil {
		logs.Errorf("request param validate failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if err := i.iamSys.Register(kt.Ctx, req.Host); err != nil {
		logs.Errorf("register to iam failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := new(pbas.InitAuthCenterResp)
	return resp, nil
}
