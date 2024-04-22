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

// Package terraform defines the terraform command
package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/httpapi"
	terraformextensionsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/tfhandler"

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
	listPath     = "/api/v1/terraforms"
	createPath   = "/api/v1/terraforms/create"
	applyPath    = "/api/v1/terraforms/apply"
	getPath      = "/api/v1/terraforms/%s"
	deletePath   = "/api/v1/terraforms/%s"
	getDiffPath  = "/api/v1/terraforms/%s/get-diff"
	getApplyPath = "/api/v1/terraforms/%s/get-apply"
	syncPath     = "/api/v1/terraforms/%s/sync"
	// nolint
	cleanPath = "/api/v1/terraforms/%s/clean"
)

type listResponse struct {
	Code int                               `json:"code"`
	Data []terraformextensionsv1.Terraform `json:"data"`
}

type stringResponse struct {
	Code int    `json:"code"`
	Data string `json:"data"`
}

// List terraforms
func (h *Handler) List(ctx context.Context, projects *[]string) {
	queryParams := make(map[string][]string)
	if len(*projects) != 0 {
		queryParams["projects"] = *projects
	}
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:        listPath,
		Method:      http.MethodGet,
		QueryParams: queryParams,
	})
	resp := &listResponse{}
	if err := json.Unmarshal(respBody, resp); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal response body failed: %s", err.Error()))
	}

	tw := utils.DefaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"NAME", "REPO", "POLICY", "SYNC", "OPERATION", "DESTROY", "LAST APPLY",
		}
	}())
	for i := range resp.Data {
		tf := resp.Data[i]
		var duration string
		if tf.Status.LastAppliedAt != nil {
			duration = utils.FriendTimeFormat(tf.Status.LastAppliedAt.Time, time.Now())
		}
		tw.Append(func() []string {
			return []string{
				tf.Name,
				tf.Spec.Repository.Repo,
				tf.Spec.SyncPolicy,
				tf.Status.SyncStatus,
				tf.Status.OperationStatus.Phase,
				fmt.Sprintf("%v", tf.Spec.DestroyResourcesOnDeletion),
				duration,
			}
		}())
	}
	tw.Render()
}

// Apply terraform with json or yaml
func (h *Handler) Apply(ctx context.Context, body []byte) {
	var jsonData map[string]interface{}
	var yamlData map[string]interface{}
	jsonErr := json.Unmarshal(body, &jsonData)
	yamlErr := yaml.Unmarshal(body, &yamlData)
	if jsonErr != nil && yamlErr != nil {
		utils.ExitError("request body not json or yaml type")
	}
	if yamlErr == nil {
		var err error
		if body, err = json.Marshal(yamlData); err != nil {
			utils.ExitError(fmt.Sprintf("yaml to json failed: %s", err.Error()))
		}
	}
	terraform := &terraformextensionsv1.Terraform{}
	if err := json.Unmarshal(body, terraform); err != nil {
		utils.ExitError(fmt.Sprintf("json unmarshal failed: %s", err.Error()))
	}
	if terraform.Spec.Project == "" {
		utils.ExitError("spec.project cannot be empty")
	}
	if !strings.HasPrefix(terraform.Name, terraform.Spec.Project) {
		terraform.Name = terraform.Spec.Project + "-" + terraform.Name
	}
	bs, _ := json.Marshal(terraform)
	req := &httpapi.TerraformApplyRequest{
		Data: string(bs),
	}
	_ = httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   applyPath,
		Method: http.MethodPost,
		Body:   req,
	})
	fmt.Println("success")
}

// Create terraform with params
func (h *Handler) Create(ctx context.Context, req *httpapi.TerraformCreateRequest) {
	if req.Name == "" {
		utils.ExitError("param 'name' cannot be empty")
	}
	if req.Project == "" {
		utils.ExitError("param 'project' cannot be empty")
	}
	if req.Repo == "" {
		utils.ExitError("param 'repo' cannot be empty")
	}
	warnings := make([]string, 0)
	if !strings.HasPrefix(req.Name, req.Project) {
		req.Name = req.Project + "-" + req.Name
		warnings = append(warnings, fmt.Sprintf("name is rewrite to '%s'", req.Name))
	}
	if req.Path == "" {
		req.Path = "./"
		warnings = append(warnings, fmt.Sprintf("repo path default set to '%s'", req.Path))
	}
	if req.Revision == "" {
		req.Revision = "HEAD"
		warnings = append(warnings, fmt.Sprintf("repo revision default set to '%s'", req.Revision))
	}
	if req.SyncPolicy == "" {
		req.SyncPolicy = "manual"
		warnings = append(warnings, fmt.Sprintf("sync-policy default set to '%s'", req.SyncPolicy))
	}
	if len(warnings) != 0 {
		color.Blue("Warning: %s\n", strings.Join(warnings, ", "))
	}
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   createPath,
		Method: http.MethodPost,
		Body:   req,
	})
	resp := &stringResponse{}
	if err := json.Unmarshal(respBody, resp); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal response body failed: %s", err.Error()))
	}
	fmt.Println(resp.Data)
}

