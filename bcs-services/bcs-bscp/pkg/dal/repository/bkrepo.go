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

package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/pkg/errors"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/repo"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

const (
	// tempDownloadURLExpireSeconds is the expire seconds for the temp download url.
	tempDownloadURLExpireSeconds = 3600
)

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

// bkrepoClient client struct
type bkrepoClient struct {
	host        string
	project     string
	client      *http.Client
	cli         *repo.Client
	repoCreated *RepoCreated
}

// RepoCreated is the created repo data with lock to keep its concurrent security
type RepoCreated struct {
	sync.Mutex
	created map[string]struct{}
}

// Set sets kv
func (r *RepoCreated) Set(name string) {
	r.Lock()
	defer r.Unlock()
	r.created[name] = struct{}{}
}

// Exist check kv
func (r *RepoCreated) Exist(name string) bool {
	r.Lock()
	defer r.Unlock()
	_, ok := r.created[name]
	return ok
}

func (c *bkrepoClient) ensureRepo(kt *kit.Kit) error {
	repoName, err := repo.GenRepoName(kt.BizID)
	if err != nil {
		return err
	}

	if c.repoCreated.Exist(repoName) {
		return nil
	}

	repoReq := &repo.CreateRepoReq{
		ProjectID:     c.project,
		Name:          repoName,
		Type:          repo.RepositoryType,
		Category:      repo.CategoryType,
		Configuration: repo.Configuration{Type: repo.RepositoryCfgType},
		Description:   fmt.Sprintf("bscp %d business repository", kt.BizID),
	}
	if err := c.cli.CreateRepo(kt.Ctx, repoReq); err != nil {
		return err
	}

	c.repoCreated.Set(repoName)
	return nil
}

// Upload file to bkrepo
func (c *bkrepoClient) Upload(kt *kit.Kit, sign string, body io.Reader) (*ObjectMetadata, error) {
	if err := c.ensureRepo(kt); err != nil {
		return nil, errors.Wrap(err, "ensure repo failed")
	}

	opt := &repo.NodeOption{Project: c.project, BizID: kt.BizID, Sign: sign}

	node, err := repo.GenNodePath(opt)
	if err != nil {
		return nil, err
	}

	rawURL := fmt.Sprintf("%s%s", c.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodPut, rawURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set(constant.RidKey, kt.Rid)
	req.Header.Set(repo.HeaderKeyOverwrite, "true")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("upload status %d != 200", resp.StatusCode)
	}

	uploadResp := new(repo.UploadResp)
	if err := json.NewDecoder(resp.Body).Decode(uploadResp); err != nil {
		return nil, errors.Wrap(err, "upload response")
	}

	if uploadResp.Code != 0 {
		return nil, errors.Errorf("upload code %d != 0", uploadResp.Code)
	}

	// cos return not have metadata
	metadata := &ObjectMetadata{
		Sha256:   uploadResp.Data.Sha256,
		ByteSize: uploadResp.Data.Size,
	}

	return metadata, nil
}

// Download download file from bkrepo
func (c *bkrepoClient) Download(kt *kit.Kit, sign string) (io.ReadCloser, int64, error) {
	node, err := repo.GenNodePath(&repo.NodeOption{Project: c.project, BizID: kt.BizID, Sign: sign})
	if err != nil {
		return nil, 0, err
	}

	rawURL := fmt.Sprintf("%s%s", c.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set(constant.RidKey, kt.Rid)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, 0, errf.ErrFileContentNotFound
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, 0, errors.Errorf("download status %d != 200", resp.StatusCode)
	}

	return resp.Body, resp.ContentLength, nil
}

// Metadata bkrepo file metadata
func (c *bkrepoClient) Metadata(kt *kit.Kit, sign string) (*ObjectMetadata, error) {
	node, err := repo.GenNodePath(&repo.NodeOption{Project: c.project, BizID: kt.BizID, Sign: sign})
	if err != nil {
		return nil, err
	}

	rawURL := fmt.Sprintf("%s%s", c.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(constant.RidKey, kt.Rid)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errf.ErrFileContentNotFound
	}

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("metadata status %d != 200", resp.StatusCode)
	}

	sha256 := resp.Header.Get("X-Checksum-Sha256")
	md5 := resp.Header.Get("X-Checksum-Md5")

	if sha256 != sign {
		return nil, errors.Errorf("metadata sha256 [%s] does not match the given signature [%s]", sha256, sign)
	}

	metadata := &ObjectMetadata{
		ByteSize: resp.ContentLength,
		Sha256:   sign,
		Md5:      md5,
	}

	return metadata, nil
}

// InitBlockUpload init block upload file
func (c *bkrepoClient) InitBlockUpload(kt *kit.Kit, sign string) (string, error) {

	if err := c.ensureRepo(kt); err != nil {
		return "", errors.Wrap(err, "ensure repo failed")
	}

	opt := &repo.NodeOption{Project: c.project, BizID: kt.BizID, Sign: sign}

	node, err := repo.GenBlockNodePath(opt)
	if err != nil {
		return "", err
	}

	rawURL := fmt.Sprintf("%s%s", c.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodPost, rawURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set(constant.RidKey, kt.Rid)
	req.Header.Set(repo.HeaderKeyOverwrite, "true")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.Errorf("init block upload status %d != 200", resp.StatusCode)
	}

	bkrepoResp := new(repo.InitBlockUploadResp)
	if err := json.NewDecoder(resp.Body).Decode(bkrepoResp); err != nil {
		return "", errors.Wrap(err, "init block upload response")
	}

	if bkrepoResp.Code != 0 {
		return "", errors.Errorf("init block upload code %d != 0", bkrepoResp.Code)
	}

	if bkrepoResp.Data == nil {
		return "", errors.New("init block upload response data is nil")
	}

	return bkrepoResp.Data.UploadID, nil
}

