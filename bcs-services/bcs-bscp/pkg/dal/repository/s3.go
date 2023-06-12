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

	"github.com/pkg/errors"
	cos "github.com/tencentyun/cos-go-sdk-v5"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/thirdparty/repo"
)

// s3Client s3 client struct
type s3Client struct {
	client *http.Client
	host   string
}

// Upload
func (s *s3Client) Upload(kt *kit.Kit, fileContentID string, body io.Reader) (*ObjectMetadata, error) {
	node, err := repo.GenS3NodeFullPath(kt.BizID, fileContentID)
	if err != nil {
		return nil, err
	}

	rawURL := fmt.Sprintf("%s/%s", s.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodPut, rawURL, body)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
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

// Download
func (s *s3Client) Download(kt *kit.Kit, fileContentID string) (io.ReadCloser, int64, error) {
	node, err := repo.GenS3NodeFullPath(kt.BizID, fileContentID)
	if err != nil {
		return nil, 0, err
	}

	rawURL := fmt.Sprintf("%s/%s", s.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, 0, errors.Errorf("download status %d != 200", resp.StatusCode)
	}

	return resp.Body, resp.ContentLength, nil
}

// Metadata
func (s *s3Client) Metadata(kt *kit.Kit, fileContentID string) (*ObjectMetadata, error) {
	node, err := repo.GenS3NodeFullPath(kt.BizID, fileContentID)
	if err != nil {
		return nil, err
	}

	rawURL := fmt.Sprintf("%s/%s", s.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

// NewS3Service new s3 service
func NewS3Service(conf cc.S3Storage) (Provider, error) {
	host := fmt.Sprintf("https://%s.%s", conf.BucketName, conf.Endpoint)

	p := &s3Client{
		host: host,
	}

	transport := &cos.AuthorizationTransport{
		SecretID:  conf.AccessKeyID,
		SecretKey: conf.SecretAccessKey,
	}

	p.client = &http.Client{Transport: transport}

	return p, nil
}
