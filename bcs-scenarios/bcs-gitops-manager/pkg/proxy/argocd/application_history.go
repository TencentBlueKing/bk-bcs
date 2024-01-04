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
 *
 */

package argocd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	argoappv1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
)

// ApplicationHistoryManifestResponse defines the response for get history manifests of application
type ApplicationHistoryManifestResponse struct {
	Code      int32                `json:"code"`
	Message   string               `json:"message"`
	RequestID string               `json:"requestID"`
	Data      *HistoryManifestData `json:"data"`
}

// HistoryManifestData defines the response data that queried by application history
type HistoryManifestData struct {
	Manifests []*HistoryManifest `json:"manifests"`
}

// HistoryManifest defines the manifest item that history managed
type HistoryManifest struct {
	Group     string                     `json:"group"`
	Kind      string                     `json:"kind"`
	Namespace string                     `json:"namespace"`
	Name      string                     `json:"name"`
	Object    *unstructured.Unstructured `json:"object"`
}

func (plugin *AppPlugin) applicationHistoryState(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	historyIDStr := r.URL.Query().Get("historyID")
	if historyIDStr == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request query param 'historyID' cannot be empty"))
	}
	historyID, err := strconv.ParseInt(historyIDStr, 10, 64)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "request query param 'historyID' %s not int", historyIDStr))
	}

	app, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName,
		iamnamespace.NameSpaceScopedView)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check application permission failed"))
	}
	applicationUID := r.URL.Query().Get("applicationUID")
	if applicationUID == "" {
		applicationUID = string(app.UID)
	}
	hm, err := plugin.db.GetApplicationHistoryManifest(appName, applicationUID, historyID)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "get application history manfiest failed"))
	}
	if hm == nil {
		return r, mw.ReturnErrorResponse(http.StatusNotFound,
			fmt.Errorf("application '%s/%s' with history '%s' not found", appName, applicationUID, historyIDStr))
	}
	resources := make([]*argoappv1.ResourceDiff, 0)
	if err = json.Unmarshal([]byte(hm.ManagedResources), &resources); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError,
			errors.Wrapf(err, "unmarshal application history managedResources failed"))
	}

	manifests := make([]*HistoryManifest, 0, len(resources))
	for i, resState := range resources {
		var obj *unstructured.Unstructured
		obj, err = resState.TargetObject()
		if err != nil {
			return r, mw.ReturnErrorResponse(http.StatusInternalServerError,
				errors.Wrapf(err, "get target object for '%d' failed", i))
		}
		manifests = append(manifests, &HistoryManifest{
			Group:     resState.Group,
			Kind:      resState.Kind,
			Namespace: resState.Namespace,
			Name:      resState.Name,
			Object:    obj,
		})
	}
	return r, mw.ReturnJSONResponse(&ApplicationHistoryManifestResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data: &HistoryManifestData{
			Manifests: manifests,
		},
	})
}
