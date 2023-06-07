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

	"github.com/bluele/gcache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/thirdparty/repo"
)

// S3Client s3 client struct
type S3Client struct {
	// repoCli s3 client.
	s3Cli *repo.ClientS3
	// s3CreatedRecords memory LRU cache used for re-create repo repository.
	s3CreatedRecords gcache.Cache
	// authorizer auth related operations.
	authorizer auth.Authorizer
}

// DownloadFile download file
func (s *S3Client) DownloadFile(w http.ResponseWriter, r *http.Request) {
	err := s.downloadFile(w, r)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
	}
}

func (s *S3Client) downloadFile(w http.ResponseWriter, r *http.Request) error {
	kt := kit.MustGetKit(r.Context())

	repoName := s.s3Cli.Config.S3.BucketName
	s3PathName, err := repo.GenRepoName(kt.BizID)
	if err != nil {
		return errors.Wrap(err, "generate S3 repository name failed")
	}

	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	fullPath, err := repo.GenS3NodeFullPath(s3PathName, sha256)
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

// UploadFile upload file
func (s S3Client) UploadFile(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())

	bizIDStr := chi.URLParam(r, "biz_id")
	bizID, err := strconv.ParseUint(bizIDStr, 10, 64)
	if err != nil {
		logs.Errorf("biz_id parse uint failed, err: %v, rid: %s", err, kt.Rid)
		fmt.Fprintf(w, errf.New(errf.InvalidParameter, err.Error()).Error())
		return
	}

	if bizID == 0 {
		fmt.Fprintf(w, errf.New(errf.InvalidParameter, "biz_id should > 0").Error())
		return
	}
	repoName := s.s3Cli.Config.S3.BucketName

	if record, err := s.s3CreatedRecords.Get(repoName); err != nil || record == nil {

		req := &repo.CreateRepoReq{
			Name:        repoName,
			Description: fmt.Sprintf("bscp %d business repository", bizID),
		}
		if err = s.s3Cli.CreateRepo(r.Context(), req); err != nil {
			logs.Errorf("create repository failed, err: %v, rid: %s", err, kt.Rid)
			fmt.Fprintf(w, errf.Error(err).Error())
			return
		}

		// set cache, to flag this biz repository already created.
		s.s3CreatedRecords.SetWithExpire(repoName, true, RepoRecordCacheExpiration)
	}
	s3pathName, err := repo.GenRepoName(uint32(bizID))
	if err != nil {
		logs.Errorf("generate s3 path name failed, err: %v, rid: %s", err, kt.Rid)
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	fullPath, err := repo.GenS3NodeFullPath(s3pathName, sha256)
	if err != nil {
		logs.Errorf("create S3 FullPath failed, err: %v, err")
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	_, err = s.s3Cli.Client.PutObject(r.Context(), repoName, fullPath, r.Body, r.ContentLength, minio.PutObjectOptions{})
	if err != nil {
		logs.Errorf("uploader S3 file failed, err: %v, err")
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	ok, _ := s.s3Cli.IsNodeExist(r.Context(), repoName, fullPath)
	if !ok {
		logs.Errorf("Failed to check artifact sha256 digest")
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	msg, _ := json.Marshal(ResponseBody{Code: 200, Message: "success"})
	w.Write(msg)
}

// FileMetadata get s3 head data
func (s S3Client) FileMetadata(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())

	config := cc.ApiServer().Repo

	bizID, _, err := GetBizIDAndAppID(nil, r)
	if err != nil {
		logs.Errorf("get biz_id and app_id from request failed, err: %v, rid: %s", err, kt.Rid)
		return
	}

	s3PathName, err := repo.GenRepoName(bizID)
	if err != nil {
		logs.Errorf("generate S3 repository name failed, err: %v, rid: %s", err, kt.Rid)
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	fullPath, err := repo.GenS3NodeFullPath(s3PathName, sha256)
	if err != nil {
		logs.Errorf("create S3 FullPath failed, err: %v, err")
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}

	fileMetadata, err := s.s3Cli.FileMetadataHead(kt.Ctx, config.S3.BucketName, fullPath)
	if err != nil {
		logs.Errorf("get file metadata information failed, err: %v, rid: %s", err)
		return
	}
	fileMetadata.Sha256 = sha256
	msg, _ := json.Marshal(fileMetadata)
	w.Write(msg)
}

// NewS3Service new s3 service
func NewS3Service(settings cc.Repository, authorizer auth.Authorizer) (FileApiType, error) {
	s3Client, err := repo.NewClientS3(&settings, metrics.Register())
	if err != nil {
		return nil, err
	}
	p := &S3Client{
		s3Cli:            s3Client,
		s3CreatedRecords: gcache.New(1000).EvictType(gcache.TYPE_LRU).Build(), // total size < 8k
		authorizer:       authorizer,
	}
	return p, nil
}
