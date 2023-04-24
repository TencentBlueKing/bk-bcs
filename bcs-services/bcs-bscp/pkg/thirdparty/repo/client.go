/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/rest/client"
	"bscp.io/pkg/tools"

	"github.com/prometheus/client_golang/prometheus"
)

// Client is repo client.
type Client struct {
	config cc.Repository
	// http client instance
	client rest.ClientInterface
	// http header info
	basicHeader http.Header
}

// NewClient new repo client.
func NewClient(repoSetting cc.Repository, reg prometheus.Registerer) (*Client, error) {
	tls := &tools.TLSConfig{
		InsecureSkipVerify: repoSetting.BkRepo.TLS.InsecureSkipVerify,
		CertFile:           repoSetting.BkRepo.TLS.CertFile,
		KeyFile:            repoSetting.BkRepo.TLS.KeyFile,
		CAFile:             repoSetting.BkRepo.TLS.CAFile,
		Password:           repoSetting.BkRepo.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &repoDiscovery{
			servers: repoSetting.BkRepo.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("Accept", "application/json")
	header.Set("Authorization", fmt.Sprintf("Platform %s", repoSetting.BkRepo.Token))
	header.Set(HeaderKeyUID, repoSetting.BkRepo.User)

	return &Client{
		config:      repoSetting,
		client:      rest.NewClient(c, "/"),
		basicHeader: header,
	}, nil
}

// ProjectID return repo project id.
func (c *Client) ProjectID() string {
	return c.config.BkRepo.Project
}

// IsProjectExist judge repo bscp project already exist.
func (c *Client) IsProjectExist(ctx context.Context) error {
	resp := c.client.Get().
		WithContext(ctx).
		SubResourcef("/repository/api/project/exist/%s", c.config.BkRepo.Project).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return resp.Err
	}

	// repo uses StatusBadRequest to mark the failure of the request, so StatusBadRequest
	// needs special handling to read out the error information in the body message.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	respBody := new(rest.BaseResp)
	if err := resp.Into(respBody); err != nil {
		return err
	}

	if respBody.Code != 0 {
		return fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}

	return nil
}

// CreateRepo create new repository in repo.
func (c *Client) CreateRepo(ctx context.Context, req *CreateRepoReq) error {
	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/repository/api/repo/create").
		WithHeaders(c.basicHeader).
		Body(req).
		Do()
	if resp.Err != nil {
		return resp.Err
	}

	// repo uses StatusBadRequest to mark the failure of the request, so StatusBadRequest
	// needs special handling to read out the error information in the body message.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	respBody := new(rest.BaseResp)
	if err := resp.Into(respBody); err != nil {
		return err
	}

	// if repo already exist, ignore this error.
	if respBody.Code != 0 && respBody.Code != errCodeRepoAlreadyExist {
		return fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}

	return nil
}

// DeleteRepo delete repository in repo. param force: whether to force deletion.
// If false, the warehouse cannot be deleted when there are files in the warehouse
func (c *Client) DeleteRepo(ctx context.Context, bizID uint32, forced bool) error {
	repoName, err := GenRepoName(bizID)
	if err != nil {
		return err
	}

	resp := c.client.Delete().
		WithContext(ctx).
		SubResourcef("/repository/api/repo/delete/%s/%s", c.config.BkRepo.Project, repoName).
		WithParam("forced", strconv.FormatBool(forced)).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return resp.Err
	}

	// repo uses StatusBadRequest to mark the failure of the request, so StatusBadRequest
	// needs special handling to read out the error information in the body message.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	respBody := new(rest.BaseResp)
	if err := resp.Into(respBody); err != nil {
		return err
	}

	if respBody.Code != 0 {
		return fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}

	return nil
}

// IsNodeExist judge repo node already exist.
func (c *Client) IsNodeExist(ctx context.Context, nodePath string) (bool, error) {
	resp := c.client.Head().
		WithContext(ctx).
		SubResourcef(nodePath).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return false, resp.Err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return false, fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return true, nil
}

// DeleteNode delete node.
func (c *Client) DeleteNode(ctx context.Context, nodePath string) error {
	resp := c.client.Delete().
		WithContext(ctx).
		SubResourcef(nodePath).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return resp.Err
	}

	// repo uses StatusBadRequest to mark the failure of the request, so StatusBadRequest
	// needs special handling to read out the error information in the body message.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	respBody := new(rest.BaseResp)
	if err := resp.Into(respBody); err != nil {
		return err
	}

	if respBody.Code != 0 {
		return fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}
	return nil
}

// QueryMetaResp query metadata repo return response.
type QueryMetaResp struct {
	Code    uint32            `json:"code"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data"`
}

// QueryMetadata query node metadata info. If node not exist, return data is {}.
func (c *Client) QueryMetadata(ctx context.Context, opt *NodeOption) (map[string]string, error) {
	repoName, err := GenRepoName(opt.BizID)
	if err != nil {
		return nil, err
	}

	fullPath, err := GenNodeFullPath(opt.Sign)
	if err != nil {
		return nil, err
	}

	resp := c.client.Get().
		WithContext(ctx).
		SubResourcef("/repository/api/metadata/%s/%s%s", opt.Project, repoName, fullPath).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return nil, resp.Err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	respBody := new(QueryMetaResp)
	if err := resp.Into(respBody); err != nil {
		return nil, err
	}

	if respBody.Code != 0 {
		return nil, fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}

	return respBody.Data, nil
}

