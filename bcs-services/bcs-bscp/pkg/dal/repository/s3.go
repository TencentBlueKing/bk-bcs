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
	"strings"

	"github.com/go-chi/render"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	cos "github.com/tencentyun/cos-go-sdk-v5"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/thirdparty/repo"
)

// s3Client s3 client struct
type s3Client struct {
	// repoCli s3 client.
	client *http.Client
	s3Cli  *repo.ClientS3
	// authorizer auth related operations.
	authorizer auth.Authorizer
	host       string
}

// DownloadFile download file
func (s *s3Client) DownloadFile(w http.ResponseWriter, r *http.Request) {
	err := s.downloadFile(w, r)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
	}
}

// UploadFile upload file
func (s *s3Client) uploadFile(r *http.Request) (*ObjectMetadata, error) {
	kt := kit.MustGetKit(r.Context())

	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	if len(sha256) != 64 {
		return nil, errors.New("not valid X-Bkapi-File-Content-Id in header")
	}

	return s.uploadFile2(kt, sha256, r.Body)
}

// UploadFile
func (s *s3Client) UploadFile(w http.ResponseWriter, r *http.Request) {
	metadata, err := s.uploadFile(r)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}
	render.JSON(w, r, rest.OKRender(metadata))

}

func (s *s3Client) uploadFile2(kt *kit.Kit, fileContentID string, body io.Reader) (*ObjectMetadata, error) {
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

func (s *s3Client) downloadFile(w http.ResponseWriter, r *http.Request) error {
	kt := kit.MustGetKit(r.Context())

	repoName := s.s3Cli.Config.S3.BucketName
	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	fullPath, err := repo.GenS3NodeFullPath(kt.BizID, sha256)
	if err != nil {
		return errors.Wrap(err, "create S3 FullPath failed")
	}

	reader, err := s.s3Cli.Client.GetObject(r.Context(), repoName, fullPath, minio.GetObjectOptions{})
	if err != nil {
		return errors.Wrap(err, "download S3 file failed")
	}
	if _, err := reader.Stat(); err != nil {
		return errors.Wrap(err, "get file stat failed")
	}

	defer reader.Close()
	if _, err := io.Copy(w, reader); err != nil {
		logs.Errorf("download file failed when io.Copy, err: %v, rid: %s", err, kt.Rid)
	}

	return nil
}

// FileMetadata get s3 head data
func (s *s3Client) FileMetadata(w http.ResponseWriter, r *http.Request) {
	metadata, err := s.fileMetadata(w, r)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}
	render.JSON(w, r, rest.OKRender(metadata))

}

func (s *s3Client) fileMetadata(w http.ResponseWriter, r *http.Request) (*ObjectMetadata, error) {
	kt := kit.MustGetKit(r.Context())

	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	if len(sha256) != 64 {
		return nil, errors.New("not valid X-Bkapi-File-Content-Id in header")
	}

	config := cc.ApiServer().Repo

	fullPath, err := repo.GenS3NodeFullPath(kt.BizID, sha256)
	if err != nil {
		return nil, errors.Wrap(err, "create S3 FullPath failed")
	}

	fileMetadata, err := s.s3Cli.FileMetadataHead(kt.Ctx, config.S3.BucketName, fullPath)
	if err != nil {
		return nil, errors.Wrap(err, "get file metadata failed")
	}

	// cos only have etag, not for validate
	metadata := &ObjectMetadata{
		ByteSize: fileMetadata.ByteSize,
		Sha256:   fileMetadata.Sha256,
	}

	return metadata, nil
}

// NewS3Service new s3 service
func NewS3Service(settings cc.Repository, authorizer auth.Authorizer) (FileApiType, error) {
	s, err := repo.NewClientS3(&settings, metrics.Register())
	if err != nil {
		return nil, err
	}

	host := fmt.Sprintf("https://%s.%s", settings.S3.BucketName, settings.S3.Endpoint)

	p := &s3Client{
		s3Cli:      s,
		authorizer: authorizer,
		host:       host,
	}

	transport := &cos.AuthorizationTransport{
		SecretID:  settings.S3.AccessKeyID,
		SecretKey: settings.S3.SecretAccessKey,
	}

	p.client = &http.Client{Transport: transport}

	return p, nil
}
