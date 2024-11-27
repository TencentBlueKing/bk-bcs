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

// Package repository xxx
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

type bkrepoStorage struct {
	client *http.Client
}

// UploadFile upload file to bkRepo
func (b *bkrepoStorage) UploadFile(ctx context.Context, localFile, filePath string) error {
	// 上传文件API PUT /generic/{project}/{repoName}/{fullPath}
	rawURL := fmt.Sprintf("%s/generic/%s/%s/%s", config.G.Repository.Bkrepo.Endpoint,
		config.G.Repository.Bkrepo.Project, config.G.Repository.Bkrepo.Repo, filePath)

	f, err := os.Open(localFile)
	if err != nil {
		return fmt.Errorf("open local file %s failed: %v", localFile, err)
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
		blog.Errorf("Upload file failed, resp: %s\n", string(body))
		return fmt.Errorf("upload file failed, Err code: %v", resp.StatusCode)
	}
	return nil
}

// UploadFileByReader upload file to bkRepo by Reader
func (b *bkrepoStorage) UploadFileByReader(ctx context.Context, r io.Reader, filePath string) error {
	// 上传文件API PUT /generic/{project}/{repoName}/{fullPath}
	rawURL := fmt.Sprintf("%s/generic/%s/%s/%s", config.G.Repository.Bkrepo.Endpoint,
		config.G.Repository.Bkrepo.Project, config.G.Repository.Bkrepo.Repo, filePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, rawURL, r)
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
		blog.Errorf("Upload file failed, resp: %s\n", string(body))
		return fmt.Errorf("upload file failed, Err code: %v", resp.StatusCode)
	}
	return nil
}

// IsExist 是否存在
func (b *bkrepoStorage) IsExist(ctx context.Context, filePath string) (bool, error) {
	rawURL := fmt.Sprintf("%s/generic/%s/%s%s", config.G.Repository.Bkrepo.Endpoint,
		config.G.Repository.Bkrepo.Project, config.G.Repository.Bkrepo.Repo, filePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return false, fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return true, nil
}

// ListFile list of files under a bkRepo folder
func (b *bkrepoStorage) ListFile(ctx context.Context, folderName string) ([]FileContent, error) {
	// 节点详情 https://github.com/TencentBlueKing/bk-repo/blob/master/docs/apidoc/node/node.md
	// GET /repository/api/node/page/{projectId}/{repoName}/{fullPath}
	// 目前一天最多一千条，pageSize限制为1000条
	rawURL := fmt.Sprintf("%s/repository/api/node/page/%s/%s/%s?pageSize=1000",
		config.G.Repository.Bkrepo.Endpoint, config.G.Repository.Bkrepo.Project, config.G.Repository.Bkrepo.Repo,
		folderName)

	var files []FileContent
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return files, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return files, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return files, fmt.Errorf("get file list failed, Err code: %v", resp.StatusCode)
	}
	listResult := new(listFileResult)
	if err := json.NewDecoder(resp.Body).Decode(listResult); err != nil {
		return files, fmt.Errorf("unmarshal resp err: %v", err)
	}
	if len(listResult.Data.Records) == 0 {
		return files, fmt.Errorf("folder %s is not exit", folderName)
	}

	for _, record := range listResult.Data.Records {
		if !record.Folder {
			files = append(files, FileContent{
				FileName: record.Name,
				Size:     formatBytes(int64(record.Size)),
			})
		}
	}
	return files, nil
}

// ListFolders list of folders under current bkRepo folder
func (b *bkrepoStorage) ListFolders(ctx context.Context, folderName string) ([]string, error) {
	// 节点详情 https://github.com/TencentBlueKing/bk-repo/blob/master/docs/apidoc/node/node.md
	// GET /repository/api/node/page/{projectId}/{repoName}/{fullPath}
	rawURL := fmt.Sprintf("%s/repository/api/node/page/%s/%s/%s", config.G.Repository.Bkrepo.Endpoint,
		config.G.Repository.Bkrepo.Project, config.G.Repository.Bkrepo.Repo, folderName)

	folders := make([]string, 0)

	// 全量列出
	pageNumber := 1
	pageSize := 1000
	for pageSize != 0 {
		newRawURL := fmt.Sprintf("%s?pageNumber=%d&pageSize=%d", rawURL, pageNumber, pageSize)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, newRawURL, nil)
		if err != nil {
			return folders, err
		}
		resp, err := b.client.Do(req)
		if err != nil {
			return folders, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return folders, fmt.Errorf("get file list failed, Err code: %v", resp.StatusCode)
		}
		listResult := new(listFileResult)
		if err := json.NewDecoder(resp.Body).Decode(listResult); err != nil {
			return folders, fmt.Errorf("unmarshal resp err: %v", err)
		}

		for _, record := range listResult.Data.Records {
			if record.Folder {
				folders = append(folders, record.Name)
			}
		}

		pageNumber++
		// 非1000条的情况下直接退出
		if len(listResult.Data.Records) != pageSize {
			pageSize = 0
		}

	}
	return folders, nil
}

// DeleteFolders delete of folders under current bkRepo folder
func (b *bkrepoStorage) DeleteFolders(ctx context.Context, folderName string) error {
	// 节点详情 https://github.com/TencentBlueKing/bk-repo/blob/master/docs/apidoc/node/node.md
	// GET /repository/api/node/page/{projectId}/{repoName}/{fullPath}
	rawURL := fmt.Sprintf("%s/repository/api/node/delete/%s/%s/%s", config.G.Repository.Bkrepo.Endpoint,
		config.G.Repository.Bkrepo.Project, config.G.Repository.Bkrepo.Repo, folderName)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("delete folder failed, Err code: %v", resp.StatusCode)
	}

	return nil
}

// DownloadFile download file from bkRepo
func (b *bkrepoStorage) DownloadFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	// 下载文件API PUT /generic/{project}/{repoName}/{fullPath}
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
		blog.Errorf("Download file failed, resp: %s\n", string(body))
		resp.Body.Close()
		return nil, fmt.Errorf("download file failed, Err code: %v", resp.StatusCode)
	}

	return resp.Body, nil
}

func newBkRepoStorage() (Provider, error) {
	transport := &bkrepoAuthTransport{
		Username:  config.G.Repository.Bkrepo.Username,
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
