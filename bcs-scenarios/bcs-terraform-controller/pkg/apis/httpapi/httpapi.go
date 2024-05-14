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

// Package httpapi defines the http api
package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	terraformextensionsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/apis/terraformextensions/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/tfhandler"
)

// TerraformHTTPServer defines the httpServer instance for terraform
type TerraformHTTPServer struct {
	op         *option.ControllerOption
	httpServer *http.Server
	mgrClient  client.Client
	tfHandler  tfhandler.TerraformHandler
}

// TerraformHTTPResponse defines the common response
type TerraformHTTPResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// NewTerraformHTTPServer create the terraform httpServer instance
func NewTerraformHTTPServer(mgrClient client.Client, tfHandler tfhandler.TerraformHandler) *TerraformHTTPServer {
	s := &TerraformHTTPServer{
		op:        option.GlobalOption(),
		mgrClient: mgrClient,
		tfHandler: tfHandler,
	}
	router := mux.NewRouter()
	router.UseEncodedPath()
	subRouter := router.PathPrefix("/api/v1/terraforms").Subrouter()
	subRouter.Methods(http.MethodGet).Path("").HandlerFunc(s.List)
	subRouter.Methods(http.MethodPost).Path("/create").HandlerFunc(s.Create)
	subRouter.Methods(http.MethodPost).Path("/apply").HandlerFunc(s.Apply)

	tfRouter := subRouter.PathPrefix("/{name}").Subrouter()
	tfRouter.Methods(http.MethodGet).Path("").HandlerFunc(s.Get)
	tfRouter.Methods(http.MethodDelete).Path("").HandlerFunc(s.Delete)
	tfRouter.Methods(http.MethodGet).Path("/get-diff").HandlerFunc(s.GetDiff)
	tfRouter.Methods(http.MethodGet).Path("/get-apply").HandlerFunc(s.GetApply)
	tfRouter.Methods(http.MethodPut).Path("/sync").HandlerFunc(s.Sync)
	tfRouter.Methods(http.MethodPut).Path("/clean").HandlerFunc(s.Clean)
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", s.op.HTTPPort),
		Handler: router,
	}
	return s
}

// Start http server
func (s *TerraformHTTPServer) Start() error {
	if err := s.httpServer.ListenAndServe(); err != nil {
		return errors.Wrapf(err, "http server error")
	}
	return nil
}

// List  the terraform by projects
func (s *TerraformHTTPServer) List(writer http.ResponseWriter, request *http.Request) {
	terraformList := &terraformextensionsv1.TerraformList{}
	if err := s.mgrClient.List(request.Context(), terraformList, &client.ListOptions{}); err != nil {
		s.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "list terraforms failed"))
		return
	}
	projects := request.URL.Query()["projects"]
	if len(projects) == 0 {
		s.httpJson(writer, &TerraformHTTPResponse{Code: 0, Data: terraformList.Items})
		return
	}
	projectMap := make(map[string]struct{})
	for i := range projects {
		projectMap[projects[i]] = struct{}{}
	}
	result := make([]terraformextensionsv1.Terraform, 0)
	for i := range terraformList.Items {
		item := terraformList.Items[i]
		if _, ok := projectMap[item.Spec.Project]; ok {
			result = append(result, item)
		}
	}
	s.httpJson(writer, &TerraformHTTPResponse{Code: 0, Data: result})
}

// TerraformApplyRequest defines the request of terraform apply
type TerraformApplyRequest struct {
	Data string `json:"data"`
}

// Apply the terraform from config file
func (s *TerraformHTTPServer) Apply(writer http.ResponseWriter, request *http.Request) {
	applyReq := &TerraformApplyRequest{}
	if err := s.readBody(request, applyReq); err != nil {
		s.httpError(writer, http.StatusBadRequest, err)
		return
	}
	data := []byte(applyReq.Data)
	terraform := &terraformextensionsv1.Terraform{}
	if err := json.Unmarshal(data, terraform); err != nil {
		s.httpError(writer, http.StatusBadRequest, errors.Wrapf(err, "unmarshal request body failed"))
	}

	k8sTerraform := &terraformextensionsv1.Terraform{}
	if err := s.mgrClient.Get(request.Context(), types.NamespacedName{
		Namespace: s.op.WorkerNamespace,
		Name:      terraform.Name,
	}, k8sTerraform); err != nil {
		if !k8serrors.IsNotFound(err) {
			s.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "check terraform failed"))
			return
		}
		if err = s.mgrClient.Create(request.Context(), &terraformextensionsv1.Terraform{
			TypeMeta: metav1.TypeMeta{
				APIVersion: terraformextensionsv1.GroupName + "/" + terraformextensionsv1.Version,
				Kind:       "Terraform",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      terraform.Name,
				Namespace: s.op.WorkerNamespace,
			},
			Spec: terraform.Spec,
		}); err != nil {
			s.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "create terraform failed"))
			return
		}
	} else {
		k8sTerraform.Spec = terraform.Spec
		if err = s.mgrClient.Update(request.Context(), k8sTerraform); err != nil {
			s.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "update terraform failed"))
			return
		}
	}
	s.httpJson(writer, &TerraformHTTPResponse{Code: 0, Data: "success"})
}

