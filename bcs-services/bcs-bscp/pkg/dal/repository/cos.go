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
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	cos "github.com/tencentyun/cos-go-sdk-v5"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/thirdparty/repo"
	"bscp.io/pkg/tools"
)

const (
	// cosSchema 请求使用 https 协议
	cosSchema = "https"
)

// cosClient tencentcloud cos client struct
type cosClient struct {
	host   string
	client *http.Client
}

// Upload upload file to cos
func (c *cosClient) Upload(kt *kit.Kit, fileContentID string, body io.Reader, contentLength int64) (*ObjectMetadata, error) {
	node, err := repo.GenS3NodeFullPath(kt.BizID, fileContentID)
	if err != nil {
		return nil, err
	}

	rawURL := fmt.Sprintf("%s%s", c.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodPut, rawURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Length", strconv.FormatInt(contentLength, 10))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set(constant.RidKey, kt.Rid)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("upload status %d != 200", resp.StatusCode)
	}

	// cos return not have metadata
	metadata := &ObjectMetadata{
		Sha256: fileContentID,
	}

	return metadata, nil
}

// Download download file from cos
func (c *cosClient) Download(kt *kit.Kit, fileContentID string) (io.ReadCloser, int64, error) {
	node, err := repo.GenS3NodeFullPath(kt.BizID, fileContentID)
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
		return nil, 0, errors.New("config item not found")
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, 0, errors.Errorf("download status %d != 200", resp.StatusCode)
	}

	return resp.Body, resp.ContentLength, nil
}

// Metadata cos file metadata
func (c *cosClient) Metadata(kt *kit.Kit, fileContentID string) (*ObjectMetadata, error) {
	node, err := repo.GenS3NodeFullPath(kt.BizID, fileContentID)
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
		return nil, errors.New("config item not found")
	}

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("download status %d != 200", resp.StatusCode)
	}

	// cos only have etag, not for validate
	metadata := &ObjectMetadata{
		ByteSize: 0,
		Sha256:   fileContentID,
	}

	return metadata, nil
}

// DownloadLink cos file download link
func (c *cosClient) DownloadLink(kt *kit.Kit, fileContentID string, fetchLimit uint32) (string, error) {
	return "", nil
}

// AsyncDownload cos
func (c *cosClient) AsyncDownload(kt *kit.Kit, fileContentID string) (string, error) {
	return "", nil
}

// AsyncDownloadStatus cos
func (c *cosClient) AsyncDownloadStatus(kt *kit.Kit, fileContentID string, taskID string) (bool, error) {
	return false, nil
}

// newCosProvider new cos provider
func newCosProvider(conf cc.S3Storage) (Provider, error) {
	host := fmt.Sprintf("%s://%s.%s", cosSchema, conf.BucketName, conf.Endpoint)

	// cos 鉴权签名
	transport := &cos.AuthorizationTransport{
		SecretID:  conf.AccessKeyID,
		SecretKey: conf.SecretAccessKey,
		Transport: tools.NewCurlLogTransport(defaultTransport),
	}

	p := &cosClient{
		host:   host,
		client: &http.Client{Transport: transport},
	}

	return p, nil
}
