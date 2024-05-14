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

// Package terraformstore xx
package terraformstore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/httpapi"
	terraformextensionsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/httputils"
)

var (
	getPath      = "/api/v1/terraforms/%s"
	getDiffPath  = "/api/v1/terraforms/%s/get-diff"
	getApplyPath = "/api/v1/terraforms/%s/get-apply"
)

// TerraformInterface defines the interface to operate terraform
type TerraformInterface interface {
	Get(ctx context.Context, name string) (*terraformextensionsv1.Terraform, error)
	GetDiff(ctx context.Context, name string) (*httpapi.TerraformGetDiffOrApplyData, error)
	GetApply(ctx context.Context, name string) (*httpapi.TerraformGetDiffOrApplyData, error)
}

type terraformHandler struct {
	op *common.TerraformConfig
}

// NewTerraformStore create the terraform store instance
func NewTerraformStore() TerraformInterface {
	op := options.GlobalOptions()
	return &terraformHandler{
		op: op.TerraformServer,
	}
}

type terraformGetResponse struct {
	Code int                              `json:"code"`
	Data *terraformextensionsv1.Terraform `json:"data"`
}

// Get return the terraform by name
func (h *terraformHandler) Get(ctx context.Context, name string) (*terraformextensionsv1.Terraform, error) {
	resp := &terraformGetResponse{}
	if err := h.get(ctx, name, fmt.Sprintf(getPath, name), resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetDiff return terraform diff
func (h *terraformHandler) GetDiff(ctx context.Context, name string) (*httpapi.TerraformGetDiffOrApplyData, error) {
	resp := &httpapi.TerraformGetDiffOrApplyResponse{}
	if err := h.get(ctx, name, fmt.Sprintf(getDiffPath, name), resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetApply return terraform apply
func (h *terraformHandler) GetApply(ctx context.Context, name string) (*httpapi.TerraformGetDiffOrApplyData, error) {
	resp := &httpapi.TerraformGetDiffOrApplyResponse{}
	if err := h.get(ctx, name, fmt.Sprintf(getApplyPath, name), resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (h *terraformHandler) get(ctx context.Context, name, path string, resp interface{}) error {
	bs, err := httputils.Send(ctx, &httputils.HTTPRequest{
		Address: h.op.Address,
		Port:    h.op.Port,
		Path:    path,
		Method:  http.MethodGet,
	})
	if err != nil {
		return errors.Wrapf(err, "get terraform '%s' failed", name)
	}
	if err = json.Unmarshal(bs, resp); err != nil {
		return errors.Wrapf(err, "get terrform '%s' unmarshal '%s' failed", name, string(bs))
	}
	return nil
}
