/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/metrics"
	"bscp.io/pkg/thirdparty/repo"
	"bscp.io/pkg/tools"
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
	resp, err := t.transport(req).RoundTrip(req)
	return resp, err
}

func (t *bkrepoAuthTransport) transport(req *http.Request) http.RoundTripper {
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
	repoCreated map[string]struct{}
}

func (c *bkrepoClient) ensureRepo(kt *kit.Kit) error {
	repoName, err := repo.GenRepoName(kt.BizID)
	if err != nil {
		return err
	}
	if _, ok := c.repoCreated[repoName]; ok {
		return nil
	}
	repoReq := &repo.CreateRepoReq{
		ProjectID:     cc.ApiServer().Repo.BkRepo.Project,
		Name:          repoName,
		Type:          repo.RepositoryType,
		Category:      repo.CategoryType,
		Configuration: repo.Configuration{Type: repo.RepositoryCfgType},
		Description:   fmt.Sprintf("bscp %d business repository", kt.BizID),
	}
	if err := c.cli.CreateRepo(kt.Ctx, repoReq); err != nil {
		return err
	}

	c.repoCreated[repoName] = struct{}{}
	return nil
}

// getNodeMetadata If the node already exists, this appID will be added to the metadata of the current node.
// If not exist, will create new metadata with this bizID and appID.
func getNodeMetadata(kt *kit.Kit, cli *repo.Client, opt *repo.NodeOption, appID uint32) (string, error) {
	metadata, err := cli.QueryMetadata(kt.Ctx, opt)
	if err != nil {
		return "", err
	}

	if len(metadata) == 0 {
		meta := repo.NodeMeta{
			BizID: opt.BizID,
			AppID: []uint32{appID},
		}

		return meta.String()
	}

	// validate already node metadata.
	bizID, exist := metadata["biz_id"]
	if !exist {
		return "", errors.New("node metadata not has biz id")
	}

	if bizID != strconv.Itoa(int(opt.BizID)) {
		return "", fmt.Errorf("node metadata %s biz id is different from the request %d biz id", bizID, opt.BizID)
	}

	appIDStr, exist := metadata["app_id"]
	if !exist {
		return "", errors.New("node metadata not has app id")
	}

	appIDs := make([]uint32, 0)
	if err = json.Unmarshal([]byte(appIDStr), &appIDs); err != nil {
		return "", fmt.Errorf("unmarshal node metadata appID failed, err: %v", err)
	}

	// judge current app if already upload this node.
	var idExist bool
	for index := range appIDs {
		if appIDs[index] == appID {
			idExist = true
			break
		}
	}

	if !idExist {
		appIDs = append(appIDs, appID)
	}

	meta := &repo.NodeMeta{
		BizID: opt.BizID,
		AppID: appIDs,
	}
	return meta.String()
}

// Upload file to bkrepo
func (c *bkrepoClient) Upload(kt *kit.Kit, sign string, body io.Reader) (*ObjectMetadata, error) {
	if err := c.ensureRepo(kt); err != nil {
		return nil, errors.Wrap(err, "ensure repo failed")
	}

	opt := &repo.NodeOption{Project: c.project, BizID: kt.BizID, Sign: sign}
	nodeMeta, err := getNodeMetadata(kt, c.cli, opt, kt.AppID)
	if err != nil {
		return nil, errors.Wrap(err, "get node metadata")
	}

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
	req.Header.Set(repo.HeaderKeyMETA, nodeMeta)
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

	metadata := &ObjectMetadata{
		ByteSize: resp.ContentLength,
		Sha256:   sign,
	}

	return metadata, nil
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

// newBKRepoProvider new bkrepo provider
func newBKRepoProvider(settings cc.Repository) (Provider, error) {
	cli, err := repo.NewClient(settings, metrics.Register())
	if err != nil {
		return nil, err
	}

	host := settings.BkRepo.Endpoints[0]

	p := &bkrepoClient{
		cli:         cli,
		host:        host,
		project:     settings.BkRepo.Project,
		repoCreated: map[string]struct{}{},
	}

	transport := &bkrepoAuthTransport{
		Username:  settings.BkRepo.Username,
		Password:  settings.BkRepo.Password,
		Transport: tools.NewCurlLogTransport(defaultTransport),
	}

	p.client = &http.Client{Transport: transport}

	return p, nil
}
