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
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
)

const (
	repositoryListHelmURI      = "/repository/api/package/page"
	repositorySearchHelmURI    = "/repository/api/package/search"
	repositoryGetHelmURI       = "/repository/api/package/info"
	repositoryDeletePackageURI = "/repository/api/package/delete"
	repositoryDeleteVersionURI = "/repository/api/version/delete"
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
	resp, err := rh.handler.get(ctx, rh.getListHelmChartURI(option), nil, nil)
	if err != nil {
		blog.Errorf("list helm chart from bk-repo get failed, %s, with projectID %s, repoName %s",
			err.Error(), rh.projectID, rh.repository)
		return nil, err
	}

	var r listPackResp
	if err = codec.DecJson(resp.Reply, &r); err != nil {
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

func (rh *repositoryHandler) getListHelmChartURI(option repo.ListOption) string {
	return fmt.Sprintf("%s/%s/%s/?pageNumber=%s&pageSize=%s&packageName=%s", repositoryListHelmURI,
		rh.projectID, rh.repository, strconv.FormatInt(option.Page, 10), strconv.FormatInt(option.Size, 10),
		option.PackageName)
}

func (rh *repositoryHandler) searchChart(ctx context.Context, option repo.ListOption) (*repo.ListChartData, error) {
	if option.Size == 0 {
		option.Size = 10
	}

	switch rh.repoType {
	case repo.RepositoryTypeHelm:
		return rh.searchHelmChart(ctx, option)
	case repo.RepositoryTypeOCI:
		return rh.searchOCIChart(ctx, option)
	default:
		return nil, fmt.Errorf("unknown repo type %d", rh.repoType)
	}
}

func (rh *repositoryHandler) searchOCIChart(ctx context.Context, option repo.ListOption) (*repo.ListChartData, error) {
	return rh.listHelmChart(ctx, option)
}

func (rh *repositoryHandler) searchHelmChart(ctx context.Context, option repo.ListOption) (*repo.ListChartData, error) {
	req := searchChartReq{
		Fields: []string{"projectId", "name", "key", "type", "latest", "downloads", "versions", "description",
			"extension", "createdBy", "createdDate", "lastModifiedBy", "lastModifiedDate"},
		Page: searchChartPage{PageNumber: int(option.Page), PageSize: int(option.Size)},
		Sort: searchChartSort{Properties: []string{"lastModifiedDate"}, Direction: SortDesc},
		Rule: searchChartRule{
			Relation: "AND",
			Rules: []searchChartRules{
				{Field: "projectId", Value: rh.projectID, Operation: "EQ"},
				{Field: "repoType", Value: "HELM", Operation: "EQ"},
				{Field: "repoName", Value: rh.repository, Operation: "EQ"},
				{Field: "name", Value: fmt.Sprintf("*%s*", option.PackageName), Operation: "MATCH"},
			},
		},
	}
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("get chart param failed, %s", err.Error())
	}
	resp, err := rh.handler.post(ctx, repositorySearchHelmURI, nil, b)
	if err != nil {
		blog.Errorf("search helm chart from bk-repo get failed, %s, with projectID %s, repoName %s",
			err.Error(), rh.projectID, rh.repository)
		return nil, err
	}

	var r listPackResp
	if err = codec.DecJson(resp.Reply, &r); err != nil {
		blog.Errorf("search helm chart from bk-repo decode failed, %s, with resp %s", err.Error(), resp.Reply)
		return nil, err
	}
	if r.Code != respCodeOK {
		blog.Errorf("search helm chart from bk-repo get resp with error code %d, message %s, traceID %s",
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

func (rh *repositoryHandler) getChartDetail(ctx context.Context, name string) (
	*repo.Chart, error) {
	switch rh.repoType {
	case repo.RepositoryTypeHelm:
		return rh.getHelmChart(ctx, name)
	case repo.RepositoryTypeOCI:
		return rh.getOCIChart(ctx, name)
	default:
		return nil, fmt.Errorf("unknown repo type %d", rh.repoType)
	}
}

func (rh *repositoryHandler) getHelmChart(ctx context.Context, name string) (*repo.Chart, error) {
	resp, err := rh.handler.get(ctx, rh.getHelmChartURI(name), nil, nil)
	if err != nil {
		blog.Errorf("get helm chart from bk-repo get failed, %s, with projectID %s, repoName %s",
			err.Error(), rh.projectID, rh.repository)
		return nil, err
	}

	var r getPackResp
	if err = codec.DecJson(resp.Reply, &r); err != nil {
		blog.Errorf("get helm chart from bk-repo decode failed, %s, with resp %s", err.Error(), resp.Reply)
		return nil, err
	}
	if r.Code != respCodeOK {
		blog.Errorf("get helm chart from bk-repo get resp with error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)
		return nil, err
	}

	return r.Data.convert2Chart(), nil
}

func (rh *repositoryHandler) getHelmChartURI(name string) string {
	return fmt.Sprintf("%s/%s/%s?packageKey=helm://%s", repositoryGetHelmURI, rh.projectID, rh.repository, name)
}

func (rh *repositoryHandler) getOCIChart(ctx context.Context, name string) (*repo.Chart, error) {
	resp, err := rh.handler.get(ctx, rh.getOCIChartURI(name), nil, nil)
	if err != nil {
		blog.Errorf("list helm chart from bk-repo get failed, %s, with projectID %s, repoName %s",
			err.Error(), rh.projectID, rh.repository)
		return nil, err
	}

	var r getPackResp
	if err = codec.DecJson(resp.Reply, &r); err != nil {
		blog.Errorf("list helm chart from bk-repo decode failed, %s, with resp %s", err.Error(), resp.Reply)
		return nil, err
	}
	if r.Code != respCodeOK {
		blog.Errorf("list helm chart from bk-repo get resp with error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)
		return nil, err
	}

	return r.Data.convert2Chart(), nil
}

func (rh *repositoryHandler) getOCIChartURI(name string) string {
	return fmt.Sprintf("%s/%s/%s?packageKey=oci://%s", repositoryGetHelmURI, rh.projectID, rh.repository, name)
}

type searchChartReq struct {
	Fields []string        `json:"select"`
	Page   searchChartPage `json:"page"`
	Sort   searchChartSort `json:"sort"`
	Rule   searchChartRule `json:"rule"`
}

type searchChartPage struct {
	PageNumber int `json:"pageNumber"`
	PageSize   int `json:"pageSize"`
}

type searchChartSort struct {
	Properties []string `json:"properties"`
	Direction  Order    `json:"direction"`
}

type searchChartRule struct {
	Rules    []searchChartRules `json:"rules"`
	Relation string             `json:"relation"`
}

type searchChartRules struct {
	Field     string `json:"field"`
	Value     string `json:"value"`
	Operation string `json:"operation"`
}

// Order define sort order
type Order string

const (
	// SortAsc 升序
	SortAsc Order = "ASC"
	// SortDesc 降序
	SortDesc Order = "DESC"
)

type getPackResp struct {
	basicResp
	Data *pack `json:"data"`
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
	CreatedDate      json.Number   `json:"createdDate"`
	LastModifiedBy   string        `json:"lastModifiedBy"`
	LastModifiedDate json.Number   `json:"lastModifiedDate"`
}

type packExtension struct {
	AppVersion string `json:"appVersion"`
}

// convert2Chart 将bk-repo HELM仓库中的package信息, 转换为chart信息
func (p *pack) convert2Chart() *repo.Chart {
	chart := &repo.Chart{
		Key:         p.Key,
		Name:        p.Name,
		Type:        p.Type,
		Version:     p.Latest,
		AppVersion:  p.Extension.AppVersion,
		Description: p.Description,
		CreateBy:    p.CreatedBy,
		UpdateBy:    p.LastModifiedBy,
	}
	if i, err := p.CreatedDate.Int64(); err != nil {
		t, err := time.Parse(time.RFC3339Nano, p.CreatedDate.String())
		if err != nil {
			chart.CreateTime = p.CreatedDate.String()
		} else {
			chart.CreateTime = t.Format(common.TimeFormat)
		}
	} else {
		chart.CreateTime = time.UnixMilli(i).Format(common.TimeFormat)
	}
	if i, err := p.LastModifiedDate.Int64(); err != nil {
		t, err := time.Parse(time.RFC3339Nano, p.LastModifiedDate.String())
		if err != nil {
			chart.UpdateTime = p.LastModifiedDate.String()
		} else {
			chart.UpdateTime = t.Format(common.TimeFormat)
		}
	} else {
		chart.UpdateTime = time.UnixMilli(i).Format(common.TimeFormat)
	}
	return chart
}

func (ch *chartHandler) deleteChart(ctx context.Context) error {
	resp, err := ch.handler.delete(ctx, ch.getDeletePackageURI(), nil, nil)
	if err != nil {
		blog.Errorf(
			"delete chart from bk-repo failed, %s, with projectID %s, repoName %s, chartName %s",
			err.Error(), ch.projectID, ch.repository, ch.chartName)
		return err
	}

	var r basicResp
	if err = codec.DecJson(resp.Reply, &r); err != nil {
		blog.Errorf(
			"delete chart from bk-repo failed, %s, with resp %s", err.Error(), resp.Reply)
		return err
	}
	if r.Code != respCodeOK {
		blog.Errorf("delete chart from bk-repo get resp with error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)
		return fmt.Errorf("delete chart err with code %d, message %s, traceID %s", r.Code, r.Message, r.TraceID)
	}
	return nil
}

func (ch *chartHandler) getDeletePackageURI() string {
	return fmt.Sprintf("%s/%s/%s?packageKey=%s://%s", repositoryDeletePackageURI, ch.projectID, ch.repository,
		ch.repoType.PackageKey(), ch.chartName)
}

func (ch *chartHandler) deleteChartVersion(ctx context.Context, version string) error {
	resp, err := ch.handler.delete(ctx, ch.getDeleteVersionURI(version), nil, nil)
	if err != nil {
		blog.Errorf(
			"delete chart version from bk-repo failed, %s, with projectID %s, repoName %s, chartName %s, version %s",
			err.Error(), ch.projectID, ch.repository, ch.chartName, version)
		return err
	}

	var r basicResp
	if err = codec.DecJson(resp.Reply, &r); err != nil {
		blog.Errorf(
			"delete chart version from bk-repo failed, %s, with resp %s", err.Error(), resp.Reply)
		return err
	}
	if r.Code != respCodeOK {
		blog.Errorf("delete chart version from bk-repo get resp with error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)
		return fmt.Errorf("delete chart version err with code %d, message %s, traceID %s", r.Code, r.Message, r.TraceID)
	}
	return nil
}

func (ch *chartHandler) getDeleteVersionURI(version string) string {
	return fmt.Sprintf("%s/%s/%s?packageKey=%s://%s&version=%s", repositoryDeleteVersionURI, ch.projectID, ch.repository,
		ch.repoType.PackageKey(), ch.chartName, version)
}
