/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package server

import (
	"encoding/json"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/metric"
)

func (s *Server) healthy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
}

type response struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Marshal return the bytes of response object
func (r *response) Marshal() []byte {
	bs, _ := json.Marshal(r)
	return bs
}

func (s *Server) responseError(r *http.Request, w http.ResponseWriter, statusCode int, err error) {
	blog.Errorf("RequestID[%s] request '%s' response err: %s", requestID(r.Context()),
		r.RequestURI, err.Error())

	if statusCode >= 500 {
		metric.RequestFailed.WithLabelValues().Inc()
	}
	w.WriteHeader(statusCode)
	resp := &response{
		Code:    1,
		Message: err.Error(),
	}
	_, _ = w.Write(resp.Marshal())
}

func (s *Server) responseSuccess(w http.ResponseWriter, obj interface{}) {
	w.WriteHeader(http.StatusOK)
	resp := &response{
		Code: 0,
		Data: obj,
	}
	_, _ = w.Write(resp.Marshal())
}

func (s *Server) responseDirect(w http.ResponseWriter, obj []byte) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(obj)
}
