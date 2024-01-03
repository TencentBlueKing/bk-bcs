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

// Package iam NOTES
package iam

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// IAM related operate.
type IAM struct {
	// data service's iamSys api
	ds pbds.DataClient
	// iam client.
	iamSys *sys.Sys
	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
}

// NewIAM new iam.
func NewIAM(ds pbds.DataClient, iamSys *sys.Sys, disableAuth bool) (*IAM, error) {
	if ds == nil {
		return nil, errf.New(errf.InvalidParameter, "data client is nil")
	}

	if iamSys == nil {
		return nil, errf.New(errf.InvalidParameter, "iam sys is nil")
	}

	i := &IAM{
		ds:          ds,
		iamSys:      iamSys,
		disableAuth: disableAuth,
	}

	return i, nil
}