// TerraformCreateRequest defines the request that create terraform
type TerraformCreateRequest struct {
	Name       string `json:"name"`
	Destroy    bool   `json:"destroy"`
	Project    string `json:"project"`
	Repo       string `json:"repo"`
	Path       string `json:"path"`
	Revision   string `json:"revision"`
	SyncPolicy string `json:"syncPolicy"`
}

// Create the terraform by form
func (s *TerraformHTTPServer) Create(writer http.ResponseWriter, request *http.Request) {
	createReq := &TerraformCreateRequest{}
	if err := s.readBody(request, createReq); err != nil {
		s.httpError(writer, http.StatusBadRequest, err)
		return
	}
	if err := s.mgrClient.Create(request.Context(), &terraformextensionsv1.Terraform{
		TypeMeta: metav1.TypeMeta{
			APIVersion: terraformextensionsv1.GroupName + "/" + terraformextensionsv1.Version,
			Kind:       "Terraform",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      createReq.Name,
			Namespace: s.op.WorkerNamespace,
		},
		Spec: terraformextensionsv1.TerraformSpec{
			DestroyResourcesOnDeletion: createReq.Destroy,
			Project:                    createReq.Project,
			Repository: terraformextensionsv1.GitRepository{
				Path:           createReq.Path,
				Repo:           createReq.Repo,
				TargetRevision: createReq.Revision,
			},
			SyncPolicy: createReq.SyncPolicy,
		},
	}); err != nil {
		s.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "create terraform failed"))
		return
	}
	s.httpJson(writer, &TerraformHTTPResponse{Code: 0, Data: "success"})
}

// Delete terraform
func (s *TerraformHTTPServer) Delete(writer http.ResponseWriter, request *http.Request) {
	obj, statusCode, err := s.getTerraform(request)
	if err != nil {
		s.httpError(writer, statusCode, err)
		return
	}
	if err = s.mgrClient.Delete(request.Context(), obj); err != nil {
		s.httpError(writer, http.StatusInternalServerError, err)
		return
	}
	s.httpJson(writer, &TerraformHTTPResponse{
		Code: 0,
		Data: obj,
	})
}

// Get return the single terraform
func (s *TerraformHTTPServer) Get(writer http.ResponseWriter, request *http.Request) {
	obj, statusCode, err := s.getTerraform(request)
	if err != nil {
		s.httpError(writer, statusCode, err)
		return
	}
	s.httpJson(writer, &TerraformHTTPResponse{
		Code: 0,
		Data: obj,
	})
}

// TerraformGetDiffOrApplyData the data that terraform get of diff
type TerraformGetDiffOrApplyData struct {
	Result    *tfhandler.TerraformPlanOrApply  `json:"result"`
	Terraform *terraformextensionsv1.Terraform `json:"terraform"`
}

// TerraformGetDiffOrApplyResponse defines the response
type TerraformGetDiffOrApplyResponse struct {
	Code int                          `json:"code"`
	Data *TerraformGetDiffOrApplyData `json:"data"`
}

// GetDiff return the diff that terraform plan result
func (s *TerraformHTTPServer) GetDiff(writer http.ResponseWriter, request *http.Request) {
	obj, statusCode, err := s.getTerraform(request)
	if err != nil {
		s.httpError(writer, statusCode, err)
		return
	}
	result, err := s.tfHandler.GetPlanResult(request.Context(), obj)
	if err != nil {
		s.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "get terraform diff failed"))
		return
	}
	s.httpJson(writer, &TerraformGetDiffOrApplyResponse{
		Code: 0,
		Data: &TerraformGetDiffOrApplyData{
			Result:    result,
			Terraform: obj,
		},
	})
}

// GetApply return the last apply result
func (s *TerraformHTTPServer) GetApply(writer http.ResponseWriter, request *http.Request) {
	obj, statusCode, err := s.getTerraform(request)
	if err != nil {
		s.httpError(writer, statusCode, err)
		return
	}
	result, err := s.tfHandler.GetLastApply(request.Context(), obj)
	if err != nil {
		s.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "get terraform diff failed"))
		return
	}
	s.httpJson(writer, &TerraformGetDiffOrApplyResponse{
		Code: 0,
		Data: &TerraformGetDiffOrApplyData{
			Result:    result,
			Terraform: obj,
		},
	})
}