// Delete delete the resource
func (h *Handler) Delete(ctx context.Context, name *string) {
	if name == nil || *name == "" {
		utils.ExitError("'name' cannot be empty")
	}
	_ = httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(deletePath, *name),
		Method: http.MethodDelete,
	})
	fmt.Println("success")
}

// Get the terraform by name
func (h *Handler) Get(ctx context.Context, name *string, output *string) {
	if name == nil || *name == "" {
		utils.ExitError("'name' cannot be empty")
	}
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(getPath, *name),
		Method: http.MethodGet,
	})
	tf := &terraformextensionsv1.Terraform{}
	if err := json.Unmarshal(respBody, tf); err != nil {
		utils.ExitError(fmt.Sprintf("unmarshal response body failed: %s", err.Error()))
	}
	if output == nil || *output == "" || (*output != "json" && *output != "yaml") {
		tw := utils.DefaultTableWriter()
		tw.SetHeader(func() []string {
			return []string{
				"NAME", "REPO", "POLICY", "SYNC", "OPERATION", "DESTROY", "LAST APPLY",
			}
		}())
		var duration string
		if tf.Status.LastAppliedAt != nil {
			duration = utils.FriendTimeFormat(tf.Status.LastAppliedAt.Time, time.Now())
		}
		tw.Append(func() []string {
			return []string{
				tf.Name,
				tf.Spec.Repository.Repo,
				tf.Spec.SyncPolicy,
				tf.Status.SyncStatus,
				tf.Status.OperationStatus.Phase,
				fmt.Sprintf("%v", tf.Spec.DestroyResourcesOnDeletion),
				duration,
			}
		}())
		tw.Render()
		os.Exit(0)
	}
	tf.SetManagedFields(nil)
	var bs []byte
	if *output == "json" {
		bs, _ = json.Marshal(tf)
	}
	if *output == "yaml" {
		serializer := k8sjson.NewSerializerWithOptions(k8sjson.DefaultMetaFactory,
			nil, nil, k8sjson.SerializerOptions{Yaml: true})
		bs, _ = runtime.Encode(serializer, tf)
	}
	fmt.Println(string(bs))
	os.Exit(0)
}

// GetDiff the terraform diff by name
func (h *Handler) GetDiff(ctx context.Context, name *string) {
	if name == nil || *name == "" {
		utils.ExitError("'name' cannot be empty")
	}
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(getDiffPath, *name),
		Method: http.MethodGet,
	})
	result := &tfhandler.TerraformPlanOrApply{}
	if err := json.Unmarshal(respBody, result); err != nil {
		utils.ExitError(err.Error())
	}
	if result.Result == "" {
		utils.ExitError(fmt.Sprintf("terraform '%s' not have diff", *name))
	}
	fmt.Printf(`##################################################################
# Revision: %s             #
# Time: %s                                      #
#----------------------------------------------------------------#`,
		result.CommitID, result.CreationTime.Format("2006-01-02 15:04:05"))
	fmt.Println()
	color.Green(result.Result)
}

// GetApply get the terraform apply by name
func (h *Handler) GetApply(ctx context.Context, name *string) {
	if name == nil || *name == "" {
		utils.ExitError("'name' cannot be empty")
	}
	respBody := httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(getApplyPath, *name),
		Method: http.MethodGet,
	})
	result := &tfhandler.TerraformPlanOrApply{}
	if err := json.Unmarshal(respBody, result); err != nil {
		utils.ExitError(err.Error())
	}
	if result.Result == "" {
		utils.ExitError(fmt.Sprintf("terraform '%s' not have apply", *name))
	}
	fmt.Printf(`##################################################################
### Revision: %s             
### Time: %s                                      
#----------------------------------------------------------------#`,
		result.CommitID, result.CreationTime.Format("2006-01-02 15:04:05"))
	fmt.Println()
	color.Green(result.Result)
}

// Sync terraform
func (h *Handler) Sync(ctx context.Context, name *string) {
	if name == nil || *name == "" {
		utils.ExitError("'name' cannot be empty")
	}
	_ = httputils.DoRequest(ctx, &httputils.HTTPRequest{
		Path:   fmt.Sprintf(syncPath, *name),
		Method: http.MethodPut,
	})
	fmt.Println("sync operation is triggered")
}
