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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	pbas "bscp.io/pkg/protocol/auth-server"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/runtime/gwparser"
	"bscp.io/pkg/thirdparty/repo"

	"github.com/bluele/gcache"
	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"
)

const (
	// defaultWriteBufferSize is default write buffer size, 4KB.
	defaultWriteBufferSize = 4 << 10

	// defaultReadBufferSize is default read buffer size, 4KB.
	defaultReadBufferSize = 4 << 10

	// repoRecordCacheExpiration repo created record cache expiration.
	repoRecordCacheExpiration = time.Hour
)

// repoProxy is http reverse proxy for bkrepo.
type repoProxy struct {
	// proxy repo http reverse proxy.
	proxy *httputil.ReverseProxy
	// repoCli repo client.
	repoCli *repo.Client
	// repoCreatedRecords memory LRU cache used for re-create repo repository.
	repoCreatedRecords gcache.Cache
	// authorizer auth related operations.
	authorizer auth.Authorizer
}

// ServeHTTP handle request.
func (p repoProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	kt, err := gwparser.Parse(r.Context(), r.Header)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, errf.Error(err).Error())
		return
	}

	authRes, needReturn := p.authorize(kt, r)
	if needReturn {
		fmt.Fprintf(w, authRes)
		return
	}

	// if request is upload file request, need to judge biz repo repository if created.
	// if not created, need to create.
	if r.Method == http.MethodPut {
		// parse biz_id.
		bizIDStr := mux.Vars(r)["biz_id"]
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

		if record, err := p.repoCreatedRecords.Get(bizID); err != nil || record == nil {
			repoName, err := repo.GenRepoName(uint32(bizID))
			if err != nil {
				logs.Errorf("generate repository name failed, err: %v, rid: %s", err, kt.Rid)
				fmt.Fprintf(w, errf.Error(err).Error())
				return
			}

			req := &repo.CreateRepoReq{
				ProjectID:     cc.ApiServer().Repo.Project,
				Name:          repoName,
				Type:          repo.RepositoryType,
				Category:      repo.CategoryType,
				Configuration: repo.Configuration{Type: repo.RepositoryCfgType},
				Description:   fmt.Sprintf("bscp %d business repository", bizID),
			}
			if err = p.repoCli.CreateRepo(r.Context(), req); err != nil {
				logs.Errorf("create repository failed, err: %v, rid: %s", err, kt.Rid)
				fmt.Fprintf(w, errf.Error(err).Error())
				return
			}

			// set cache, to flag this biz repository already created.
			p.repoCreatedRecords.SetWithExpire(bizID, true, repoRecordCacheExpiration)
		}
	}

	p.proxy.ServeHTTP(w, r)
}

// authorize the request, returns error response and if the response needs return.
func (p repoProxy) authorize(kt *kit.Kit, r *http.Request) (string, bool) {
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
	err = p.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		respJson, _ := json.Marshal(resp)
		return string(respJson), true
	}

	return "", false
}

// authResp http response with need apply permission.
type authResp struct {
	Code       int32               `json:"code"`
	Message    string              `json:"message"`
	Permission *pbas.IamPermission `json:"permission,omitempty"`
}

// newRepoProxy creates a new ReverseProxy for repo.
func newRepoProxy(authorizer auth.Authorizer) (*repoProxy, error) {
	settings := cc.ApiServer().Repo
	repoCli, err := repo.NewClient(&settings, metrics.Register())
	if err != nil {
		return nil, err
	}

	p := &repoProxy{
		proxy: &httputil.ReverseProxy{
			// Director must be a function which modifies the request into a new Request
			// to be sent using Transport. Its response is then copied back to the original
			// client unmodified. Director must not access the provided Request after returning.
			Director: newRepoDirector(repoCli),

			// The transport used to perform proxy requests. If nil,
			// http.DefaultTransport is used.
			Transport: &http.Transport{
				Proxy:               http.ProxyFromEnvironment,
				Dial:                (&net.Dialer{Timeout: 10 * time.Second}).Dial,
				MaxConnsPerHost:     200,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     time.Minute,
				WriteBufferSize:     defaultWriteBufferSize,
				ReadBufferSize:      defaultReadBufferSize,
			},

			// Modify the response returned by the product library and convert it to the bscp response body
			ModifyResponse: modifyResponse,
		},
		repoCli:            repoCli,
		repoCreatedRecords: gcache.New(1000).EvictType(gcache.TYPE_LRU).Build(), // total size < 8k
		authorizer:         authorizer,
	}

	return p, nil
}

