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

package service

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"k8s.io/klog/v2"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/dal/repository"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/rest"
)

// repoService is http handler for repo services.
type repoService struct {
	// authorizer auth related operations.
	authorizer auth.Authorizer
	provider   repository.Provider
}

// UploadFile upload to repo provider
func (s *repoService) UploadFile(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())

	sign, err := repository.GetFileSign(r)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}

	metadata, err := s.provider.Upload(kt, sign, r.Body)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}

	render.Render(w, r, rest.OKRender(metadata))
}

// DownloadFile download file from provider repo
func (s *repoService) DownloadFile(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())

	sign, err := repository.GetFileSign(r)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}

	body, contentLength, err := s.provider.Download(kt, sign)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}
	defer body.Close()

	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, body)
	if err != nil {
		klog.ErrorS(err, "download file", "sign", sign)
	}
}

// FileMetadata get repo head data
func (s *repoService) FileMetadata(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())

	sign, err := repository.GetFileSign(r)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}

	metadata, err := s.provider.Metadata(kt, sign)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}

	render.Render(w, r, rest.OKRender(metadata))
}

func newRepoService(settings cc.Repository, authorizer auth.Authorizer) (*repoService, error) {
	provider, err := repository.NewProvider(settings)
	if err != nil {
		return nil, err
	}

	repo := &repoService{
		authorizer: authorizer,
		provider:   provider,
	}

	return repo, nil
}