// BlockUpload upload one block of the file
func (c *bkrepoClient) BlockUpload(kt *kit.Kit, sign string, uploadID string, blockNum uint32, body io.Reader) error {

	node, err := repo.GenNodePath(&repo.NodeOption{Project: c.project, BizID: kt.BizID, Sign: sign})
	if err != nil {
		return err
	}

	rawURL := fmt.Sprintf("%s%s", c.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodPut, rawURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set(constant.RidKey, kt.Rid)
	req.Header.Set(repo.HeaderKeyUploadID, uploadID)
	req.Header.Set(repo.HeaderKeySequence, strconv.Itoa(int(blockNum)))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("block upload status %d != 200", resp.StatusCode)
	}

	bkrepoResp := new(repo.BkRepoBaseResp)
	if err := json.NewDecoder(resp.Body).Decode(bkrepoResp); err != nil {
		return errors.Wrap(err, "block upload response")
	}

	if bkrepoResp.Code != 0 {
		return errors.Errorf("block upload code %d != 0", bkrepoResp.Code)
	}

	return nil
}

// CompleteBlockUpload complete block upload and return metadata
func (c *bkrepoClient) CompleteBlockUpload(kt *kit.Kit, sign string, uploadID string) (*ObjectMetadata, error) {

	node, err := repo.GenBlockNodePath(&repo.NodeOption{Project: c.project, BizID: kt.BizID, Sign: sign})
	if err != nil {
		return nil, err
	}

	rawURL := fmt.Sprintf("%s%s", c.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodPut, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(constant.RidKey, kt.Rid)
	req.Header.Set(repo.HeaderKeyUploadID, uploadID)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("complete block upload status %d != 200", resp.StatusCode)
	}

	uploadResp := new(repo.UploadResp)
	if err := json.NewDecoder(resp.Body).Decode(uploadResp); err != nil {
		return nil, errors.Wrap(err, "complete block upload response")
	}

	if uploadResp.Code != 0 {
		return nil, errors.Errorf("complete block upload code %d != 0", uploadResp.Code)
	}

	return c.Metadata(kt, sign)
}

// URIDecorator ..
func (c *bkrepoClient) URIDecorator(bizID uint32) DecoratorInter {
	return newUriDecoratorInter(bizID)
}

// DownloadLink bkrepo file download link
func (c *bkrepoClient) DownloadLink(kt *kit.Kit, sign string, fetchLimit uint32) (string, error) {
	repoName, err := repo.GenRepoName(kt.BizID)
	if err != nil {
		return "", err
	}

	objPath, err := repo.GenNodeFullPath(sign)
	if err != nil {
		return "", err
	}

	// get file download url.
	url, err := c.cli.GenerateTempDownloadURL(kt.Ctx, &repo.GenerateTempDownloadURLReq{
		ProjectID:     c.project,
		RepoName:      repoName,
		FullPathSet:   []string{objPath},
		ExpireSeconds: uint32(tempDownloadURLExpireSeconds),
		Permits:       fetchLimit,
		Type:          "DOWNLOAD",
	})

	if err != nil {
		return "", errors.Wrap(err, "generate temp download url failed")
	}

	return url, nil
}

// AsyncDownload bkrepo
func (c *bkrepoClient) AsyncDownload(kt *kit.Kit, sign string) (string, error) {
	return "", nil
}

// AsyncDownloadStatus bkrepo
func (c *bkrepoClient) AsyncDownloadStatus(kt *kit.Kit, sign string, taskID string) (bool, error) {
	return false, nil
}

// newBKRepoClient new bkrepo client
func newBKRepoClient(settings cc.Repository) (BaseProvider, error) {
	cli, err := repo.NewClient(settings, metrics.Register())
	if err != nil {
		return nil, err
	}

	host := settings.BkRepo.Endpoints[0]

	p := &bkrepoClient{
		cli:     cli,
		host:    host,
		project: settings.BkRepo.Project,
		repoCreated: &RepoCreated{
			created: make(map[string]struct{}),
		},
	}

	transport := &bkrepoAuthTransport{
		Username:  settings.BkRepo.Username,
		Password:  settings.BkRepo.Password,
		Transport: tools.NewCurlLogTransport(defaultTransport),
	}

	p.client = &http.Client{Transport: transport}

	return p, nil
}

// newBKRepoProvider new bkrepo provider
func newBKRepoProvider(settings cc.Repository) (Provider, error) {
	p, err := newBKRepoClient(settings)
	if err != nil {
		return nil, err
	}

	var c VariableCacher
	c, err = newVariableCacher(settings.RedisCluster, p)
	if err != nil {
		return nil, err
	}

	return &repoProvider{
		BaseProvider:   p,
		VariableCacher: c,
	}, nil
}