// modifyResponse modify the response returned by the product library and convert it to the bscp response body.
func modifyResponse(res *http.Response) error {
	switch res.Request.Method {
	case http.MethodPut:
		return modifyUploadResp(res)

	case http.MethodGet:
		return modifyDownloadResp(res)

	default:
		return fmt.Errorf("unknown request method %s", res.Request.Method)
	}
}

// modifyDownloadResp modify download file api response to convert bscp response body.
func modifyDownloadResp(res *http.Response) error {
	rid := res.Request.Header.Get(constant.RidKey)

	switch res.StatusCode {
	case http.StatusOK:
		return nil

	case http.StatusNotFound:
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		code := gjson.GetBytes(body, "code").Int()
		msg := gjson.GetBytes(body, "message").String()

		if code == repo.ErrCodeNodeNotExist {
			res.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(errf.New(errf.RecordNotFound, ""+
				"file content not found").Error())))

			return nil
		}

		logs.Errorf("repo proxy download file failed, code: %d, msg: %s, rid: %s", code, msg, rid)
		res.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(errf.New(errf.Unknown, "repo proxy "+
			"download file failed, state code: 404").Error())))

		return nil

	default:
		logs.Errorf("repo proxy download file failed, code: %d, msg: %s, rid: %s", res.StatusCode, res.Status, rid)
		res.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(errf.New(errf.Unknown, fmt.Sprintf(
			"repo proxy download file failed, state code: %d", res.StatusCode)).Error())))
	}

	return nil
}

// modifyUploadResp bscp needs to log the repo upload interface response of the agent and
// convert it to the response body of bscp to avoid disclosure of internal repo configuration information.
func modifyUploadResp(res *http.Response) error {
	rid := res.Request.Header.Get(constant.RidKey)

	switch res.StatusCode {
	case http.StatusOK:
		if err := successResp(res, rid); err != nil {
			return err
		}
		return nil

	case http.StatusBadRequest:
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		code := gjson.GetBytes(body, "code").Int()
		msg := gjson.GetBytes(body, "message").String()

		// repeatedly create node(uploading files) is a successful response to bscp.
		if code != repo.ErrCodeNodeAlreadyExist {
			res.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(errf.New(errf.Unknown, fmt.Sprintf("repo proxy "+
				"upload file failed, code: %d, msg: %s", code, msg)).Error())))

		} else {
			res.StatusCode = http.StatusOK
			res.Status = "OK"

			if err := successResp(res, rid); err != nil {
				return err
			}
		}
		return nil

	default:
		logs.Errorf("repo proxy upload file failed, code: %d, msg: %s, rid: %s",
			res.StatusCode, res.Status, rid)
		res.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(errf.New(errf.Unknown, fmt.Sprintf(
			"repo proxy upload file failed, state code: %d", res.StatusCode)).Error())))
	}

	return nil
}

