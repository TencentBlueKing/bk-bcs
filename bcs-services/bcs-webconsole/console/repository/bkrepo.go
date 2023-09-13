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

package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

type bkrepoStorage struct {
	client *http.Client
}

// UploadFile upload file to bkRepo
func (b *bkrepoStorage) UploadFile(ctx context.Context, localFile, filePath string) error {
	//上传文件API PUT /generic/{project}/{repoName}/{fullPath}
	rawURL := fmt.Sprintf("%s/generic/%s/%s/%s", config.G.Repository.Bkrepo.Endpoint,
		config.G.Repository.Bkrepo.Project, config.G.Repository.Bkrepo.Repo, filePath)

	f, err := os.Open(localFile)
	if err != nil {
		return fmt.Errorf("Open local file %s failed: %v\n", localFile, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, rawURL, f)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-BKREPO-OVERWRITE", "true")

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		klog.Errorf("Upload file failed, resp: %s\n", string(body))
		return fmt.Errorf("Upload file failed, Err code: %v\n", resp.StatusCode)
	}
	return nil
}

// ListFile list of files under a bkRepo folder
func (b *bkrepoStorage) ListFile(ctx context.Context, folderName string) ([]string, error) {
	//节点详情 https://github.com/TencentBlueKing/bk-repo/blob/master/docs/apidoc/node/node.md
	//GET /repository/api/node/page/{projectId}/{repoName}/{fullPath}
	rawURL := fmt.Sprintf("%s/repository/api/node/page/%s/%s/%s", config.G.Repository.Bkrepo.Endpoint,
		config.G.Repository.Bkrepo.Project, config.G.Repository.Bkrepo.Repo, folderName)

	files := make([]string, 0)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		return files, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return files, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return files, fmt.Errorf("Get file list failed, Err code: %v\n", resp.StatusCode)
	}
	listResult := new(listFileResult)
	if err := json.NewDecoder(resp.Body).Decode(listResult); err != nil {
		return files, fmt.Errorf("Unmarshal resp err: %v\n", err)
	}
	if len(listResult.Data.Records) <= 0 {
		return files, fmt.Errorf("folder %s is not exit", folderName)
	}

	for _, record := range listResult.Data.Records {
		if !record.Folder {
			files = append(files, record.Name)
		}
	}
	return files, nil
}

// DownloadFile download file from bkRepo
func (b *bkrepoStorage) DownloadFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	//下载文件API PUT /generic/{project}/{repoName}/{fullPath}
	rawURL := fmt.Sprintf("%s/generic/%s/%s/%s?download=true", config.G.Repository.Bkrepo.Endpoint,
		config.G.Repository.Bkrepo.Project, config.G.Repository.Bkrepo.Repo, filePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		klog.Errorf("Download file failed, resp: %s\n", string(body))
		resp.Body.Close()
		return nil, fmt.Errorf("Download file failed, Err code: %v\n", resp.StatusCode)
	}

	return resp.Body, nil
}

func newBkRepoStorage() (Provider, error) {
	transport := &bkrepoAuthTransport{
		Username:  config.G.Repository.Bkrepo.UserName,
		Password:  config.G.Repository.Bkrepo.Password,
		Transport: http.DefaultTransport,
	}
	p := &bkrepoStorage{
		client: &http.Client{Transport: transport},
	}
	return p, nil
}

// bkrepoAuthTransport 给请求增加 Authorization header
type bkrepoAuthTransport struct {
	Username  string
	Password  string
	Transport http.RoundTripper
}

// RoundTrip Transport
func (t *bkrepoAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.Username, t.Password)
	resp, err := t.transport().RoundTrip(req)
	return resp, err
}

func (t *bkrepoAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// 查询文件信息(节点详情)结果
// https://github.com/TencentBlueKing/bk-repo/blob/master/docs/apidoc/node/node.md
type listFileResult struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    listFileData `json:"data"`
}

type listFileData struct {
	PageNumber   int                `json:"pageNumber"`
	PageSize     int                `json:"pageSize"`
	TotalRecords int                `json:"totalRecords"`
	TotalPages   int                `json:"totalPages"`
	Records      []*listFileRecords `json:"records"`
}

type listFileRecords struct {
	ProjectId string `json:"projectId"`
	RepoName  string `json:"repoName"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	FullPath  string `json:"fullPath"`
	Folder    bool   `json:"folder"`
	Size      int    `json:"size"`
}
