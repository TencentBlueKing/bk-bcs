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

package secret

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/vaultplugin-server/handler"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/pkg/errors"
)

// ServerOptions for revese proxy
type ServerOptions struct {
	Address string
	Port    string
}

// NewServerProxy create proxy instance
func NewServerProxy(opt *ServerOptions) *ServerProxy {
	return &ServerProxy{
		option: opt,
	}
}

// ServerProxy for secretServer
type ServerProxy struct {
	option *ServerOptions
}

// ProxySecretRequest proxy for secret server
func (s *ServerProxy) ProxySecretRequest(req *http.Request) *handler.SecretResponse {
	realPath := strings.TrimPrefix(req.URL.RequestURI(), common.GitOpsProxyURL)
	// !force https link
	fullPath := fmt.Sprintf("http://%s:%s%s", s.option.Address, s.option.Port, realPath)

	rbody, err := s.send(context.TODO(), fullPath, req.Method, nil, req.Body, req.Header)
	if err != nil {
		return &handler.SecretResponse{
			Code:    handler.ErrHttpCode,
			Message: fmt.Sprintf("send request when ProxySecretRequest, err: %s", err),
		}
	}

	response := &handler.SecretResponse{}
	blog.Info("[ProxySecretRequest] send request return body: %s", string(rbody))
	err = json.Unmarshal(rbody, response)
	if err != nil {
		return &handler.SecretResponse{
			Code:    handler.ErrHttpCode,
			Message: fmt.Sprintf("error unmarshal resp when ProxySecretRequest, err: %s", err),
		}
	}
	return response
}

// InitSecretRequest hard code path and method for initSecret
func (s *ServerProxy) InitSecretRequest(project string) error {
	realPath := "/api/v1/secrets/init"
	method := http.MethodPost
	fullPath := fmt.Sprintf("http://%s:%s%s?project=%s", s.option.Address, s.option.Port, realPath, project)

	rbody, err := s.send(context.TODO(), fullPath, method, nil, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "send request when InitSecretRequest")
	}

	response := &handler.SecretResponse{}
	err = json.Unmarshal(rbody, response)
	if err != nil {
		return errors.Wrapf(err, "error unmarshal resp when InitSecretRequest")
	}

	if response.Code != handler.SuccessHttpCode {
		return fmt.Errorf(response.Message)
	}

	return nil
}

// GetInitSecretRequest hard code path and method for initSecret
func (s *ServerProxy) GetInitSecretRequest(project string) (string, error) {
	realPath := "/api/v1/secrets/annotation"
	method := http.MethodGet
	fullPath := fmt.Sprintf("http://%s:%s%s?project=%s", s.option.Address, s.option.Port, realPath, project)

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	rbody, err := s.send(ctx, fullPath, method, nil, nil, nil)
	if err != nil {
		return "", errors.Wrapf(err, "send request when GetInitSecretRequest")
	}

	response := &handler.SecretResponse{}
	err = json.Unmarshal(rbody, response)
	if err != nil {
		return "", errors.Wrapf(err, "error unmarshal resp when GetInitSecretRequest")
	}
	if response.Code != handler.SuccessHttpCode {
		return "", fmt.Errorf(response.Message)
	}

	if str, ok := response.Data.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("response.Data.(string) error, data is not str [%s]", response.Data)
}

func (s *ServerProxy) send(ctx context.Context, fullPath, method string, queryParams map[string]string, body io.ReadCloser,
	header map[string][]string) ([]byte, error) {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, fullPath, body)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, fullPath, nil)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "create http request failed when proxy send")
	}

	for k, v := range header {
		req.Header.Set(k, v[0])
	}

	if queryParams != nil {
		query := req.URL.Query()
		for k, v := range queryParams {
			query.Set(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http request failed when proxy send")
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body failed when proxy send")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http response code not 200 but %d, resp: %s",
			resp.StatusCode, string(respBody))
	}
	return respBody, err
}
