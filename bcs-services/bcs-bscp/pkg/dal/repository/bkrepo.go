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
	"strings"

	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/metrics"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/thirdparty/repo"
)

// bkrepoAuthTransport 给请求增加 Authorization header
type bkrepoAuthTransport struct {
	Username  string
	Password  string
	Transport http.RoundTripper
}

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

// bkrepo s3 client struct
type bkrepo struct {
	// repoCli s3 client.
	client *http.Client
	cli    *repo.Client
	// authorizer auth related operations.
	authorizer  auth.Authorizer
	host        string
	project     string
	repoCreated map[string]struct{}
}

// DownloadFile download file
func (s *bkrepo) DownloadFile(w http.ResponseWriter, r *http.Request) {
	err := s.downloadFile(w, r)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
	}
}

// UploadFile upload file
func (s *bkrepo) uploadFile(r *http.Request) (*ObjectMetadata, error) {
	kt := kit.MustGetKit(r.Context())

	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	if len(sha256) != 64 {
		return nil, errors.New("not valid X-Bkapi-File-Content-Id in header")
	}

	return s.uploadFile2(kt, sha256, r.Body)
}

// UploadFile
func (s *bkrepo) UploadFile(w http.ResponseWriter, r *http.Request) {
	metadata, err := s.uploadFile(r)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}
	render.JSON(w, r, rest.OKRender(metadata))

}

func (s *bkrepo) uploadFile2(kt *kit.Kit, fileContentID string, body io.Reader) (*ObjectMetadata, error) {
	node, err := repo.GenNodePath(&repo.NodeOption{Project: s.project, BizID: kt.BizID, Sign: fileContentID})
	if err != nil {
		return nil, err
	}

	rawURL := fmt.Sprintf("%s%s", s.host, node)
	req, err := http.NewRequestWithContext(kt.Ctx, http.MethodPut, rawURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(repo.HeaderKeyOverwrite, "true")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Errorf("upload status %d != 200", resp.StatusCode)
	}

	uploadResp := new(repo.UploadResp)
	if err := json.NewDecoder(resp.Body).Decode(uploadResp); err != nil {
		return nil, errors.Wrap(err, "upload respones")
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

func (s *bkrepo) downloadFile(w http.ResponseWriter, r *http.Request) error {
	kt := kit.MustGetKit(r.Context())

	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	if len(sha256) != 64 {
		return errors.New("not valid X-Bkapi-File-Content-Id in header")
	}

	resp, contentLength, err := s.downloadFil2(kt, sha256)
	if err != nil {
		return err
	}
	defer resp.Close()

	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.Header().Set("Content-Type", "application/octet-stream; charset=UTF-8")
	_, err = io.Copy(w, resp)
	if err != nil {
		klog.ErrorS(err, "download file", "fileContentID", sha256)
	}

	return nil
}

func (s *bkrepo) downloadFil2(kt *kit.Kit, fileContentID string) (io.ReadCloser, int64, error) {
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

// FileMetadata get s3 head data
func (s *bkrepo) FileMetadata(w http.ResponseWriter, r *http.Request) {
	metadata, err := s.fileMetadata(w, r)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}
	render.JSON(w, r, rest.OKRender(metadata))

}

func (s *bkrepo) fileMetadata(w http.ResponseWriter, r *http.Request) (*ObjectMetadata, error) {
	kt := kit.MustGetKit(r.Context())

	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	if len(sha256) != 64 {
		return nil, errors.New("not valid X-Bkapi-File-Content-Id in header")
	}

	return s.fileMetadata2(kt, sha256)
}

func (s *bkrepo) fileMetadata2(kt *kit.Kit, fileContentID string) (*ObjectMetadata, error) {
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
	fmt.Println(resp.Header)

	// cos only have etag, not for validate
	metadata := &ObjectMetadata{
		ByteSize: 0,
		Sha256:   fileContentID,
	}

	return metadata, nil
}

// NewBKRepoService new s3 service
func NewBKRepoService(settings cc.Repository, authorizer auth.Authorizer) (FileApiType, error) {
	s, err := repo.NewClient(settings, metrics.Register())
	if err != nil {
		return nil, err
	}

	host := settings.BkRepo.Endpoints[0]

	p := &bkrepo{
		cli:        s,
		authorizer: authorizer,
		host:       host,
		project:    settings.BkRepo.Project,
	}

	transport := &bkrepoAuthTransport{
		Username: settings.BkRepo.Username,
		Password: settings.BkRepo.Password,
	}

	p.client = &http.Client{Transport: transport}

	return p, nil
}