func successResp(res *http.Response, rid string) error {
	logs.Infof("repo proxy upload file success, rid: %s", rid)
	resp := &rest.BaseResp{
		Code: errf.OK,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	res.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	res.Header.Set("Content-Length", strconv.Itoa(len(b)))
	return nil
}

// newRepoDirector returns a director for repo.
func newRepoDirector(cli *repo.Client) func(req *http.Request) {
	return func(req *http.Request) {
		config := cc.ApiServer().Repo
		kt, err := gwparser.Parse(req.Context(), req.Header)
		if err != nil {
			logs.Errorf("gateway parser failed, err: %v, rid: %s", err, kt.Rid)
			return
		}

		addr, err := config.OneEndpoint()
		if err != nil {
			logs.Errorf("get repo address failed, err: %v, rid: %s", err, kt.Rid)
			return
		}

		// set scheme(http) and addr.
		elmHost := strings.Split(addr, "://")
		if len(elmHost) != 2 {
			logs.Errorf("repo address format does not conform to the regulations, addr: %s, rid: %s", addr, kt.Rid)
			return
		}
		req.URL.Scheme = elmHost[0]
		req.Host = elmHost[1]
		req.URL.Host = elmHost[1]

		bizID, appID, err := getBizIDAndAppID(kt, req)
		if err != nil {
			logs.Errorf("get biz_id and app_id from request failed, err: %v, rid: %s", err, kt.Rid)
			return
		}

		sha256 := strings.ToLower(req.Header.Get(constant.ContentIDHeaderKey))
		opt := &repo.NodeOption{
			Project: config.Project,
			BizID:   bizID,
			Sign:    sha256,
		}
		req.URL.Path, err = repo.GenNodePath(opt)
		if err != nil {
			logs.Errorf("generate node path failed, err: %v, rid: %s", err, kt.Rid)
			return
		}

		// set rid, this is the rid for internal positioning requests.
		// this field is not supported by repo and will not be used.
		req.Header.Set(constant.RidKey, kt.Rid)

		// set repo header.
		req.Header.Set("Authorization", "Platform "+config.Token)
		req.Header.Set(repo.HeaderKeyUID, config.User)
		req.Header.Set(repo.HeaderKeySHA256, sha256)

		// if it is an upload request, you need to set the upload node metadata.
		if req.Method == http.MethodPut {
			metadata, err := getNodeMetadata(kt, cli, opt, appID)
			if err != nil {
				logs.Errorf("get node metadata failed, err: %v, rid: %s", err, kt.Rid)
				return
			}

			req.Header.Set(repo.HeaderKeyMETA, metadata)
			// the contents of the files under the same business may be duplicated,
			// and the metadata information needs to be written by overwriting.
			req.Header.Set(repo.HeaderKeyOverwrite, "true")
		}
	}
}

// getNodeMetadata If the node already exists, this appID will be added to the metadata of the current node.
// If not exist, will create new metadata with this bizID and appID.
func getNodeMetadata(kt *kit.Kit, cli *repo.Client, opt *repo.NodeOption, appID uint32) (string, error) {
	metadata, err := cli.QueryMetadata(kt.Ctx, opt)
	if err != nil {
		return "", err
	}

	if len(metadata) == 0 {
		meta := repo.NodeMeta{
			BizID: opt.BizID,
			AppID: []uint32{appID},
		}

		return meta.String()
	}

	// validate already node metadata.
	bizID, exist := metadata["biz_id"]
	if !exist {
		return "", errors.New("node metadata not has biz id")
	}

	if bizID != strconv.Itoa(int(opt.BizID)) {
		return "", fmt.Errorf("node metadata %s biz id is different from the request %d biz id", bizID, opt.BizID)
	}

	appIDStr, exist := metadata["app_id"]
	if !exist {
		return "", errors.New("node metadata not has app id")
	}

	appIDs := make([]uint32, 0)
	if err = json.Unmarshal([]byte(appIDStr), &appIDs); err != nil {
		return "", fmt.Errorf("unmarshal node metadata appID failed, err: %v", err)
	}

	// judge current app if already upload this node.
	var idExist bool
	for index := range appIDs {
		if appIDs[index] == appID {
			idExist = true
			break
		}
	}

	if !idExist {
		appIDs = append(appIDs, appID)
	}

	meta := &repo.NodeMeta{
		BizID: opt.BizID,
		AppID: appIDs,
	}
	return meta.String()
}

// getBizIDAndAppID get biz_id and app_id from req path.
func getBizIDAndAppID(kt *kit.Kit, req *http.Request) (uint32, uint32, error) {
	bizIDStr := mux.Vars(req)["biz_id"]
	bizID, err := strconv.ParseUint(bizIDStr, 10, 64)
	if err != nil {
		logs.Errorf("biz id parse uint failed, err: %v, rid: %s", err, kt.Rid)
		return 0, 0, err
	}

	if bizID == 0 {
		return 0, 0, errf.New(errf.InvalidParameter, "biz_id should > 0")
	}

	appIDStr := mux.Vars(req)["app_id"]
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
