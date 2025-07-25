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

package bkrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cm "github.com/chartmuseum/helm-push/pkg/chartmuseum"
	"github.com/chartmuseum/helm-push/pkg/helm"
	"k8s.io/helm/pkg/chartutil"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
)

const (
	// repositoryListHelmURI list chart
	repositoryListHelmURI = "/repository/api/package/page"
	// repositorySearchHelmURI search chart
	repositorySearchHelmURI = "/repository/api/package/search"
	// get helm chart detail
	repositoryGetHelmURI = "/repository/api/package/info"
	// delete helm chart
	repositoryDeletePackageURI = "/helm/ext/package/delete"
	// delete helm chart version
	repositoryDeleteVersionURI = "/helm/ext/version/delete"

	// upload chart timeout
	timeout = 10
)

// list chart
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

// list oci chart
func (rh *repositoryHandler) listOCIChart(ctx context.Context, option repo.ListOption) (*repo.ListChartData, error) {
	return rh.listHelmChart(ctx, option)
}

// list helm chart
func (rh *repositoryHandler) listHelmChart(ctx context.Context, option repo.ListOption) (*repo.ListChartData, error) {
	resp, err := rh.handler.get(ctx, rh.getListHelmChartURI(option), nil, nil)
	if err != nil {
		blog.Errorf("list helm chart from bk-repo get failed, %s, with projectID %s, repoName %s",
			err.Error(), rh.projectID, rh.repository)
		return nil, err
	}

	var r listPackResp
	if err = json.Unmarshal(resp.Reply, &r); err != nil {
		blog.Errorf("list helm chart from bk-repo decode failed, %s, with resp %s", err.Error(), resp.Reply)
		return nil, err
	}
	if r.Code != respCodeOK {
		blog.Errorf("list helm chart from bk-repo get resp with error code %d, message %s, traceID %s",
			r.Code, r.Message, r.TraceID)
		return nil, err
	}

	// parse chart
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

// get list helm chart uri
func (rh *repositoryHandler) getListHelmChartURI(option repo.ListOption) string {
	return fmt.Sprintf("%s/%s/%s/?pageNumber=%s&pageSize=%s&packageName=%s", repositoryListHelmURI,
		rh.projectID, rh.repository, strconv.FormatInt(option.Page, 10), strconv.FormatInt(option.Size, 10),
		option.PackageName)
}

// search chart
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

// searchOCIChart
func (rh *repositoryHandler) searchOCIChart(ctx context.Context, option repo.ListOption) (*repo.ListChartData, error) {
	return rh.listHelmChart(ctx, option)
}

// searchHelmChart
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

	// parse list resp
	var r listPackResp
	if err = json.Unmarshal(resp.Reply, &r); err != nil {
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

// get chart detail
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

// get helm chart
func (rh *repositoryHandler) getHelmChart(ctx context.Context, name string) (*repo.Chart, error) {
	resp, err := rh.handler.get(ctx, rh.getHelmChartURI(name), nil, nil)
	if err != nil {
		blog.Errorf("get helm chart from bk-repo get failed, %s, with projectID %s, repoName %s",
			err.Error(), rh.projectID, rh.repository)
		return nil, err
	}

	var r getPackResp
	if err = json.Unmarshal(resp.Reply, &r); err != nil {
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

// getHelmChartURI
func (rh *repositoryHandler) getHelmChartURI(name string) string {
	return fmt.Sprintf("%s/%s/%s?packageKey=helm://%s", repositoryGetHelmURI, rh.projectID, rh.repository, name)
}

// getOCIChart
func (rh *repositoryHandler) getOCIChart(ctx context.Context, name string) (*repo.Chart, error) {
	resp, err := rh.handler.get(ctx, rh.getOCIChartURI(name), nil, nil)
	if err != nil {
		blog.Errorf("list helm chart from bk-repo get failed, %s, with projectID %s, repoName %s",
			err.Error(), rh.projectID, rh.repository)
		return nil, err
	}

	var r getPackResp
	if err = json.Unmarshal(resp.Reply, &r); err != nil {
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

// getOCIChartURI
func (rh *repositoryHandler) getOCIChartURI(name string) string {
	return fmt.Sprintf("%s/%s/%s?packageKey=oci://%s", repositoryGetHelmURI, rh.projectID, rh.repository, name)
}

// uploadChart
func (rh *repositoryHandler) uploadChart(_ context.Context, option repo.UploadOption) error {
	chart, err := chartutil.LoadArchive(option.Content)
	if err != nil {
		return fmt.Errorf("failed load chart, %s", err)
	}
	helmChart := &helm.Chart{Chart: chart}
	// 设置自定义版本
	if option.Version != "" {
		helmChart.SetVersion(option.Version)
	}

	// 创建临时目录
	tmp, err := os.MkdirTemp("", "helm-push-")
	if err != nil {
		return fmt.Errorf("error creates a new temporary directory in the directory dir, %s", err)
	}
	defer func(path string) {
		err = os.RemoveAll(path)
		if err != nil {
			blog.Errorf("failed to remove temporary directory, %s: %s", path, err.Error())
		}
	}(tmp)
	chartPackagePath, err := helm.CreateChartPackage(helmChart, tmp)
	if err != nil {
		return fmt.Errorf("creates chart package in directory error, %s", err)
	}

	// new chartmusuem client
	cmClient, err := cm.NewClient(
		cm.URL(rh.getRepoURL()),
		cm.Username(rh.user.Name),
		cm.Password(rh.user.Password),
		cm.Timeout(timeout),
	)
	if err != nil {
		return fmt.Errorf("creates client fail, %s", err)
	}

	// 上传chart
	chartPackage, err := cmClient.UploadChartPackage(chartPackagePath, option.Force)
	if err != nil {
		return fmt.Errorf("uploads a chart package fail, %s", err)
	}
	if chartPackage.StatusCode != 201 {
		return fmt.Errorf("uploads a chart package response error, %s", chartPackage.Status)
	}
	return nil
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

// pack helm package
type pack struct {
	ProjectID        string          `json:"projectId"`
	Name             string          `json:"name"`
	Key              string          `json:"key"`
	Type             string          `json:"type"`
	Latest           string          `json:"latest"`
	Downloads        int64           `json:"downloads"`
	Versions         int             `json:"versions"`
	Description      string          `json:"description"`
	VersionTag       interface{}     `json:"versionTag"`
	Extension        packExtension   `json:"extension"`
	HistoryVersion   []string        `json:"historyVersion"`
	CreatedBy        string          `json:"createdBy"`
	CreatedDate      json.RawMessage `json:"createdDate"`
	LastModifiedBy   string          `json:"lastModifiedBy"`
	LastModifiedDate json.RawMessage `json:"lastModifiedDate"`
}

type packExtension struct {
	AppVersion string `json:"appVersion"`
}

// convert2Chart 将bk-repo HELM仓库中的package信息, 转换为chart信息
func (p *pack) convert2Chart() *repo.Chart {
	if p == nil {
		return &repo.Chart{}
	}
	chart := &repo.Chart{
		Key:         p.Key,
		Name:        p.Name,
		Type:        p.Type,
		Version:     p.Latest,
		AppVersion:  p.Extension.AppVersion,
		Description: p.Description,
		CreateBy:    p.CreatedBy,
		UpdateBy:    p.LastModifiedBy,
		CreateTime:  getDateFromRawMessage(p.CreatedDate),
		UpdateTime:  getDateFromRawMessage(p.LastModifiedDate),
	}
	return chart
}

// 返回结构中可能包含 int 格式时间戳和 string 格式时间
func getDateFromRawMessage(raw json.RawMessage) string {
	data, err := raw.MarshalJSON()
	if err != nil {
		return "null"
	}
	if !utf8.Valid(data) {
		return "null"
	}
	i, err := strconv.Atoi(string(data))
	if err != nil {
		return string(data)
	}
	return time.UnixMilli(int64(i)).Format(common.TimeFormat)
}

// deleteChart
func (ch *chartHandler) deleteChart(ctx context.Context) error {
	resp, err := ch.handler.delete(ctx, ch.getDeletePackageURI(), nil, nil)
	if err != nil {
		blog.Errorf(
			"delete chart from bk-repo failed, %s, with projectID %s, repoName %s, chartName %s",
			err.Error(), ch.projectID, ch.repository, ch.chartName)
		return err
	}

	var r basicResp
	if err = json.Unmarshal(resp.Reply, &r); err != nil {
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

// getDeletePackageURI
func (ch *chartHandler) getDeletePackageURI() string {
	return fmt.Sprintf("%s/%s/%s?packageKey=%s://%s", repositoryDeletePackageURI, ch.projectID, ch.repository,
		ch.repoType.PackageKey(), ch.chartName)
}

// deleteChartVersion
func (ch *chartHandler) deleteChartVersion(ctx context.Context, version string) error {
	resp, err := ch.handler.delete(ctx, ch.getDeleteVersionURI(version), nil, nil)
	if err != nil {
		blog.Errorf(
			"delete chart version from bk-repo failed, %s, with projectID %s, repoName %s, chartName %s, version %s",
			err.Error(), ch.projectID, ch.repository, ch.chartName, version)
		return err
	}

	var r basicResp
	if err = json.Unmarshal(resp.Reply, &r); err != nil {
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

// getDeleteVersionURI
func (ch *chartHandler) getDeleteVersionURI(version string) string {
	u, _ := url.ParseRequestURI(fmt.Sprintf("%s/%s/%s", repositoryDeleteVersionURI, ch.projectID, ch.repository))
	q := u.Query()
	q.Add("packageKey", fmt.Sprintf("%s://%s", ch.repoType.PackageKey(), ch.chartName))
	q.Add("version", version)
	u.RawQuery = q.Encode()
	return u.String()
}