// Sync terraform
func (s *TerraformHTTPServer) Sync(writer http.ResponseWriter, request *http.Request) {
	revision := request.URL.Query().Get("revision")
	name := mux.Vars(request)["name"]
	if name == "" {
		s.httpError(writer, http.StatusBadRequest, errors.Errorf("query param 'name' cannot be empty"))
		return
	}
	obj, statusCode, err := s.getTerraform(request)
	if err != nil {
		s.httpError(writer, statusCode, err)
		return
	}
	if obj.Status.SyncStatus == terraformextensionsv1.SyncedStatus {
		s.httpError(writer, http.StatusBadRequest, errors.Errorf("terraform is synced, no need sync again"))
		return
	}
	lastRevision := obj.Status.LastPlannedRevision
	if lastRevision == "" {
		s.httpError(writer, http.StatusBadRequest, errors.Errorf("terraform status.lastPlannedRevision is empty, "+
			"should retry again after have plan result"))
		return
	}
	if revision != "" && revision != lastRevision {
		s.httpError(writer, http.StatusBadRequest, errors.Errorf("query param 'revision' not same as "+
			"status.lastPlannedRevision '%s'", lastRevision))
		return
	}

	rawPatch := client.RawPatch(k8stypes.JSONPatchType,
		[]byte(fmt.Sprintf(`[{"op":"add","path":"/metadata/annotations/%s", "value":"%s"}]`,
			terraformextensionsv1.TerraformOperationSync, obj.Status.LastPlannedRevision)))
	if err = s.mgrClient.Patch(request.Context(), obj, rawPatch); err != nil {
		s.httpError(writer, http.StatusInternalServerError, errors.Wrapf(err, "patch terraform sync annotation failed"))
		return
	}
	s.httpJson(writer, &TerraformHTTPResponse{Code: 0, Data: "sync is in progressing"})
}

// Clean the terraform resources
func (s *TerraformHTTPServer) Clean(writer http.ResponseWriter, request *http.Request) {
	name := mux.Vars(request)["name"]
	if name == "" {
		s.httpError(writer, http.StatusBadRequest, errors.Errorf("query param 'name' cannot be empty"))
		return
	}
	obj, statusCode, err := s.getTerraform(request)
	if err != nil {
		s.httpError(writer, statusCode, err)
		return
	}
	rawPatch := client.RawPatch(k8stypes.JSONPatchType,
		[]byte(fmt.Sprintf(`[{"op":"add","path":"/metadata/annotations/%s", "value":"%s"}]`,
			terraformextensionsv1.TerraformOperationClean, "clean")))
	if err = s.mgrClient.Patch(request.Context(), obj, rawPatch); err != nil {
		s.httpError(writer, http.StatusInternalServerError, errors.
			Wrapf(err, "patch terraform clean annotation failed"))
		return
	}
	s.httpJson(writer, &TerraformHTTPResponse{Code: 0, Data: "clean is in progressing"})

}

func (s *TerraformHTTPServer) httpError(rw http.ResponseWriter, statusCode int, err error) {
	http.Error(rw, err.Error(), statusCode)
}

func (s *TerraformHTTPServer) httpJson(rw http.ResponseWriter, obj interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	content, _ := json.Marshal(obj)
	rw.Write(content)
}

func (s *TerraformHTTPServer) readBody(req *http.Request, result interface{}) error {
	bs, err := io.ReadAll(req.Body)
	if err != nil {
		return errors.Wrapf(err, "read request body failed")
	}
	if err = json.Unmarshal(bs, &result); err != nil {
		return errors.Wrapf(err, "unmarshal request body '%s' failed", string(bs))
	}
	return nil
}

func (s *TerraformHTTPServer) getTerraform(request *http.Request) (*terraformextensionsv1.Terraform, int, error) {
	name := mux.Vars(request)["name"]
	if name == "" {
		return nil, http.StatusBadRequest, errors.Errorf("query param 'name' cannot be empty")
	}
	obj := &terraformextensionsv1.Terraform{}
	if err := s.mgrClient.Get(request.Context(), types.NamespacedName{
		Namespace: s.op.WorkerNamespace,
		Name:      name,
	}, obj); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, http.StatusNotFound, errors.Errorf("terraform '%s' not found", name)
		}
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "query terraform failed")
	}
	return obj, http.StatusOK, nil
}
