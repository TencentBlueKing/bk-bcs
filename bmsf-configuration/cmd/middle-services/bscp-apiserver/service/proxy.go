/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
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
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/bluele/gcache"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"bk-bscp/cmd/middle-services/bscp-apiserver/modules/metrics"
	"bk-bscp/pkg/bkrepo"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

const (
	// defaultWriteBufferSize is default write buffer size, 4KB.
	defaultWriteBufferSize = 4 << 10

	// defaultReadBufferSize is default read buffer size, 4KB.
	defaultReadBufferSize = 4 << 10

	// defaultProxyScheme is default proxy scheme.
	defaultProxyScheme = "http"

	// interfaceFileContentPath is path for bscp gateway commit file content interface.
	interfaceFileContentPath = "api/v2/file/content/biz/"
)

type syncBKRepoRecordType int

const (
	// syncBKRepoRecordTypeNone mesns have not sync bkrepo.
	syncBKRepoRecordTypeNone syncBKRepoRecordType = 0

	// syncBKRepoRecordTypeProjCreated mesns have only synchronized bkrepo project.
	syncBKRepoRecordTypeProjCreated syncBKRepoRecordType = 1

	// syncBKRepoRecordTypeRepoCreated mesns have only synchronized bkrepo repository.
	syncBKRepoRecordTypeRepoCreated syncBKRepoRecordType = 2

	// syncBKRepoRecordTypeCreated mesns have both synchronized bkrepo project and repository.
	syncBKRepoRecordTypeCreated syncBKRepoRecordType = 3
)

// NewBKRepoDirector returns a director for bkrepo.
func NewBKRepoDirector(host, token string) func(req *http.Request) {
	return func(req *http.Request) {
		// NOTE: file upload only support http now.
		req.URL.Scheme = defaultProxyScheme
		req.Host = host
		req.URL.Host = host

		req.URL.Path = strings.Replace(req.URL.Path,
			interfaceFileContentPath, bkrepo.GENERICAPIPATH+bkrepo.BSCPBIZIDPREFIX, 1)

		req.Header.Set("Authorization", fmt.Sprintf("Platform %s", token))
		req.Header.Set(bkrepo.BKRepoUIDHeaderKey, req.Header.Get(common.UserHeaderKey))
		req.Header.Set(bkrepo.BKRepoSHA256HeaderKey, strings.ToUpper(req.Header.Get(common.ContentIDHeaderKey)))
		req.Header.Set(bkrepo.BKRepoOverwriteHeaderKey, req.Header.Get(common.ContentOverwriteHeaderKey))

		logger.Info("director file content commit now, %+v", req.URL)
	}
}

// BKRepoReverseProxy is http reverse proxy for bkrepo.
type BKRepoReverseProxy struct {
	viper     *viper.Viper
	proxy     *httputil.ReverseProxy
	collector *metrics.Collector

	// memory LRU cache used for re-create bkrepo project/repository, bizid -> syncBKRepoRecordType.
	syncBKRepoRecords gcache.Cache
}

// NewBKRepoReverseProxy creates a new ReverseProxy for bkrepo.
func NewBKRepoReverseProxy(viper *viper.Viper, director func(*http.Request),
	collector *metrics.Collector) *BKRepoReverseProxy {

	return &BKRepoReverseProxy{
		viper: viper,
		proxy: &httputil.ReverseProxy{
			// Director must be a function which modifies the request into a new Request
			// to be sent using Transport. Its response is then copied back to the original
			// client unmodified. Director must not access the provided Request after returning.
			Director: director,

			// The transport used to perform proxy requests. If nil,
			// http.DefaultTransport is used.
			Transport: &http.Transport{
				Proxy:               http.ProxyFromEnvironment,
				Dial:                (&net.Dialer{Timeout: viper.GetDuration("bkrepo.dialerTimeout")}).Dial,
				MaxConnsPerHost:     viper.GetInt("bkrepo.maxConnsPerHost"),
				MaxIdleConnsPerHost: viper.GetInt("bkrepo.maxIdleConnsPerHost"),
				IdleConnTimeout:     viper.GetDuration("bkrepo.idleConnTimeout"),
				WriteBufferSize:     defaultWriteBufferSize,
				ReadBufferSize:      defaultReadBufferSize,
			},
		},
		collector:         collector,
		syncBKRepoRecords: gcache.New(viper.GetInt("bkrepo.recordCacheSize")).EvictType(gcache.TYPE_LRU).Build(),
	}
}

