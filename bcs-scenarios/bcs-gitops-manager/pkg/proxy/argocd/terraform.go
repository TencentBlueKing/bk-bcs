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

package argocd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/httpapi"
	terraformextensionsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/terraformstore"
)

// TerraformPlugin defines the terraform plugin
type TerraformPlugin struct {
	*mux.Router
	middleware     mw.MiddlewareInterface
	terraformStore terraformstore.TerraformInterface
}

// Init the terraform route
func (plugin *TerraformPlugin) Init() error {
	plugin.Path("").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.listHandler))
	plugin.Path("/apply").Methods(http.MethodPost).Handler(plugin.middleware.HttpWrapper(plugin.applyHandler))
	plugin.Path("/create").Methods(http.MethodPost).Handler(plugin.middleware.HttpWrapper(plugin.createHandler))

	tfRouter := plugin.PathPrefix("/{name}").Subrouter()
	tfRouter.Path("").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.getHandler))
	tfRouter.Path("").Methods(http.MethodDelete).Handler(plugin.middleware.HttpWrapper(plugin.deleteHandler))
	tfRouter.Path("/get-diff").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.getDiffHandler))
	tfRouter.Path("/get-apply").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.getApplyHandler))
	tfRouter.Path("/sync").Methods(http.MethodPut).Handler(plugin.middleware.HttpWrapper(plugin.syncHandler))
	tfRouter.Path("/clean").Methods(http.MethodPut).Handler(plugin.middleware.HttpWrapper(plugin.cleanHandler))
	return nil
}

func (plugin *TerraformPlugin) applyHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	applyReq := &httpapi.TerraformApplyRequest{}
	if err = json.Unmarshal(body, applyReq); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal request body failed"))
	}
	data := []byte(applyReq.Data)
	var jsonData map[string]interface{}
	var yamlData map[string]interface{}
	jsonErr := json.Unmarshal(data, &jsonData)
	yamlErr := yaml.Unmarshal(data, &yamlData)
	if jsonErr != nil && yamlErr != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("request body not json or yaml type"))
	}
	if yamlErr == nil {
		if data, err = json.Marshal(yamlData); err != nil {
			return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "request body yamltojson failed"))
		}
	}
	terraform := &terraformextensionsv1.Terraform{}
	if err = json.Unmarshal(data, terraform); err != nil {
		err = errors.Wrapf(err, "request body json unmarshal data '%s' failed", applyReq.Data)
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, err)
	}
	if terraform.Spec.Project == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("spec.project cannot be mepty"))
	}
	if !strings.HasPrefix(terraform.Name, terraform.Spec.Project) {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Errorf("name must have preffix '%s-'", terraform.Spec.Project))
	}
	_, status, err := plugin.middleware.CheckProjectPermission(r.Context(), terraform.Spec.Project, iam.ProjectView)
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	// nolint
	tfbs, _ := json.Marshal(terraform)
	updatedBody, _ := json.Marshal(&httpapi.TerraformApplyRequest{Data: string(tfbs)})
	r.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
	length := len(updatedBody)
	r.Header.Set("Content-Length", strconv.Itoa(length))
	r.ContentLength = int64(length)
	return r, mw.ReturnTerraformReverse()
}

func (plugin *TerraformPlugin) listHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projectList, statusCode, err := plugin.middleware.ListProjects(r.Context())
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "list projects failed"))
	}
	allProjects := make(map[string]struct{})
	for i := range projectList.Items {
		allProjects[projectList.Items[i].Name] = struct{}{}
	}
	projects := r.URL.Query()["projects"]
	queryProjects := make([]string, 0, len(allProjects))
	if len(projects) == 0 {
		for proj := range allProjects {
			queryProjects = append(queryProjects, proj)
		}
	} else {
		for i := range projects {
			if _, ok := allProjects[projects[i]]; ok {
				queryProjects = append(queryProjects, projects[i])
			}
		}
	}
	if len(queryProjects) == 0 {
		return r, mw.ReturnJSONResponse(httpapi.TerraformHTTPResponse{
			Code: 0,
			Data: []terraformextensionsv1.Terraform{},
		})
	}
	values := r.URL.Query()
	values.Set("projects", queryProjects[0])
	for i := 1; i < len(queryProjects); i++ {
		values.Add("projects", queryProjects[i])
	}
	r.URL.RawQuery = values.Encode()
	return r, mw.ReturnTerraformReverse()
}

func (plugin *TerraformPlugin) createHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	createReq := &httpapi.TerraformCreateRequest{}
	if err = json.Unmarshal(body, createReq); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal request body failed"))
	}
	if createReq.Project == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("'project' cannot be mepty"))
	}
	if !strings.HasPrefix(createReq.Name, createReq.Project) {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Errorf("name must have preffix '%s-'", createReq.Project))
	}
	_, status, err := plugin.middleware.CheckProjectPermission(r.Context(), createReq.Project, iam.ProjectView)
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	repos, status, err := plugin.middleware.ListRepositories(r.Context(), []string{createReq.Project}, false)
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	matched := false
	for _, repo := range repos.Items {
		if createReq.Repo == repo.Repo {
			matched = true
			break
		}
	}
	if !matched {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.
			Errorf("'repo' not exist in project '%s'", createReq.Project))
	}
	// nolint
	updatedBody, _ := json.Marshal(createReq)
	r.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
	length := len(updatedBody)
	r.Header.Set("Content-Length", strconv.Itoa(length))
	r.ContentLength = int64(length)
	return r, mw.ReturnTerraformReverse()
}

func (plugin *TerraformPlugin) getHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	name := mux.Vars(r)["name"]
	tf, err := plugin.terraformStore.Get(r.Context(), name)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	_, status, err := plugin.middleware.CheckProjectPermission(r.Context(), tf.Spec.Project, iam.ProjectView)
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	return r, mw.ReturnJSONResponse(tf)
}

func (plugin *TerraformPlugin) deleteHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	name := mux.Vars(r)["name"]
	tf, err := plugin.terraformStore.Get(r.Context(), name)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	_, status, err := plugin.middleware.CheckProjectPermission(r.Context(), tf.Spec.Project, iam.ProjectView)
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	return r, mw.ReturnTerraformReverse()
}

func (plugin *TerraformPlugin) getDiffHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	name := mux.Vars(r)["name"]
	data, err := plugin.terraformStore.GetDiff(r.Context(), name)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	_, status, err := plugin.middleware.CheckProjectPermission(r.Context(),
		data.Terraform.Spec.Project, iam.ProjectView)
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	return r, mw.ReturnJSONResponse(data.Result)
}

func (plugin *TerraformPlugin) getApplyHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	name := mux.Vars(r)["name"]
	data, err := plugin.terraformStore.GetApply(r.Context(), name)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	_, status, err := plugin.middleware.CheckProjectPermission(r.Context(),
		data.Terraform.Spec.Project, iam.ProjectView)
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	return r, mw.ReturnJSONResponse(data.Result)
}

func (plugin *TerraformPlugin) syncHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	name := mux.Vars(r)["name"]
	tf, err := plugin.terraformStore.Get(r.Context(), name)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	_, status, err := plugin.middleware.CheckProjectPermission(r.Context(), tf.Spec.Project, iam.ProjectEdit)
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	return r, mw.ReturnTerraformReverse()
}

func (plugin *TerraformPlugin) cleanHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	name := mux.Vars(r)["name"]
	tf, err := plugin.terraformStore.Get(r.Context(), name)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	_, status, err := plugin.middleware.CheckProjectPermission(r.Context(), tf.Spec.Project, iam.ProjectEdit)
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	return r, mw.ReturnTerraformReverse()
}
