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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	cos "github.com/tencentyun/cos-go-sdk-v5"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/repo"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

const (
	// cosSchema 请求使用 https 协议
	cosSchema = "https"
)

// cosClient tencentcloud cos client struct
type cosClient struct {
	conf        *cc.S3Storage
	host        string
	client      *http.Client
	innerClient *cos.Client
}

// SyncManager implements HAEnhancer interface
func (c *cosClient) SyncManager() *SyncManager {
	return nil
}

// Upload uploads file to cos
func (c *cosClient) Upload(kt *kit.Kit, sign string, body io.Reader) (*ObjectMetadata, error) {
	node, err := repo.GenS3NodeFullPath(kt.BizID, sign)
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

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("upload status %d != 200", resp.StatusCode)
	}

	// cos return not have metadata
	metadata := &ObjectMetadata{
		Sha256: sign,
	}

	return metadata, nil
}

// Download downloads file from cos
func (c *cosClient) Download(kt *kit.Kit, sign string) (io.ReadCloser, int64, error) {
	node, err := repo.GenS3NodeFullPath(kt.BizID, sign)
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

// Metadata cos file metadata
func (c *cosClient) Metadata(kt *kit.Kit, sign string) (*ObjectMetadata, error) {
	node, err := repo.GenS3NodeFullPath(kt.BizID, sign)
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

	md5 := strings.Trim(resp.Header.Get("Etag"), "\"")

	// cos only have etag, not for validate
	metadata := &ObjectMetadata{
		ByteSize: resp.ContentLength,
		Sha256:   sign,
		Md5:      md5,
	}

	return metadata, nil
}

func (c *cosClient) InitMultipartUpload(kt *kit.Kit, sign string) (string, error) {
	// TODO: InitMultipartUpload not implemented in cos
	return "", fmt.Errorf("not implemented")
}

func (c *cosClient) MultipartUpload(kt *kit.Kit, sign string, uploadID string, partNum uint32, body io.Reader) error {
	// TODO: MultipartUpload not implemented in cos
	return fmt.Errorf("not implemented")
}

func (c *cosClient) CompleteMultipartUpload(kt *kit.Kit, sign string, uploadID string) (*ObjectMetadata, error) {
	// TODO: CompleteMultipartUpload not implemented in cos
	return nil, fmt.Errorf("not implemented")
}

// URIDecorator ..
func (c *cosClient) URIDecorator(bizID uint32) DecoratorInter {
	return newUriDecoratorInter(bizID)
}

// DownloadLink cos file download link
func (c *cosClient) DownloadLink(kt *kit.Kit, sign string, fetchLimit uint32) ([]string, error) {
	node, err := repo.GenS3NodeFullPath(kt.BizID, sign)
	if err != nil {
		return nil, err
	}

	opt := &cos.PresignedURLOptions{
		Query:  &url.Values{},
		Header: &http.Header{},
	}

	// cos sdk 已经包含根目录, 需要去重
	node = strings.TrimLeft(node, "/")
	u, err := c.innerClient.Object.GetPresignedURL(kt.Ctx, http.MethodGet, node, c.conf.AccessKeyID,
		c.conf.SecretAccessKey, time.Hour, opt)
	if err != nil {
		return nil, err
	}

	return []string{u.String()}, nil
}

// AsyncDownload cos
func (c *cosClient) AsyncDownload(kt *kit.Kit, sign string) (string, error) {
	return "", errNotImplemented
}

// AsyncDownloadStatus cos
func (c *cosClient) AsyncDownloadStatus(kt *kit.Kit, sign string, taskID string) (bool, error) {
	return false, errNotImplemented
}

// newCosClient new cos client
func newCosClient(conf cc.S3Storage) (BaseProvider, error) {
	host := fmt.Sprintf("%s://%s.%s", cosSchema, conf.BucketName, conf.Endpoint)

	// cos 鉴权签名
	transport := &cos.AuthorizationTransport{
		SecretID:  conf.AccessKeyID,
		SecretKey: conf.SecretAccessKey,
		Transport: tools.NewCurlLogTransport(defaultTransport),
	}

	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	p := &cosClient{
		host:        host,
		conf:        &conf,
		client:      &http.Client{Transport: transport},
		innerClient: cos.NewClient(&cos.BaseURL{BucketURL: u}, nil),
	}

	return p, nil
}

// newCosProvider new cos provider
func newCosProvider(repo cc.BaseRepo, redis cc.RedisCluster) (Provider, error) {
	p, err := newCosClient(repo.S3)
	if err != nil {
		return nil, err
	}

	var c VariableCacher
	c, err = newVariableCacher(redis, p)
	if err != nil {
		return nil, err
	}

	return &repoProvider{
		BaseProvider:   p,
		HAEnhancer:     p.(*cosClient),
		VariableCacher: c,
	}, nil
}
