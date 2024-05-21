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

package handler

import (
	"context"
	"encoding/json"

	bcscommon "github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/middleware"
	"go-micro.dev/v4/metadata"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/common"
)

func getUserInfo(ctx context.Context) (*middleware.AuthUser, uint32, string) {
	md, _ := metadata.FromContext(ctx)
	data, ok := md.Get(string(middleware.AuthUserKey))
	user := &middleware.AuthUser{}
	_ = json.Unmarshal([]byte(data), user)
	if !ok || user.Username == "" {
		return nil, bcscommon.BcsErrCommHttpParametersFailed, bcscommon.BcsErrCommHttpParametersFailedStr
	}

	return user, common.CodeSuccess, common.MsgSuccess
}

// stringInSlice returns true if given string in slice
func stringInSlice(s string, l []string) bool {
	for _, objStr := range l {
		if s == objStr {
			return true
		}
	}
	return false
}

func getUint64Value(p *uint64) uint64 {
	if p == nil {
		return 0
	}
	return *p
}