// FileMetadataHead get head data.
func (c *Client) FileMetadataHead(ctx context.Context, nodePath string) (*FileMetadataValue, error) {
	resp := c.client.Head().
		WithContext(ctx).
		SubResourcef(nodePath).
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return nil, resp.Err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code: %d", resp.StatusCode)
	}
	fileSize := resp.Header.Get("Content-Length")
	size, _ := strconv.Atoi(fileSize)
	message := &FileMetadataValue{
		ByteSize: int64(size),
		Sha256:   resp.Header.Get("X-Checksum-Sha256"),
	}

	return message, nil
}

type FileMetadataValue struct {
	ByteSize int64  `json:"byte_size"`
	Sha256   string `json:"sha256"`
}

// GenerateTempDownloadURL generate temp download url.
func (c *Client) GenerateTempDownloadURL(ctx context.Context, req *GenerateTempDownloadURLReq) (string, error) {
	resp := c.client.Post().
		WithContext(ctx).
		SubResourcef("/generic/temporary/url/create").
		WithHeaders(c.basicHeader).
		Do()
	if resp.Err != nil {
		return "", resp.Err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("response status code: %d", resp.StatusCode)
	}
	respBody := new(GenerateTempDownloadURLResp)
	if err := resp.Into(respBody); err != nil {
		return "", err
	}

	if respBody.Code != 0 {
		return "", fmt.Errorf("code: %d, message: %s", respBody.Code, respBody.Message)
	}

	if len(respBody.Data) != 1 {
		return "", fmt.Errorf("invalid response data")
	}

	return respBody.Data[0].URL, nil
}

// Client is s3 client.
type ClientS3 struct {
	Config *cc.Repository
	// http client instance
	Client *minio.Client
}

// NewClient new s3 client.
func NewClientS3(repoSetting *cc.Repository, reg prometheus.Registerer) (*ClientS3, error) {
	minioClient, err := minio.New(repoSetting.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(repoSetting.S3.AccessKeyID, repoSetting.S3.SecretAccessKey, ""),
		Secure: repoSetting.S3.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		MetricOpts: client.MetricOption{Register: reg},
	}
	if c.MetricOpts.Register != nil {

		var buckets []float64
		if len(c.MetricOpts.DurationBuckets) == 0 {
			// set default buckets
			buckets = []float64{10, 30, 50, 70, 100, 200, 300, 400, 500, 1000, 2000, 5000}
		} else {
			// use user defined buckets
			buckets = c.MetricOpts.DurationBuckets
		}

		requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "bscp_http_requests_duration_millisecond",
			Help:    "third party api request duration millisecond.",
			Buckets: buckets,
		}, []string{"handler", "status_code", "dimension"})

		if err := c.MetricOpts.Register.Register(requestDuration); err != nil {
			if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
				requestDuration = are.ExistingCollector.(*prometheus.HistogramVec)
			} else {
				panic(err)
			}
		}
	}
	return &ClientS3{
		Config: repoSetting,
		Client: minioClient,
	}, nil
}

// CreateRepo create new repository in s3.
func (c *ClientS3) CreateRepo(ctx context.Context, s3 *CreateRepoReq) error {
	found, _ := c.Client.BucketExists(context.Background(), s3.Name)

	if found {
		return nil
	}
	if err := c.Client.MakeBucket(ctx, s3.Name, minio.MakeBucketOptions{}); err != nil {
		return err
	}
	for !found {
		found, _ = c.Client.BucketExists(context.Background(), s3.Name)
	}
	return nil
}

// DeleteRepo delete repository in repo. param force: whether to force deletion.
// If false, the warehouse cannot be deleted when there are files in the warehouse
func (c *ClientS3) DeleteRepo(ctx context.Context, bizID uint32, forced bool) error {
	err := c.Client.RemoveBucket(ctx, c.Config.S3.BucketName)
	if err != nil {
		return err
	}
	return nil
}

// IsNodeExist judge repo node already exist.
func (c *ClientS3) IsNodeExist(ctx context.Context, bucketName, nodePath string) (bool, error) {
	_, err := c.Client.StatObject(ctx, bucketName, nodePath, minio.StatObjectOptions{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// FileMetadataHead get head data
func (c *ClientS3) FileMetadataHead(ctx context.Context, bucketName, nodePath string) (*FileMetadataValue, error) {

	resp, err := c.Client.StatObject(ctx, bucketName, nodePath, minio.StatObjectOptions{Checksum: true})
	if err != nil {
		return nil, err
	}
	fileSize := resp.Size
	message := &FileMetadataValue{
		ByteSize: fileSize,
	}

	return message, nil
}

// DeleteNode delete node.
func (c *ClientS3) DeleteNode(ctx context.Context, bucketName, nodePath string) error {
	err := c.Client.RemoveObject(ctx, bucketName, nodePath, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

// QueryMetadata query node metadata info. If node not exist, return data is {}.
func (c *ClientS3) QueryMetadata(ctx context.Context, opt *NodeOption) (map[string]string, error) {
	bucketName, err := GenRepoName(uint32(opt.BizID))
	state, err := c.Client.StatObject(ctx, bucketName, opt.Sign, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	var dataMeta map[string]string
	stateJson, _ := json.Marshal(state)
	json.Unmarshal(stateJson, dataMeta)
	return dataMeta, nil
}
