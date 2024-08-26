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

package secretutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/httputils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
)

// Handler defines the handler for secret
type Handler struct {
	sync.Mutex
	store map[string]map[string]map[string]string
}

// NewHandler create the secret handler instance
func NewHandler() *Handler {
	return &Handler{
		store: make(map[string]map[string]map[string]string),
	}
}

var (
	keyRegex = regexp.MustCompile(`<path:(.*?)/data/(.*?)#(.*?)>`)
)

// GetSecret get secret with key
func (h *Handler) GetSecret(key string) (string, error) {
	matches := keyRegex.FindStringSubmatch(key)
	if len(matches) != 4 {
		return "", errors.Errorf("invalid secret key: %s", key)
	}
	proj := matches[1]
	secret := matches[2]
	secretKey := matches[3]
	result := h.loadSecret(proj, secret, secretKey)
	if result != "" {
		return result, nil
	}
	secretData, err := h.getSecretFromManager(proj, secret)
	if err != nil {
		return "", errors.Wrapf(err, "get secret '%s/%s' from manager failed", proj, secret)
	}
	h.saveSecret(proj, secret, secretData)
	return secretData[secretKey], nil
}

var (
	metadataUrl = "/api/v1/secrets/%s/%s/metadata"
	secretUrl   = "/api/v1/secrets/%s/%s?version=%d"
)

type metadataResp struct {
	Data struct {
		CurrentVersion int `json:"CurrentVersion"`
	} `json:"data"`
}

type secretResp struct {
	Data map[string]string `json:"data"`
}

func (h *Handler) getSecretFromManager(proj, secret string) (map[string]string, error) {
	metadataBody := httputils.DoRequest(context.Background(), &httputils.HTTPRequest{
		Path:   fmt.Sprintf(metadataUrl, proj, secret),
		Method: http.MethodGet,
	})
	mr := new(metadataResp)
	if err := json.Unmarshal(metadataBody, mr); err != nil {
		utils.ExitError(fmt.Sprintf("get secret '%s/%s' metdata failed: %s'", proj, secret, err.Error()))
	}
	secretBody := httputils.DoRequest(context.Background(), &httputils.HTTPRequest{
		Path:   fmt.Sprintf(secretUrl, proj, secret, mr.Data.CurrentVersion),
		Method: http.MethodGet,
	})
	sr := new(secretResp)
	if err := json.Unmarshal(secretBody, sr); err != nil {
		utils.ExitError(fmt.Sprintf("get secret '%s/%s/%d' result failed: %s'", proj, secret,
			mr.Data.CurrentVersion, err.Error()))
	}
	return sr.Data, nil
}

func (h *Handler) loadSecret(proj, secret, secretKey string) string {
	h.Lock()
	defer h.Unlock()
	v1, ok := h.store[proj]
	if !ok {
		return ""
	}
	v2, ok := v1[secret]
	if !ok {
		return ""
	}
	v3, ok := v2[secretKey]
	if !ok {
		return ""
	}
	return v3
}

func (h *Handler) saveSecret(proj, secret string, secretData map[string]string) {
	h.Lock()
	defer h.Unlock()
	_, ok := h.store[proj]
	if !ok {
		h.store[proj] = make(map[string]map[string]string)
	}
	h.store[proj][secret] = secretData
}
