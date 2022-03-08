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

package handler

import (
	"context"

	actionChart "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/actions/chart"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// ListChart provide the actions to list charts
func (hm *HelmManager) ListChart(ctx context.Context,
	req *helmmanager.ListChartReq, resp *helmmanager.ListChartResp) error {

	defer recorder(ctx, "ListChart", req, resp)()
	action := actionChart.NewListChartAction(hm.model, hm.platform)
	return action.Handle(ctx, req, resp)
}

// ListChartVersion provide the actions to list chart versions
func (hm *HelmManager) ListChartVersion(ctx context.Context,
	req *helmmanager.ListChartVersionReq, resp *helmmanager.ListChartVersionResp) error {

	defer recorder(ctx, "ListChartVersion", req, resp)()
	action := actionChart.NewListChartVersionAction(hm.model, hm.platform)
	return action.Handle(ctx, req, resp)
}

// GetChartDetail provide the actions the get chart detail
func (hm *HelmManager) GetChartDetail(ctx context.Context,
	req *helmmanager.GetChartDetailReq, resp *helmmanager.GetChartDetailResp) error {

	defer recorder(ctx, "GetChartDetail", req, resp)()
	action := actionChart.NewGetChartDetailAction(hm.model, hm.platform)
	return action.Handle(ctx, req, resp)
}
