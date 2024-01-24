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

package client

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/pkg"
	utilschart "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/chart"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/httpx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	urlChartList          = "/projects/%s/repos/%s/charts"
	urlChartDetailGet     = "/projects/%s/repos/%s/charts/%s"
	urlChartDelete        = "/projects/%s/repos/%s/charts/%s"
	urlChartVersionList   = "/projects/%s/repos/%s/charts/%s/versions"
	urlChartVersionDetail = "/projects/%s/repos/%s/charts/%s/versions/%s"
	urlChartVersionDelete = "/projects/%s/repos/%s/charts/%s/versions/%s"
	urlChartCreate        = "/projects/%s/repos/%s/charts/upload"
)

// Chart return a pkg.ChartClient instance
func (c *Client) Chart() pkg.ChartClient {
	return &chart{Client: c}
}

type chart struct {
	*Client
}

// Versions list chart version
func (ct *chart) Versions(ctx context.Context, req *helmmanager.ListChartVersionV1Req) (
	*helmmanager.ChartVersionListData, error) {

	projectCode := req.GetProjectCode()
	if projectCode == "" {
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
		urlPrefix+fmt.Sprintf(urlChartVersionList, projectCode, repo, chartName)+"?"+
			ct.listChartVersionQuery(req).Encode(),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.ListChartVersionV1Resp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("list chart version get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

func (ct *chart) listChartVersionQuery(req *helmmanager.ListChartVersionV1Req) url.Values {
	query := url.Values{}
	if req.Page != nil {
		query.Set("page", strconv.FormatInt(int64(req.GetPage()), 10))
	}
	if req.Size != nil {
		query.Set("size", strconv.FormatInt(int64(req.GetSize()), 10))
	}
	return query
}

// Detail get chart detail
func (ct *chart) GetChartDetail(ctx context.Context, req *helmmanager.GetChartDetailV1Req) (*helmmanager.Chart, error) {
	projectCode := req.GetProjectCode()
	if projectCode == "" {
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
		urlPrefix+fmt.Sprintf(urlChartDetailGet, projectCode, repo, chartName),
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
		return nil, fmt.Errorf("get chart detail result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

// Detail get version detail
func (ct *chart) GetVersionDetail(ctx context.Context, req *helmmanager.GetVersionDetailV1Req) (
	*helmmanager.ChartDetail, error) {
	projectCode := req.GetProjectCode()
	if projectCode == "" {
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
	version := req.GetVersion()
	if version == "" {
		return nil, fmt.Errorf("chart version can not be empty")
	}

	resp, err := ct.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartVersionDetail, projectCode, repo, chartName, version),
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
		return nil, fmt.Errorf("get chart detail get result code %d, message: %s", r.GetCode(), r.GetMessage())
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

	data, _ := json.Marshal(req)

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
	projectCode := req.GetProjectCode()
	if projectCode == "" {
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
	data, _ := json.Marshal(req)

	resp, err := ct.delete(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartVersionDelete, projectCode, repo, chartName, version),
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

// PushChart push chart
func (ct *chart) PushChart(ctx context.Context, req *utilschart.PushChart) error {
	if req == nil {
		return fmt.Errorf("push chart request is empty")
	}

	if req.RepoName == "" {
		return fmt.Errorf("repository name can not be empty")
	}

	if req.ProjectCode == "" {
		return fmt.Errorf("repository projectCode can not be empty")
	}

	fi, err := os.Stat(req.FilePath)
	if err != nil {
		return err
	}

	// 文件夹需要压缩，生成的tgz文件需要删除
	if fi.IsDir() {
		req.FilePath, err = getTgz(req.FilePath)
		if err != nil {
			return err
		}

		defer func(path string) {
			err = os.Remove(req.FilePath)
			if err != nil {
				fmt.Printf("Failed to remove tgz file %s: %s", path, err)
			}
		}(req.FilePath)
	}

	data, header, err := getFileData(req.FilePath)
	if err != nil {
		return err
	}

	resp, err := ct.post(ctx, urlUploadChartPrefix+fmt.Sprintf(urlChartCreate, req.ProjectCode, req.RepoName)+
		ct.createChartQuery(req.Version, req.Force), header, data)
	if err != nil {
		return err
	}

	var r httpx.BaseResponse
	if err = json.Unmarshal(resp.Reply, &r); err != nil {
		return err
	}

	if r.Code != resultCodeSuccess {
		return fmt.Errorf("push chart get result code %d, message: %s", r.Code, r.Message)
	}

	return nil
}

// 将文件夹进行压缩
func getTgz(filePath string) (string, error) {
	// 转成所在系统合法路径
	name := filepath.FromSlash(filePath)

	if validChart, err := chartutil.IsChartDir(name); !validChart {
		return "", err
	}
	// 获取绝对路径
	chartPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	// 获取文件夹路径及名称
	dir, name := filepath.Split(chartPath)
	tgzPath := filepath.Join(dir, name+".tgz")
	file, err := os.Create(tgzPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer file.Close() // nolint

	// 创建 gzip.Writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close() // nolint

	// 创建 tar.Writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close() // nolint

	// 遍历源文件夹中的所有文件和子文件夹
	err = filepath.Walk(chartPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking through directory: %w", err)
		}
		// 压缩包里面的放置需要以路径分离后chartPath的文件夹名称为初始文件夹
		_, name := filepath.Split(chartPath)
		// 获取相对路径
		relativePath, err := filepath.Rel(chartPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		relativePath = filepath.Join(name, relativePath)

		// 创建 tar.Header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}

		// 设置相对路径
		header.Name = relativePath

		// 写入 header
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// 如果是文件，将文件内容写入 tar.Writer
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return fmt.Errorf("failed to write file content to tar writer: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error compressing to TGZ: %w", err)
	}

	return tgzPath, nil
}

// 生成文件流及返回Header
func getFileData(filePath string) ([]byte, http.Header, error) {
	if filePath == "" {
		return nil, nil, fmt.Errorf("empty param data")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// 创建一个新的multipart请求体
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	header := http.Header{}
	header.Add("Content-Type", writer.FormDataContentType())
	part, err := writer.CreateFormFile("chart", filePath)
	if err != nil {
		return nil, nil, err
	}

	_, err = io.Copy(part, f)
	if err != nil {
		return nil, nil, err
	}

	// 完成multipart请求体的创建
	err = writer.Close()
	if err != nil {
		return nil, nil, err
	}
	return body.Bytes(), header, nil
}

func (ct *chart) createChartQuery(version string, force bool) string {
	query := url.Values{}
	if version != "" {
		query.Set("version", version)
	}
	if force {
		query.Set("force", "true")
	}
	if len(query) == 0 {
		return ""
	}
	return "?" + query.Encode()
}

// List chart
func (ct *chart) List(ctx context.Context, req *helmmanager.ListChartV1Req) (*helmmanager.ChartListData, error) {
	if req == nil {
		return nil, fmt.Errorf("list chart request is empty")
	}

	projectCode := req.GetProjectCode()
	if projectCode == "" {
		return nil, fmt.Errorf("chart project can not be empty")
	}
	repositoryName := req.GetRepoName()
	if repositoryName == "" {
		return nil, fmt.Errorf("chart repository can not be empty")
	}

	resp, err := ct.get(
		ctx,
		urlPrefix+fmt.Sprintf(urlChartList, projectCode, repositoryName)+"?"+ct.listChartQuery(req).Encode(),
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var r helmmanager.ListChartV1Resp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return nil, err
	}

	if r.GetCode() != resultCodeSuccess {
		return nil, fmt.Errorf("list chart get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return r.Data, nil
}

func (ct *chart) listChartQuery(req *helmmanager.ListChartV1Req) url.Values {
	query := url.Values{}
	if req.Page != nil {
		query.Set("page", strconv.FormatInt(int64(req.GetPage()), 10))
	}
	if req.Size != nil {
		query.Set("size", strconv.FormatInt(int64(req.GetSize()), 10))
	}
	return query
}
