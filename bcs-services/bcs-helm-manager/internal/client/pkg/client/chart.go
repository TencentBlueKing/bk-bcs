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

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	urlChartList          = "/helmmanager/v1/chart/%s/%s"
	urlChartListV1        = "/helmmanager/v1/chart/%s/%s"
	urlChartDelete        = "/helmmanager/v1/projects/%s/repos/%s/charts/%s"
	urlChartVersionDelete = "/helmmanager/v1/projects/%s/repos/%s/charts/%s/versions/%s"
	urlChartVersionList   = "/helmmanager/v1/chart/%s/%s/%s/version"
	urlChartDetailGet     = "/helmmanager/v1/chart/%s/%s/%s/detail/%s"
	urlChartDetailV1Get   = "/helmmanager/v1/projects/%s/repos/%s/charts/%s"
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
func (ct *chart) Detail(ctx context.Context, req *helmmanager.GetVersionDetailReq) (*helmmanager.ChartDetail, error) {
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

	var r helmmanager.GetVersionDetailResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("get chart detail get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// Detail get chart detail v1
func (ct *chart) GetChartDetailV1(ctx context.Context, req *helmmanager.GetChartDetailV1Req) (*helmmanager.Chart, error) {
	projectID := req.GetProjectCode()
	if projectID == "" {
		return nil, fmt.Errorf("chart project can not be empty")
	}
	repo := req.GetRepoName()
	if repo == "" {
		return nil, fmt.Errorf("chart repository can not be empty")
	}
	chartName := req.GetName()
	if chartName == "" {
		return nil, fmt.Errorf("chart name can not be empty")
	}

	resp, err := ct.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartDetailV1Get, projectID, repo, chartName),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.GetChartDetailV1Resp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("get chart detail v1 get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// Detail get version detail v1
func (ct *chart) GetVersionDetailV1(ctx context.Context, req *helmmanager.GetVersionDetailV1Req) (*helmmanager.ChartDetail, error) {
	projectID := req.GetProjectCode()
	if projectID == "" {
		return nil, fmt.Errorf("chart project can not be empty")
	}
	repo := req.GetRepoName()
	if repo == "" {
		return nil, fmt.Errorf("chart repository can not be empty")
	}
	chartName := req.GetName()
	if chartName == "" {
		return nil, fmt.Errorf("chart name can not be empty")
	}

	resp, err := ct.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartDetailV1Get, projectID, repo, chartName),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.GetVersionDetailV1Resp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("get chart detail v1 get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// Detail delete chart
func (ct *chart) DeleteChart(ctx context.Context, req *helmmanager.DeleteChartReq) error {
	projectID := req.GetProjectCode()
	if projectID == "" {
		return fmt.Errorf("chart project can not be empty")
	}
	repo := req.GetRepoName()
	if repo == "" {
		return fmt.Errorf("chart repository can not be empty")
	}
	chartName := req.GetName()
	if chartName == "" {
		return fmt.Errorf("chart name can not be empty")
	}

	var data []byte
	_ = codec.EncJson(req, &data)

	resp, err := ct.delete(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartDelete, projectID, repo, chartName),
		nil,
		data,
	)
	if err != nil {
		return err
	}

	var r helmmanager.DeleteChartResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("delete chart get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}

// Detail delete chart version
func (ct *chart) DeleteChartVersion(ctx context.Context, req *helmmanager.DeleteChartVersionReq) error {
	projectID := req.GetProjectCode()
	if projectID == "" {
		return fmt.Errorf("chart project can not be empty")
	}
	repo := req.GetRepoName()
	if repo == "" {
		return fmt.Errorf("chart repository can not be empty")
	}
	chartName := req.GetName()
	if chartName == "" {
		return fmt.Errorf("chart name can not be empty")
	}
	version := req.GetVersion()
	if version == "" {
		return fmt.Errorf("chart version can not be empty")
	}
	var data []byte
	_ = codec.EncJson(req, &data)

	resp, err := ct.delete(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartVersionDelete, projectID, repo, chartName, version),
		nil,
		data,
	)
	if err != nil {
		return err
	}

	var r helmmanager.DeleteChartVersionResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("delete chart version get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}

// List chart
func (ct *chart) ListV1(ctx context.Context, req *helmmanager.ListChartV1Req) (*helmmanager.ChartListData, error) {
	if req == nil {
		return nil, fmt.Errorf("list chart v1 request is empty")
	}

	projectID := req.GetProjectCode()
	if projectID == "" {
		return nil, fmt.Errorf("chart project can not be empty")
	}
	repositoryName := req.GetRepoName()
	if repositoryName == "" {
		return nil, fmt.Errorf("chart repository can not be empty")
	}
	name := req.GetName()
	if name == "" {
		return nil, fmt.Errorf("chart name can not be empty")
	}

	resp, err := ct.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartListV1, projectID, repositoryName)+"?"+ct.listV1ChartQuery(req).Encode(),
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
		return nil, fmt.Errorf("list chart v1 get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

func (ct *chart) listV1ChartQuery(req *helmmanager.ListChartV1Req) url.Values {
	query := url.Values{}
	if req.Page != nil {
		query.Set("page", strconv.FormatInt(int64(req.GetPage()), 10))
	}
	if req.Size != nil {
		query.Set("size", strconv.FormatInt(int64(req.GetSize()), 10))
	}
	return query
}
