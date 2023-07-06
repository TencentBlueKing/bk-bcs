package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bluele/gcache"
	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	pbas "bscp.io/pkg/protocol/auth-server"
	"bscp.io/pkg/thirdparty/repo"
)

const (
	// repoRecordCacheExpiration repo created record cache expiration.
	RepoRecordCacheExpiration = time.Hour
)

// FileApiType file api type
type FileApiType interface {
	DownloadFile(w http.ResponseWriter, r *http.Request)
	FileMetadata(w http.ResponseWriter, r *http.Request)
	UploadFile(w http.ResponseWriter, r *http.Request)
}

// DownloadFile download file
func (s S3Client) DownloadFile(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())

	authRes, needReturn := s.authorize(kt, r)
	if needReturn {
		fmt.Fprintf(w, authRes)
		return
	}

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
	s3PathName, err := repo.GenRepoName(uint32(bizID))
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
	reader, err := s.s3Cli.Client.GetObject(r.Context(), repoName, fullPath, minio.GetObjectOptions{})
	if err != nil {
		logs.Errorf("download S3 file failed, err: %v, err")
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	io.Copy(w, reader)
}

// UploadFile upload file
func (s S3Client) UploadFile(w http.ResponseWriter, r *http.Request) {
	kt := kit.MustGetKit(r.Context())

	authRes, needReturn := s.authorize(kt, r)
	if needReturn {
		fmt.Fprintf(w, authRes)
		return
	}

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

	authRes, needReturn := s.authorize(kt, r)
	if needReturn {
		fmt.Fprintf(w, authRes)
		return
	}
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

type ResponseBody struct {
	Code    int
	Message string
}

// S3Client s3 client struct
type S3Client struct {
	// repoCli s3 client.
	s3Cli *repo.ClientS3
	// s3CreatedRecords memory LRU cache used for re-create repo repository.
	s3CreatedRecords gcache.Cache
	// authorizer auth related operations.
	authorizer auth.Authorizer
}

// authorize the request, returns error response and if the response needs return.
func (s S3Client) authorize(kt *kit.Kit, r *http.Request) (string, bool) {
	bizID, appID, err := GetBizIDAndAppID(kt, r)
	if err != nil {
		logs.Errorf("get biz_id and app_id from request failed, err: %v, rid: %s", err, kt.Rid)
		return errf.New(errf.InvalidParameter, err.Error()).Error(), true
	}

	var authRes *meta.ResourceAttribute
	switch r.Method {
	case http.MethodPut:
		authRes = &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Content, Action: meta.Upload,
			ResourceID: appID}, BizID: bizID}
	case http.MethodGet:
		authRes = &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Content, Action: meta.Download,
			ResourceID: appID}, BizID: bizID}
	}

	resp := new(AuthResp)
	err = s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		respJson, _ := json.Marshal(resp)
		return string(respJson), true
	}

	return "", false
}

// AuthResp http response with need apply permission.
type AuthResp struct {
	Code       int32               `json:"code"`
	Message    string              `json:"message"`
	Permission *pbas.IamPermission `json:"permission,omitempty"`
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

// GetBizIDAndAppID get biz_id and app_id from req path.
func GetBizIDAndAppID(kt *kit.Kit, req *http.Request) (uint32, uint32, error) {
	bizIDStr := chi.URLParam(req, "biz_id")
	bizID, err := strconv.ParseUint(bizIDStr, 10, 64)
	if err != nil {
		logs.Errorf("biz id parse uint failed, err: %v, rid: %s", err, kt.Rid)
		return 0, 0, err
	}

	if bizID == 0 {
		return 0, 0, errf.New(errf.InvalidParameter, "biz_id should > 0")
	}

	appIDStr := chi.URLParam(req, "app_id")
	appID, err := strconv.ParseUint(appIDStr, 10, 64)
	if err != nil {
		logs.Errorf("app id parse uint failed, err: %v, rid: %s", err, kt.Rid)
		return 0, 0, err
	}

	if appID == 0 {
		return 0, 0, errf.New(errf.InvalidParameter, "app_id should > 0")
	}

	return uint32(bizID), uint32(appID), nil
}
