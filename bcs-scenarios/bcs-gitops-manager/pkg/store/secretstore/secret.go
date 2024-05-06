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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/httputils"
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
func NewSecretStore(op *common.SecretStoreOptions) SecretInterface {
	return &secretStore{
		op: op,
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
	bs, err := httputils.Send(ctx, &httputils.HTTPRequest{
		Address: s.op.Address,
		Port:    s.op.Port,
		Path:    initPath,
		Method:  http.MethodPost,
		QueryParams: map[string]string{
			"project": project,
		},
	})
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
	bs, err := httputils.Send(ctx, &httputils.HTTPRequest{
		Address: s.op.Address,
		Port:    s.op.Port,
		Path:    getSecretPath, // nolint
		Method:  http.MethodGet,
		QueryParams: map[string]string{
			"project": project,
		},
	})
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
	bs, err := httputils.Send(ctx, &httputils.HTTPRequest{
		Address: s.op.Address,
		Port:    s.op.Port,
		Path:    fmt.Sprintf(listProjectSecretsPath, project),
		Method:  http.MethodGet,
		QueryParams: map[string]string{
			"path": "/",
		},
	})
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