// ServeHTTP proxy http request to bkrepo.
func (p *BKRepoReverseProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rtime := time.Now()
	kit := common.HTTPRequestKit(req)

	logger.V(2).Infof("FileContent[%s][%s]| input[%+v]", kit.Rid, req.Method, req.URL)

	defer func() {
		cost := p.collector.StatRequest(fmt.Sprintf("FileContent-%s", req.Method), http.StatusOK, rtime, time.Now())
		logger.V(2).Infof("FileContent[%s][%s]| output[%dms] [%+v]", kit.Rid, req.Method, cost, req.URL)
	}()

	// mux limit the file content path in target path format, and you can get biz_id
	// from mux vars base on the request.
	bizID := mux.Vars(req)["biz_id"]

	if req.Method == "PUT" {
		// NOTE: check target business repo project/repository in upload request.

		syncBKRepoFlag := syncBKRepoRecordTypeNone
		if record, err := p.syncBKRepoRecords.Get(bizID); err == nil && record != nil {
			if flag, ok := record.(syncBKRepoRecordType); ok {
				syncBKRepoFlag = flag
			}
		}

		if syncBKRepoFlag < syncBKRepoRecordTypeProjCreated {
			err := bkrepo.CreateProject(
				fmt.Sprintf("%s://%s", defaultProxyScheme, p.viper.GetString("bkrepo.host")),
				&bkrepo.Auth{Token: p.viper.GetString("bkrepo.token"), UID: kit.User},
				&bkrepo.CreateProjectReq{
					Name:        bizID,
					DisplayName: bizID,
					Description: "bscp-configs"},
				p.viper.GetDuration("bkrepo.timeout"))

			if err != nil {
				logger.Warnf("FileContent[%s][%s]| [%+v], create bkrepo project failed, %+v",
					kit.Rid, req.Method, req.URL, err)
			} else {
				p.syncBKRepoRecords.SetWithExpire(bizID, syncBKRepoRecordTypeProjCreated,
					p.viper.GetDuration("bkrepo.recordCacheExpiration"))

				logger.V(2).Infof("FileContent[%s][%s]| [%+v], create bkrepo project success",
					kit.Rid, req.Method, req.URL)
			}
		}

		if syncBKRepoFlag < syncBKRepoRecordTypeRepoCreated {
			err := bkrepo.CreateRepo(
				fmt.Sprintf("%s://%s", defaultProxyScheme, p.viper.GetString("bkrepo.host")),
				&bkrepo.Auth{Token: p.viper.GetString("bkrepo.token"), UID: kit.User},
				&bkrepo.CreateRepoReq{
					ProjectID:     bizID,
					Name:          bkrepo.CONFIGSREPONAME,
					Type:          bkrepo.REPOTYPE,
					Category:      bkrepo.CATEGORY,
					Configuration: bkrepo.Configuration{Type: bkrepo.REPOCFGTYPE},
					Description:   "bscp-configs"},
				p.viper.GetDuration("bkrepo.timeout"))

			if err != nil {
				logger.Warnf("FileContent[%s][%s]| [%+v], create bkrepo repository failed, %+v",
					kit.Rid, req.Method, req.URL, err)
			} else {
				p.syncBKRepoRecords.SetWithExpire(bizID, syncBKRepoRecordTypeRepoCreated,
					p.viper.GetDuration("bkrepo.recordCacheExpiration"))

				logger.V(2).Infof("FileContent[%s][%s]| [%+v], create bkrepo repository success",
					kit.Rid, req.Method, req.URL)
			}
		}
	}

	p.proxy.ServeHTTP(w, req)
}

// Verify parses request metadata and check auth info.
func (p *BKRepoReverseProxy) Verify(req *http.Request) error {
	// check headers.
	operator := req.Header.Get(common.UserHeaderKey)
	if len(operator) == 0 {
		return errors.New("file content operator info in header is required")
	}
	sha256 := strings.ToUpper(req.Header.Get(common.ContentIDHeaderKey))
	if len(sha256) == 0 {
		return errors.New("file content sha256 id info in header is required")
	}
	req.URL.Path += fmt.Sprintf("/%s/%s", bkrepo.CONFIGSREPONAME, sha256)

	return nil
}
