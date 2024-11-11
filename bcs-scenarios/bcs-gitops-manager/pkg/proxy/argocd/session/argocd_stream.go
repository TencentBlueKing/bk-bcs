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
 */

package session

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils/jsonq"
)

// ArgoStreamSession defines the session of argo stream
type ArgoStreamSession struct {
	option *options.Options
	store  store.Store
	token  string
}

// NewArgoStreamSession create the argo stream session instance
func NewArgoStreamSession() *ArgoStreamSession {
	s := &ArgoStreamSession{
		option: options.GlobalOptions(),
		store:  store.GlobalStore(),
	}
	s.token = s.store.GetToken(context.Background())
	return s
}

// ServeHTTP http.Handler implementation
func (s *ArgoStreamSession) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	requestID := req.Context().Value(traceconst.RequestIDHeaderKey).(string)
	// backend real path with encoded format
	realPath := strings.TrimPrefix(req.URL.RequestURI(), common.GitOpsProxyURL)
	// !force https link
	fullPath := fmt.Sprintf("https://%s%s", s.option.GitOps.Service, realPath)
	blog.Infof("RequestID[%s] GitOps stream proxy: %s", requestID, fullPath)
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	if err := s.forwardStreamToArgo(rw, req, fullPath); err != nil {
		if utils.IsContextCanceled(err) || utils.IsContextDeadlineExceeded(err) {
			rw.WriteHeader(http.StatusOK)
			return
		}
		blog.Errorf("RequestID[%s] GitOps stream proxy %s failed: %s", requestID, fullPath, err.Error())
	} else {
		blog.Infof("RequestID[%s] GitOps stream proxy %s success", requestID, fullPath)
	}
}

func (s *ArgoStreamSession) forwardStreamToArgo(rw http.ResponseWriter, req *http.Request, fullPath string) error {
	requestID := req.Context().Value(traceconst.RequestIDHeaderKey).(string)
	fieldsQuery := req.URL.Query().Get("fields")
	var fields []string
	if fieldsQuery != "" {
		fields = strings.Split(fieldsQuery, ",")
	}
	// NOCC:Server Side Request Forgery(只是代码封装，所有 URL都是可信的)
	forwardReq, err := http.NewRequestWithContext(req.Context(), req.Method, fullPath, req.Body)
	if err != nil {
		return errors.Wrapf(err, "create forward request failed")
	}
	forwardReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
	forwardReq.Header.Set("Token", s.token)
	forwardReq.Header.Set("Content-Type", "text/event-stream")
	httpClient := http.DefaultClient
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint
	}
	resp, err := httpClient.Do(forwardReq)
	if err != nil {
		return errors.Wrapf(err, "do forward request failed")
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		lineBytes := scanner.Bytes()
		if len(fields) != 0 {
			var afterReserve []byte
			if afterReserve, err = jsonq.ReserveFieldBytes(lineBytes, fields); err != nil {
				blog.Errorf("RequestID[%s] GitOps stream reserve(fields: '%v', target: '%s') field failed: %s",
					requestID, fields, string(lineBytes), err.Error())
			} else {
				lineBytes = afterReserve
			}
		}
		// nolint
		rw.Write([]byte(fmt.Sprintf("data: %s\n\n", string(lineBytes))))
		rw.(http.Flusher).Flush()
	}
	return nil
}
