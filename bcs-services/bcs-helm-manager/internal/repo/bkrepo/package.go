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
	repositoryListHelmUri = "/repository/api/package/page"
)

func (rh *repositoryHandler) listChart(ctx context.Context, option repo.ListOption) (*repo.ListChartData, error) {
	if option.Size == 0 {
		option.Size = 10
	}

	switch rh.repoType {
	case repo.RepositoryTypeHelm:
		return rh.listHelmChart(ctx, option)
	case repo.RepositoryTypeOCI:
		return rh.listOCIChart(ctx, option)
	default:
		return nil, fmt.Errorf("unknown repo type %d", rh.repoType)
	}
}

func (rh *repositoryHandler) listOCIChart(ctx context.Context, option repo.ListOption) (*repo.ListChartData, error) {
	return rh.listHelmChart(ctx, option)
}

func (rh *repositoryHandler) listHelmChart(ctx context.Context, option repo.ListOption) (*repo.ListChartData, error) {
	resp, err := rh.handler.get(ctx, rh.getListHelmChartUri(option), nil, nil)
	if err != nil {
		blog.Errorf("list helm chart from bk-repo get failed, %s, with projectID %s, repoName %s",
			err.Error(), rh.projectID, rh.repository)
		return nil, err
	}

	var r listPackResp
	if err := codec.DecJson(resp.Reply, &r); err != nil {
		blog.Errorf("list helm chart from bk-repo decode failed, %s, with resp %s", err.Error(), resp.Reply)
		return nil, err
	}
	if r.Code != respCodeOK {
		blog.Errorf("list helm chart from bk-repo get resp with error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)
		return nil, err
	}

	var data []*repo.Chart
	for _, item := range r.Data.Records {
		data = append(data, item.convert2Chart())
	}
	return &repo.ListChartData{
		Total:  r.Data.TotalRecords,
		Page:   r.Data.PageNumber,
		Size:   r.Data.PageSize,
		Charts: data,
	}, nil
}

func (rh *repositoryHandler) getListHelmChartUri(option repo.ListOption) string {
	return repositoryListHelmUri + "/" + rh.projectID + "/" + rh.repository + "/" +
		"?pageNumber=" + strconv.FormatInt(option.Page, 10) +
		"&pageSize=" + strconv.FormatInt(option.Size, 10)
}

type listPackResp struct {
	basicResp
	Data *listPackData `json:"data"`
}

type listPackData struct {
	basicRecord
	Records []*pack `json:"records"`
}

type pack struct {
	ProjectID        string        `json:"projectId"`
	RepoName         string        `json:"repoName"`
	Name             string        `json:"name"`
	Key              string        `json:"key"`
	Type             string        `json:"type"`
	Latest           string        `json:"latest"`
	Downloads        int64         `json:"downloads"`
	Versions         int           `json:"versions"`
	Description      string        `json:"description"`
	VersionTag       interface{}   `json:"versionTag"`
	Extension        packExtension `json:"extension"`
	HistoryVersion   []string      `json:"historyVersion"`
	CreatedBy        string        `json:"createdBy"`
	CreatedDate      string        `json:"createdDate"`
	LastModifiedBy   string        `json:"lastModifiedBy"`
	LastModifiedDate string        `json:"lastModifiedDate"`
}

type packExtension struct {
	AppVersion string `json:"appVersion"`
}

// convert2Chart 将bk-repo HELM仓库中的package信息, 转换为chart信息
func (p *pack) convert2Chart() *repo.Chart {
	return &repo.Chart{
		Key:         p.Key,
		Name:        p.Name,
		Type:        p.Type,
		Version:     p.Latest,
		AppVersion:  p.Extension.AppVersion,
		Description: p.Description,
		CreateBy:    p.CreatedBy,
		UpdateBy:    p.LastModifiedBy,
		CreateTime:  p.CreatedDate,
		UpdateTime:  p.LastModifiedDate,
	}
}
