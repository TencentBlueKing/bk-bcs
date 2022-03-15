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

package client

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	urlChartList        = "/helmmanager/v1/chart/%s/%s"
	urlChartVersionList = "/helmmanager/v1/chart/%s/%s/%s/version"
	urlChartDetailGet   = "/helmmanager/v1/chart/%s/%s/%s/detail/%s"
)

// Chart return a pkg.ChartClient instance
func (c *Client) Chart() pkg.ChartClient {
	return &chart{Client: c}
}

type chart struct {
	*Client
}

// List chart
func (ct *chart) List(ctx context.Context, req *helmmanager.ListChartReq) (*helmmanager.ChartListData, error) {
	if req == nil {
		return nil, fmt.Errorf("list chart request is empty")
	}

	req.Operator = common.GetStringP(ct.conf.Operator)
	projectID := req.GetProjectID()
	if projectID == "" {
		return nil, fmt.Errorf("chart project can not be empty")
	}
	repo := req.GetRepository()
	if repo == "" {
		return nil, fmt.Errorf("chart repository can not be empty")
	}

	resp, err := ct.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartList, projectID, repo)+"?"+ct.listChartQuery(req).Encode(),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.ListChartResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("list chart get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

func (ct *chart) listChartQuery(req *helmmanager.ListChartReq) url.Values {
	query := url.Values{}
	if req.Page != nil {
		query.Set("page", strconv.FormatInt(int64(req.GetPage()), 10))
	}
	if req.Size != nil {
		query.Set("size", strconv.FormatInt(int64(req.GetSize()), 10))
	}
	if req.Operator != nil {
		query.Set("operator", req.GetOperator())
	}
	return query
}

// Versions list chart version
func (ct *chart) Versions(ctx context.Context, req *helmmanager.ListChartVersionReq) (
	*helmmanager.ChartVersionListData, error) {

	req.Operator = common.GetStringP(ct.conf.Operator)
	projectID := req.GetProjectID()
	if projectID == "" {
		return nil, fmt.Errorf("chart project can not be empty")
	}
	repo := req.GetRepository()
	if repo == "" {
		return nil, fmt.Errorf("chart repository can not be empty")
	}
	chartName := req.GetName()
	if chartName == "" {
		return nil, fmt.Errorf("chart name can not be empty")
	}

	resp, err := ct.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartVersionList, projectID, repo, chartName)+"?"+
			ct.listChartVersionQuery(req).Encode(),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.ListChartVersionResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("list chart version get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

func (ct *chart) listChartVersionQuery(req *helmmanager.ListChartVersionReq) url.Values {
	query := url.Values{}
	if req.Page != nil {
		query.Set("page", strconv.FormatInt(int64(req.GetPage()), 10))
	}
	if req.Size != nil {
		query.Set("size", strconv.FormatInt(int64(req.GetSize()), 10))
	}
	if req.Operator != nil {
		query.Set("operator", req.GetOperator())
	}
	return query
}

// Detail get chart version detail
func (ct *chart) Detail(ctx context.Context, req *helmmanager.GetChartDetailReq) (*helmmanager.ChartDetail, error) {
	req.Operator = common.GetStringP(ct.conf.Operator)
	projectID := req.GetProjectID()
	if projectID == "" {
		return nil, fmt.Errorf("chart project can not be empty")
	}
	repo := req.GetRepository()
	if repo == "" {
		return nil, fmt.Errorf("chart repository can not be empty")
	}
	chartName := req.GetName()
	if chartName == "" {
		return nil, fmt.Errorf("chart name can not be empty")
	}
	version := req.GetVersion()
	if version == "" {
		return nil, fmt.Errorf("version can not be empty")
	}

	resp, err := ct.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartDetailGet, projectID, repo, chartName, version)+"?operator="+req.GetOperator(),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.GetChartDetailResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("get chart detail get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}
