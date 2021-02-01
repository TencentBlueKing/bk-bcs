/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	pbapiserver "bk-bscp/internal/protocol/apiserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/json"
	"bk-bscp/pkg/logger"
)

// setupFilters setups all api filters here. All request would cross here, and we filter request base on URL.
func (s *APIServer) setupFilters(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rtime := time.Now()
		kit := common.HTTPRequestKit(r)
		httpStatusCode := http.StatusOK

		// NOTE: validate ESB request metadata in kit. Validate basic metadata in the request kit.
		// Ignore swagger request if the switch is open. And return bad request error when
		// the request kit is invalid.
		if err := kit.Validate(); err != nil {
			// swagger.
			if s.viper.GetBool("server.api.open") && strings.Contains(r.RequestURI, "swagger") {
				mux.ServeHTTP(w, r)
				return
			}

			// healthz.
			if strings.Contains(r.RequestURI, "healthz") {
				mux.ServeHTTP(w, r)
				return
			}
			logger.Errorf("Filter| bad request, metadata:%+v, method:%s, uri:%s, remote_addr:%s, err: %+v",
				kit, r.Method, r.RequestURI, r.RemoteAddr, err)

			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, err.Error())
			return
		}

		// NOTE: add the request rid to response header, and handle normal requests now.
		w.Header().Set(common.RidHeaderKey, kit.Rid)

		logger.V(2).Infof("Filter[%s]-input| method:%s, uri:%s, appcode:%s, user:%s, remote_addr:%s",
			kit.Rid, r.Method, r.RequestURI, kit.AppCode, kit.User, r.RemoteAddr)

		defer func() {
			cost := s.collector.StatRequest(fmt.Sprintf("Filter-%s", r.Method), httpStatusCode, rtime, time.Now())
			logger.V(2).Infof("Filter[%s]-output[%dms]| method:%s, uri:%s, appcode:%s, user:%s, remote_addr:%s, %d %s",
				kit.Rid, cost, r.Method, r.RequestURI, kit.AppCode, kit.User, r.RemoteAddr, httpStatusCode,
				http.StatusText(httpStatusCode))
		}()

		// filter and handle normal http requests.
		if errCode, errMsg := s.filter(r); errCode != pbcommon.ErrCode_E_OK {
			response, err := json.MarshalPB(&pbapiserver.CommonAPIResponse{
				Result:  false,
				Code:    errCode,
				Message: errMsg,
			})
			if err != nil {
				httpStatusCode = http.StatusInternalServerError
				w.WriteHeader(httpStatusCode)
				fmt.Fprintf(w, fmt.Sprintf("marshal filter error response failed, %+v", err))
				return
			}
			if _, err := w.Write([]byte(response)); err != nil {
				httpStatusCode = http.StatusInternalServerError
				w.WriteHeader(httpStatusCode)
				fmt.Fprintf(w, fmt.Sprintf("marshal filter error response failed, %+v", err))
				return
			}
		} else {
			// NOTE:filter done, and then handle request mux.
			// DO NOT change this action, all requests should be handled by the real mux.
			// The filters only hijack requests to authserver or other sides, but it do not
			// reject any request.
			mux.ServeHTTP(w, r)
		}
	})
}

// filter filters some bypass logics including the auth check.
func (s *APIServer) filter(r *http.Request) (pbcommon.ErrCode, string) {
	// authorize filter.
	if errCode, errMsg := s.authorize(r); errCode != pbcommon.ErrCode_E_OK {
		return errCode, errMsg
	}
	return pbcommon.ErrCode_E_OK, "OK"
}
