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

// Package secretstore defines the function for vaultplugin
package secretstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

// SecretInterface defines the interface of secret
type SecretInterface interface {
	InitProjectSecret(ctx context.Context, project string) error
	GetProjectSecret(ctx context.Context, project string) (string, error)
	ListProjectSecrets(ctx context.Context, project string) ([]string, error)
}

type secretStore struct {
	op *common.SecretStoreOptions
}

// NewSecretStore will create the instance of SecretStore
func NewSecretStore() SecretInterface {
	op := options.GlobalOptions()
	return &secretStore{
		op: op.SecretServer,
	}
}

const (
	initPath = "/api/v1/secrets/init"
)

type secretResponse struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// InitProjectSecret init the secret when project is open
func (s *secretStore) InitProjectSecret(ctx context.Context, project string) error {
	hr := &httpRequest{
		path:   initPath,
		method: http.MethodPost,
		queryParams: map[string]string{
			"project": project,
		},
	}
	bs, err := s.send(ctx, hr)
	if err != nil {
		return errors.Wrapf(err, "init secret for project '%s' failed", project)
	}
	response := new(secretResponse)
	if err = json.Unmarshal(bs, response); err != nil {
		return errors.Wrapf(err, "init secret for project '%s' unmarshal '%s' failed",
			project, string(bs))
	}
	if response.Code != 0 {
		return errors.Errorf("init secret for project '%s' response code not 0 but %d: %s",
			project, response.Code, response.Message)
	}
	return nil
}

const (
	// nolint
	getSecretPath = "/api/v1/secrets/annotation"
)

// GetProjectSecret get the secret by project name
func (s *secretStore) GetProjectSecret(ctx context.Context, project string) (string, error) {
	hr := &httpRequest{
		path:   getSecretPath, // nolint
		method: http.MethodGet,
		queryParams: map[string]string{
			"project": project,
		},
	}
	bs, err := s.send(ctx, hr)
	if err != nil {
		return "", errors.Wrapf(err, "get project '%s' secret failed", project)
	}
	response := new(secretResponse)
	if err = json.Unmarshal(bs, response); err != nil {
		return "", errors.Wrapf(err, "get project '%s' secret unmarshal '%s' failed", project, string(bs))
	}
	if response.Code != 0 {
		return "", errors.Errorf("get project '%s' secret resp code not 0 but %d: %s",
			project, response.Code, response.Message)
	}
	if str, ok := response.Data.(string); ok {
		return str, nil
	}
	return "", errors.Errorf("get project '%s' secret response convert to string failed", project)
}

const (
	// nolint
	listProjectSecretsPath = "/api/v1/secrets/%s/list"
)

func (s *secretStore) ListProjectSecrets(ctx context.Context, project string) ([]string, error) {
	hr := &httpRequest{
		path:   fmt.Sprintf(listProjectSecretsPath, project),
		method: http.MethodGet,
		queryParams: map[string]string{
			"path": "/",
		},
	}
	bs, err := s.send(ctx, hr)
	if err != nil {
		return nil, errors.Wrapf(err, "list project '%s' secret failed", project)
	}
	response := new(secretResponse)
	if err = json.Unmarshal(bs, response); err != nil {
		return nil, errors.Wrapf(err, "init secret for project '%s' unmarshal '%s' failed",
			project, string(bs))
	}
	if response.Code != 0 {
		return nil, errors.Errorf("list secrets for project '%s' response code not 0 but %d: %s",
			project, response.Code, response.Message)
	}
	if response.Data == nil {
		return nil, nil
	}
	if secrets, ok := response.Data.([]interface{}); ok {
		result := make([]string, 0, len(secrets))
		for i := range secrets {
			result = append(result, secrets[i].(string))
		}
		return result, nil
	}
	return nil, errors.Errorf("list secrets for project '%s' convert failed", project)
}

type httpRequest struct {
	path        string
	method      string
	queryParams map[string]string
	body        interface{}
	header      map[string]string
}

func (s *secretStore) send(ctx context.Context, hr *httpRequest) ([]byte, error) {
	var req *http.Request
	var err error

	urlStr := fmt.Sprintf("http://%s:%s%s", s.op.Address, s.op.Port, hr.path) // nolint
	if hr.body != nil {
		var body []byte
		body, err = json.Marshal(hr.body)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal body failed")
		}
		req, err = http.NewRequestWithContext(ctx, hr.method, urlStr, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, hr.method, urlStr, nil)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "create http request failed")
	}
	for k, v := range hr.header {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")

	if hr.queryParams != nil {
		query := req.URL.Query()
		for k, v := range hr.queryParams {
			query.Set(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if !utils.IsContextCanceled(err) {
			metric.ManagerSecretOperateFailed.WithLabelValues().Inc()
		}
		return nil, errors.Wrap(err, "http request failed when proxy send")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body failed when proxy send")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("http response code not 200 but %d, resp: %s",
			resp.StatusCode, string(respBody))
	}
	return respBody, err
}
