package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bluele/gcache"
	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/repository"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	"bscp.io/pkg/runtime/gwparser"
	"bscp.io/pkg/thirdparty/repo"
)

func (cs S3Client) DownloadFile(w http.ResponseWriter, r *http.Request) {
	kt, err := gwparser.Parse(r.Context(), r.Header)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}

	authRes, needReturn := cs.authorize(kt, r)
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

	repoName, err := repo.GenS3Name(uint32(bizID))
	if err != nil {
		logs.Errorf("generate S3 repository name failed, err: %v, rid: %s", err, kt.Rid)
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	fullPath, err := repo.GenS3NodeFullPath(sha256)
	if err != nil {
		logs.Errorf("create S3 FullPath failed, err: %v, err")
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	reader, err := cs.s3Cli.Client.GetObject(r.Context(), repoName, fullPath, minio.GetObjectOptions{})
	if err != nil {
		logs.Errorf("download S3 file failed, err: %v, err")
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	io.Copy(w, reader)
}

func (cs S3Client) UploadFile(w http.ResponseWriter, r *http.Request) {
	kt, err := gwparser.Parse(r.Context(), r.Header)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}

	authRes, needReturn := cs.authorize(kt, r)
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
	repoName, err := repo.GenS3Name(uint32(bizID))
	if err != nil {
		logs.Errorf("generate s3 repository name failed, err: %v, rid: %s", err, kt.Rid)
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}

	if record, err := cs.s3CreatedRecords.Get(bizID); err != nil || record == nil {

		req := &repo.CreateRepoReq{
			Name:        repoName,
			Description: fmt.Sprintf("bscp %d business repository", bizID),
		}
		if err = cs.s3Cli.CreateRepo(r.Context(), req); err != nil {
			logs.Errorf("create repository failed, err: %v, rid: %s", err, kt.Rid)
			fmt.Fprintf(w, errf.Error(err).Error())
			return
		}

		// set cache, to flag this biz repository already created.
		cs.s3CreatedRecords.SetWithExpire(bizID, true, repoRecordCacheExpiration)
	}

	sha256 := strings.ToLower(r.Header.Get(constant.ContentIDHeaderKey))
	fullPath, err := repo.GenS3NodeFullPath(sha256)
	if err != nil {
		logs.Errorf("create S3 FullPath failed, err: %v, err")
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	_, err = cs.s3Cli.Client.PutObject(r.Context(), repoName, fullPath, r.Body, r.ContentLength, minio.PutObjectOptions{})
	if err != nil {
		logs.Errorf("uploader S3 file failed, err: %v, err")
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	ok, _ := cs.s3Cli.IsNodeExist(r.Context(), repoName, fullPath)
	if !ok {
		logs.Errorf("Failed to check artifact sha256 digest")
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}
	msg, _ := json.Marshal(ResponseBody{Code: 200, Message: "success"})
	w.Write(msg)
}

type ResponseBody struct {
	Code    int
	Message string
}

type S3Client struct {
	// repoCli s3 client.
	s3Cli *repo.ClientS3
	// s3CreatedRecords memory LRU cache used for re-create repo repository.
	s3CreatedRecords gcache.Cache
	// authorizer auth related operations.
	authorizer auth.Authorizer
}

// authorize the request, returns error response and if the response needs return.
func (cs S3Client) authorize(kt *kit.Kit, r *http.Request) (string, bool) {
	bizID, appID, err := getBizIDAndAppID(kt, r)
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

	resp := new(authResp)
	err = cs.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		respJson, _ := json.Marshal(resp)
		return string(respJson), true
	}

	return "", false
}

func NewS3Service(settings cc.Repository, authorizer auth.Authorizer) (repository.FileApiType, error) {
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
