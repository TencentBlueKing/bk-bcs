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

// Package secret defines the secret command
package secret

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/secret"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/httputils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-cli/pkg/utils"
)

// Handler defines the handler of terraform
type Handler struct {
	op *options.GitOpsOptions
}

// NewHandler create the terraform handler instance
func NewHandler() *Handler {
	return &Handler{
		op: options.GlobalOption(),
	}
}

var (
	listPath     = "/api/v1/secrets/%s/list?&path=%%2F"
	createPath   = "/api/v1/secrets/%s/%s"
	metadataPath = "/api/v1/secrets/%s/%s/metadata"
	versionPath  = "/api/v1/secrets/%s/%s?&version=%d"
	deletePath   = "/api/v1/secrets/%s/%s"
)

type listResponse struct {
	Code int      `json:"code"`
	Data []string `json:"data"`
}

// List secrets
func (h *Handler) List(ctx context.Context, proj string) {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(listPath, proj),
		Method: http.MethodGet,
	})
	resp := &listResponse{}
	if err := json.Unmarshal(respBody, resp); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal response body failed: %s", err.Error()))
	}

	tw := utils.DefaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"NAME",
		}
	}())
	for i := range resp.Data {
		tw.Append(func() []string {
			return []string{
				resp.Data[i],
			}
		}())
	}
	tw.Render()
}

// VersionData defines the version data
type VersionData struct {
	Data map[string]string `json:"data"`
}

// Create the secret
func (h *Handler) Create(ctx context.Context, proj string, name string, data map[string]string) {
	_ = httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(createPath, proj, name),
		Method: http.MethodPost,
		Body: &VersionData{
			Data: data,
		},
	})
	fmt.Println("success")
}

// GetMetadata get secret metadata
func (h *Handler) GetMetadata(ctx context.Context, proj string, name string) {
	resp := h.getMetadata(ctx, proj, name)
	fmt.Printf(`##################################################################
### CurrentVersion: %d                                           
### CreateTime: %s                                
### UpdateTime: %s                                
#----------------------------------------------------------------#`,
		resp.Data.CurrentVersion, resp.Data.CreateTime.Format("2006-01-02 15:04:05"),
		resp.Data.UpdatedTime.Format("2006-01-02 15:04:05"))
	fmt.Println()
	tw := utils.DefaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"VERSION", "CREATE_TIME",
		}
	}())
	for _, v := range resp.Data.Version {
		tw.Append(func() []string {
			return []string{
				strconv.Itoa(v.Version), v.CreatedTime.Format("2006-01-02 15:04:05"),
			}
		}())
	}
	tw.Render()
}

// GetLatestVersion get the secret last version details
func (h *Handler) GetLatestVersion(ctx context.Context, proj string, name string) map[string]string {
	metadata := h.getMetadata(ctx, proj, name)
	version := metadata.Data.CurrentVersion
	versionResp := h.getVersion(ctx, proj, name, version)
	return versionResp.Data
}

// GetVersion get the version details
func (h *Handler) GetVersion(ctx context.Context, proj string, name string, version int) {
	if version == 0 {
		metadata := h.getMetadata(ctx, proj, name)
		version = metadata.Data.CurrentVersion
		fmt.Println(">>> Used latest version: ", version)
	} else {
		fmt.Println(">>> Used specified version: ", version)
	}
	versionResp := h.getVersion(ctx, proj, name, version)
	tw := utils.DefaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"KEY", "VALUE", "PLACEHOLDER",
		}
	}())
	for k, v := range versionResp.Data {
		tw.Append(func() []string {
			return []string{
				k, v, fmt.Sprintf("<path:%s/data/%s#%s>", proj, name, k),
			}
		}())
	}
	tw.Render()
}

// Delete secret
func (h *Handler) Delete(ctx context.Context, proj string, name string) {
	_ = httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(deletePath, proj, name),
		Method: http.MethodDelete,
	})
	fmt.Println("success")
}

type metadataResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    *secret.SecretMetadata `json:"data"`
}

func (h *Handler) getMetadata(ctx context.Context, proj, name string) *metadataResponse {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(metadataPath, proj, name),
		Method: http.MethodGet,
	})
	resp := &metadataResponse{}
	if err := json.Unmarshal(respBody, resp); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal metadata '%s' failed: %s", string(respBody), err.Error()))
	}
	return resp
}

type versionResponse struct {
	Code int               `json:"code"`
	Data map[string]string `json:"data"`
}

func (h *Handler) getVersion(ctx context.Context, proj, name string, version int) *versionResponse {
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(versionPath, proj, name, version),
		Method: http.MethodGet,
	})
	resp := &versionResponse{}
	if err := json.Unmarshal(respBody, resp); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal version '%s' failed: %s", string(respBody), err.Error()))
	}
	return resp
}
