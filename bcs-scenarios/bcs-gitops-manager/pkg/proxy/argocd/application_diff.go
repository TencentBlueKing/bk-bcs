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
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
)

// ApplicationDiffRequest defines the diff request
type ApplicationDiffRequest struct {
	ApplicationName string   `json:"applicationName"`
	Revision        string   `json:"revision"`
	Revisions       []string `json:"revisions"`
}

// ApplicationDiffResponse defines the diff response
type ApplicationDiffResponse struct {
	Code      int32                `json:"code"`
	Message   string               `json:"message"`
	RequestID string               `json:"requestID"`
	Data      *ApplicationDiffData `json:"data"`
}

// ApplicationDiffData defines the data of application diff
type ApplicationDiffData struct {
	Cluster       string `json:"cluster,omitempty"`
	RepoUrl       string `json:"repoUrl,omitempty"`
	LocalRevision string `json:"localRevision,omitempty"`
	LiveRevision  string `json:"liveRevision,omitempty"`

	Result []*ApplicationDiffManifest `json:"result"`
}

// ApplicationDiffManifest defines the diff manifest
type ApplicationDiffManifest struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Group     string `json:"group,omitempty"`
	Version   string `json:"version,omitempty"`
	Kind      string `json:"kind,omitempty"`

	// diff result
	Local *unstructured.Unstructured `json:"local,omitempty"`
	Live  *unstructured.Unstructured `json:"live,omitempty"`
}

func (plugin *AppPlugin) applicationDiff(r *http.Request) (*http.Request, *mw.HttpResponse) {
	req, err := buildDiffRequest(r)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "build request failed"))
	}
	blog.Infof("RequestID[%s] received diff request: %v", mw.RequestID(r.Context()), req)
	application, err := plugin.checkApplicationDiffRequest(r.Context(), req)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "build application from request failed"))
	}
	result, err := plugin.handleApplicationDiff(r.Context(), req, application.DeepCopy())
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "handle application diff manifests failed"))
	}
	return r, mw.ReturnJSONResponse(&ApplicationDiffResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data: &ApplicationDiffData{
			Cluster:       application.Spec.Destination.Server,
			RepoUrl:       application.Spec.Source.RepoURL,
			LocalRevision: req.Revision,
			LiveRevision:  application.Status.Sync.Revision,
			Result:        result,
		},
	})
}

func (plugin *AppPlugin) handleApplicationDiff(ctx context.Context, req *ApplicationDiffRequest,
	application *v1alpha1.Application) ([]*ApplicationDiffManifest, error) {
	// 获取应用当前 live 的 manifest
	originalApplication := application.DeepCopy()
	liveManifestMap, err := plugin.applicationManifests(ctx, originalApplication)
	if err != nil {
		return nil, errors.Wrapf(err, "get live manifest failed")
	}
	// 获取应用要对比的目标 revision 的 manifest
	if len(req.Revisions) != 0 {
		for i := range application.Spec.Sources {
			application.Spec.Sources[i].TargetRevision = req.Revisions[i]
		}
	} else if req.Revision != "" {
		application.Spec.Source.TargetRevision = req.Revision
	}
	manifestMap, err := plugin.applicationManifests(ctx, application)
	if err != nil {
		return nil, errors.Wrapf(err, "get target manifest failed")
	}

	// 对比目标 revision 和 live 的 revision
	results := make([]*ApplicationDiffManifest, 0)
	for k, decodeObj := range manifestMap {
		liveDecodeObj, ok := liveManifestMap[k]
		if !ok {
			continue
		}
		if err = plugin.storage.ApplicationNormalizeWhenDiff(originalApplication,
			decodeObj, liveDecodeObj, true); err != nil {
			return nil, errors.Wrapf(err, "application normlize failed when diff")
		}
		results = append(results, &ApplicationDiffManifest{
			Namespace: decodeObj.GetNamespace(),
			Name:      decodeObj.GetName(),
			Group:     decodeObj.GroupVersionKind().Group,
			Version:   decodeObj.GroupVersionKind().Version,
			Kind:      decodeObj.GetKind(),
			Local:     decodeObj,
			Live:      liveDecodeObj,
		})
		delete(manifestMap, k)
		delete(liveManifestMap, k)
	}
	for _, obj := range manifestMap {
		obj.SetManagedFields(nil)
		results = append(results, &ApplicationDiffManifest{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
			Group:     obj.GroupVersionKind().Group,
			Version:   obj.GroupVersionKind().Version,
			Kind:      obj.GetKind(),
			Local:     obj,
		})
	}
	for _, obj := range liveManifestMap {
		obj.SetManagedFields(nil)
		results = append(results, &ApplicationDiffManifest{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
			Group:     obj.GroupVersionKind().Group,
			Version:   obj.GroupVersionKind().Version,
			Kind:      obj.GetKind(),
			Live:      obj,
		})
	}
	return results, nil
}

func (plugin *AppPlugin) checkApplicationDiffRequest(ctx context.Context,
	req *ApplicationDiffRequest) (*v1alpha1.Application, error) {
	application, statusCode, err := plugin.middleware.CheckApplicationPermission(ctx, req.ApplicationName,
		iamnamespace.NameSpaceScopedView)
	if statusCode != http.StatusOK {
		return nil, errors.Wrapf(err, "check application permission failed")
	}
	if application.Spec.HasMultipleSources() {
		if len(req.Revisions) != 0 && len(req.Revisions) != len(application.Spec.Sources) {
			return nil, errors.Errorf("application has '%d' resources, but request body 'revisions' only "+
				"have '%d'. Or you can set empty 'revisions', we will use default revisions in application",
				len(application.Spec.Sources), len(req.Revisions))
		}
		return application, nil
	}
	if application.Spec.Source == nil {
		return nil, errors.Errorf("application spec.source is nil")
	}
	return application, nil
}

func buildDiffRequest(r *http.Request) (*ApplicationDiffRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read body failed")
	}
	req := new(ApplicationDiffRequest)
	if err = json.Unmarshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal request body failed")
	}
	if req.ApplicationName == "" {
		return nil, errors.Errorf("applicationName or applicationManifests should in request")
	}
	return req, nil
}
