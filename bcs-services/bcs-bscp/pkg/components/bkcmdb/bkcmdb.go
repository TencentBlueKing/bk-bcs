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

// Package bkcmdb provides bkcmdb client.
package bkcmdb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/components"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/config"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/cmdb"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/types"
)

// SearchBusiness 组件化的函数
func SearchBusiness(ctx context.Context, params *cmdb.SearchBizParams) (*cmdb.SearchBizResp, error) {
	url := fmt.Sprintf("%s/api/c/compapi/v2/cc/search_business/", config.G.Base.BKPaaSHost)

	// SearchBizParams is esb search cmdb business parameter.
	type esbSearchBizParams struct {
		*types.CommParams
		*cmdb.SearchBizParams
	}

	req := &esbSearchBizParams{
		CommParams: &types.CommParams{
			AppCode:   config.G.Base.AppCode,
			AppSecret: config.G.Base.AppSecret,
			UserName:  "admin",
		},
		SearchBizParams: params,
	}

	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetBody(req).
		Post(url)

	if err != nil {
		return nil, err
	}

	bizList := &cmdb.SearchBizResp{}
	if err := json.Unmarshal(resp.Body(), bizList); err != nil {
		return nil, err
	}
	return bizList, nil

}
