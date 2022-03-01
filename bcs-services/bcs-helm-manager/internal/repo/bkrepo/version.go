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

package bkrepo

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
)

const (
	chartVersionListHelmUri     = "/repository/api/version/page"
	chartVersionDownloadHelmUri = "/repository/api/version/download"
)

func (ch *chartHandler) listChartVersion(ctx context.Context, option repo.ListOption) (
	*repo.ListChartVersionData, error) {

	if option.Size == 0 {
		option.Size = 10
	}

	switch ch.repoType {
	case repo.RepositoryTypeHelm:
		return ch.listHelmChartVersion(ctx, option)
	default:
		return nil, fmt.Errorf("unknown repo type %d", ch.repoType)
	}
}

func (ch *chartHandler) listHelmChartVersion(ctx context.Context, option repo.ListOption) (
	*repo.ListChartVersionData, error) {

	resp, err := ch.handler.get(ctx, ch.getListHelmChartVersionUri(option), nil, nil)
	if err != nil {
		blog.Errorf(
			"list helm chart version from bk-repo get failed, %s, with projectID %s, repoName %s, chartName %s",
			err.Error(), ch.projectID, ch.repository, ch.chartName)
		return nil, err
	}

	var r listPackVersionResp
	if err := codec.DecJson(resp.Reply, &r); err != nil {
		blog.Errorf(
			"list helm chart version from bk-repo decode failed, %s, with resp %s", err.Error(), resp.Reply)
		return nil, err
	}
	if r.Code != respCodeOK {
		blog.Errorf("list helm chart version from bk-repo get resp with error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)
		return nil, err
	}

	var data []*repo.ChartVersion
	for _, item := range r.Data.Records {
		chart := item.convert2Chart()
		chart.Name = ch.chartName
		data = append(data, chart)
	}
	return &repo.ListChartVersionData{
		Total:    r.Data.TotalRecords,
		Page:     r.Data.PageNumber,
		Size:     r.Data.PageSize,
		Versions: data,
	}, nil
}

func (ch *chartHandler) getListHelmChartVersionUri(option repo.ListOption) string {
	return chartVersionListHelmUri + "/" + ch.projectID + "/" + ch.repository + "/" +
		"?packageKey=helm://" + ch.chartName +
		"&pageNumber=" + strconv.FormatInt(option.Page, 10) +
		"&pageSize=" + strconv.FormatInt(option.Size, 10)
}

type listPackVersionResp struct {
	basicResp
	Data *listPackVersionData `json:"data"`
}

type listPackVersionData struct {
	basicRecord
	Records []*packVersion `json:"records"`
}

type packVersion struct {
	Name             string              `json:"name"`
	Size             int64               `json:"size"`
	Downloads        int64               `json:"downloads"`
	Metadata         packVersionMetaData `json:"metadata"`
	CreatedBy        string              `json:"createdBy"`
	CreatedDate      string              `json:"createdDate"`
	LastModifiedBy   string              `json:"lastModifiedBy"`
	LastModifiedDate string              `json:"lastModifiedDate"`
}

type packVersionMetaData struct {
	AppVersion  string `json:"appVersion"`
	Description string `json:"description"`
}

// convert2Chart 将bk-repo HELM仓库中的package version信息, 转换为chart version信息
func (pv *packVersion) convert2Chart() *repo.ChartVersion {
	return &repo.ChartVersion{
		Version:     pv.Name,
		AppVersion:  pv.Metadata.AppVersion,
		Description: pv.Metadata.Description,
		CreateBy:    pv.CreatedBy,
		UpdateBy:    pv.LastModifiedBy,
		CreateTime:  pv.CreatedDate,
		UpdateTime:  pv.LastModifiedDate,
	}
}

func (ch *chartHandler) getChartVersionDetail(ctx context.Context, version string) (*repo.ChartDetail, error) {
	switch ch.repoType {
	case repo.RepositoryTypeHelm:
		return ch.getHelmChartVersionDetail(ctx, version)
	default:
		return nil, fmt.Errorf("unknown repo type %d", ch.repoType)
	}
}

func (ch *chartHandler) downloadChartVersion(ctx context.Context, version string) ([]byte, error) {
	switch ch.repoType {
	case repo.RepositoryTypeHelm:
		return ch.downloadHelmChartVersionOrigin(ctx, version)
	default:
		return nil, fmt.Errorf("unknown repo type %d", ch.repoType)
	}
}

func (ch *chartHandler) downloadHelmChartVersionOrigin(ctx context.Context, version string) ([]byte, error) {
	resp, err := ch.get(ctx, ch.getDownloadHelmChartVersionUri(version), nil, nil)
	if err != nil {
		blog.Errorf("download helm chart version origin from bk-repo get failed, %s, "+
			"with projectID %s, repoName %s, chartName %s, version %s",
			err.Error(), ch.projectID, ch.repository, ch.chartName, version)
		return nil, err
	}
	var r downloadChartVersionResp
	if err := codec.DecJson(resp.Reply, &r); err == nil && r.Code != respCodeOK {
		blog.Errorf("download helm chart version origin from bk-repo get resp with "+
			"error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)
		return nil, err
	}

	return resp.Reply, nil
}

func (ch *chartHandler) getHelmChartVersionDetail(ctx context.Context, version string) (*repo.ChartDetail, error) {
	contents, err := ch.downloadHelmChartVersionOrigin(ctx, version)
	if err != nil {
		blog.Errorf("get helm chart version detail get origin contents failed, %s, "+
			"with projectID %s, repoName %s, chartName %s, version %s",
			err.Error(), ch.projectID, ch.repository, ch.chartName, version)
		return nil, err
	}

	detail := &repo.ChartDetail{
		Name:    ch.chartName,
		Version: version,
	}
	if err = detail.LoadContentFromTgz(contents); err != nil {
		blog.Errorf("get helm chart version detail from bk-repo load from gzip file failed, %s, "+
			"with projectID %s, repoName %s, chartName %s, version %s",
			err.Error(), ch.projectID, ch.repository, ch.chartName, version)
		return nil, err
	}

	return detail, nil
}

func (ch *chartHandler) getDownloadHelmChartVersionUri(version string) string {
	return chartVersionDownloadHelmUri + "/" + ch.projectID + "/" + ch.repository + "/" +
		"?packageKey=helm://" + ch.chartName +
		"&version=" + version
}

type downloadChartVersionResp struct {
	basicResp
}
